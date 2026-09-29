package site

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
)

func TestBundledIconCompilation(t *testing.T) {
	svg := []byte(`<svg viewBox="0 0 24 24"><path d="M5 12h14"/></svg>`)
	markup, err := CompileIconMarkup(`{icon plus size="18"}`, svg)
	if err != nil || !strings.Contains(markup, `width="18"`) || !strings.Contains(markup, `aria-hidden="true"`) {
		t.Fatalf("%s: %v", markup, err)
	}
	if _, err := CompileIconMarkup(`{icon plus size="{app.size}"}`, svg); err == nil {
		t.Fatal("accepted dynamic bundled icon")
	}
	if _, err := CompileIconMarkup(`{icon plus}`, []byte(`<svg><script>alert(1)</script></svg>`)); err == nil {
		t.Fatal("accepted executable SVG")
	}
}

func TestEmptyCollectionPhotoVisiblePlaceholder(t *testing.T) {
	row := multiRow{PhotoFields: map[string]bool{"photo": true}, Fields: map[string]string{}}
	src, err := multiPhoto(row, "photo", "src")
	if err != nil {
		t.Fatal(err)
	}
	data, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(fmt.Sprint(src), "data:image/svg+xml;base64,"))
	if err != nil || !strings.Contains(string(data), `<circle`) || !strings.Contains(string(data), `viewBox="0 0 160 160"`) {
		t.Fatalf("placeholder: %s, %v", data, err)
	}
}
