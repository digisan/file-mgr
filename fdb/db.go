package fdb

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	badger "github.com/dgraph-io/badger/v3"
	. "github.com/digisan/go-generics/v2"
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

// [id] is prefix, could remove many fi
func (db *FDB) RemoveFileItems(id string, lock bool) error {
	if lock {
		db.Lock()
		defer db.Unlock()
	}

	if len(id) < 32 {
		return errors.New("id length MUST greater than 32")
	}

	id = strings.ToLower(id)
	return db.dbFile.Update(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(id)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			if err := txn.Delete(it.Item().KeyCopy(nil)); err != nil {
				lk.WarnOnErr("%v", err)
				return err
			}
		}
		return nil
	})
}

// exactly update ONE fi
func (db *FDB) UpdateFileItem(fi *FileItem) error {
	db.Lock()
	defer db.Unlock()

	if fi.prevPath == "" {
		fi.prevPath = fi.Path
		defer func() { fi.prevPath = "" }()
	}

	// exactly remove ONE fi
	if err := db.RemoveFileItems(fi.Id, false); err != nil {
		return err
	}

	// exactly update ONE fi
	return db.dbFile.Update(func(txn *badger.Txn) error {
		return txn.Set(fi.Marshal())
	})
}

func (db *FDB) FirstFileItem(id string) (*FileItem, bool, error) {
	db.Lock()
	defer db.Unlock()

	id = strings.ToLower(id)
	var err error
	fi, ret := &FileItem{}, false
	err = db.dbFile.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(id)
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

	return nil, false, err
}

func (db *FDB) SearchFileItems(ftype string, groups ...string) (fis []*FileItem, err error) {
	if ftype != "" && NotIn(ftype, FileTypes()...) {
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
	id = strings.ToLower(id)
	fi, ok, err := db.FirstFileItem(id)
	return err == nil && ok && fi != nil
}
