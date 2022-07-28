package filemgr

import (
	"fmt"
	"path/filepath"

	fd "github.com/digisan/gotk/filedir"
	"github.com/jtguibas/cinema"
)

// https://pkg.go.dev/github.com/jtguibas/cinema#section-readme

func videoCrop(fpath, note string) (fcrop string) {
	x, y, w, h := 0, 0, 0, 0
	if n, err := fmt.Sscanf(note, "crop:%d-%d-%d-%d", &x, &y, &w, &h); err == nil && n == 4 {
		fcrop = fd.ChangeFileName(fpath, "", "-crop")
		if ext := filepath.Ext(fcrop); len(ext) == 0 {
			fcrop += ".mp4"
		}
		video, err := cinema.Load(fpath)
		if err != nil {
			return ""
		}
		video.Crop(x, y, w, h)
		if err := video.Render(fcrop); err != nil {
			return ""
		}
	}
	return fcrop
}
