package fdb

import (
	"fmt"
	"testing"
	"time"

	lk "github.com/digisan/logkit"
	"github.com/google/uuid"
)

var (
	id = uuid.New().String()
)

func TestUpdateFile(t *testing.T) {
	InitDB("../data")
	defer CloseDB()

	fi := &FileItem{
		Id:        id,
		Path:      "a/b/c/d",
		Tm:        time.Now(),
		GroupList: "",
		Note:      "this is a note test",
	}
	fmt.Println(fi)
	fmt.Println(UpdateFileItem(fi))
}

func TestLoadFile(t *testing.T) {
	InitDB("../data")
	defer CloseDB()

	fi, ok, err := FirstFileItem("7cfb1626-c129-473a-9b19-58a27da5b837")
	fmt.Println(fi, ok, err)
}

func TestListFile(t *testing.T) {
	InitDB("../data")
	defer CloseDB()

	fis, err := ListFileItems(func(fi *FileItem) bool {
		return true
	})
	lk.FailOnErr("%v", err)
	fmt.Println(len(fis))
	for _, fi := range fis {
		fmt.Println("--------------------------------")
		fmt.Println(fi)
	}
}
