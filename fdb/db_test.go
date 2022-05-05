package fdb

import (
	"fmt"
	"testing"
	"time"

	lk "github.com/digisan/logkit"
)

func TestUpdateFile(t *testing.T) {

	fdb := GetDB("../data")
	defer fdb.Close()

	fi := &FileItem{
		Id:        "id",
		Path:      "a/b/c/d",
		Tm:        time.Now(),
		GroupList: "",
		Note:      "this is a note test",
	}
	fmt.Println(fi)

	fdb.UpdateFileItem(fi)
}

func TestLoadFile(t *testing.T) {

	fdb := GetDB("../data")
	defer fdb.Close()

	fi, ok, err := fdb.FirstFileItem("id")
	fmt.Println(fi, ok, err)
}

func TestListFile(t *testing.T) {

	fdb := GetDB("../data")
	defer fdb.Close()

	fis, err := fdb.ListFileItems(func(fi *FileItem) bool {
		return true
	})
	lk.FailOnErr("%v", err)
	fmt.Println(len(fis))
	for _, fi := range fis {
		fmt.Println("--------------------------------")
		fmt.Println(fi)
	}
}
