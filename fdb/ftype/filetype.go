package ftype

import "github.com/digisan/go-generics/str"

const (	
	Text       = "text"
	Image      = "image"
	Audio      = "audio"
	Video      = "video"
	Executable = "executable?"
	Binary     = "binary"
	Unknown    = "unknown"
)

func AllFileType() []string {
	return []string{Unknown, Text, Image, Audio, Video, Executable, Binary}
}

func TypeOK(fType string) bool {
	return str.In(fType, AllFileType()...)
}
