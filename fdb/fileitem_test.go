package fdb

import (
	"fmt"
	"testing"
	"time"
)

func TestFileItem(t *testing.T) {
	fi := FileItem{Tm: time.Now(), Path: "a/b/c", Note: "this is a note test"}
	fmt.Println(fi)

	dbKey, dbVal := fi.Marshal(nil)
	fmt.Println(dbKey, dbVal)

	fi.Unmarshal(dbKey, dbVal)
	fmt.Println(fi)
}
