package filemgr

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"

	"github.com/jtguibas/cinema"
)

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

// func roi4gray(img image.Image, left, top, right, bottom int) *image.Gray {
// 	rect := image.Rect(0, 0, right-left, bottom-top)
// 	gray := image.NewGray(rect)
// 	draw.Draw(gray, rect, img, image.Point{left, top}, draw.Src)
// 	return gray
// }

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

// return "width,height"
func GetImageSize(fpath string) (string, error) {
	img, err := loadImage(fpath)
	if err != nil || img == nil {
		return "", err
	}
	return fmt.Sprintf("%d,%d", img.Bounds().Dx(), img.Bounds().Dy()), nil
}

// sudo apt install ffmpeg
// https://pkg.go.dev/github.com/jtguibas/cinema#section-readme

// return "width,height"
func GetVideoSize(fpath string) (string, error) {
	video, err := cinema.Load(fpath)
	if err != nil || video == nil {
		return "", err
	}
	return fmt.Sprintf("%d,%d", video.Width(), video.Height()), nil
}
