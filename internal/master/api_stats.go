package master

import (
    "encoding/json"
    "net/http"
    "sort"
    "strconv"
    "time"

    "github.com/dgraph-io/badger/v4"
    "github.com/gin-gonic/gin"
    "v2wall/internal/db"
    "v2wall/internal/ipgeo"
    "v2wall/internal/logwriter"
)

// IPStat 用于列表展示的去重 IP 统计
type IPStat struct {
	IP          string `json:"ip"`
	Count       uint64 `json:"count"`
	LastSeen    int64  `json:"last_seen"`    // Unix 纳秒
	LastSeenStr string `json:"last_seen_str"` // 可读时间
	Info        string `json:"info"`          // ip2region 原始数据
}

// LogEntryView 原始日志展示（从 Badger 读出）
type LogEntryView struct {
	Time      int64  `json:"time"`
	TimeStr   string `json:"time_str"`
	Method    string `json:"method"`
	URL       string `json:"url"`
	UserAgent string `json:"user_agent"`
	Referer   string `json:"referer"`
}

// handleGetIPs 获取去重 IP 列表，支持分页
func handleGetIPs(bdb *badger.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
		if page < 1 {
			page = 1
		}
		if size < 1 || size > 100 {
			size = 20
		}

		var stats []IPStat

		err := bdb.View(func(txn *badger.Txn) error {
			opts := badger.DefaultIteratorOptions
			opts.Prefix = []byte(db.PrefixCounter)
			opts.PrefetchValues = true
			it := txn.NewIterator(opts)
			defer it.Close()

			for it.Rewind(); it.Valid(); it.Next() {
				item := it.Item()
				key := item.Key()
				ip := string(key[len(db.PrefixCounter):])

				count, _ := db.GetCounter(txn, ip)
				lastSeen, _ := db.GetLastSeen(txn, ip)
				info, _ := db.GetIPInfo(txn, ip)

				stats = append(stats, IPStat{
					IP:          ip,
					Count:       count,
					LastSeen:    lastSeen,
					LastSeenStr: time.Unix(0, lastSeen).Format(time.RFC3339),
					Info:        info,
				})
			}
			return nil
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan data"})
			return
		}

		// 按最后访问时间倒序排序
		sort.Slice(stats, func(i, j int) bool {
			return stats[i].LastSeen > stats[j].LastSeen
		})

		total := len(stats)
		start := (page - 1) * size
		if start > total {
			start = total
		}
		end := start + size
		if end > total {
			end = total
		}
		paged := stats[start:end]

		c.JSON(http.StatusOK, gin.H{
			"total": total,
			"page":  page,
			"size":  size,
			"data":  paged,
		})
	}
}

// handleGetLogs 获取某个 IP 的原始日志，按时间倒序分页
func handleGetLogs(bdb *badger.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.Query("ip")
		if ip == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ip required"})
			return
		}
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
		if page < 1 {
			page = 1
		}
		if size < 1 || size > 100 {
			size = 20
		}

		var logs []LogEntryView
		var total int

		err := bdb.View(func(txn *badger.Txn) error {
			prefix := []byte(db.PrefixLog + ip + ":")
			opts := badger.DefaultIteratorOptions
			opts.Prefix = prefix
			opts.PrefetchValues = true
			opts.Reverse = true // 时间倒序（键尾是纳秒时间戳）
			it := txn.NewIterator(opts)
			defer it.Close()

			// 先统计总数
			for it.Rewind(); it.Valid(); it.Next() {
				total++
			}

			// 重新迭代到对应页
			skip := (page - 1) * size
			it.Rewind()
			for i := 0; i < skip && it.Valid(); i++ {
				it.Next()
			}
			count := 0
			for ; it.Valid() && count < size; it.Next() {
				item := it.Item()
				err := item.Value(func(v []byte) error {
					var le logwriter.LogEntry
					if json.Unmarshal(v, &le) != nil {
						return nil
					}
					logs = append(logs, LogEntryView{
						Time:      le.Time,
						TimeStr:   time.Unix(0, le.Time).Format(time.RFC3339Nano),
						Method:    le.Method,
						URL:       le.URL,
						UserAgent: le.UserAgent,
						Referer:   le.Referer,
					})
					return nil
				})
				if err != nil {
					continue
				}
				count++
			}
			return nil
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "read logs failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"total": total,
			"page":  page,
			"size":  size,
			"data":  logs,
		})
	}
}

// handleGetIPInfo 获取单个 IP 的 ip2region 原始数据
func handleGetIPInfo(bdb *badger.DB, searcher *ipgeo.Searcher) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.Query("ip")
		if ip == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ip required"})
			return
		}
		var info string
		err := bdb.View(func(txn *badger.Txn) error {
			var err error
			info, err = db.GetIPInfo(txn, ip)
			return err
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ip": ip, "info": info})
	}
}
