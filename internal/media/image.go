// Package media handles bounded photo processing without external executables.
package media

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"regexp"
	"unicode/utf8"
)

const MaxFile = 8 << 20
const MaxTotal = 64 << 20
const MaxPhotos = 100

type Photo struct {
	File   string `json:"file"`
	Name   string `json:"name"`
	Alt    string `json:"alt"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

var Filename = regexp.MustCompile(`^[0-9a-f]{64}\.(jpg|png)$`)

func Validate(photos map[string]Photo) error {
	if len(photos) > MaxPhotos {
		return fmt.Errorf("photo library is limited to 100 photos")
	}
	for id, p := range photos {
		if !regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]{9}$`).MatchString(id) || !Filename.MatchString(p.File) || len(p.Name) > 200 || len(p.Alt) > 1000 || !utf8.ValidString(p.Name+p.Alt) || p.Width < 1 || p.Height < 1 || p.Width > 2048 || p.Height > 2048 {
			return fmt.Errorf("invalid photo metadata")
		}
	}
	return nil
}
func CheckBytes(p Photo, b []byte) error {
	return CheckReader(p, bytes.NewReader(b), int64(len(b)))
}

// CheckReader validates encoded bytes without retaining the full photo in memory.
func CheckReader(p Photo, r io.ReaderAt, size int64) error {
	if !Filename.MatchString(p.File) {
		return fmt.Errorf("invalid photo filename")
	}
	if size < 0 || size > MaxFile {
		return fmt.Errorf("photo exceeds 8 MiB")
	}
	h := sha256.New()
	n, err := io.Copy(h, io.NewSectionReader(r, 0, size))
	if err != nil {
		return err
	}
	if n != size || hex.EncodeToString(h.Sum(nil)) != p.File[:64] {
		return fmt.Errorf("photo checksum mismatch")
	}
	c, format, err := image.DecodeConfig(io.NewSectionReader(r, 0, size))
	if err != nil {
		return err
	}
	ext := "png"
	if format == "jpeg" {
		ext = "jpg"
	}
	if format != "png" && format != "jpeg" || p.File[65:] != ext || c.Width != p.Width || c.Height != p.Height {
		return fmt.Errorf("photo metadata mismatch")
	}
	return nil
}

// Process accepts JPEG/PNG only. Decode dimensions before allocating pixels,
// resize to 2048px, then re-encode to discard embedded metadata and extra payloads.
func Process(data []byte, name, alt string) (Photo, []byte, error) {
	p := Photo{Name: name, Alt: alt}
	if len(data) > MaxFile || len(name) > 200 || len(alt) > 1000 || !utf8.ValidString(name+alt) {
		return p, nil, fmt.Errorf("photo or description exceeds limits")
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || format != "jpeg" && format != "png" {
		return p, nil, fmt.Errorf("choose a JPEG or PNG photo")
	}
	if config.Width < 1 || config.Height < 1 || config.Width > 20000 || config.Height > 20000 || int64(config.Width)*int64(config.Height) > 20000000 {
		return p, nil, fmt.Errorf("photo exceeds 20 megapixels")
	}
	src, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return p, nil, fmt.Errorf("photo could not be decoded")
	}
	if format == "jpeg" {
		src = orient(src, jpegOrientation(data))
	}
	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	if w > 2048 || h > 2048 {
		if w >= h {
			h = max(1, h*2048/w)
			w = 2048
		} else {
			w = max(1, w*2048/h)
			h = 2048
		}
		src = downsample(src, w, h)
	}
	var out bytes.Buffer
	ext := "png"
	if format == "jpeg" {
		ext = "jpg"
		err = jpeg.Encode(&out, src, &jpeg.Options{Quality: 85})
	} else {
		err = png.Encode(&out, src)
	}
	if err != nil {
		return p, nil, err
	}
	if out.Len() > MaxFile {
		return p, nil, fmt.Errorf("processed photo exceeds 8 MiB")
	}
	sum := sha256.Sum256(out.Bytes())
	p.File = hex.EncodeToString(sum[:]) + "." + ext
	p.Width = w
	p.Height = h
	return p, out.Bytes(), nil
}

// Area averaging avoids nearest-neighbour aliasing when reducing photographs.
// Premultiplied channels preserve transparent PNG edges.
func downsample(src image.Image, w, h int) image.Image {
	b := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		y0, y1 := y*b.Dy()/h, (y+1)*b.Dy()/h
		for x := 0; x < w; x++ {
			x0, x1 := x*b.Dx()/w, (x+1)*b.Dx()/w
			var r, g, bl, a, n uint64
			for sy := y0; sy < y1; sy++ {
				for sx := x0; sx < x1; sx++ {
					rr, gg, bb, aa := src.At(b.Min.X+sx, b.Min.Y+sy).RGBA()
					r += uint64(rr)
					g += uint64(gg)
					bl += uint64(bb)
					a += uint64(aa)
					n++
				}
			}
			dst.SetRGBA(x, y, color.RGBA{uint8(r / n >> 8), uint8(g / n >> 8), uint8(bl / n >> 8), uint8(a / n >> 8)})
		}
	}
	return dst
}
