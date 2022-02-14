package fdb

import (
	"bytes"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	ft "github.com/digisan/file-mgr/fdb/ftype"
	lk "github.com/digisan/logkit"
)

type FileItem struct {
	Path        string // id
	Tm          string // timestamp
	Status      string // "received", "applying", "approved", etc
	GroupList   string // "group1^^group2^^...^^groupN"
	Description string // "description..."
	RefBy       string // *****^^*****^^...
}

const (
	SEP     = "||"
	SEP_GRP = "^^"
	SEP_REF = "^^"
)

// db key order
const (
	MOK_Status int = iota
	MOK_Path
	MOK_Tm
	MOK_GroupList
	MOK_Description
	MOK_END
)

func (fi *FileItem) KeyFieldAddr(mok int) *string {
	mFldAddr := map[int]*string{
		MOK_Status:      &fi.Status,
		MOK_Path:        &fi.Path,
		MOK_Tm:          &fi.Tm,
		MOK_GroupList:   &fi.GroupList,
		MOK_Description: &fi.Description,
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
			fld := typ.Field(i)
			val := val.Field(i)
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

func (fi *FileItem) Type() string {
	dir := filepath.Dir(fi.Path)
	typedir := filepath.Base(dir)
	lk.FailOnErrWhen(!ft.TypeOK(typedir), "%v", fmt.Errorf("file type is invalid"))
	return typedir
}

func (fi *FileItem) Name() string {
	return filepath.Base(fi.Path)
}
