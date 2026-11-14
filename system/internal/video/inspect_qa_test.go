package video

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"os"
	"testing"
)

func qaFixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile("testdata/" + name + ".mp4")
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestQAInspectRealH264Faststart(t *testing.T) {
	for _, name := range []string{"valid", "silent"} {
		t.Run(name, func(t *testing.T) {
			b := qaFixture(t, name)
			m, err := Inspect(bytes.NewReader(b), int64(len(b)))
			if err != nil {
				t.Fatal(err)
			}
			h := sha256.Sum256(b)
			if m.Width != 32 || m.Height != 24 || m.Duration < 0.29 || m.Duration > 0.4 || m.Size != int64(len(b)) || m.File != hex.EncodeToString(h[:])+".mp4" {
				t.Fatalf("unexpected metadata: %+v", m)
			}
		})
	}
}

func TestQAInspectRejectsUnsupportedOrMalformed(t *testing.T) {
	good := qaFixture(t, "valid")
	zeroSize := bytes.Clone(good)
	binary.BigEndian.PutUint32(zeroSize[:4], 1) // Extended header missing a valid extent.
	overrun := bytes.Clone(good)
	binary.BigEndian.PutUint32(overrun[:4], uint32(len(good)+100))
	for name, b := range map[string][]byte{
		"no faststart":         qaFixture(t, "no-faststart"),
		"wrong codec":          qaFixture(t, "wrong-codec"),
		"wrong audio":          qaFixture(t, "wrong-audio"),
		"empty":                nil,
		"text":                 []byte("not a video"),
		"truncated header":     good[:6],
		"truncated payload":    good[:len(good)-1],
		"invalid extended box": zeroSize,
		"box beyond file":      overrun,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := Inspect(bytes.NewReader(b), int64(len(b))); err == nil {
				t.Fatal("accepted invalid video")
			}
		})
	}
	if _, err := Inspect(bytes.NewReader(good), 1<<40); err == nil {
		t.Fatal("accepted oversized video")
	}
	if _, err := Inspect(bytes.NewReader(good), -1); err == nil {
		t.Fatal("accepted negative size")
	}
}

func TestQAInspectRejectsShortReaderAndMissingTiming(t *testing.T) {
	good := qaFixture(t, "valid")
	t.Run("short payload reader", func(t *testing.T) {
		if _, err := Inspect(bytes.NewReader(good[:len(good)-1]), int64(len(good))); err == nil {
			t.Fatal("accepted reader shorter than declared file size")
		}
	})
	t.Run("missing time-to-sample table", func(t *testing.T) {
		bad := bytes.Replace(bytes.Clone(good), []byte("stts"), []byte("free"), 1)
		if bytes.Equal(bad, good) {
			t.Fatal("fixture lacks timing table")
		}
		if _, err := Inspect(bytes.NewReader(bad), int64(len(bad))); err == nil {
			t.Fatal("accepted video track without time-to-sample table")
		}
	})
}
