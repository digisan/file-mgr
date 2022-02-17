package filemgr

import (
	"fmt"
	"os"
	"testing"

	"github.com/digisan/file-mgr/fdb"
	ft "github.com/digisan/file-mgr/fdb/ftype"
	"github.com/digisan/file-mgr/fdb/status"
	lk "github.com/digisan/logkit"
)

func TestLoadFileItem(t *testing.T) {
	db := fdb.GetDB("./data/fdb")
	defer db.Close()

	fis, err := db.SearchFileItems("")
	lk.FailOnErr("%v", err)
	fmt.Println(fis)
}

func TestSaveFile(t *testing.T) {

	us, err := UseUser("qing miao")
	lk.FailOnErr("%v", err)

	data, err := os.ReadFile("./go.sum")
	lk.FailOnErr("%v", err)
	err = us.SaveFile("go.sum", "this is a test 1", data)
	lk.FailOnErr("%v", err)
	fmt.Println(us)

	//
	fmt.Println("-----------------------------")

	us, err = UseUser("qing")
	lk.FailOnErr("%v", err)

	data, err = os.ReadFile("./go.sum")
	lk.FailOnErr("%v", err)
	err = us.SaveFile("go.sum", "this is a test 2", data)
	lk.FailOnErr("%v", err)
	fmt.Println(us)
}

func TestFileItemDB(t *testing.T) {

	us, err := UseUser("qing miao")
	lk.FailOnErr("%v", err)
	// fmt.Println(us)

	fis := us.SearchFileItem(ft.Text)

	fis[0].AddRefBy("abc", "def", "def", "ghi")
	fis[0].RmRefBy("abc")
	fis[0].SetStatus(status.Applying)
	lk.FailOnErr("%v", fis[0].SetGroup(0, "GRP0"))

	us.Update(fis[0])

	fmt.Println(us)
}

func TestCheck(t *testing.T) {

	us, err := UseUser("qing miao")
	lk.FailOnErr("%v", err)
	fmt.Println(us.SelfCheck(true))
	fmt.Println(us)

	us, err = UseUser("qing")
	lk.FailOnErr("%v", err)
	fmt.Println(us.SelfCheck(true))
	fmt.Println(us)
}
