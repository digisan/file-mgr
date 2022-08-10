package filemgr

import (
	"fmt"
	"testing"
)

func TestVideoCrop(t *testing.T) {
	fmt.Println(videoCrop("./samples/Screencast", "crop:100-200-500-400"))
	fmt.Println(videoCrop("./samples/Screencast1.mp4", "crop:100-200-500-400"))
}

func TestImageCrop(t *testing.T) {
	fmt.Println(imageCrop("./samples/moon", "crop:100-200-500-400"))
}
