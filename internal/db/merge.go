package db

import (
	"encoding/binary"

	"github.com/dgraph-io/badger/v4"
)

// getUint64 从字节切片中解码 uint64（大端序）
func getUint64(val []byte) uint64 {
	if len(val) == 0 {
		return 0
	}
	return binary.BigEndian.Uint64(val)
}

// putUint64 将 uint64 编码为字节切片（大端序）
func putUint64(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
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
		val = getUint64(v)
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
		ts = getUint64(v)
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
