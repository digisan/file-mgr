package fdb

import (
	"os"
	"path/filepath"

	. "github.com/digisan/go-generics/v2"
	fd "github.com/digisan/gotk/file-dir"
	lk "github.com/digisan/logkit"
	"github.com/h2non/filetype"
)

const (
	Document   = "document"
	Image      = "image"
	Audio      = "audio"
	Video      = "video"
	Archive    = "archive"
	Executable = "executable?"
	Binary     = "binary"
	Unknown    = "unknown"
	Any        = "any"
)

func FileTypes() []string {
	return []string{Document, Image, Audio, Video, Archive, Executable, Binary, Unknown}
}

func TypeOK(fType string) bool {
	return In(fType, FileTypes()...)
}

func FileType(f *os.File) string {
	head := make([]byte, 261)
	f.Read(head)
	switch {
	case filetype.IsImage(head):
		return Image
	case filetype.IsVideo(head):
		return Video
	case filetype.IsAudio(head):
		return Audio
	case filetype.IsDocument(head):
		return Document
	case filetype.IsArchive(head):
		return Archive
	case filetype.IsApplication(head):
		return Executable
	default:
		return Unknown
	}
}

func GetFileType(fPath string) string {
	if !fd.FileExists(fPath) {
		return ""
	}
	f, err := os.Open(fPath)
	lk.FailOnErr("%v", err)
	defer f.Close()
	fType := FileType(f)
	if fType == Unknown {
		switch filepath.Ext(fPath) {
		case ".txt": // add more if needed
			fType = Document
		case ".exe":
			fType = Executable
		}
	}
	return fType
}
