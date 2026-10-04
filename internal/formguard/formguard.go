// Package formguard protects anonymous mini-app forms, not admin sessions.
// A Guard belongs to one site's runtime, never its immutable content snapshot.
package formguard

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"flea/internal/devaccess"
)

const Lifetime = 24 * time.Hour
const Field = "_csrf"

// ErrInvalid deliberately contains no request data or secrets.
var ErrInvalid = fmt.Errorf("Please reload the form and try again.")

type Guard struct {
	site  string
	local bool
	once  sync.Once
	key   [32]byte
	err   error
	now   func() time.Time
}

// New does no I/O and creates no secrets. The first rendered form initializes
// the signing key. Keep this instance across template reloads.
func New(site string, local bool) *Guard {
	return &Guard{site: site, local: local, now: time.Now}
}

// SiteID is trusted runtime identity, not a browser-selected site identifier.
func (g *Guard) SiteID() string { return g.site }

// Binding reads an existing valid visitor cookie without issuing a new one.
// Its digest may bind private runtime data, but is never exposed to templates.
func (g *Guard) Binding(r *http.Request) (string, bool) {
	origin, err := g.origin(r)
	if err != nil {
		return "", false
	}
	if err := g.init(); err != nil {
		return "", false
	}
	payload, ok := g.read(r, origin)
	if !ok {
		return "", false
	}
	return base64.RawURLEncoding.EncodeToString(g.mac("binding", origin, payload)), true
}

func (g *Guard) init() error {
	g.once.Do(func() { _, g.err = rand.Read(g.key[:]) })
	if g.err != nil {
		return fmt.Errorf("Form protection is unavailable.")
	}
	return nil
}

func (g *Guard) cookieName() string {
	if !g.local {
		return "__Host-flea-form"
	}
	id := sha256.Sum256([]byte(g.site))
	return fmt.Sprintf("flea_dev_form_%x", id[:8])
}

// origin assumes the host router has selected the site in production. Local
// requests additionally need an actual loopback peer or the CLI's LAN gate.
func (g *Guard) origin(r *http.Request) (string, error) {
	scheme := "https"
	if g.local {
		scheme = "http"
	} else if r.TLS == nil {
		return "", ErrInvalid
	}
	u, err := url.Parse(scheme + "://" + r.Host)
	if err != nil || u.Host != r.Host || u.Hostname() == "" || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
		return "", ErrInvalid
	}
	if g.local && !devaccess.LAN(r) {
		peer, _, err := net.SplitHostPort(r.RemoteAddr)
		ip := net.ParseIP(peer)
		host := net.ParseIP(u.Hostname())
		if err != nil || ip == nil || !ip.IsLoopback() || (u.Hostname() != "localhost" && (host == nil || !host.IsLoopback())) {
			return "", ErrInvalid
		}
	}
	return scheme + "://" + r.Host, nil
}

func (g *Guard) mac(kind, origin string, payload []byte) []byte {
	h := hmac.New(sha256.New, g.key[:])
	for _, part := range [][]byte{[]byte("flea-form-v1"), []byte(kind), []byte(g.site), []byte(origin), payload} {
		var size [8]byte
		binary.BigEndian.PutUint64(size[:], uint64(len(part)))
		h.Write(size[:])
		h.Write(part)
	}
	return h.Sum(nil)
}

func (g *Guard) read(r *http.Request, origin string) ([]byte, bool) {
	var value string
	n := 0
	for _, c := range r.Cookies() {
		if c.Name == g.cookieName() {
			value = c.Value
			n++
		}
	}
	if n != 1 || len(value) != 96 {
		return nil, false
	}
	b, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil || len(b) != 72 || !hmac.Equal(b[40:], g.mac("cookie", origin, b[:40])) {
		return nil, false
	}
	expiry := int64(binary.BigEndian.Uint64(b[:8]))
	now := g.now().Unix()
	if expiry <= now || expiry > now+int64(Lifetime.Seconds()) {
		return nil, false
	}
	return b[:40], true
}

// Token returns a token and, only when necessary, a cookie to commit after a
// successful page render. It never writes response headers itself.
func (g *Guard) Token(r *http.Request) (string, *http.Cookie, error) {
	origin, err := g.origin(r)
	if err != nil {
		return "", nil, err
	}
	if err := g.init(); err != nil {
		return "", nil, err
	}
	payload, ok := g.read(r, origin)
	var cookie *http.Cookie
	if !ok {
		payload = make([]byte, 40)
		expiry := g.now().Add(Lifetime).Truncate(time.Second)
		binary.BigEndian.PutUint64(payload[:8], uint64(expiry.Unix()))
		if _, err := rand.Read(payload[8:]); err != nil {
			return "", nil, fmt.Errorf("Form protection is unavailable.")
		}
		value := append(append([]byte(nil), payload...), g.mac("cookie", origin, payload)...)
		cookie = &http.Cookie{Name: g.cookieName(), Value: base64.RawURLEncoding.EncodeToString(value), Path: "/", HttpOnly: true, Secure: !g.local, SameSite: http.SameSiteLaxMode, MaxAge: int(Lifetime.Seconds()), Expires: expiry}
	}
	return base64.RawURLEncoding.EncodeToString(g.mac("form", origin, payload)), cookie, nil
}

// Verify accepts parsed POST fields only. The caller must first bound and parse
// the body. URL query parameters must never be merged into fields.
func (g *Guard) Verify(r *http.Request, fields url.Values) error {
	origin, err := g.origin(r)
	if err != nil || r.Method != http.MethodPost || len(r.Header.Values("Origin")) != 1 || r.Header.Get("Origin") != origin {
		return ErrInvalid
	}
	values := fields[Field]
	if len(values) != 1 || len(values[0]) != 43 {
		return ErrInvalid
	}
	if _, exists := r.URL.Query()[Field]; exists {
		return ErrInvalid
	}
	if err := g.init(); err != nil {
		return err
	}
	payload, ok := g.read(r, origin)
	if !ok {
		return ErrInvalid
	}
	got, err := base64.RawURLEncoding.DecodeString(values[0])
	if err != nil || !hmac.Equal(got, g.mac("form", origin, payload)) {
		return ErrInvalid
	}
	return nil
}
