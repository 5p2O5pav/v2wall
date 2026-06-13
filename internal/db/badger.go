package db

import (
	"github.com/dgraph-io/badger/v4"
	"github.com/dgraph-io/badger/v4/options"
)

func OpenDB(path string) (*badger.DB, error) {
	opts := badger.DefaultOptions(path).
		WithNumVersionsToKeep(1).
		WithCompression(options.ZSTD).
		WithLoggingLevel(badger.WARNING)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return db, nil
}
