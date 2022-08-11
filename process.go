package filemgr

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	fd "github.com/digisan/gotk/filedir"
	"github.com/jtguibas/cinema"
)

// note must be 'crop:x,y,w,h'
func imageCrop(fpath, note string) (fcrop string, err error) {
	x, y, w, h := 0, 0, 0, 0
	if n, err := fmt.Sscanf(note, "crop:%d,%d,%d,%d", &x, &y, &w, &h); err == nil && n == 4 {
		img, err := loadImage(fpath)
		if err != nil {
			return "", err
		}
		roi := roi4rgba(img, x, y, x+w, y+h)
		fcrop = fd.ChangeFileName(fpath, "", "-crop")
		fcrop = strings.TrimSuffix(fcrop, filepath.Ext(fcrop)) + ".png"
		if _, err := savePNG(roi, fcrop); err != nil {
			return "", err
		}
		return fcrop, nil
	}
	return "", errors.New("note must be 'crop:x,y,w,h' to crop image")
}

/////////////////////////////////////////////////////////////////////////////////

// sudo apt install ffmpeg
// https://pkg.go.dev/github.com/jtguibas/cinema#section-readme

// note must be 'crop:x,y,w,h'
func videoCrop(fpath, note string) (fcrop string, err error) {
	x, y, w, h := 0, 0, 0, 0
	if n, err := fmt.Sscanf(note, "crop:%d,%d,%d,%d", &x, &y, &w, &h); err == nil && n == 4 {
		fcrop = fd.ChangeFileName(fpath, "", "-crop")
		fcrop = strings.TrimSuffix(fcrop, filepath.Ext(fcrop)) + ".mp4"
		video, err := cinema.Load(fpath)
		if err != nil {
			return "", err
		}
		video.Crop(x, y, w, h)
		if err := video.Render(fcrop); err != nil {
			return "", err
		}
		return fcrop, nil
	}
	return "", errors.New("note must be 'crop:x,y,w,h' to crop video")
}
