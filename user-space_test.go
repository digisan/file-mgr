package filemgr

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/digisan/file-mgr/fdb"
	lk "github.com/digisan/logkit"
	"github.com/google/uuid"
)

var (
	id = uuid.New().String()
)

func TestLoadFileItem(t *testing.T) {

	fdb.InitDB("./data/user-fdb")
	defer fdb.CloseDB()

	fis, err := fdb.SearchFileItems("video")
	lk.FailOnErr("%v", err)
	fmt.Println(fis)
}

func TestDelFileItem(t *testing.T) {

	SetFileMgrRoot("./data")

	us, err := UseUser("qing miao")
	lk.FailOnErr("%v", err)

	fmt.Println(us.DelFileItem("ab2f12c80f6341789528aebb7c0e1324"))
}

func TestSetFileItemGroup(t *testing.T) {

	SetFileMgrRoot("./data")

	us, err := UseUser("qing miao")
	lk.FailOnErr("%v", err)

	fmt.Println(us.SetFIGroup("ab2f12c80f6341789528aebb7c0e1324", 2, "G3"))
	us.SelfCheck(true)
}

func TestSetNote(t *testing.T) {

	SetFileMgrRoot("./data")

	us, err := UseUser("qing miao")
	lk.FailOnErr("%v", err)

	fmt.Println(us.SetFINote("ab2f12c80f6341789528aebb7c0e1324", "This is a Set Note 2"))
}

func TestSaveFileV2(t *testing.T) {

	SetFileMgrRoot("./data")

	us, err := UseUser("qing miao")
	lk.FailOnErr("%v", err)

	// "./samples/key.txt" overwrites "./samples/key", but in reality, it shouldn't happen unless same files both arrive in same millisecond
	// for i, fpath := range []string{"./samples/moon", "./samples/moonpdf", "./samples/moondoc", "./samples/Screencast", "./samples/key", "./samples/key.txt"} {
	for i, fpath := range []string{"./samples/moon", "./samples/moonpdf", "./samples/moondoc", "./samples/Screencast", "./samples/key.txt"} {

		file, err := os.Open(fpath)
		lk.FailOnErr("%v", err)
		defer file.Close()

		fname := filepath.Base(fpath)
		path, err := us.SaveFile(fname, fmt.Sprintf("this is a test note %d", i), file, "group0", "group1", "group2")
		lk.FailOnErr("%v", err)

		fmt.Println("---path:", path)
		// fmt.Println("---us:", us)

		fmt.Println("-----------------------------------")
	}
}

func TestListAllFI(t *testing.T) {
	SetFileMgrRoot("./data")

	fdb.InitDB("./data/user-fdb")
	defer fdb.CloseDB()

	fis, err := fdb.ListFileItems(nil)
	if err != nil {
		panic(err)
	}
	for _, fi := range fis {
		fmt.Println(fi)
	}
}

func TestPathContent(t *testing.T) {

	SetFileMgrRoot("./data")

	us, err := UseUser("qing miao")
	lk.FailOnErr("%v", err)

	lvl1 := us.PathContent("")
	fmt.Println("root:", lvl1)

	for _, path := range [][]string{
		{},
		{"group0"},
		{"group0", "group1"},
		{"group0", "group1", "group2"},
		{"group0", "group1", "group2", "document"},
		{"group0", "group1", "group2", "video"},
		{"group0", "group1", "group2", "image"},
		{"group0", "group1", "group2", "archive"},
		{"group0", "group1", "group2", "unknown"},
	} {
		fmt.Println("2022-07/"+filepath.Join(path...), "==>>", us.PathContent("2022-07", path...))
	}
}

func TestFileContent(t *testing.T) {

	SetFileMgrRoot("./data")

	us, err := UseUser("qing miao")
	lk.FailOnErr("%v", err)

	// fi := us.FileItemsByPath("G0/group1/group2/document/go.mod")
	// fmt.Println("fi:", fi)

	fmt.Println()

	id := "6bfb019bc8c2c4d4e97f978e62b05f3a"
	fis, err := us.FileItems(id)
	lk.FailOnErr("%v", err)
	if len(fis) > 0 {
		fmt.Println("fis[0]:", fis[0])
	} else {
		fmt.Printf("Couldn't find file item @%s\n", id)
	}

	data, err := us.FirstFileContent(id)
	lk.FailOnErr("%v", err)
	fmt.Println(string(data))

	us.SelfCheck(true)
}

