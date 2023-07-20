package fdb

import (
	"sync"

	badger "github.com/dgraph-io/badger/v4"
	lk "github.com/digisan/logkit"
)

type DBGrp struct {
	sync.Mutex
	File *badger.DB
}

var (
	once  sync.Once
	DbGrp *DBGrp // global, for keeping single instance
)

func open(dir string) *badger.DB {
	opt := badger.DefaultOptions("").WithInMemory(true)
	if dir != "" {
		opt = badger.DefaultOptions(dir)
		opt.Logger = nil
	}
	db, err := badger.Open(opt)
	lk.FailOnErr("%v", err)
	return db
}

func InitDB(dir string) *DBGrp {
	if DbGrp == nil {
		once.Do(func() {
			DbGrp = &DBGrp{
				File: open(dir),
			}
		})
	}
	return DbGrp
}

func CloseDB() {
	DbGrp.Lock()
	defer DbGrp.Unlock()

	if DbGrp.File != nil {
		lk.FailOnErr("%v", DbGrp.File.Close())
		DbGrp.File = nil
	}
}
