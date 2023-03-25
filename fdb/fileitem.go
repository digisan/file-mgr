package fdb

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	badger "github.com/dgraph-io/badger/v3"
	bh "github.com/digisan/db-helper/badger"
	fd "github.com/digisan/gotk/file-dir"
	lk "github.com/digisan/logkit"
)

const (
	PS      = string(os.PathSeparator)
	SEP     = "^^"
	SEP_GRP = "^"
)

type FileItem struct {
	// key
	Id string `json:"id"` // id, for linkage
	// value
	Path      string `json:"path"` // file path on real local disk
	prevPath  string
	Tm        time.Time `json:"time"`   // timestamp
	GroupList string    `json:"groups"` // "group1^group2^...^groupN", [once changed, => change Path, => move file]
	Note      string    `json:"note"`   // "note..."
}

func (fi FileItem) String() string {
	if fi.Path != "" {
		sb := strings.Builder{}
		typ := reflect.TypeOf(fi)
		val := reflect.ValueOf(fi)
		sb.WriteString("{\n")
		for i := 0; i < typ.NumField(); i++ {
			fld, val := typ.Field(i), val.Field(i)
			sb.WriteString(fmt.Sprintf("\t%-12s %v\n", fld.Name+":", val))
		}
		sb.WriteString("}\n")
		return sb.String()
	}
	return "[Empty FileItem]\n"
}

// db key order
const (
	KO_Id int = iota
	KO_END
)

func (fi *FileItem) KeyFieldAddr(mok int) any {
	mFldAddr := map[int]any{
		KO_Id: &fi.Id,
	}
	return mFldAddr[mok]
}

// db value order
const (
	VO_Path int = iota
	VO_Tm
	VO_GroupList
	VO_Note
	VO_END
)

func (fi *FileItem) ValFieldAddr(mov int) any {
	mFldAddr := map[int]any{
		VO_Path:      &fi.Path,
		VO_Tm:        &fi.Tm,
		VO_GroupList: &fi.GroupList,
		VO_Note:      &fi.Note,
	}
	return mFldAddr[mov]
}

///////////////////////////////////////////////////

func (fi *FileItem) BadgerDB() *badger.DB {
	return DbGrp.File
}

func (fi *FileItem) Key() []byte {
	var (
		sb = &strings.Builder{}
	)
	for i := 0; i < KO_END; i++ {
		if i > 0 {
			sb.WriteString(SEP)
		}
		switch v := fi.KeyFieldAddr(i).(type) {
		case *string:
			sb.WriteString(*v)
		default:
			panic("need more type for marshaling key")
		}
	}
	return []byte(sb.String())
}

func (fi *FileItem) Value() []byte {
	var (
		sb = &strings.Builder{}
	)
	for i := 0; i < VO_END; i++ {
		if i > 0 {
			sb.WriteString(SEP)
		}
		switch v := fi.ValFieldAddr(i).(type) {
		case *string:
			sb.WriteString(*v)
		case *time.Time:
			encoding, err := (*v).MarshalBinary()
			lk.FailOnErr("%v", err)
			sb.Write(encoding)
		default:
			panic("need more type for marshaling value")
		}
	}
	return []byte(sb.String())
}

func (fi *FileItem) Marshal(at any) (forKey, forValue []byte) {
	return fi.Key(), fi.Value()
}

func (fi *FileItem) Unmarshal(dbKey, dbVal []byte) (any, error) {
	params := []struct {
		in        []byte
		fnFldAddr func(int) any
	}{
		{dbKey, fi.KeyFieldAddr},
		{dbVal, fi.ValFieldAddr},
	}
	for idx, param := range params {
		for i, seg := range bytes.Split(param.in, []byte(SEP)) {
			if (idx == 0 && i == KO_END) || (idx == 1 && i == VO_END) {
				break
			}
			switch v := param.fnFldAddr(i).(type) {
			case *string:
				*v = string(seg)
			case *time.Time:
				t := &time.Time{}
				lk.FailOnErr("%v @ %v", t.UnmarshalBinary(seg), seg)
				*v = *t
			default:
				panic("Unmarshal Error Type")
			}
		}
	}
	return fi, nil
}

///////////////////////////////////////////////////

func (fi *FileItem) ID() string {
	return fi.Id
}

func (fi *FileItem) Type() string {
	dir := filepath.Dir(fi.Path)
	typeDir := filepath.Base(dir)
	if !fd.IsSupportedFileType(typeDir) {
		return "file type is unregistered"
	}
	return typeDir
}

func (fi *FileItem) Name() string {
	return filepath.Base(fi.Path)
}

