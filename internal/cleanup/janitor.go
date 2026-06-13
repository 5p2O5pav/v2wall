package cleanup

import (
	"log"
	"time"

	"github.com/dgraph-io/badger/v4"
	"v2wall/internal/db"
)

// Janitor 定期清理过期数据
type Janitor struct {
	DB            *badger.DB
	RetentionDays int
	StopCh        chan struct{}
}

func NewJanitor(bdb *badger.DB, retentionDays int) *Janitor {
	return &Janitor{
		DB:            bdb,
		RetentionDays: retentionDays,
		StopCh:        make(chan struct{}),
	}
}

func (j *Janitor) Run(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			j.clean()
		case <-j.StopCh:
			return
		}
	}
}

func (j *Janitor) clean() {
	cutoff := time.Now().AddDate(0, 0, -j.RetentionDays).UnixNano()
	log.Printf("[janitor] cleaning data before %d", cutoff)

	err := j.DB.Update(func(txn *badger.Txn) error {
		// 遍历 last_seen 键，删除过期的所有关联键
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte(db.PrefixLastSeen)
		opts.PrefetchValues = false
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
			if ts < cutoff {
				// 删除该 IP 下的所有日志键
				logPrefix := []byte(db.PrefixLog + ip + ":")
				logIt := txn.NewIterator(badger.DefaultIteratorOptions)
				for logIt.Seek(logPrefix); logIt.ValidForPrefix(logPrefix); logIt.Next() {
					txn.Delete(logIt.Item().Key())
				}
				logIt.Close()

				// 删除 counter, last_seen, ipinfo
				txn.Delete(db.CounterKey(ip))
				txn.Delete(key) // last_seen 就是当前 key
				txn.Delete(db.IPInfoKey(ip))
			}
		}
		return nil
	})
	if err != nil {
		log.Printf("[janitor] cleanup error: %v", err)
	}
}

func (j *Janitor) Stop() {
	close(j.StopCh)
}
