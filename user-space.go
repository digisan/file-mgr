package filemgr

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/digisan/file-mgr/fdb"
	. "github.com/digisan/go-generics/v2"
	fd "github.com/digisan/gotk/filedir"
	gio "github.com/digisan/gotk/io"
	"github.com/digisan/gotk/strs"
	lk "github.com/digisan/logkit"
)

const (
	PS = string(os.PathSeparator)
)

var (
	rootSP = "data/user-space"
	rootDB = "data/user-fdb"
)

type UserSpace struct {
	UName    string              // user unique name
	UserPath string              // user space path, usually is "root/name/"
	FIs      []*fdb.FileItem     // all fileitems belong to this user
	IDs      map[string]struct{} // fileitem which is group loaded in memory
}

func (us UserSpace) String() string {
	sb := &strings.Builder{}
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("%-13s%s\n", "UName:", us.UName))
	sb.WriteString(fmt.Sprintf("%-13s%s\n", "UserPath:", us.UserPath))
	sb.WriteString("FileItems: [\n")
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

// including 'InitDB'
func InitFileMgr(root string) {
	if root = filepath.Clean(root); len(root) != 0 {
		rootSP = filepath.Join(root, filepath.Base(rootSP))
		rootDB = filepath.Join(root, filepath.Base(rootDB))
	}
	fdb.InitDB(rootDB)
}

func DisposeFileMgr() {
	fdb.CloseDB()
}

func UseUser(name string) (*UserSpace, error) {
	us := &UserSpace{
		UName: name,
		IDs:   make(map[string]struct{}),
	}
	us.init()
	return us.loadFI(true)
}

func (us *UserSpace) init() *UserSpace {
	us.UserPath = filepath.Join(rootSP, us.UName)
	us.UserPath = strings.TrimSuffix(us.UserPath, PS) + PS
	if !fd.DirExists(us.UserPath) {
		gio.MustCreateDir(us.UserPath)
	}
	return us
}

// db
func (us *UserSpace) loadFI(selfcheck bool) (*UserSpace, error) {
	if selfcheck {
		if err := us.SelfCheck(false); err != nil {
			return nil, err
		}
	}
	var err error
	us.FIs, err = fdb.ListFileItems(func(fi *fdb.FileItem) bool {
		return us.Own(fi)
	})
	for _, fi := range us.FIs {
		us.IDs[fi.Id+fi.Path] = struct{}{}
	}
	return us, err
}

func (us *UserSpace) hasMemFI(fi *fdb.FileItem) bool {
	_, ok := us.IDs[fi.Id+fi.Path]
	return ok
}

////////////////////////////////////////////////////////////

// db
func (us *UserSpace) UpdateFileItem(fi *fdb.FileItem, selfcheck bool) error {
	defer func() {
		if selfcheck {
			lk.FailOnErr("%v", us.SelfCheck(false))
		}
	}()
	if us.Own(fi) {
		return fdb.UpdateFileItem(fi)
	}
	return fmt.Errorf("%v does NOT belong to %v", *fi, *us)
}

// return storage path & error
func (us *UserSpace) SaveFile(fname, note string, r io.Reader, groups ...string) (string, error) {

	now := time.Now()

	base, ext := "", ""
	if !strings.Contains(fname, ".") {
		base, ext = fname, ""
	} else {
		ext = strs.SplitPartFromLast(fname, ".", 1)
		base = strs.SplitPartFromLast(fname, ".", 2)
	}
	fname = fmt.Sprintf("%s-%v.%s", base, now.Unix(), ext)
	fname = strings.TrimSuffix(fname, ".")

	// /root/name/group0/.../groupX/type/file
	grppath := filepath.Join(groups...)                                       // /group0/.../groupX/
	path := filepath.Join(us.UserPath, time.Now().Format("2006-01"), grppath) // /root/name/2006-01/group0/.../groupX/
	gio.MustCreateDir(path)                                                   // mkdir /root/name/2006-01/group0/.../groupX/
	oldpath := filepath.Join(path, fname)                                     // /root/name/2006-01/group0/.../groupX/file
	oldFile, err := os.Create(oldpath)
	if err != nil {
		return "", err
	}
	defer oldFile.Close()
	if _, err = io.Copy(oldFile, r); err != nil {
		return "", err
	}

	fType := fdb.GetFileType(oldpath)

	// further process after uploading
	switch fType {
	case "video":
		if p := videoCrop(oldpath, note); len(p) != 0 {
			if err := os.RemoveAll(oldpath); err != nil {
				return "", err
			}
			oldpath = p
		}
	case "image":
	}

	newpath := filepath.Join(path, fType)   // /root/name/2006-01/group0/.../groupX/type/
	gio.MustCreateDir(newpath)              // /root/name/2006-01/group0/.../groupX/type/
	newpath = filepath.Join(newpath, fname) // /root/name/2006-01/group0/.../groupX/type/file

	if err = os.Rename(oldpath, newpath); err == nil {
		data, err := os.ReadFile(newpath)
		if err != nil {
			return "", err
		}
		fi := &fdb.FileItem{
			Id:        strings.ToLower(fmt.Sprintf("%x-%v", md5.Sum(data), now.UnixMilli())), // sha1.Sum, sha256.Sum256
			Path:      newpath,
			Tm:        now,
			GroupList: strings.Join(groups, fdb.SEP_GRP),
			Note:      note,
		}
		if !us.hasMemFI(fi) {
			if err = us.UpdateFileItem(fi, true); err == nil {
				us.FIs = append(us.FIs, fi)
			}
		}
	}
	return newpath, err
}

