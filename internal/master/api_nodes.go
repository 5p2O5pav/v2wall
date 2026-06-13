package master

import (
	"encoding/json"
	"net/http"

	"github.com/dgraph-io/badger/v4"
	"github.com/gin-gonic/gin"
	"github.com/yourorg/v2wall/internal/db"
)

type NodeInfo struct {
	ID              string `json:"id"`
	LastHeartbeat   int64  `json:"last_heartbeat"`   // Unix 纳秒
	HoneypotEnabled bool   `json:"honeypot_enabled"`
	LastReportTime  int64  `json:"last_report_time"` // 最后上报日志时间，0 表示从未上报
}

func handleGetNodes(bdb *badger.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var nodes []NodeInfo
		err := bdb.View(func(txn *badger.Txn) error {
			opts := badger.DefaultIteratorOptions
			opts.Prefix = []byte(db.PrefixNode)
			it := txn.NewIterator(opts)
			defer it.Close()
			for it.Rewind(); it.Valid(); it.Next() {
				item := it.Item()
				var node NodeInfo
				err := item.Value(func(v []byte) error {
					return json.Unmarshal(v, &node)
				})
				if err != nil {
					continue
				}
				nodes = append(nodes, node)
			}
			return nil
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": nodes})
	}
}
