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
	// Use the net/http package's handy Detect ContentType function. Always returns a valid
	// content-type by returning "application/octet-stream" if no others seemed to match.
	contentType := http.DetectContentType(buffer)
	return contentType, nil
}

// adding more...
var (
	mContType = map[string]string{
		"text/plain; charset=utf-8": ft.Document,
		"application/pdf":           ft.Document,
		"application/octet-stream":  ft.Binary,
		"application/x-gzip":        ft.Archive,
	}
	mBinType = map[string]string{
		"":       ft.Executable,
		".exe":   ft.Executable,
		".rmvb":  ft.Video,
		".mp4":   ft.Video,
		".ogg":   ft.Video,
		".ogv":   ft.Video,
		".webm":  ft.Video,
		".md":    ft.Document,
		".mod":   ft.Document,
		".sum":   ft.Document,
		".doc":   ft.Document,
		".docx":  ft.Document,
		".json":  ft.Document,
		".xml":   ft.Document,
		".yaml":  ft.Document,
		".toml":  ft.Document,
		".html":  ft.Document,
		".xhtml": ft.Document,
		".js":    ft.Document,
		".ejs":   ft.Document,
		".txt":   ft.Document,
		".gz":    ft.Archive,
		".png":   ft.Photo,
		".jpg":   ft.Photo,
		".jpeg":  ft.Photo,
		".gif":   ft.Photo,
		".psd":   ft.Photo,
	}
)

func FileType(f *os.File, fname string) string {
	// Get the content
	contentType, err := fileContentType(f)
	lk.FailOnErr("%v", err)
	if t, ok := mContType[contentType]; ok {
		if t == ft.Binary {
			ext := filepath.Ext(fname)
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
