package ftype

import "github.com/digisan/go-generics/str"

const (
	All        = "all"
	Text       = "text"
	Image      = "image"
	Audio      = "audio"
	Video      = "video"
	Executable = "executable?"
	Binary     = "binary"
	Unknown    = "unknown"
)

func AllFileType() []string {
	return []string{Text, Image, Audio, Video, Executable, Binary, Unknown}
}

func TypeOK(fType string) bool {
	return str.In(fType, AllFileType()...)
}
