package filemgr

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	fd "github.com/digisan/gotk/file-dir"
	"github.com/jtguibas/cinema"
)

// note must be 'crop:x,y,w,h'
func imageCrop(fPath, note, outFmt string) (fCrop string, err error) {
	x, y, w, h := 0, 0, 0, 0
	if n, err := fmt.Sscanf(note, "crop:%d,%d,%d,%d", &x, &y, &w, &h); err == nil && n == 4 {
		img, err := loadImage(fPath)
		if err != nil {
			return "", err
		}

		roi := roi4rgba(img, x, y, x+w, y+h)
		fCrop = fd.ChangeFileName(fPath, "", "-crop")
		fCrop = strings.TrimSuffix(fCrop, filepath.Ext(fCrop))

		switch outFmt {
		case ".png", "png":
			fCrop += ".png"
			if _, err := savePNG(roi, fCrop); err != nil {
				return "", err
			}
		case ".jpg", "jpg":
			fCrop += ".jpg"
			if _, err := saveJPG(roi, fCrop); err != nil {
				return "", err
			}
		default:
			fCrop += ".png"
			if _, err := savePNG(roi, fCrop); err != nil {
				return "", err
			}
		}
		return fCrop, nil
	}
	return "", errors.New("note must be 'crop:x,y,w,h' to crop image")
}

/////////////////////////////////////////////////////////////////////////////////

// sudo apt install ffmpeg
// https://pkg.go.dev/github.com/jtguibas/cinema#section-readme

// note must be 'crop:x,y,w,h'
func videoCrop(fPath, note string) (fCrop string, err error) {
	x, y, w, h := 0, 0, 0, 0
	if n, err := fmt.Sscanf(note, "crop:%d,%d,%d,%d", &x, &y, &w, &h); err == nil && n == 4 {
		fCrop = fd.ChangeFileName(fPath, "", "-crop")
		fCrop = strings.TrimSuffix(fCrop, filepath.Ext(fCrop)) + ".mp4"
		video, err := cinema.Load(fPath)
		if err != nil {
			return "", err
		}
		video.Crop(x, y, w, h)
		if err := video.Render(fCrop); err != nil {
			return "", err
		}
		return fCrop, nil
	}
	return "", errors.New("note must be 'crop:x,y,w,h' to crop video")
}