// func TestSaveFile(t *testing.T) {

// 	SetFileMgrRoot("./data")

// 	us0, err := UseUser("qing miao")
// 	lk.FailOnErr("%v", err)

// 	file0, err := os.Open("./go.sum")
// 	lk.FailOnErr("%v", err)
// 	defer file0.Close()

// 	path, err := us0.SaveFile("go.sum", "this is a test 1", file0, "group0", "group1", "group2")
// 	lk.FailOnErr("%v", err)
// 	fmt.Println("path:", path)
// 	fmt.Println("us0:", us0)

// 	//
// 	fmt.Println("-----------------------------")

// 	us1, err := UseUser("qing")
// 	lk.FailOnErr("%v", err)

// 	file1, err := os.Open("./go.sum")
// 	lk.FailOnErr("%v", err)
// 	defer file1.Close()

// 	path, err = us1.SaveFile("go.sum", "this is a test 2", file1, "GROUP00", "GROUP01", "GROUP02")
// 	lk.FailOnErr("%v", err)
// 	fmt.Println("path:", path)

// 	file2, err := os.Open("./go.mod")
// 	lk.FailOnErr("%v", err)
// 	defer file2.Close()

// 	path, err = us1.SaveFile("go.mod", "this is a test 3", file2, "GROUP00", "GROUP01", "GROUP03")
// 	lk.FailOnErr("%v", err)
// 	fmt.Println("path:", path)

// 	fmt.Println("us1:", us1)

// 	fmt.Println("-----------------------------")

// 	/////////////////

// 	lvl1 := us1.PathContent("")
// 	fmt.Println("root:", lvl1)

// 	for _, path := range [][]string{
// 		{},
// 		{"GROUP00"},
// 		{"GROUP00", "GROUP01"},
// 		{"GROUP00", "GROUP01", "GROUP03"},
// 		{"GROUP00", "GROUP01", "GROUP02"},
// 		{"GROUP00", "GROUP01", "GROUP03", "document"},
// 		{"GROUP00", "GROUP01", "GROUP02", "document"},
// 	} {
// 		fmt.Println("2022-07/"+filepath.Join(path...), us1.PathContent("2022-07", path...))
// 	}

// 	////////////////////////////////////////////////

// 	// fi := us1.FileItemsByPath("GROUP00/GROUP01/GROUP02/document/go.sum")
// 	// fmt.Println("fi:", fi)

// 	id := "8301ceb3ea3b5bf311fcab06f304ae14"
// 	fis, err := us1.FileItems(id)
// 	lk.FailOnErr("%v", err)
// 	if len(fis) > 0 {
// 		fmt.Println("fis[0]:", fis[0])
// 	} else {
// 		fmt.Printf("Couldn't find file item @v%s\n", id)
// 	}

// 	data, err := us1.FirstFileContent(id)
// 	lk.FailOnErr("%v", err)
// 	fmt.Println(string(data))
// }

func TestUpdateFileItem(t *testing.T) {

	SetFileMgrRoot("./data")

	us, err := UseUser("qing miao")
	lk.FailOnErr("%v", err)
	// fmt.Println(us)

	fis := us.SearchFileItem(fdb.Any, "*", "*", "*2")
	// lk.FailOnErrWhen(len(fis) == 0, "%v", fmt.Errorf("fis not found"))

	for _, fi := range fis {
		path, err := fi.SetGroup(0, "GRP0")
		lk.FailOnErr("%v @ %v", path, err)
		lk.FailOnErr("%v", us.UpdateFileItem(fi, true))
	}

	us.SelfCheck(true) // remove empty directories
	fmt.Println(us)
}

func TestCheck(t *testing.T) {

	SetFileMgrRoot("./data")

	us, err := UseUser("qing miao")
	lk.FailOnErr("%v", err)
	fmt.Println(us.SelfCheck(true))
	fmt.Println(us)
}
