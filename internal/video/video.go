// Package video validates web-ready MP4 uploads without decoding or transcoding.
package video

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"regexp"
	"unicode/utf8"
)

const MaxFile int64 = 25 << 20
const MaxTotal int64 = 64 << 20
const MaxVideos = 50

var Filename = regexp.MustCompile(`^[0-9a-f]{64}\.mp4$`)
var ID = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9]{9}$`)

type Metadata struct {
	File     string  `json:"file"`
	Size     int64   `json:"size"`
	Width    int     `json:"width"`
	Height   int     `json:"height"`
	Duration float64 `json:"duration"`
}
type Video struct {
	Metadata
	Name   string `json:"name"`
	Poster string `json:"poster,omitempty"`
}

func Validate(values map[string]Video) error {
	if len(values) > MaxVideos {
		return fmt.Errorf("video library is limited to 50 videos")
	}
	for id, v := range values {
		if !ID.MatchString(id) || !Filename.MatchString(v.File) || v.Size <= 0 || v.Size > MaxFile || v.Width < 1 || v.Height < 1 || v.Width > 8192 || v.Height > 8192 || math.IsNaN(v.Duration) || math.IsInf(v.Duration, 0) || v.Duration <= 0 || len(v.Name) > 200 || !utf8.ValidString(v.Name) || (v.Poster != "" && !ID.MatchString(v.Poster)) {
			return fmt.Errorf("invalid video metadata")
		}
	}
	return nil
}

type box struct {
	kind             string
	start, body, end int64
}

func boxes(r io.ReaderAt, start, end int64) ([]box, error) {
	var out []box
	for start < end {
		if len(out) >= 10000 || end-start < 8 {
			return nil, fmt.Errorf("invalid MP4 box structure")
		}
		var h [16]byte
		if _, e := r.ReadAt(h[:8], start); e != nil {
			return nil, e
		}
		n := int64(binary.BigEndian.Uint32(h[:4]))
		header := int64(8)
		if n == 1 {
			if end-start < 16 {
				return nil, fmt.Errorf("truncated MP4 box")
			}
			if _, e := r.ReadAt(h[8:], start+8); e != nil {
				return nil, e
			}
			u := binary.BigEndian.Uint64(h[8:])
			if u > math.MaxInt64 {
				return nil, fmt.Errorf("oversized MP4 box")
			}
			n = int64(u)
			header = 16
		}
		if n == 0 {
			n = end - start
		}
		if n < header || n > end-start {
			return nil, fmt.Errorf("truncated MP4 box")
		}
		out = append(out, box{string(h[4:8]), start, start + header, start + n})
		start += n
	}
	return out, nil
}
func one(r io.ReaderAt, parent box, kind string) (box, error) {
	children, e := boxes(r, parent.body, parent.end)
	if e != nil {
		return box{}, e
	}
	var found box
	count := 0
	for _, b := range children {
		if b.kind == kind {
			found = b
			count++
		}
	}
	if count != 1 {
		return box{}, fmt.Errorf("MP4 requires one %s box", kind)
	}
	return found, nil
}
func contents(r io.ReaderAt, b box, limit int64) ([]byte, error) {
	if b.end-b.body > limit {
		return nil, fmt.Errorf("MP4 metadata is too large")
	}
	v := make([]byte, b.end-b.body)
	_, e := r.ReadAt(v, b.body)
	return v, e
}

// Inspect checks box bounds, fast-start layout, sample tables, codec declarations
// and decoder configuration. It is not a full bitstream decoder or malware scan.
func Inspect(r io.ReaderAt, size int64) (Metadata, error) {
	var result Metadata
	if size <= 0 || size > MaxFile {
		return result, fmt.Errorf("choose an optimized MP4 up to 25 MiB")
	}
	top, e := boxes(r, 0, size)
	if e != nil {
		return result, e
	}
	var moov, mdat box
	types := map[string]int{}
	for _, b := range top {
		types[b.kind]++
		switch b.kind {
		case "ftyp":
			v, e := contents(r, b, 1024)
			if e != nil {
				return result, e
			}
			if len(v) < 8 || (len(v)-8)%4 != 0 {
				return result, fmt.Errorf("invalid MP4 file type")
			}
			ok := false
			for i := 0; i+4 <= len(v); i += 4 {
				if i == 4 {
					continue
				}
				switch string(v[i : i+4]) {
				case "isom", "iso2", "mp41", "mp42", "avc1":
					ok = true
				}
			}
			if !ok {
				return result, fmt.Errorf("use an MP4 container")
			}
		case "moov":
			moov = b
		case "mdat":
			mdat = b
		case "free", "skip", "wide":
		default:
			return result, fmt.Errorf("unsupported MP4 layout; export a non-fragmented web MP4")
		}
	}
	if types["ftyp"] != 1 || types["moov"] != 1 || types["mdat"] != 1 || moov.end > mdat.start || mdat.end == mdat.body {
		return result, fmt.Errorf("export an MP4 with web optimization (fast start) enabled")
	}
	if moov.end-moov.body > 2<<20 {
		return result, fmt.Errorf("MP4 metadata exceeds 2 MiB")
	}
	children, e := boxes(r, moov.body, moov.end)
	if e != nil {
		return result, e
	}
	videoCount, audioCount := 0, 0
	for _, track := range children {
		if track.kind == "mvex" {
			return result, fmt.Errorf("fragmented MP4 is unsupported")
		}
		if track.kind != "trak" {
			continue
		}
		mdia, e := one(r, track, "mdia")
		if e != nil {
			return result, e
		}
		handler, e := one(r, mdia, "hdlr")
		if e != nil {
			return result, e
		}
		h, e := contents(r, handler, 4096)
		if e != nil || len(h) < 12 {
			return result, fmt.Errorf("invalid MP4 track handler")
		}
		kind := string(h[8:12])
		if kind != "vide" && kind != "soun" {
			return result, fmt.Errorf("MP4 must contain only video and optional audio tracks")
		}
		minf, e := one(r, mdia, "minf")
		if e != nil {
			return result, e
		}
		dinf, e := one(r, minf, "dinf")
		if e != nil {
			return result, e
		}
		dref, e := one(r, dinf, "dref")
		if e != nil {
			return result, e
		}
		dr, e := contents(r, dref, 4096)
		if e != nil || len(dr) < 20 || binary.BigEndian.Uint32(dr[4:8]) != 1 || binary.BigEndian.Uint32(dr[8:12]) != 12 || string(dr[12:16]) != "url " || binary.BigEndian.Uint32(dr[16:20]) != 1 {
			return result, fmt.Errorf("external MP4 media references are unsupported")
		}
		stbl, e := one(r, minf, "stbl")
		if e != nil {
			return result, e
		}
		stsd, e := one(r, stbl, "stsd")
		if e != nil {
			return result, e
		}
		v, e := contents(r, stsd, 1<<20)
		if e != nil || len(v) < 16 || binary.BigEndian.Uint32(v[4:8]) != 1 {
			return result, fmt.Errorf("MP4 requires one codec per track")
		}
		entries, e := boxes(r, stsd.body+8, stsd.end)
		if e != nil || len(entries) != 1 {
			return result, fmt.Errorf("invalid MP4 codec entry")
		}
		entry := entries[0]
		b, e := contents(r, entry, 1<<20)
		if e != nil || len(b) < 8 || binary.BigEndian.Uint16(b[6:8]) != 1 {
			return result, fmt.Errorf("invalid MP4 sample entry")
		}
		if kind == "vide" {
			videoCount++
			if videoCount > 1 || entry.kind != "avc1" || len(b) < 78 {
				return result, fmt.Errorf("use H.264 video in MP4, not HEVC or other codecs")
			}
			result.Width = int(binary.BigEndian.Uint16(b[24:26]))
			result.Height = int(binary.BigEndian.Uint16(b[26:28]))
			if result.Width < 1 || result.Height < 1 || result.Width > 8192 || result.Height > 8192 {
				return result, fmt.Errorf("invalid video dimensions")
			}
			avcc, e := one(r, box{body: entry.body + 78, end: entry.end}, "avcC")
			if e != nil {
				return result, e
			}
			av, e := contents(r, avcc, 65536)
			if e != nil || !validAVC(av) {
				return result, fmt.Errorf("use 8-bit H.264 Baseline, Main or High profile")
			}
			mdhd, e := one(r, mdia, "mdhd")
			if e != nil {
				return result, e
			}
			md, e := contents(r, mdhd, 128)
			if e != nil || len(md) < 20 {
				return result, fmt.Errorf("invalid video duration")
			}
			var scale uint32
			var duration uint64
			if md[0] == 0 {
				scale = binary.BigEndian.Uint32(md[12:16])
				duration = uint64(binary.BigEndian.Uint32(md[16:20]))
			} else if md[0] == 1 && len(md) >= 32 {
				scale = binary.BigEndian.Uint32(md[20:24])
				duration = binary.BigEndian.Uint64(md[24:32])
			} else {
				return result, fmt.Errorf("invalid video duration")
			}
			if scale == 0 || duration == 0 {
				return result, fmt.Errorf("invalid video duration")
			}
			result.Duration = float64(duration) / float64(scale)
		} else {
			audioCount++
			if audioCount > 1 || entry.kind != "mp4a" || len(b) < 28 || binary.BigEndian.Uint16(b[8:10]) != 0 {
				return result, fmt.Errorf("use AAC audio or export without audio")
			}
			esds, e := one(r, box{body: entry.body + 28, end: entry.end}, "esds")
			if e != nil {
				return result, e
			}
			es, e := contents(r, esds, 65536)
			if e != nil || !validAAC(es) {
				return result, fmt.Errorf("use AAC-LC audio or export without audio")
			}
		}
		if e := validateSamples(r, stbl, mdat); e != nil {
			return result, e
		}
	}
	if videoCount != 1 {
		return result, fmt.Errorf("MP4 needs one H.264 video track")
	}
	h := sha256.New()
	n, e := io.Copy(h, io.NewSectionReader(r, 0, size))
	if e != nil {
		return result, e
	}
	if n != size {
		return result, fmt.Errorf("truncated MP4 data")
	}
	result.File = fmt.Sprintf("%x.mp4", h.Sum(nil))
	result.Size = size
	return result, nil
}

func validAVC(b []byte) bool {
	if len(b) < 7 || b[0] != 1 || (b[1] != 66 && b[1] != 77 && b[1] != 100) || b[4]&3 == 2 {
		return false
	}
	n := int(b[5] & 31)
	if n == 0 {
		return false
	}
	i := 6
	read := func(count int) bool {
		for j := 0; j < count; j++ {
			if i+2 > len(b) {
				return false
			}
			size := int(binary.BigEndian.Uint16(b[i : i+2]))
			i += 2
			if size == 0 || size > len(b)-i {
				return false
			}
			i += size
		}
		return true
	}
	if !read(n) || i >= len(b) {
		return false
	}
	n = int(b[i])
	i++
	if n == 0 || !read(n) {
		return false
	}
	if i < len(b) {
		if b[1] != 100 || len(b)-i < 4 || b[i]&3 != 1 || b[i+1]&7 != 0 || b[i+2]&7 != 0 {
			return false
		}
		n = int(b[i+3])
		i += 4
		if !read(n) {
			return false
		}
	}
	return i == len(b)
}
func descriptor(b []byte) (byte, []byte, []byte, bool) {
	if len(b) < 2 {
		return 0, nil, nil, false
	}
	tag := b[0]
	n, i := 0, 1
	for j := 0; j < 4; j++ {
		if i >= len(b) {
			return 0, nil, nil, false
		}
		v := b[i]
		i++
		n = n<<7 | int(v&127)
		if v&128 == 0 {
			if n > len(b)-i {
				return 0, nil, nil, false
			}
			return tag, b[i : i+n], b[i+n:], true
		}
	}
	return 0, nil, nil, false
}
func validAAC(b []byte) bool {
	if len(b) < 4 {
		return false
	}
	tag, es, _, ok := descriptor(b[4:])
	if !ok || tag != 3 || len(es) < 3 {
		return false
	}
	flags := es[2]
	i := 3
	if flags&128 != 0 {
		i += 2
	}
	if flags&64 != 0 {
		if i >= len(es) {
			return false
		}
		i += 1 + int(es[i])
	}
	if flags&32 != 0 {
		i += 2
	}
	if i > len(es) {
		return false
	}
	tag, dc, _, ok := descriptor(es[i:])
	if !ok || tag != 4 || len(dc) < 13 || dc[0] != 0x40 || dc[1]>>2 != 5 {
		return false
	}
	tag, asc, _, ok := descriptor(dc[13:])
	if !ok || tag != 5 || len(asc) < 2 {
		return false
	}
	frequency := (asc[0]&7)<<1 | asc[1]>>7
	channels := (asc[1] >> 3) & 15
	return asc[0]>>3 == 2 && frequency < 13 && channels > 0 && channels <= 7
}

func validateSamples(r io.ReaderAt, stbl, mdat box) error {
	read := func(kind string) ([]byte, error) {
		b, e := one(r, stbl, kind)
		if e != nil {
			return nil, e
		}
		return contents(r, b, 2<<20)
	}
	sz, e := read("stsz")
	if e != nil || len(sz) < 12 {
		return fmt.Errorf("invalid MP4 sample sizes")
	}
	fixed := uint64(binary.BigEndian.Uint32(sz[4:8]))
	count := int(binary.BigEndian.Uint32(sz[8:12]))
	if count < 1 || count > 1000000 || (fixed == 0 && len(sz) != 12+count*4) {
		return fmt.Errorf("invalid MP4 sample count")
	}
	sc, e := read("stsc")
	if e != nil || len(sc) < 20 {
		return fmt.Errorf("invalid MP4 chunk mapping")
	}
	n := int(binary.BigEndian.Uint32(sc[4:8]))
	if n < 1 || n > 100000 || len(sc) != 8+n*12 {
		return fmt.Errorf("invalid MP4 chunk mapping")
	}
	co, e := read("stco")
	wide := false
	if e != nil {
		co, e = read("co64")
		wide = true
	}
	if e != nil || len(co) < 8 {
		return fmt.Errorf("invalid MP4 chunk offsets")
	}
	chunks := int(binary.BigEndian.Uint32(co[4:8]))
	step := 4
	if wide {
		step = 8
	}
	if chunks < 1 || chunks > 1000000 || len(co) != 8+chunks*step {
		return fmt.Errorf("invalid MP4 chunk offsets")
	}
	for j := 0; j < n; j++ {
		first := int(binary.BigEndian.Uint32(sc[8+j*12:]))
		per := binary.BigEndian.Uint32(sc[12+j*12:])
		desc := binary.BigEndian.Uint32(sc[16+j*12:])
		if first < 1 || first > chunks || per == 0 || desc != 1 || (j == 0 && first != 1) || (j > 0 && first <= int(binary.BigEndian.Uint32(sc[8+(j-1)*12:]))) {
			return fmt.Errorf("invalid MP4 chunk mapping")
		}
	}
	index, mapping := 0, 0
	for chunk := 1; chunk <= chunks; chunk++ {
		if mapping+1 < n && chunk >= int(binary.BigEndian.Uint32(sc[8+(mapping+1)*12:])) {
			mapping++
		}
		per := int(binary.BigEndian.Uint32(sc[12+mapping*12:]))
		if per > count-index {
			return fmt.Errorf("invalid MP4 sample mapping")
		}
		var length uint64
		for j := 0; j < per; j++ {
			v := fixed
			if v == 0 {
				v = uint64(binary.BigEndian.Uint32(sz[12+index*4:]))
			}
			if v == 0 {
				return fmt.Errorf("empty MP4 sample")
			}
			length += v
			index++
		}
		i := 8 + (chunk-1)*step
		offset := uint64(binary.BigEndian.Uint32(co[i:]))
		if wide {
			offset = binary.BigEndian.Uint64(co[i:])
		}
		if offset < uint64(mdat.body) || offset > uint64(mdat.end) || length > uint64(mdat.end)-offset {
			return fmt.Errorf("MP4 samples fall outside media data")
		}
	}
	if index != count {
		return fmt.Errorf("MP4 sample count mismatch")
	}
	timing, e := read("stts")
	if e != nil || len(timing) < 8 {
		return fmt.Errorf("invalid MP4 sample timing")
	}
	entries := int(binary.BigEndian.Uint32(timing[4:8]))
	if entries < 1 || entries > 1000000 || len(timing) != 8+entries*8 {
		return fmt.Errorf("invalid MP4 sample timing")
	}
	var samples uint64
	for i := 0; i < entries; i++ {
		n := binary.BigEndian.Uint32(timing[8+i*8:])
		delta := binary.BigEndian.Uint32(timing[12+i*8:])
		if n == 0 || delta == 0 {
			return fmt.Errorf("invalid MP4 sample timing")
		}
		samples += uint64(n)
	}
	if samples != uint64(count) {
		return fmt.Errorf("MP4 timing sample count mismatch")
	}
	return nil
}
