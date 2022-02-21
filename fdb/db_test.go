package fdb

import (
	"fmt"
	"testing"
	"time"

	lk "github.com/digisan/logkit"
)

func TestUpdateFile(t *testing.T) {

	SetDbRoot("../data")

	fdb := GetDB()
	defer fdb.Close()

	fi := &FileItem{
		Id:        "id",
		Path:      "a/b/c/d",
		Tm:        time.Now().String(),
		Status:    "received",
		GroupList: "",
		Note:      "this is a note test",
		RefBy:     "ID111",
	}
	fmt.Println(fi)

	fdb.UpdateFileItem(fi)
}

func TestLoadFile(t *testing.T) {

	SetDbRoot("../data")

	fdb := GetDB()
	defer fdb.Close()

	fi, ok, err := fdb.LoadFileItem("id", "received")
	fmt.Println(fi, ok, err)
}

func TestListFile(t *testing.T) {

	SetDbRoot("../data")

	fdb := GetDB()
	defer fdb.Close()

	fis, err := fdb.ListFileItems(func(fi *FileItem) bool {
		return true
	})
	lk.FailOnErr("%v", err)
	for _, fi := range fis {
		fmt.Println("--------------------------------")
		fmt.Println(fi)
	}
}
