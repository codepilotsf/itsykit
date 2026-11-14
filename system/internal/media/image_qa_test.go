package media

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"strings"
	"testing"
)

func TestQAPhotoProcessResizeTransparencyAndPayload(t *testing.T) {
	for _, format := range []string{"jpeg", "png"} {
		t.Run(format, func(t *testing.T) {
			src := image.NewNRGBA(image.Rect(0, 0, 4096, 512))
			for y := 0; y < 512; y++ {
				for x := 0; x < 4096; x++ {
					src.SetNRGBA(x, y, color.NRGBA{R: 200, G: 100, B: 50, A: 128})
				}
			}
			var b bytes.Buffer
			var e error
			if format == "jpeg" {
				e = jpeg.Encode(&b, src, nil)
			} else {
				e = png.Encode(&b, src)
			}
			if e != nil {
				t.Fatal(e)
			}
			input := append(b.Bytes(), []byte("secret trailing metadata")...)
			p, data, e := Process(input, "Example", "Description")
			if e != nil {
				t.Fatal(e)
			}
			if p.Width != 2048 || p.Height != 256 || bytes.Contains(data, []byte("secret trailing metadata")) {
				t.Fatal("resize or metadata stripping failed", p)
			}
			if e = CheckBytes(p, data); e != nil {
				t.Fatal(e)
			}
			out, _, e := image.Decode(bytes.NewReader(data))
			if e != nil {
				t.Fatal(e)
			}
			if format == "png" {
				_, _, _, a := out.At(10, 10).RGBA()
				if a < 32000 || a > 34000 {
					t.Fatal("transparency lost", a)
				}
			}
		})
	}
}
func TestQAPhotoInvalidInputs(t *testing.T) {
	var tooWide bytes.Buffer
	if err := png.Encode(&tooWide, image.NewRGBA(image.Rect(0, 0, 20001, 1))); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Process(tooWide.Bytes(), "Too wide", ""); err == nil {
		t.Fatal("accepted excessive dimensions")
	}
	for _, data := range [][]byte{nil, []byte("<svg/>"), bytes.Repeat([]byte{'x'}, MaxFile+1)} {
		if _, _, e := Process(data, "name", ""); e == nil {
			t.Fatal("accepted invalid input")
		}
	}
	var b bytes.Buffer
	if e := png.Encode(&b, image.NewRGBA(image.Rect(0, 0, 1, 1))); e != nil {
		t.Fatal(e)
	}
	for _, name := range []string{strings.Repeat("a", 201), string([]byte{0xff})} {
		if _, _, e := Process(b.Bytes(), name, ""); e == nil {
			t.Fatal("invalid metadata accepted")
		}
	}
	digest := sha256.Sum256(b.Bytes())
	if e := CheckBytes(Photo{File: hex.EncodeToString(digest[:]), Width: 1, Height: 1}, b.Bytes()); e == nil {
		t.Fatal("invalid filename accepted")
	}
}

func TestQAOrientationTransforms(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 2, 3))
	for i := 0; i < 6; i++ {
		src.SetNRGBA(i%2, i/2, color.NRGBA{R: uint8(i + 1), A: 255})
	}
	expected := map[int][]uint8{1: {1, 2, 3, 4, 5, 6}, 2: {2, 1, 4, 3, 6, 5}, 3: {6, 5, 4, 3, 2, 1}, 4: {5, 6, 3, 4, 1, 2}, 5: {1, 3, 5, 2, 4, 6}, 6: {5, 3, 1, 6, 4, 2}, 7: {6, 4, 2, 5, 3, 1}, 8: {2, 4, 6, 1, 3, 5}}
	for rotation, want := range expected {
		got := orient(src, rotation)
		for i, v := range want {
			r, _, _, _ := got.At(i%got.Bounds().Dx(), i/got.Bounds().Dx()).RGBA()
			if uint8(r>>8) != v {
				t.Fatalf("orientation %d pixel %d: got %d want %d", rotation, i, r>>8, v)
			}
		}
	}
}
