package editor

import (
	"crypto/sha256"
	_ "embed"
	"encoding/base64"
	"itsy/internal/devaccess"
	"itsy/internal/snips"
	"itsy/internal/viewer"
	"itsy/templates"
	"net"
	"net/http"
	"net/url"
	"strings"
)

// Only the canonical, data-free preference initializer may run inline.
var preferenceScriptHash = func() string {
	markup, err := templates.Files.ReadFile("site/admin/_toolbar.html")
	if err != nil {
		panic(err)
	}
	_, tail, ok := strings.Cut(string(markup), "<script>")
	if !ok {
		panic("missing preference initializer")
	}
	script, _, ok := strings.Cut(tail, "</script>")
	if !ok {
		panic("unclosed preference initializer")
	}
	sum := sha256.Sum256([]byte(script))
	return " 'sha256-" + base64.StdEncoding.EncodeToString(sum[:]) + "'"
}()

//go:embed inline/squire.js
var inlineSquire []byte

//go:embed inline/purify.js
var inlinePurify []byte

//go:embed inline/owner.js
var inlineOwner []byte

//go:embed inline/collections.js
var inlineCollections []byte

//go:embed inline/media.js
var inlineMedia []byte

//go:embed inline/owner.html
var inlineUI string

// ServePublic authenticates before requesting the owner-only rendering path.
// Anonymous pages are untouched, including scripts and reload behavior.
func (h *Handler) ServePublic(w http.ResponseWriter, r *http.Request, public http.Handler) {
	if h.options.Public == nil || (r.Method != "GET" && r.Method != "HEAD") || h.getCookie(r, "session") == "" {
		public.ServeHTTP(w, r)
		return
	}
	if h.options.Local {
		host := r.Host
		if name, _, err := net.SplitHostPort(host); err == nil {
			host = name
		}
		ip := net.ParseIP(host)
		if !devaccess.LAN(r) && host != "localhost" && (ip == nil || !ip.IsLoopback()) {
			public.ServeHTTP(w, r)
			return
		}
	} else if r.TLS == nil {
		public.ServeHTTP(w, r)
		return
	}
	s, logged, err := h.current(r)
	if err != nil {
		w.Header().Set("Cache-Control", "private, no-store")
		h.sessionError(w, err)
		return
	}
	if !logged {
		public.ServeHTTP(w, r)
		return
	}
	r = r.WithContext(viewer.With(r.Context(), viewer.Admin{Email: s.email, CSRF: s.csrf}))
	h.options.Public(w, r, func(body []byte, store *snips.Store, version string) ([]byte, bool) {
		// Renew only actual page navigation, not assets or background reload polls.
		if !h.renew(w, r) {
			return nil, false
		}
		scheme := "https"
		if h.options.Local {
			scheme = "http"
		}
		assetURL := (&url.URL{Scheme: scheme, Host: r.Host, Path: "/admin/assets/editor.js"}).String()
		policy := "default-src 'none'; style-src 'self' https: 'unsafe-inline'; font-src 'self' https: data:; img-src 'self' https: data: blob:; media-src 'self' https: blob:; connect-src 'self'; form-action 'self'; frame-ancestors 'none'; base-uri 'none'; script-src " + assetURL
		w.Header().Set("Content-Security-Policy", policy+preferenceScriptHash)
		w.Header().Set("Cache-Control", "private, no-store")
		w.Header().Set("Referrer-Policy", "same-origin")
		w.Header().Set("X-Frame-Options", "DENY")
		return body, true
	})
}
