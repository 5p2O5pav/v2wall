package logwriter

import (
	"encoding/json"
	"fmt"

	"github.com/dgraph-io/badger/v4"
	"github.com/dgraph-io/badger/v4/y"
	"v2wall/internal/db"
	"v2wall/internal/ipgeo"
)

type LogEntry struct {
	IP        string `json:"ip"`
	Time      int64  `json:"time"`       // 纳秒时间戳
	Method    string `json:"method"`
	URL       string `json:"url"`
	UserAgent string `json:"user_agent"`
	Referer   string `json:"referer"`
}

// WriteLog 写入一条访问日志，如果 IP 首次出现则查询 ip2region。
// 返回错误（通常忽略，记录即可）
func WriteLog(bdb *badger.DB, entry LogEntry, ipSearcher *ipgeo.Searcher, enableV4, enableV6 bool) error {
	// 准备值
	logValue, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	return bdb.Update(func(txn *badger.Txn) error {
		// 1. 存储原始日志
		logKey := db.LogKey(entry.IP, entry.Time)
		if err := txn.Set(logKey, logValue); err != nil {
			return fmt.Errorf("set log: %w", err)
		}

		// 2. 原子递增计数器（使用 MergeOperator）
		counterKey := db.CounterKey(entry.IP)
		err := txn.Set(counterKey, y.U64ToBytes(1))
		if err != nil {
			return fmt.Errorf("merge counter: %w", err)
		}

		// 3. 更新最后访问时间
		lastSeenKey := db.LastSeenKey(entry.IP)
		lastSeenVal := y.U64ToBytes(uint64(entry.Time))
		if err := txn.Set(lastSeenKey, lastSeenVal); err != nil {
			return fmt.Errorf("set last_seen: %w", err)
		}

		// 4. 首次 IP 查询地理信息（仅在 ipinfo 键不存在时）
		ipInfoKey := db.IPInfoKey(entry.IP)
		_, err = txn.Get(ipInfoKey)
		if err == badger.ErrKeyNotFound {
			// 查询
			var region string
			if ipSearcher != nil {
				region, err = ipSearcher.Search(entry.IP)
				if err != nil {
					region = "" // 查询失败留空
				}
			} else {
				region = ""
			}
			// 存储地理信息
			if err := txn.Set(ipInfoKey, []byte(region)); err != nil {
				return fmt.Errorf("set ipinfo: %w", err)
			}
		} else if err != nil {
			return fmt.Errorf("check ipinfo: %w", err)
		}
		return nil
	})
}
