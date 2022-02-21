package fdb

import (
	"fmt"
	"strings"
	"sync"

	badger "github.com/dgraph-io/badger/v3"
	ft "github.com/digisan/file-mgr/fdb/ftype"
	"github.com/digisan/file-mgr/fdb/status"
	"github.com/digisan/go-generics/str"
	lk "github.com/digisan/logkit"
)

var once sync.Once

type FDB struct {
	sync.Mutex
	dbPath string
	dbFile *badger.DB
}

var fDB *FDB // global, for keeping single instance

func open(dir string) *badger.DB {
	opt := badger.DefaultOptions("").WithInMemory(true)
	if dir != "" {
		opt = badger.DefaultOptions(dir)
	}
	db, err := badger.Open(opt)
	lk.FailOnErr("%v", err)
	return db
}

func GetDB(dir string) *FDB {
	if fDB == nil {
		once.Do(func() {
			fDB = &FDB{
				dbPath: dir,
				dbFile: open(dir),
			}
		})
	}
	return fDB
}

func (db *FDB) Close() {
	db.Lock()
	defer db.Unlock()

	if db.dbFile != nil {
		lk.FailOnErr("%v", db.dbFile.Close())
		db.dbFile = nil
	}
}

///////////////////////////////////////////////////////////////

func (db *FDB) RemoveFileItem(id, path string, lock bool) error {
	if lock {
		db.Lock()
		defer db.Unlock()
	}

	prefixList := [][]byte{}
	for _, stat := range status.AllStatus() {
		prefix := stat + SEP + id + SEP + path + SEP
		prefixList = append(prefixList, []byte(prefix))
	}

	return db.dbFile.Update(func(txn *badger.Txn) (err error) {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for _, prefix := range prefixList {
			for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
				if err = txn.Delete(it.Item().KeyCopy(nil)); err != nil {
					lk.WarnOnErr("%v", err)
					return err
				}
			}
		}
		return err
	})
}

func (db *FDB) UpdateFileItem(fi *FileItem) error {
	db.Lock()
	defer db.Unlock()

	if fi.PPath == "" {
		fi.PPath = fi.Path
		defer func() { fi.PPath = "" }()
	}
	if err := db.RemoveFileItem(fi.Id, fi.PPath, false); err != nil {
		return err
	}
	return db.dbFile.Update(func(txn *badger.Txn) error {
		return txn.Set(fi.Marshal())
	})
}

func (db *FDB) LoadFileItem(id, stat string) (*FileItem, bool, error) {
	db.Lock()
	defer db.Unlock()

	prefixList := [][]byte{}
	if stat == "" {
		for _, stat := range status.AllStatus() {
			prefix := stat + SEP + id + SEP
			prefixList = append(prefixList, []byte(prefix))
		}
	} else {
		prefix := stat + SEP + id + SEP
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

func (db *FDB) SearchFileItems(ftype string, groups ...string) (fis []*FileItem, err error) {
	if ftype != "" && str.NotIn(ftype, ft.AllFileType()...) {
		return nil, fmt.Errorf("file type [%s] is unregistered", ftype)
	}
	return db.ListFileItems(func(fi *FileItem) bool {
		if ftype != "" {
			return fi.Type() == ftype && strings.HasPrefix(fi.GroupList, strings.Join(groups, SEP_GRP))
		}
		if ftype == "" && len(groups) > 0 {
			return strings.HasPrefix(fi.GroupList, strings.Join(groups, SEP_GRP))
		}
		return true
	})
}

func (db *FDB) ListFileItems(filter func(*FileItem) bool) (fis []*FileItem, err error) {
	db.Lock()
	defer db.Unlock()

	err = db.dbFile.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			item.Value(func(v []byte) error {
				fi := &FileItem{}
				fi.Unmarshal(item.Key(), v)
				if filter(fi) {
					fis = append(fis, fi)
				}
				return nil
			})
		}
		return nil
	})
	return
}

func (db *FDB) IsExisting(id string) bool {
	fi, ok, err := db.LoadFileItem(id, "")
	return err == nil && ok && fi.Status != status.Deleted
}
