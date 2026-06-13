package master

import (
	"net"
	"net/http"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/gin-gonic/gin"
	"v2wall/internal/db"
)

// handleSyncBlacklist 返回当前需要阻断的 IP 数组
func handleSyncBlacklist(bdb *badger.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var ips []string
		err := bdb.View(func(txn *badger.Txn) error {
			// 获取黑名单配置，决定保留天数
			cfg, err := getBlacklistConfig(txn)
			if err != nil {
				return err
			}
			cutoff := time.Now().AddDate(0, 0, -cfg.RetentionDays).UnixNano()

			opts := badger.DefaultIteratorOptions
			opts.Prefix = []byte(db.PrefixLastSeen)
			it := txn.NewIterator(opts)
			defer it.Close()

			for it.Rewind(); it.Valid(); it.Next() {
				item := it.Item()
				key := item.Key()
				ip := string(key[len(db.PrefixLastSeen):])

				ts, err := db.GetLastSeen(txn, ip)
				if err != nil {
					continue
				}
				// 过期的不返回
				if ts < cutoff {
					continue
				}
				// 检查是否在白名单中
				if isWhitelisted(txn, ip) {
					continue
				}
				ips = append(ips, ip)
			}
			return nil
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"blacklist": ips})
	}
}

// isWhitelisted 检查 IP 是否匹配任意白名单 CIDR
func isWhitelisted(txn *badger.Txn, ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	opts := badger.DefaultIteratorOptions
	opts.Prefix = []byte(db.PrefixWhitelist)
	it := txn.NewIterator(opts)
	defer it.Close()
	for it.Rewind(); it.Valid(); it.Next() {
		item := it.Item()
		var entry WhitelistEntry
		err := item.Value(func(v []byte) error {
			return json.Unmarshal(v, &entry)
		})
		if err != nil {
			continue
		}
		_, cidr, err := net.ParseCIDR(entry.CIDR)
		if err != nil {
			continue
		}
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}

// handleSyncWhitelist 返回白名单 CIDR 数组
func handleSyncWhitelist(bdb *badger.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var cidrs []string
		err := bdb.View(func(txn *badger.Txn) error {
			opts := badger.DefaultIteratorOptions
			opts.Prefix = []byte(db.PrefixWhitelist)
			it := txn.NewIterator(opts)
			defer it.Close()
			for it.Rewind(); it.Valid(); it.Next() {
				item := it.Item()
				var entry WhitelistEntry
				err := item.Value(func(v []byte) error {
					return json.Unmarshal(v, &entry)
				})
				if err == nil {
					cidrs = append(cidrs, entry.CIDR)
				}
			}
			return nil
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"whitelist": cidrs})
	}
}

// handleSyncHeartbeat 接收被控端心跳
func handleSyncHeartbeat(bdb *badger.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			NodeID          string `json:"node_id" binding:"required"`
			HoneypotEnabled bool   `json:"honeypot_enabled"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		now := time.Now().UnixNano()
		node := NodeInfo{
			ID:              req.NodeID,
			LastHeartbeat:   now,
			HoneypotEnabled: req.HoneypotEnabled,
			// LastReportTime 保持不变，这里不更新
		}

		// 读取旧的 LastReportTime
		err := bdb.View(func(txn *badger.Txn) error {
			key := db.NodeKey(req.NodeID)
			item, err := txn.Get(key)
			if err == badger.ErrKeyNotFound {
				return nil
			}
			if err != nil {
				return err
			}
			return item.Value(func(v []byte) error {
				var old NodeInfo
				if json.Unmarshal(v, &old) == nil {
					node.LastReportTime = old.LastReportTime
				}
				return nil
			})
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		err = bdb.Update(func(txn *badger.Txn) error {
			val, _ := json.Marshal(node)
			return txn.Set(db.NodeKey(req.NodeID), val)
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "heartbeat received"})
	}
}
