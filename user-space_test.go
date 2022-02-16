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

func TestSaveFile(t *testing.T) {

	us := NewUserSpace("qing")
	lk.FailOnErrWhen(us == nil, "%v", fmt.Errorf("us init error"))
	fmt.Println(us)

	data, err := os.ReadFile("./go.sum")
	lk.FailOnErr("%v", err)
	err = us.SaveFile("go.sum", "this is a test", data)
	lk.FailOnErr("%v", err)
}

func TestFDB(t *testing.T) {

	db := fdb.GetDB("./data/fdb")
	defer db.Close()

	fis, err := db.SearchFileItems(ft.Text, "")
	lk.FailOnErr("%v", err)

	fis[0].AddRefBy("abc", "def", "def", "ghi")
	fis[0].RmRefBy("abc")
	fis[0].SetStatus(status.Applying)
	lk.FailOnErr("%v", fis[0].SetGroup(0, "GRP6"))

	db.UpdateFileItem(fis[0])

	for _, fi := range fis {
		fmt.Println(fi)
	}
}

func TestLoadFileItem(t *testing.T) {
	db := fdb.GetDB("./data/fdb")
	defer db.Close()

	fis, err := db.SearchFileItems("")
	lk.FailOnErr("%v", err)
	fmt.Println(fis)

	// fi, ok, err := db.LoadFileItem("cf7851b71a462087ce36705f182c50ff", "")
	// fmt.Println(fi, ok, err)
}
