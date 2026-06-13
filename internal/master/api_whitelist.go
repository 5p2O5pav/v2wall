package master

import (
	"net"
	"net/http"

	"github.com/dgraph-io/badger/v4"
	"github.com/gin-gonic/gin"
	"github.com/yourorg/v2wall/internal/db"
)

type WhitelistEntry struct {
	CIDR   string `json:"cidr"`
	Remark string `json:"remark"`
}

// handleGetWhitelist 获取所有白名单
func handleGetWhitelist(bdb *badger.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var entries []WhitelistEntry
		err := bdb.View(func(txn *badger.Txn) error {
			opts := badger.DefaultIteratorOptions
			opts.Prefix = []byte(db.PrefixWhitelist)
			it := txn.NewIterator(opts)
			defer it.Close()
			for it.Rewind(); it.Valid(); it.Next() {
				item := it.Item()
				err := item.Value(func(v []byte) error {
					var entry WhitelistEntry
					if json.Unmarshal(v, &entry) == nil {
						entries = append(entries, entry)
					}
					return nil
				})
				if err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": entries})
	}
}

// handleAddWhitelist 添加白名单
func handleAddWhitelist(bdb *badger.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var entry WhitelistEntry
		if err := c.ShouldBindJSON(&entry); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// 校验 CIDR
		_, _, err := net.ParseCIDR(entry.CIDR)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid CIDR"})
			return
		}

		key := db.WhitelistKey(entry.CIDR)
		val, _ := json.Marshal(entry)
		err = bdb.Update(func(txn *badger.Txn) error {
			return txn.Set(key, val)
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "added"})
	}
}

// handleDeleteWhitelist 删除白名单
func handleDeleteWhitelist(bdb *badger.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id") // 实际上是 CIDR
		key := db.WhitelistKey(id)
		err := bdb.Update(func(txn *badger.Txn) error {
			return txn.Delete(key)
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	}
}
