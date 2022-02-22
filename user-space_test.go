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

	SetFileMgrRoot("./data/user-space", "./data/fdb")

	us0, err := UseUser("qing miao")
	lk.FailOnErr("%v", err)

	file0, err := os.Open("./go.sum")
	lk.FailOnErr("%v", err)
	defer file0.Close()

	err = us0.SaveFile("go.sum", "this is a test 1", file0, "group0", "group1", "group2")
	lk.FailOnErr("%v", err)
	fmt.Println(us0)

	//
	fmt.Println("-----------------------------")

	us1, err := UseUser("qing")
	lk.FailOnErr("%v", err)

	file1, err := os.Open("./go.sum")
	lk.FailOnErr("%v", err)
	defer file1.Close()

	err = us1.SaveFile("go.sum", "this is a test 2", file1, "GROUP00", "GROUP01", "GROUP02")
	lk.FailOnErr("%v", err)

	file2, err := os.Open("./go.mod")
	lk.FailOnErr("%v", err)
	defer file2.Close()

	err = us1.SaveFile("go.mod", "this is a test 3", file2, "GROUP00", "GROUP01", "GROUP03")
	lk.FailOnErr("%v", err)

	fmt.Println(us1)

	/////////////////

	lvl1 := us1.PathContent("")
	fmt.Println("root:", lvl1)

	lvl1 = us1.PathContent("GROUP00")
	fmt.Println("GROUP00:", lvl1)

	lvl1 = us1.PathContent("GROUP00/GROUP01")
	fmt.Println("GROUP00/GROUP01:", lvl1)

	lvl1 = us1.PathContent("GROUP00/GROUP01/GROUP03")
	fmt.Println("GROUP00/GROUP01/GROUP03:", lvl1)

	lvl1 = us1.PathContent("GROUP00/GROUP01/GROUP02")
	fmt.Println("GROUP00/GROUP01/GROUP02:", lvl1)

	lvl1 = us1.PathContent("GROUP00/GROUP01/GROUP03/document")
	fmt.Println("GROUP00/GROUP01/GROUP03/document:", lvl1)

	lvl1 = us1.PathContent("GROUP00/GROUP01/GROUP02/document")
	fmt.Println("GROUP00/GROUP01/GROUP02/document:", lvl1)

	////////////////////////////////////////////////

	fi := us1.FileItemByPath("GROUP00/GROUP01/GROUP02/document/go.sum")
	fmt.Println(fi)

	fis := us1.FileItemByID("cf7851b71a462087ce36705f182c50ff")
	fmt.Println(fis[0])

	// fi.AddRefBy
	// fi.RmRefBy
	// fi.AddNote
	// fi.SetStatus
	// fi.SetGroup
	// us1.Update(fi, true)
}

func TestFileItemDB(t *testing.T) {

	SetFileMgrRoot("./data/user-space", "./data/fdb")

	us, err := UseUser("qing miao")
	lk.FailOnErr("%v", err)
	// fmt.Println(us)

	fis := us.SearchFileItem(ft.All, "*", "*", "*2")

	lk.FailOnErrWhen(len(fis) == 0, "%v", fmt.Errorf("fis not found"))

	fis[0].AddRefBy("abc", "def", "def", "ghi")
	fis[0].RmRefBy("abc")
	fis[0].SetStatus(status.Approved)
	lk.FailOnErr("%v", fis[0].SetGroup(0, "GRP1"))

	lk.FailOnErr("%v", us.Update(fis[0], true))
	// us.SelfCheck(true) // remove empty directories

	fmt.Println(us)
}

func TestCheck(t *testing.T) {

	SetFileMgrRoot("./data/user-space", "./data/fdb")

	us, err := UseUser("qing miao")
	lk.FailOnErr("%v", err)
	fmt.Println(us.SelfCheck(true))
	fmt.Println(us)

	us, err = UseUser("qing")
	lk.FailOnErr("%v", err)
	fmt.Println(us.SelfCheck(true))
	fmt.Println(us)
}
