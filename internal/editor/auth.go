// Package editor serves the small owner dashboard. It has no deployment powers.
package editor

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"strings"
	"time"
)

const codeLife = 10 * time.Minute
const sessionLife = 8 * time.Hour
const maxState = 2048

type challenge struct {
	email, host string
	digest      [32]byte
	expires     time.Time
	attempts    int
}
type session struct {
	email, host, csrf string
	expires           time.Time
}
type bucket struct {
	count   int
	expires time.Time
}

func token() (string, error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}
func newCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
func key(value string) [32]byte { return sha256.Sum256([]byte(value)) }
func equal(a, b string) bool {
	return len(a) == len(b) && subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
func authority(r *http.Request) string { return strings.ToLower(r.Host) }

func (h *Handler) cookieName(kind string) string {
	if h.options.Local {
		return "flea_dev_" + kind
	}
	return "__Host-flea-" + kind
}
func (h *Handler) cookie(w http.ResponseWriter, kind, value string, age int) {
	http.SetCookie(w, &http.Cookie{Name: h.cookieName(kind), Value: value, Path: "/", HttpOnly: true, Secure: !h.options.Local, SameSite: http.SameSiteStrictMode, MaxAge: age})
}
func (h *Handler) getCookie(r *http.Request, kind string) string {
	c, err := r.Cookie(h.cookieName(kind))
	if err != nil || len(c.Value) != 64 {
		return ""
	}
	return c.Value
}

// Called under mu. Bounded, expiring state prevents untrusted email/IP values
// from growing memory indefinitely. Forwarded headers are deliberately ignored.
func (h *Handler) clean(now time.Time) {
	for k, v := range h.challenges {
		if !now.Before(v.expires) {
			delete(h.challenges, k)
		}
	}
	for k, v := range h.sessions {
		if !now.Before(v.expires) {
			delete(h.sessions, k)
		}
	}
	for k, v := range h.limits {
		if !now.Before(v.expires) {
			delete(h.limits, k)
		}
	}
}
func (h *Handler) allow(k string, count int, period time.Duration, now time.Time) bool {
	b, ok := h.limits[k]
	if !ok {
		if len(h.limits) >= maxState {
			return false
		}
		b.expires = now.Add(period)
	}
	if b.count >= count {
		return false
	}
	b.count++
	h.limits[k] = b
	return true
}
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
func (h *Handler) current(r *http.Request) (session, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clean(h.now())
	s, ok := h.sessions[key(h.getCookie(r, "session"))]
	return s, ok && s.host == authority(r) && h.allowed[s.email]
}

func (h *Handler) requestCode(w http.ResponseWriter, r *http.Request) {
	login := h.getCookie(r, "login")
	if login == "" || !equal(login, r.FormValue("csrf")) {
		http.Error(w, "Please reload the login page.", 403)
		return
	}
	email := strings.ToLower(strings.TrimSpace(r.FormValue("email")))
	if len(email) > 254 {
		email = ""
	}
	code, err := newCode()
	if err != nil {
		http.Error(w, "Please try again later.", 503)
		return
	}
	now := h.now()
	h.mu.Lock()
	h.clean(now)
	permitted := h.allow("ip:"+clientIP(r), 10, 15*time.Minute, now) && h.allow("global", 40, 15*time.Minute, now)
	permitted = permitted && h.allow("email:"+email, 3, 15*time.Minute, now)
	if !permitted || len(h.challenges) >= maxState {
		h.mu.Unlock()
		http.Error(w, "Too many requests. Please try again in 15 minutes.", 429)
		return
	}
	h.challenges[key(login)] = challenge{email: email, host: authority(r), digest: key(login + code), expires: now.Add(codeLife)}
	h.mu.Unlock()
	if h.allowed[email] {
		message := emailMessage(email, code, authority(r))
		// Respond independently of provider timing so the allowlist cannot be
		// inferred from a network round trip. The rate limiter bounds workers.
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := h.options.Sender.Send(ctx, message); err != nil {
				h.mu.Lock()
				if c, ok := h.challenges[key(login)]; ok && c.digest == key(login+code) {
					delete(h.challenges, key(login))
				}
				h.mu.Unlock()
				if h.options.Logger != nil {
					h.options.Logger.Print("Admin login email could not be delivered. Check the email provider configuration.")
				}
			}
		}()
	}
	// The same response is used for unlisted addresses and provider failures.
	h.page(w, view{Mode: "verify", CSRF: login, Message: "If this address has access, a code will arrive shortly. It expires in 10 minutes."})
}

func (h *Handler) verifyCode(w http.ResponseWriter, r *http.Request) {
	login := h.getCookie(r, "login")
	if login == "" || !equal(login, r.FormValue("csrf")) {
		http.Error(w, "Please reload the login page.", 403)
		return
	}
	t, err := token()
	if err != nil {
		http.Error(w, "Please try again later.", 503)
		return
	}
	csrf, err := token()
	if err != nil {
		http.Error(w, "Please try again later.", 503)
		return
	}
	now := h.now()
	h.mu.Lock()
	h.clean(now)
	c, ok := h.challenges[key(login)]
	got := key(login + r.FormValue("code"))
	valid := ok && c.host == authority(r) && c.attempts < 5 && h.allowed[c.email] && subtle.ConstantTimeCompare(c.digest[:], got[:]) == 1
	if ok {
		c.attempts++
		h.challenges[key(login)] = c
		if c.attempts >= 5 || valid {
			delete(h.challenges, key(login))
		}
	}
	if valid && len(h.sessions) < maxState {
		// Rotate rather than reuse a browser-supplied session identifier.
		delete(h.sessions, key(h.getCookie(r, "session")))
		h.sessions[key(t)] = session{email: c.email, host: c.host, csrf: csrf, expires: now.Add(sessionLife)}
	} else {
		valid = false
	}
	h.mu.Unlock()
	if !valid {
		w.WriteHeader(401)
		h.page(w, view{Mode: "verify", CSRF: login, Message: "That code is invalid or expired. Try again, or request a new code."})
		return
	}
	h.cookie(w, "session", t, int(sessionLife.Seconds()))
	h.cookie(w, "login", "", -1)
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