// 'fh' --- FormFile("param"), return storage path & error
func (us *UserSpace) SaveFormFile(fh *multipart.FileHeader, note string, groups ...string) (string, error) {
	file, err := fh.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()
	return us.SaveFile(fh.Filename, note, file, groups...)
}

func (us *UserSpace) Own(fi *fdb.FileItem) bool {
	return strings.Contains(fi.Path, us.UserPath)
}

func (us *UserSpace) SelfCheck(rmEmptyDir bool) error {
	for i, fi := range us.FIs {
		if !fd.FileExists(fi.Path) {
			return fmt.Errorf("%d - [%s] file does NOT exist on disk", i, fi.Path)
		}
	}
	if rmEmptyDir {
		_, dirs, err := fd.WalkFileDir(us.UserPath, true)
		if err != nil {
			return err
		}
		for _, dir := range dirs {
		NEXT:
			empty, err := fd.IsDirEmpty(dir)
			if err != nil {
				return err
			}
			if empty {
				err = os.RemoveAll(dir)
				if err != nil {
					return err
				}
				dir = filepath.Dir(dir)
				goto NEXT
			}
		}
	}
	return nil
}

func (us *UserSpace) SearchFileItem(ftype string, groups ...string) (fis []*fdb.FileItem) {
	regs := make([]*regexp.Regexp, 0, len(groups))
	for _, grp := range groups {
		ltr := strings.ReplaceAll(grp, `*`, `[\d\w\s]*`)
		ltr = strings.ReplaceAll(ltr, `?`, `[\d\w\s]?`)
		regs = append(regs, regexp.MustCompile(ltr))
	}
NEXT:
	for _, fi := range us.FIs {
		if ftype == fdb.Any || ftype == fi.Type() {
			grouplist := strings.Split(fi.GroupList, fdb.SEP_GRP)
			if len(regs) == len(grouplist) {
				for i, reg := range regs {
					if reg.FindString(grouplist[i]) == "" {
						continue NEXT
					}
				}
				fis = append(fis, fi)
			}
		}
	}
	return
}

// tmYM: such as "2022-04"
func (us *UserSpace) PathContent(tmYM string, grps ...string) (content []string) {
	path := filepath.Join(tmYM, filepath.Join(grps...))
	fullpath := strings.TrimSuffix(filepath.Join(us.UserPath, path), PS) + PS
	for _, fi := range us.FIs {
		if strings.HasPrefix(fi.Path, fullpath) {
			segs := strings.Split(strings.TrimPrefix(fi.Path, fullpath), PS)
			if len(segs) > 0 {
				content = append(content, segs[0])
			}
		}
	}
	return Settify(content...)
}

func (us *UserSpace) FileItems(id string) (fis []*fdb.FileItem, err error) {
	if len(id) < 32 {
		return nil, errors.New("id length MUST greater than 32")
	}
	id = strings.ToLower(id)
	for _, fi := range us.FIs {
		if strings.HasPrefix(fi.Id, id) {
			fis = append(fis, fi)
		}
	}
	return
}

func (us *UserSpace) FirstFileContent(id string) ([]byte, error) {
	fis, err := us.FileItems(id)
	if err != nil {
		return nil, err
	}
	if len(fis) > 0 {
		data, err := os.ReadFile(fis[0].Path)
		lk.WarnOnErr("%v", err)
		return data, nil
	}
	return nil, nil
}

func (us *UserSpace) DelFileItem(id string) error {
	fis, err := us.FileItems(id)
	if err != nil {
		return err
	}
	for _, fi := range fis {
		if _, err := fdb.RemoveFileItems(fi.ID(), true); err != nil {
			lk.WarnOnErr("%v", err)
			return err
		}
		if err := fd.Remove(fi.Path, true); err != nil {
			lk.WarnOnErr("%v", err)
			return err
		}
	}
	return nil
}

func (us *UserSpace) SetFIGroup(fId string, iGrp int, nameGrp string) error {
	for i, fi := range us.FIs {
		if strings.HasPrefix(fi.ID(), fId) {
			_, err := us.FIs[i].SetGroup(iGrp, nameGrp)
			if err != nil {
				return err
			}
			if err = us.UpdateFileItem(us.FIs[i], true); err != nil {
				return err
			}
		}
	}
	return nil
}

func (us *UserSpace) SetFINote(fId, note string) error {
	for i, fi := range us.FIs {
		if strings.HasPrefix(fi.ID(), fId) {
			oriNote := us.FIs[i].Note
			us.FIs[i].SetNote(note)
			if err := us.UpdateFileItem(us.FIs[i], true); err != nil {
				us.FIs[i].SetNote(oriNote)
				return err
			}
		}
	}
	return nil
}
