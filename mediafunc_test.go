package filemgr

import (
	"fmt"
	"testing"
)

func TestImageSize(t *testing.T) {
	fmt.Println(GetImageSize("./samples/moon"))
	fmt.Println(GetImageSize("./samples/jpg.jpeg"))
}

func TestVideoSize(t *testing.T) {
	fmt.Println(GetVideoSize("./samples/Screencast"))
	fmt.Println(GetVideoSize("./samples/Screencast1.mp4"))
}
