package ftype

import "github.com/digisan/go-generics/str"

const (
	All        = "all"
	Document   = "document"
	Photo      = "photo"
	Audio      = "audio"
	Video      = "video"
	Archive    = "archive"
	Executable = "executable?"
	Binary     = "binary"
	Unknown    = "unknown"
)

func AllFileType() []string {
	return []string{Document, Photo, Audio, Video, Archive, Executable, Binary, Unknown}
}

func TypeOK(fType string) bool {
	return str.In(fType, AllFileType()...)
}