// type value as `<video><source src="movie.mp4" type="video/mp4"> ...`
func (fi *FileItem) MediaType() string {
	ext := strings.TrimSuffix(filepath.Ext(fi.Path), ".")
	switch fi.Type() {
	case "photo":
		return "image/" + ext // apng gif ico cur jpg jpeg jfif pjpeg pjp png svg
	case "audio":
		return "audio/" + ext // mid midi rm ram wma aac wav ogg mp3 mp4
	case "video":
		return "video/" + ext // mpg mpeg avi wmv mov rm ram swf flv ogg webm mp4
	default:
		return ""
	}
}

// Need updating DB immediately
func (fi *FileItem) SetNote(note string) {
	fi.Note = note
}

// Need updating DB immediately
func (fi *FileItem) SetGroup(grpIdx int, grpName string) (string, error) {
	oldGrpPath := strings.ReplaceAll(fi.GroupList, SEP_GRP, PS)
	fi.prevPath = fi.Path

	if !fd.FileExists(fi.prevPath) {
		return "", fmt.Errorf("[%s] file is NOT existing", fi.prevPath)
	}

	grps := strings.Split(fi.GroupList, SEP_GRP)
	switch {
	case grpIdx < len(grps):
		grps[grpIdx] = grpName
	case grpIdx >= len(grps):
		grps = append(grps, grpName)
	}
	fi.GroupList = strings.Join(grps, SEP_GRP)
	fi.GroupList = strings.TrimPrefix(fi.GroupList, SEP_GRP)
	fi.GroupList = strings.TrimSuffix(fi.GroupList, SEP_GRP) // GroupList Update

	// [once changed, => change Path, => move file]
	if oldGrpPath != "" {
		newGrpPath := strings.ReplaceAll(fi.GroupList, SEP_GRP, PS)
		fi.Path = strings.ReplaceAll(fi.Path, oldGrpPath, newGrpPath) // Path Update
	} else {
		file, dir := filepath.Base(fi.Path), filepath.Dir(fi.Path) // sample.txt & user-space/name/text
		head := filepath.Dir(dir)                                  // user-space/name
		tail := filepath.Join(filepath.Base(dir), file)            // text/sample.txt
		fi.Path = filepath.Join(head, fi.GroupList, tail)          // user-space/name/groupX.../text/sample.txt , Path Update
	}
	fd.MustCreateDir(filepath.Dir(fi.Path))
	return fi.Path, os.Rename(fi.prevPath, fi.Path)
}

///////////////////////////////////////////////////

// [id] is prefix, could remove many fi
func RemoveFileItems(id string, lock bool) (int, error) {
	if lock {
		DbGrp.Lock()
		defer DbGrp.Unlock()
	}

	if len(id) < 32 {
		return 0, errors.New("id length MUST greater than 32")
	}
	return bh.DeleteObjects[FileItem]([]byte(strings.ToLower(id)))
}

// exactly update ONE fi
func UpdateFileItem(fi *FileItem) error {
	DbGrp.Lock()
	defer DbGrp.Unlock()

	if fi.prevPath == "" {
		fi.prevPath = fi.Path
		defer func() { fi.prevPath = "" }()
	}

	// exactly remove ONE fi
	if _, err := RemoveFileItems(fi.Id, false); err != nil {
		return err
	}
	return bh.UpsertOneObject(fi)
}

func FirstFileItem(id string) (*FileItem, bool, error) {
	DbGrp.Lock()
	defer DbGrp.Unlock()

	fi, err := bh.GetFirstObject[FileItem]([]byte(strings.ToLower(id)), nil)
	if err != nil {
		return nil, false, err
	}
	if fi == nil {
		return fi, false, nil
	}
	return fi, fi.Path != "", nil
}

func ListFileItems(filter func(*FileItem) bool) ([]*FileItem, error) {
	DbGrp.Lock()
	defer DbGrp.Unlock()

	return bh.GetObjects([]byte(""), filter)
}

func IsExisting(id string) bool {
	fi, ok, err := FirstFileItem(strings.ToLower(id))
	return err == nil && ok && fi != nil
}

func SearchFileItems(fType string, groups ...string) (fis []*FileItem, err error) {
	if fType != "" && !fd.IsSupportedFileType(fType) {
		return nil, fmt.Errorf("file type [%s] is unregistered", fType)
	}
	return ListFileItems(func(fi *FileItem) bool {
		if fType != "" {
			return fi.Type() == fType && strings.HasPrefix(fi.GroupList, strings.Join(groups, SEP_GRP))
		}
		if fType == "" && len(groups) > 0 {
			return strings.HasPrefix(fi.GroupList, strings.Join(groups, SEP_GRP))
		}
		return true
	})
}
