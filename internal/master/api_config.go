package master

import (
	"encoding/json"
	"net/http"

	"github.com/dgraph-io/badger/v4"
	"github.com/gin-gonic/gin"
	"github.com/yourorg/v2wall/internal/db"
)

// BlacklistConfig 黑名单配置，存储在 badger 中
type BlacklistConfig struct {
	RetentionDays    int `json:"retention_days"`     // 保留天数
	SyncIntervalSec  int `json:"sync_interval_sec"`  // 同步间隔（秒）
	LastSyncTime     int64 `json:"last_sync_time"`   // 最后同步时间（Unix 纳秒）
}

func getBlacklistConfig(txn *badger.Txn) (BlacklistConfig, error) {
	var cfg BlacklistConfig
	item, err := txn.Get(db.BlacklistConfigKey())
	if err == badger.ErrKeyNotFound {
		// 默认值
		return BlacklistConfig{
			RetentionDays:   30,
			SyncIntervalSec: 60,
			LastSyncTime:    0,
		}, nil
	}
	if err != nil {
		return cfg, err
	}
	err = item.Value(func(v []byte) error {
		return json.Unmarshal(v, &cfg)
	})
	return cfg, err
}

func setBlacklistConfig(txn *badger.Txn, cfg BlacklistConfig) error {
	val, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	return txn.Set(db.BlacklistConfigKey(), val)
}

func handleGetConfig(bdb *badger.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var cfg BlacklistConfig
		err := bdb.View(func(txn *badger.Txn) error {
			var e error
			cfg, e = getBlacklistConfig(txn)
			return e
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, cfg)
	}
}

func handleUpdateConfig(bdb *badger.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var newCfg BlacklistConfig
		if err := c.ShouldBindJSON(&newCfg); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		err := bdb.Update(func(txn *badger.Txn) error {
			// 保留原有的 LastSyncTime
			oldCfg, err := getBlacklistConfig(txn)
			if err != nil {
				return err
			}
			newCfg.LastSyncTime = oldCfg.LastSyncTime
			return setBlacklistConfig(txn, newCfg)
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "updated"})
	}
}
