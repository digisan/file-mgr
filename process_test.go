package filemgr

import "testing"

func TestVideoCrop(t *testing.T) {
	videoCrop("./samples/Screencast", "crop:1-2-500-400")
}
