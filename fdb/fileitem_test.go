package fdb

import (
	"fmt"
	"testing"
	"time"
)

func TestFileItem(t *testing.T) {
	fi := FileItem{Tm: time.Now().String(), Path: "a/b/c", Status: "received", Note: "this is a note test", RefBy: "ID111"}
	fmt.Println(fi)

	dbKey, dbVal := fi.Marshal()
	fmt.Println(dbKey, dbVal)

	fi.Unmarshal(dbKey, dbVal)
	fmt.Println(fi)
}
