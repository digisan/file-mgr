package fdb

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	fd "github.com/digisan/gotk/filedir"
	gio "github.com/digisan/gotk/io"
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

// db key order
const (
	MOK_Id int = iota
	MOK_END
)

func (fi *FileItem) KeyFieldAddr(mok int) any {
	mFldAddr := map[int]any{
		MOK_Id: &fi.Id,
	}
	return mFldAddr[mok]
}

// db value order
const (
	MOV_Path int = iota
	MOV_Tm
	MOV_GroupList
	MOV_Note
	MOV_END
)

func (fi *FileItem) ValFieldAddr(mov int) any {
	mFldAddr := map[int]any{
		MOV_Path:      &fi.Path,
		MOV_Tm:        &fi.Tm,
		MOV_GroupList: &fi.GroupList,
		MOV_Note:      &fi.Note,
	}
	return mFldAddr[mov]
}

///////////////////////////////////////////////////

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

func (fi *FileItem) Marshal() (forKey, forValue []byte) {
	params := []struct {
		end       int
		fnFldAddr func(int) any
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
			switch v := param.fnFldAddr(i).(type) {
			case *string:
				sb.WriteString(*v)
			case *time.Time:
				encoding, err := (*v).MarshalBinary()
				lk.FailOnErr("%v", err)
				sb.Write(encoding)
			default:
				panic("Marshal Error Type")
			}

		}
		*param.out = []byte(sb.String())
	}
	return
}

func (fi *FileItem) Unmarshal(dbKey, dbVal []byte) {
	params := []struct {
		in        []byte
		fnFldAddr func(int) any
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
	for idx, param := range params {
		for i, seg := range bytes.Split(param.in, []byte(SEP)) {
			if (idx == 0 && i == MOK_END) || (idx == 1 && i == MOV_END) {
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
}

///////////////////////////////////////////////////

func (fi *FileItem) ID() string {
	return fi.Id
}

func (fi *FileItem) Type() string {
	dir := filepath.Dir(fi.Path)
	typedir := filepath.Base(dir)
	lk.FailOnErrWhen(!TypeOK(typedir), "%v", fmt.Errorf("file type is unregistered"))
	return typedir
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
	gio.MustCreateDir(filepath.Dir(fi.Path))
	return fi.Path, os.Rename(fi.prevPath, fi.Path)
}
