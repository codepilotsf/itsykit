package editor

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"fmt"
	"net/http"
	"strings"
	"time"
)

//go:embed inline/blog.js
var blogJS []byte

//go:embed inline/squire.js
var inlineSquire []byte

//go:embed inline/purify.js
var inlinePurify []byte
var editorAssets = func() map[string][]byte {
	blog := append([]byte(nil), inlinePurify...)
	blog = append(blog, '\n', ';', '\n')
	blog = append(blog, inlineSquire...)
	blog = append(blog, '\n', ';', '\n')
	blog = append(blog, blogJS...)
	return map[string][]byte{"blog.js": blog}
}()

// MediaValidationError is a safe validation explanation without operational details.
type MediaValidationError struct{ Message string }

func (e *MediaValidationError) Error() string { return e.Message }

var mediaSlot = make(chan struct{}, 1)

func serveEditorAsset(w http.ResponseWriter, r *http.Request) bool {
	if !strings.HasPrefix(r.URL.Path, "/admin/assets/") {
		return false
	}
	data, ok := editorAssets[strings.TrimPrefix(r.URL.Path, "/admin/assets/")]
	if !ok || r.URL.RawPath != "" {
		http.NotFound(w, r)
		return true
	}
	if r.Method != "GET" && r.Method != "HEAD" {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "Method not allowed", 405)
		return true
	}
	w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=0, must-revalidate")
	w.Header().Set("ETag", fmt.Sprintf(`"%x"`, sha256.Sum256(data)))
	http.ServeContent(w, r, r.URL.Path, time.Time{}, bytes.NewReader(data))
	return true
}
