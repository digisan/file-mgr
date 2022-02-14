package fdb

import (
	"fmt"
	"testing"
	"time"
)

func TestUpdateFile(t *testing.T) {
	fdb := GetDB("../data")
	defer fdb.Close()

	fi := &FileItem{
		Path:        "a/b/c",
		Tm:          time.Now().String(),
		Status:      "received",
		GroupList:   "",
		Description: "this is a description test",
		RefBy:       "ID111",
	}
	fmt.Println(fi)

	fdb.UpdateFile(fi)
}
