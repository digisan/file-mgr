package fdb

import (
	"sync"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/digisan/file-mgr/fdb/status"
	lk "github.com/digisan/logkit"
)

var (
	mDbUsing = &sync.Map{}
)

type FDB struct {
	sync.Mutex
	dbPath string
	dbFile *badger.DB
}

func open(dir string) *badger.DB {
	defer func() { mDbUsing.Store(dir, true) }()
	opt := badger.DefaultOptions("").WithInMemory(true)
	if dir != "" {
		opt = badger.DefaultOptions(dir)
	}
	db, err := badger.Open(opt)
	lk.FailOnErr("%v", err)
	return db
}

func GetDB(dir string) *FDB {
	val, ok := mDbUsing.Load(dir)
	if ok && val.(bool) {
		return nil
	}
	return &FDB{
		dbPath: dir,
		dbFile: open(dir),
	}
}

func (db *FDB) Close() {
	defer func() { mDbUsing.Store(db.dbPath, false) }()
	db.Lock()
	defer db.Unlock()
	lk.FailOnErr("%v", db.dbFile.Close())
}

///////////////////////////////////////////////////////////////

func (db *FDB) RemoveFile(path string, lock bool) error {
	if lock {
		db.Lock()
		defer db.Unlock()
	}

	prefixList := [][]byte{}
	for _, stat := range status.AllStatus() {
		prefix := stat + SEP + path + SEP
		prefixList = append(prefixList, []byte(prefix))
	}

	return db.dbFile.Update(func(txn *badger.Txn) (err error) {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for _, prefix := range prefixList {
			for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
				if err = txn.Delete(it.Item().KeyCopy(nil)); err != nil {
					return err
				}
			}
		}
		return err
	})
}

func (db *FDB) UpdateFile(fi *FileItem) error {
	db.Lock()
	defer db.Unlock()

	if err := db.RemoveFile(fi.Path, false); err != nil {
		return err
	}
	return db.dbFile.Update(func(txn *badger.Txn) error {
		return txn.Set(fi.Marshal())
	})
}

func (db *FDB) LoadFile(path string) (*FileItem, bool, error) {
	db.Lock()
	defer db.Unlock()

	prefixList := [][]byte{}
	for _, stat := range status.AllStatus() {
		prefix := stat + SEP + path + SEP
		prefixList = append(prefixList, []byte(prefix))
	}

	var err error
	for _, prefix := range prefixList {
		fi, ret := &FileItem{}, false
		err = db.dbFile.View(func(txn *badger.Txn) error {
			it := txn.NewIterator(badger.DefaultIteratorOptions)
			defer it.Close()
			if it.Seek(prefix); it.ValidForPrefix(prefix) {
				item := it.Item()
				k := item.Key()
				return item.Value(func(v []byte) error {
					// fmt.Printf("key=%s, value=%s\n", k, v)
					ret = true
					fi.Unmarshal(k, v)
					return nil
				})
			}
			return nil
		})
		if ret {
			return fi, fi.Path != "", err
		}
	}

	return nil, false, err
}
