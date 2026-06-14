package db

import (
    "encoding/binary"
    "github.com/dgraph-io/badger/v4"
)

// getUint64 从value中解码uint64值
func getUint64(val []byte) uint64 {
    if len(val) == 0 {
        return 0
    }
    return binary.BigEndian.Uint64(val)
}

// putUint64 将uint64值编码为字节切片
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
    // 实现与GetCounter类似，但将val转换为int64
    // ...（此处省略具体代码，参考现有实现）
    return int64(val), err
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
