package filemgr

import (
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/digisan/file-mgr/fdb"
	"github.com/digisan/file-mgr/fdb/status"
	fd "github.com/digisan/gotk/filedir"
	gio "github.com/digisan/gotk/io"
)

type UserSpace struct {
	UName    string
	UserPath string
}

var (
	rootfdb = "./data/fdb"
	root    = "./data/user-space"
)

func SetRoot(rtus, rtfdb string) {
	if rtus != "" {
		root = rtus
	}
	if rtfdb != "" {
		rootfdb = rtfdb
	}
}

func NewUserSpace(name string) *UserSpace {
	return (&UserSpace{UName: name}).init()
}

func (us *UserSpace) init() *UserSpace {
	us.UserPath = filepath.Join(root, us.UName)
	if !fd.DirExists(us.UserPath) {
		gio.MustCreateDir(us.UserPath)
	}
	return us
}

func (us *UserSpace) SaveFile(filename, note string, data []byte, groups ...string) error {

	// /root/name/group0/.../groupX/type/file
	grppath := filepath.Join(groups...)         // /group0/.../groupX/
	path := filepath.Join(us.UserPath, grppath) // /root/name/group0/.../groupX/
	gio.MustCreateDir(path)
	oldpath := filepath.Join(path, filename) // /root/name/group0/.../groupX/file
	err := os.WriteFile(oldpath, data, os.ModePerm)
	if err != nil {
		return err
	}

	fType := fdb.GetFileType(oldpath)
	newpath := filepath.Join(path, fType)      // /root/name/group0/.../groupX/type/
	gio.MustCreateDir(newpath)                 // /root/name/group0/.../groupX/type/
	newpath = filepath.Join(newpath, filename) // /root/name/group0/.../groupX/type/file
	err = os.Rename(oldpath, newpath)
	if err == nil {
		us.Update(&fdb.FileItem{
			Id:        fmt.Sprintf("%x", md5.Sum(data)), // sha1.Sum, sha256.Sum256
			Path:      newpath,
			Tm:        time.Now().String(),
			Status:    status.Received,
			GroupList: strings.Join(groups, fdb.SEP_GRP),
			Note:      note,
			RefBy:     "",
		})
	}
	return err
}

func (us *UserSpace) Update(fi *fdb.FileItem) error {
	db := fdb.GetDB(rootfdb)
	defer db.Close()
	if strings.Contains(fi.Path, "/"+us.UName+"/") {
		return db.UpdateFileItem(fi)
	}
	return fmt.Errorf("%v does NOT belong to %v", *fi, *us)
}
