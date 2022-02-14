package fdb

import (
	"net/http"
	"os"
	"path/filepath"

	ft "github.com/digisan/file-mgr/fdb/ftype"
	fd "github.com/digisan/gotk/filedir"
	lk "github.com/digisan/logkit"
)

func fileContentType(f *os.File) (string, error) {
	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)
	_, err := f.Read(buffer)
	if err != nil {
		return "", err
	}
	// Use the net/http package's handy DectectContentType function. Always returns a valid
	// content-type by returning "application/octet-stream" if no others seemed to match.
	contentType := http.DetectContentType(buffer)
	return contentType, nil
}

var (
	mContType = map[string]string{
		"text/plain; charset=utf-8": ft.Text,
		"application/pdf":           ft.Text,
		"application/octet-stream":  ft.Binary,
		"application/x-gzip":        ft.Binary,
	}
	mBinType = map[string]string{
		"":      ft.Executable,
		".rmvb": ft.Video,
		".exe":  ft.Executable,
		".md":   ft.Text,
		".mod":  ft.Text,
		".sum":  ft.Text,
		".gz":   ft.Binary,
	}
)

func FileType(f *os.File, filename string) string {
	// Get the content
	contentType, err := fileContentType(f)
	lk.FailOnErr("%v", err)
	if t, ok := mContType[contentType]; ok {
		if t == ft.Binary {
			ext := filepath.Ext(filename)
			if t, ok := mBinType[ext]; ok {
				return t
			}
			lk.Log("New Binary Type@ %v", ext)
		}
		return t
	}
	lk.Warn("New Type@ [%v], must be added to 'filetype.go'", contentType)
	return ft.Unknown
}

func GetFileType(path string) string {
	if !fd.FileExists(path) {
		return ""
	}
	// Open File
	f, err := os.Open(path)
	lk.FailOnErr("%v", err)
	defer f.Close()
	return FileType(f, filepath.Base(path))
}
