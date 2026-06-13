package master

import (
	"net/http"

	"github.com/dgraph-io/badger/v4"
	"github.com/gin-gonic/gin"
	"github.com/yourorg/v2wall/internal/config"
	"github.com/yourorg/v2wall/internal/ipgeo"
	"github.com/yourorg/v2wall/internal/logwriter"
)

// handleReportLogs 接收被控端上报的日志数组，写入 BadgerDB
func handleReportLogs(bdb *badger.DB, searcher *ipgeo.Searcher, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var entries []logwriter.LogEntry
		if err := c.ShouldBindJSON(&entries); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json array"})
			return
		}

		processed := 0
		for _, entry := range entries {
			err := logwriter.WriteLog(bdb, entry, searcher, cfg.Master.EnableIPv4, cfg.Master.EnableIPv6)
			if err == nil {
				processed++
			}
			// 忽略单条错误，继续处理
		}
		c.JSON(http.StatusOK, gin.H{"processed": processed})
	}
}
