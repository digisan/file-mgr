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

var (
	rootDb = "data/fdb"
	rootSp = "data/user-space"
)

type UserSpace struct {
	UName    string
	UserPath string
	FIs      []*fdb.FileItem
}

func (us UserSpace) String() string {
	sb := &strings.Builder{}
	sb.WriteString(fmt.Sprintf("%-13s%s\n", "UName:", us.UName))
	sb.WriteString(fmt.Sprintf("%-13s%s\n", "UserPath:", us.UserPath))
	sb.WriteString("[\n")
	for i, fi := range us.FIs {
		if i == len(us.FIs)-1 {
			sb.WriteString(fmt.Sprint(fi))
			break
		}
		sb.WriteString(fmt.Sprintln(fi))
	}
	sb.WriteString("]\n")
	return sb.String()
}

func SetRoot(rtSp, rtDb string) {
	if rtSp != "" {
		rootSp = filepath.Clean(rtSp)
	}
	if rtDb != "" {
		rootDb = filepath.Clean(rtDb)
	}
}

func UseUser(name string) (*UserSpace, error) {
	us := &UserSpace{
		UName: name,
	}
	us.init()
	err := us.loadFI()
	return us, err
}

func (us *UserSpace) init() *UserSpace {
	us.UserPath = filepath.Join(rootSp, us.UName)
	us.UserPath = strings.TrimSuffix(us.UserPath, "/") + "/"
	if !fd.DirExists(us.UserPath) {
		gio.MustCreateDir(us.UserPath)
	}
	return us
}

// db
func (us *UserSpace) loadFI() (err error) {
	db := fdb.GetDB(rootDb)
	defer db.Close()
	us.FIs, err = db.ListFileItems(func(fi *fdb.FileItem) bool {
		return us.Has(fi)
	})
	return err
}

////////////////////////////////////////////////////////////

// db
func (us *UserSpace) Update(fi *fdb.FileItem) error {
	db := fdb.GetDB(rootDb)
	defer db.Close()

	if us.Has(fi) {
		return db.UpdateFileItem(fi)
	}
	return fmt.Errorf("%v does NOT belong to %v", *fi, *us)
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
		fi := &fdb.FileItem{
			Id:        fmt.Sprintf("%x", md5.Sum(data)), // sha1.Sum, sha256.Sum256
			Path:      newpath,
			Tm:        time.Now().String(),
			Status:    status.Received,
			GroupList: strings.Join(groups, fdb.SEP_GRP),
			Note:      note,
			RefBy:     "",
		}
		if err = us.Update(fi); err == nil {
			us.FIs = append(us.FIs, fi)
		}
	}
	return err
}

func (us *UserSpace) Has(fi *fdb.FileItem) bool {
	return strings.Contains(fi.Path, us.UserPath)
}

func (us *UserSpace) SelfCheck(rmEmptyDir bool) error {
	for _, fi := range us.FIs {
		if !fd.FileExists(fi.Path) {
			return fmt.Errorf("[%s] file does NOT exist on disk", fi.Path)
		}
	}
	if rmEmptyDir {
		_, dirs, err := fd.WalkFileDir(us.UserPath, true)
		if err != nil {
			return err
		}
		for _, dir := range dirs {
			empty, err := fd.DirIsEmpty(dir)
			if err != nil {
				return err
			}
			if empty {
				return os.RemoveAll(dir)
			}
		}
	}
	return nil
}

func (us *UserSpace) SearchFileItem(ftype string, groups ...string) (fis []*fdb.FileItem) {
	for _, fi := range us.FIs {
		switch {
		case ftype != "" && len(groups) > 0:
			if fi.Type() == ftype && fi.GroupList == strings.Join(groups, fdb.SEP_GRP) {
				fis = append(fis, fi)
			}
		case ftype == "" && len(groups) > 0:
			if fi.GroupList == strings.Join(groups, fdb.SEP_GRP) {
				fis = append(fis, fi)
			}
		case ftype != "" && len(groups) == 0:
			if fi.Type() == ftype {
				fis = append(fis, fi)
			}
		default:
			fis = us.FIs
		}
	}
	return
}
