package filemgr

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/digisan/file-mgr/fdb"
	ft "github.com/digisan/file-mgr/fdb/ftype"
	lk "github.com/digisan/logkit"
)

func TestLoadFileItem(t *testing.T) {

	db := fdb.GetDB("./data/fdb")
	defer db.Close()

	fis, err := db.SearchFileItems("")
	lk.FailOnErr("%v", err)
	fmt.Println(fis)
}

func TestDelFileItem(t *testing.T) {

	SetFileMgrRoot("./data/user-space", "./data/fdb")

	db := fdb.GetDB("./data/fdb")
	defer db.Close()

	us, err := UseUser("qing miao")
	lk.FailOnErr("%v", err)

	fmt.Println(us.DelFileItem("ab2f12c80f6341789528aebb7c0e1324"))
}

func TestSetFileItemGroup(t *testing.T) {

	SetFileMgrRoot("./data/user-space", "./data/fdb")

	us, err := UseUser("qing miao")
	lk.FailOnErr("%v", err)

	us.SetFIGroup("ab2f12c80f6341789528aebb7c0e1324", 2, "G3")

	us.SelfCheck(true)
}

func TestSetNote(t *testing.T) {

	SetFileMgrRoot("./data/user-space", "./data/fdb")

	us, err := UseUser("qing miao")
	lk.FailOnErr("%v", err)

	us.SetFINote("ab2f12c80f6341789528aebb7c0e1324", "This is a Set Note 2")

}

func TestSaveFileV2(t *testing.T) {

	SetFileMgrRoot("./data/user-space", "./data/fdb")

	us, err := UseUser("qing miao")
	lk.FailOnErr("%v", err)

	///////////////////////////////////////////

	for i, fname := range []string{"go.mod", "go.sum"} {

		file, err := os.Open(fname)
		lk.FailOnErr("%v", err)
		defer file.Close()

		path, err := us.SaveFile(fname, fmt.Sprintf("this is a test %d", i), file, "group0", "group1", "group2")
		lk.FailOnErr("%v", err)

		fmt.Println("---path:", path)
		// fmt.Println("---us:", us)

		fmt.Println("-----------------------------------")
	}

	/////////////////////////////////////////////////////////////////////////

	lvl1 := us.PathContent("")
	fmt.Println("root:", lvl1)

	for _, path := range [][]string{
		{},
		{"G0"},
		{"G0", "group1"},
		{"G0", "group1", "group2"},
		{"G0", "group1", "group2", "document"},
	} {
		fmt.Println("2022-05/"+filepath.Join(path...), us.PathContent("2022-05", path...))
	}

	/////////////////////////////////////////////////////////////////////////

	fmt.Println()

	// fi := us.FileItemsByPath("G0/group1/group2/document/go.mod")
	// fmt.Println("fi:", fi)

	fmt.Println()

	id := "ab2f12c80f6341789528aebb7c0e1324"
	fis := us.FileItems(id)
	if len(fis) > 0 {
		fmt.Println("fis[0]:", fis[0])
	} else {
		fmt.Printf("Couldn't find file item @%s\n", id)
	}

	fmt.Println(string(us.FirstFileContent(id)))

	us.SelfCheck(true)
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

	for _, path := range [][]string{
		{},
		{"GROUP00"},
		{"GROUP00", "GROUP01"},
		{"GROUP00", "GROUP01", "GROUP03"},
		{"GROUP00", "GROUP01", "GROUP02"},
		{"GROUP00", "GROUP01", "GROUP03", "document"},
		{"GROUP00", "GROUP01", "GROUP02", "document"},
	} {
		fmt.Println("2022-05/"+filepath.Join(path...), us1.PathContent("2022-05", path...))
	}

	////////////////////////////////////////////////

	// fi := us1.FileItemsByPath("GROUP00/GROUP01/GROUP02/document/go.sum")
	// fmt.Println("fi:", fi)

	id := "8301ceb3ea3b5bf311fcab06f304ae14"
	fis := us1.FileItems(id)
	if len(fis) > 0 {
		fmt.Println("fis[0]:", fis[0])
	} else {
		fmt.Printf("Couldn't find file item @v%s\n", id)
	}

	fmt.Println(string(us1.FirstFileContent(id)))

	// fi.AddNote
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

	path, err := fis[0].SetGroup(0, "GRP1")
	lk.FailOnErr("%v @ %v", path, err)

	lk.FailOnErr("%v", us.UpdateFileItem(fis[0], true))
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
