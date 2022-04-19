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

	path, err := us0.SaveFile("go.sum", "this is a test 1", file0, "group0", "group1", "group2")
	lk.FailOnErr("%v", err)
	fmt.Println("path:", path)
	fmt.Println("us0:", us0)

	//
	fmt.Println("-----------------------------")

	us1, err := UseUser("qing")
	lk.FailOnErr("%v", err)

	file1, err := os.Open("./go.sum")
	lk.FailOnErr("%v", err)
	defer file1.Close()

	path, err = us1.SaveFile("go.sum", "this is a test 2", file1, "GROUP00", "GROUP01", "GROUP02")
	lk.FailOnErr("%v", err)
	fmt.Println("path:", path)

	file2, err := os.Open("./go.mod")
	lk.FailOnErr("%v", err)
	defer file2.Close()

	path, err = us1.SaveFile("go.mod", "this is a test 3", file2, "GROUP00", "GROUP01", "GROUP03")
	lk.FailOnErr("%v", err)
	fmt.Println("path:", path)

	fmt.Println("us1:", us1)

	fmt.Println("-----------------------------")

	/////////////////

	lvl1 := us1.PathContent("")
	fmt.Println("root:", lvl1)

	for _, path := range []string{
		"GROUP00",
		"GROUP00/GROUP01",
		"GROUP00/GROUP01/GROUP03",
		"GROUP00/GROUP01/GROUP02",
		"GROUP00/GROUP01/GROUP03/document",
		"GROUP00/GROUP01/GROUP02/document",
	} {
		fmt.Println(path, us1.PathContent(path))
	}

	////////////////////////////////////////////////

	fi := us1.FileItemByPath("GROUP00/GROUP01/GROUP02/document/go.sum")
	fmt.Println("fi:", fi)

	id := "04a17805dfdebf30b46875371a3c7d28"
	fis := us1.FileItemByID(id)
	if len(fis) > 0 {
		fmt.Println("fis[0]:", fis[0])
	} else {
		fmt.Printf("Couldn't find file item @v%s\n", id)
	}

	fmt.Println(string(us1.FileContentByID(id)))

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
