package fdb

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	ft "github.com/digisan/file-mgr/fdb/ftype"
	"github.com/digisan/file-mgr/fdb/status"
	"github.com/digisan/go-generics/str"
	fd "github.com/digisan/gotk/filedir"
	gio "github.com/digisan/gotk/io"
	lk "github.com/digisan/logkit"
)

type FileItem struct {
	Id        string // id
	Path      string // file path
	Tm        string // timestamp
	Status    string // "received", "applying", "approved", etc
	GroupList string // "group1^group2^...^groupN", [once changed, => change Path, => move file]
	Note      string // "note..."
	RefBy     string // refcode1^refcode2^...
}

const (
	SEP     = "||"
	SEP_GRP = "^"
	SEP_REF = "^"
)

// db key order
const (
	MOK_Status int = iota
	MOK_Id
	MOK_Path
	MOK_Tm
	MOK_GroupList
	MOK_Note
	MOK_END
)

func (fi *FileItem) KeyFieldAddr(mok int) *string {
	mFldAddr := map[int]*string{
		MOK_Status:    &fi.Status,
		MOK_Id:        &fi.Id,
		MOK_Path:      &fi.Path,
		MOK_Tm:        &fi.Tm,
		MOK_GroupList: &fi.GroupList,
		MOK_Note:      &fi.Note,
	}
	return mFldAddr[mok]
}

// db value order
const (
	MOV_RefBy int = iota
	MOV_END
)

func (fi *FileItem) ValFieldAddr(mov int) *string {
	mFldAddr := map[int]*string{
		MOV_RefBy: &fi.RefBy,
	}
	return mFldAddr[mov]
}

///////////////////////////////////////////////////

func (fi FileItem) String() string {
	if fi.Path != "" {
		sb := strings.Builder{}
		typ := reflect.TypeOf(fi)
		val := reflect.ValueOf(fi)
		for i := 0; i < typ.NumField(); i++ {
			fld, val := typ.Field(i), val.Field(i)
			sb.WriteString(fmt.Sprintf("%-12s %v\n", fld.Name+":", val.String()))
		}
		return sb.String()
	}
	return "[Empty FileItem]\n"
}

func (fi *FileItem) Marshal() (forKey, forValue []byte) {
	params := []struct {
		end       int
		fnFldAddr func(int) *string
		out       *[]byte
	}{
		{
			end:       MOK_END,
			fnFldAddr: fi.KeyFieldAddr,
			out:       &forKey,
		},
		{
			end:       MOV_END,
			fnFldAddr: fi.ValFieldAddr,
			out:       &forValue,
		},
	}
	for _, param := range params {
		sb := &strings.Builder{}
		for i := 0; i < param.end; i++ {
			if i > 0 {
				sb.WriteString(SEP)
			}
			sb.WriteString(*param.fnFldAddr(i))
		}
		*param.out = []byte(sb.String())
	}
	return
}

func (fi *FileItem) Unmarshal(dbKey, dbVal []byte) {
	params := []struct {
		in        []byte
		fnFldAddr func(int) *string
	}{
		{
			in:        dbKey,
			fnFldAddr: fi.KeyFieldAddr,
		},
		{
			in:        dbVal,
			fnFldAddr: fi.ValFieldAddr,
		},
	}
	for _, param := range params {
		for i, seg := range bytes.Split(param.in, []byte(SEP)) {
			*param.fnFldAddr(i) = string(seg)
		}
	}
}

///////////////////////////////////////////////////

func (fi *FileItem) ID() string {
	return fi.Id
}

func (fi *FileItem) Type() string {
	dir := filepath.Dir(fi.Path)
	typedir := filepath.Base(dir)
	lk.FailOnErrWhen(!ft.TypeOK(typedir), "%v", fmt.Errorf("file type is unregistered"))
	return typedir
}

func (fi *FileItem) Name() string {
	return filepath.Base(fi.Path)
}

func (fi *FileItem) SetStatus(stat string) error {
	if str.NotIn(stat, status.AllStatus()...) {
		return fmt.Errorf("status [%v] is unregistered", stat)
	}
	fi.Status = stat
	return nil
}

func (fi *FileItem) SetGroup(idx int, grp string) error {
	oldGrpPath := strings.ReplaceAll(fi.GroupList, SEP_GRP, "/")
	oldPath := fi.Path

	if !fd.FileExists(oldPath) {
		return fmt.Errorf("[%s] file is NOT existing", oldPath)
	}

	grps := strings.Split(fi.GroupList, SEP_GRP)
	switch {
	case idx < len(grps):
		grps[idx] = grp
	case idx >= len(grps):
		grps = append(grps, grp)
	}
	fi.GroupList = strings.Join(grps, SEP_GRP)
	fi.GroupList = strings.TrimPrefix(fi.GroupList, SEP_GRP)
	fi.GroupList = strings.TrimSuffix(fi.GroupList, SEP_GRP)

	// [once changed, => change Path, => move file]
	if oldGrpPath != "" {
		newGrpPath := strings.ReplaceAll(fi.GroupList, SEP_GRP, "/")
		fi.Path = strings.ReplaceAll(fi.Path, oldGrpPath, newGrpPath) // Path update
	} else {
		file := filepath.Base(fi.Path)
		dir := filepath.Dir(fi.Path)
		head := filepath.Dir(dir)                         // user-space/name
		tail := filepath.Join(filepath.Base(dir), file)   // text/sample.txt
		fi.Path = filepath.Join(head, fi.GroupList, tail) // user-space/name/groupX/text/sample.txt
	}
	gio.MustCreateDir(filepath.Dir(fi.Path))
	return os.Rename(oldPath, fi.Path)
}

func (fi *FileItem) AddRefBy(refCodes ...string) {
	for _, refCode := range refCodes {
		if fi.RefBy == "" {
			fi.RefBy = refCode
		} else {
			if strings.Contains(fi.RefBy, SEP_REF+refCode+SEP_REF) ||
				strings.HasPrefix(fi.RefBy, refCode+SEP_REF) ||
				strings.HasSuffix(fi.RefBy, SEP_REF+refCode) {
				continue
			}
			fi.RefBy += SEP_REF + refCode
		}
	}
}

func (fi *FileItem) RmRefBy(refCodes ...string) {
	for _, refCode := range refCodes {
		fi.RefBy = strings.ReplaceAll(fi.RefBy, SEP_REF+refCode+SEP_REF, SEP_REF)
		fi.RefBy = strings.TrimPrefix(fi.RefBy, refCode+SEP_REF)
		fi.RefBy = strings.TrimSuffix(fi.RefBy, SEP_REF+refCode)
	}
}
