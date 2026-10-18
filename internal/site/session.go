package site

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"
	"time"

	"flea/internal/appapi"
)

// App-owned session bridge. Flea owns the cookie name and attributes and keeps
// the opaque token out of templates, logs, URLs and public JSON. The token only
// ever travels to the app as trusted transport metadata.

// sessionCookieName isolates sessions by app and by site. Production uses the
// __Host- prefix (Secure, Path=/, no Domain). Development names include a site
// identifier so sites sharing a loopback host never collide.
func (s *Server) sessionCookieName(app string) string {
	if s.production {
		return "__Host-flea-session-" + app
	}
	id := sha256.Sum256([]byte(s.forms.SiteID()))
	return fmt.Sprintf("flea_dev_session_%x_%s", id[:8], app)
}

// sessionToken returns the browser's opaque token only when exactly one
// well-formed cookie for this site and app is present. Duplicate, malformed,
// cross-app and cross-site cookies yield no token.
//
// The raw Cookie headers are inspected rather than r.Cookies(), which silently
// drops malformed fragments; a valid cookie accompanied by a malformed
// duplicate of the same name must fail closed.
func (s *Server) sessionToken(r *http.Request, app string) string {
	name := s.sessionCookieName(app)
	token := ""
	count := 0
	for _, header := range r.Header.Values("Cookie") {
		for _, part := range strings.Split(header, ";") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			eq := strings.IndexByte(part, '=')
			if eq < 0 {
				// A bare fragment naming our cookie is a malformed duplicate.
				if part == name {
					count++
				}
				continue
			}
			key := strings.TrimSpace(part[:eq])
			if key != name {
				continue
			}
			token = part[eq+1:]
			count++
		}
	}
	if count != 1 || !appapi.ValidSessionToken(token) {
		return ""
	}
	return token
}

func (s *Server) sessionCookie(app, token string, maxAge int) *http.Cookie {
	cookie := &http.Cookie{Name: s.sessionCookieName(app), Value: token, Path: "/", HttpOnly: true, Secure: s.production, SameSite: http.SameSiteLaxMode, MaxAge: maxAge}
	if maxAge > 0 {
		cookie.Expires = time.Now().Add(time.Duration(maxAge) * time.Second)
	}
	return cookie
}

// applySession writes at most one Set-Cookie for the calling app. It must run
// before the response status or redirect is written. Concurrent responses for
// the same app carry one session cookie each; ordinary browser cookie rules
// decide the winner, and the last committed response takes effect.
func (s *Server) applySession(w http.ResponseWriter, app string, directive appapi.SessionDirective) {
	if !directive.Present {
		return
	}
	if directive.Clear {
		http.SetCookie(w, s.sessionCookie(app, "", -1))
		return
	}
	http.SetCookie(w, s.sessionCookie(app, directive.Token, directive.MaxAge))
}
