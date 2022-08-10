package filemgr

import (
	"errors"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	fd "github.com/digisan/gotk/filedir"
	"github.com/jtguibas/cinema"
)

// https://pkg.go.dev/github.com/jtguibas/cinema#section-readme

func videoCrop(fpath, note string) (fcrop string, err error) {
	x, y, w, h := 0, 0, 0, 0
	if n, err := fmt.Sscanf(note, "crop:%d-%d-%d-%d", &x, &y, &w, &h); err == nil && n == 4 {
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
	return "", errors.New("note must be 'crop:x-y-w-h' to crop video")
}

/////////////////////////////////////////////////////////////////////////////////

func loadImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func roi4gray(img image.Image, left, top, right, bottom int) *image.Gray {
	rect := image.Rect(0, 0, right-left, bottom-top)
	gray := image.NewGray(rect)
	draw.Draw(gray, rect, img, image.Point{left, top}, draw.Src)
	return gray
}

func roi4rgba(img image.Image, left, top, right, bottom int) *image.RGBA {
	rect := image.Rect(0, 0, right-left, bottom-top)
	rgba := image.NewRGBA(rect)
	draw.Draw(rgba, rect, img, image.Point{left, top}, draw.Src)
	return rgba
}

// func saveJPG(img image.Image, path string) (image.Image, error) {
// 	out, err := os.Create(path)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer out.Close()

// 	var opts jpeg.Options
// 	opts.Quality = 100
// 	if err := jpeg.Encode(out, img, &opts); err != nil {
// 		return nil, err
// 	}
// 	return img, nil
// }

func savePNG(img image.Image, path string) (image.Image, error) {
	out, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	defer out.Close()

	if err := png.Encode(out, img); err != nil {
		return nil, err
	}
	return img, nil
}

func imageCrop(fpath, note string) (fcrop string, err error) {
	x, y, w, h := 0, 0, 0, 0
	if n, err := fmt.Sscanf(note, "crop:%d-%d-%d-%d", &x, &y, &w, &h); err == nil && n == 4 {
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
	return "", errors.New("note must be 'crop:x-y-w-h' to crop image")
}
