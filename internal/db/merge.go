package db

import (
	"encoding/binary"
	"math"

	"github.com/dgraph-io/badger/v4"
	"github.com/dgraph-io/badger/v4/y"
)

// RegisterCounterMerge 为 counter 键注册加法合并函数。
// 在数据库打开后调用一次。
func RegisterCounterMerge(db *badger.DB) {
	db.RegisterMergeOperator(
		func(existing, new []byte) []byte {
			existingVal := y.BytesToU64(existing)
			newVal := y.BytesToU64(new)
			return y.U64ToBytes(existingVal + newVal)
		},
	)
}

// GetCounter 读取 counter 值
func GetCounter(txn *badger.Txn, ip string) (uint64, error) {
	item, err := txn.Get(CounterKey(ip))
	if err == badger.ErrKeyNotFound {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	var val uint64
	err = item.Value(func(v []byte) error {
		val = y.BytesToU64(v)
		return nil
	})
	return val, err
}

// GetLastSeen 读取最后访问时间（纳秒）
func GetLastSeen(txn *badger.Txn, ip string) (int64, error) {
	item, err := txn.Get(LastSeenKey(ip))
	if err == badger.ErrKeyNotFound {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	var ts uint64
	err = item.Value(func(v []byte) error {
		ts = y.BytesToU64(v)
		return nil
	})
	return int64(ts), err
}

// GetIPInfo 获取地理信息
func GetIPInfo(txn *badger.Txn, ip string) (string, error) {
	item, err := txn.Get(IPInfoKey(ip))
	if err == badger.ErrKeyNotFound {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	var info string
	err = item.Value(func(v []byte) error {
		info = string(v)
		return nil
	})
	return info, err
}
