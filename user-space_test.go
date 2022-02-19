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

	file0, err := os.Open("./go.sum")
	lk.FailOnErr("%v", err)
	defer file0.Close()

	err = us.SaveFile("go.sum", "this is a test 1", file0, "group0", "group1", "group2")
	lk.FailOnErr("%v", err)
	fmt.Println(us)

	//
	fmt.Println("-----------------------------")

	us, err = UseUser("qing")
	lk.FailOnErr("%v", err)

	file1, err := os.Open("./go.sum")
	lk.FailOnErr("%v", err)
	defer file1.Close()

	err = us.SaveFile("go.sum", "this is a test 2", file1, "GROUP00", "GROUP01", "GROUP02")
	lk.FailOnErr("%v", err)
	fmt.Println(us)
}

func TestFileItemDB(t *testing.T) {

	us, err := UseUser("qing miao")
	lk.FailOnErr("%v", err)
	// fmt.Println(us)

	fis := us.SearchFileItem(ft.All, "g*0", "gr*1", "*2")

	lk.FailOnErrWhen(len(fis) == 0, "%v", fmt.Errorf("fis not found"))

	fis[0].AddRefBy("abc", "def", "def", "ghi")
	fis[0].RmRefBy("abc")
	fis[0].SetStatus(status.Applying)
	lk.FailOnErr("%v", fis[0].SetGroup(1, "GRP1"))

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
