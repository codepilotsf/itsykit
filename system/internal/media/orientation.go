package media

import (
	"encoding/binary"
	"image"
	"image/color"
)

// Read only the bounded EXIF orientation field. No original metadata is kept.
func jpegOrientation(data []byte) int {
	for pos := 2; pos+4 <= len(data); {
		if data[pos] != 0xff {
			return 1
		}
		marker := data[pos+1]
		if marker == 0xda || marker == 0xd9 {
			return 1
		}
		if marker == 0xff {
			pos++
			continue
		}
		size := int(binary.BigEndian.Uint16(data[pos+2 : pos+4]))
		if size < 2 || pos+2+size > len(data) {
			return 1
		}
		payload := data[pos+4 : pos+2+size]
		pos += 2 + size
		if marker != 0xe1 || len(payload) < 14 || string(payload[:6]) != "Exif\x00\x00" {
			continue
		}
		t := payload[6:]
		var order binary.ByteOrder
		switch string(t[:2]) {
		case "II":
			order = binary.LittleEndian
		case "MM":
			order = binary.BigEndian
		default:
			return 1
		}
		if order.Uint16(t[2:4]) != 42 {
			return 1
		}
		offset := uint64(order.Uint32(t[4:8]))
		if offset+2 > uint64(len(t)) {
			return 1
		}
		count := int(order.Uint16(t[offset : offset+2]))
		for i := 0; i < count && i < 256; i++ {
			at := offset + 2 + uint64(i)*12
			if at+12 > uint64(len(t)) {
				return 1
			}
			field := t[at : at+12]
			if order.Uint16(field[:2]) == 0x112 && order.Uint16(field[2:4]) == 3 && order.Uint32(field[4:8]) == 1 {
				v := int(order.Uint16(field[8:10]))
				if v >= 1 && v <= 8 {
					return v
				}
				return 1
			}
		}
	}
	return 1
}

type oriented struct {
	src      image.Image
	rotation int
}

func orient(src image.Image, rotation int) image.Image {
	if rotation < 2 || rotation > 8 {
		return src
	}
	return oriented{src, rotation}
}
func (o oriented) ColorModel() color.Model { return o.src.ColorModel() }
func (o oriented) Bounds() image.Rectangle {
	b := o.src.Bounds()
	w, h := b.Dx(), b.Dy()
	if o.rotation >= 5 {
		w, h = h, w
	}
	return image.Rect(0, 0, w, h)
}
func (o oriented) At(x, y int) color.Color {
	b := o.src.Bounds()
	w, h := b.Dx(), b.Dy()
	switch o.rotation {
	case 2:
		x = w - 1 - x
	case 3:
		x, y = w-1-x, h-1-y
	case 4:
		y = h - 1 - y
	case 5:
		x, y = y, x
	case 6:
		x, y = y, h-1-x
	case 7:
		x, y = w-1-y, h-1-x
	case 8:
		x, y = w-1-y, x
	}
	return o.src.At(b.Min.X+x, b.Min.Y+y)
}
