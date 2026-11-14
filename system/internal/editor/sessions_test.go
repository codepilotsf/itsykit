package editor

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestPersistentRollingSessions(t *testing.T) {
	h, mail := fixture(t)
	now := time.Now().UTC().Truncate(time.Second)
	h.now = func() time.Time { return now }
	h.options.SessionIdleDays = 2
	c := login(t, h, mail)
	otherDevice := login(t, h, mail)
	data, err := os.ReadFile(h.options.SessionPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), c.Value) || strings.Contains(string(data), otherDevice.Value) {
		t.Fatal("raw token persisted")
	}
	info, _ := os.Stat(h.options.SessionPath)
	if info.Mode().Perm() != 0600 {
		t.Fatal(info.Mode())
	}
	// A new handler represents either a process restart or another release.
	restarted := New(h.options)
	restarted.now = func() time.Time { return now }
	now = now.Add(36 * time.Hour)
	w := request(restarted, "GET", "/admin", "example.test", nil, c)
	if w.Code != 200 || !strings.Contains(w.Body.String(), "Blog is not enabled") {
		t.Fatal(w.Code, w.Body.String())
	}
	renewed := cookie(t, w, h.cookieName("session"))
	if renewed.MaxAge != 2*86400 || !renewed.Expires.Equal(now.Add(48*time.Hour)) {
		t.Fatal(renewed)
	}
	// The active device outlives its original expiry; the other one does not.
	now = now.Add(24 * time.Hour)
	restarted = New(h.options)
	restarted.now = func() time.Time { return now }
	w = request(restarted, "GET", "/admin", "example.test", nil, c)
	if w.Code != 200 || !strings.Contains(w.Body.String(), "Blog is not enabled") {
		t.Fatal("rolling session expired early")
	}
	w = request(restarted, "GET", "/admin", "example.test", nil, otherDevice)
	if strings.Contains(w.Body.String(), "Blog is not enabled") {
		t.Fatal("idle device was renewed by another device")
	}
	csrf := restarted.sessions[key(c.Value)].csrf
	w = request(restarted, "POST", "/admin/logout", "example.test", url.Values{"csrf": {csrf}}, c)
	if w.Code != 303 {
		t.Fatal(w.Code, w.Body.String())
	}
	// Even the original handler must read revocation from disk.
	w = request(h, "GET", "/admin", "example.test", nil, c)
	if strings.Contains(w.Body.String(), "Blog is not enabled") {
		t.Fatal("stale handler resurrected logout")
	}
}

func TestSessionOwnerRemovalAndShorterIdleWindow(t *testing.T) {
	for _, mode := range []string{"removed", "shorter"} {
		t.Run(mode, func(t *testing.T) {
			h, mail := fixture(t)
			c := login(t, h, mail)
			o := h.options
			if mode == "removed" {
				o.Emails = nil
				if err := PruneSessions(o); err != nil {
					t.Fatal(err)
				}
				o.Emails = h.options.Emails // Re-adding must not revive old tokens.
			} else {
				o.SessionIdleDays = 1
			}
			next := New(o)
			future := time.Now().Add(25 * time.Hour)
			next.now = func() time.Time { return future }
			w := request(next, "GET", "/admin", "example.test", nil, c)
			if strings.Contains(w.Body.String(), "Blog is not enabled") {
				t.Fatal("revoked/expired session survived")
			}
		})
	}
}

func TestSessionStorageFailsClosed(t *testing.T) {
	for _, mode := range []string{"corrupt", "oversize", "symlink", "hardlink", "permissions", "directory", "write"} {
		t.Run(mode, func(t *testing.T) {
			h, mail := fixture(t)
			c := login(t, h, mail)
			name := h.options.SessionPath
			original, _ := os.ReadFile(name)
			switch mode {
			case "corrupt":
				os.WriteFile(name, []byte(`{broken`), 0600)
			case "oversize":
				os.WriteFile(name, []byte(strings.Repeat(" ", (2<<20)+1)), 0600)
			case "symlink", "hardlink":
				other := filepath.Join(t.TempDir(), "secret")
				if err := os.WriteFile(other, original, 0600); err != nil {
					t.Fatal(err)
				}
				if err := os.Remove(name); err != nil {
					t.Fatal(err)
				}
				link := os.Symlink
				if mode == "hardlink" {
					link = os.Link
				}
				if err := link(other, name); err != nil {
					t.Fatal(err)
				}
			case "permissions":
				os.Chmod(name, 0644)
			case "directory":
				os.Remove(name)
				os.Mkdir(name, 0700)
			case "write":
				parent := filepath.Dir(name)
				os.Chmod(parent, 0500)
				t.Cleanup(func() { os.Chmod(parent, 0700) })
			}
			w := request(h, "GET", "/admin", "example.test", nil, c)
			if w.Code != 503 || w.Code == 303 || len(w.Result().Cookies()) != 0 {
				t.Fatal(mode, w.Code, w.Body.String())
			}
		})
	}
}

func TestActivityRequiresSessionOriginAndCSRF(t *testing.T) {
	h, mail := fixture(t)
	c := login(t, h, mail)
	initial := h.sessions[key(c.Value)]
	now := initial.lastActive.Add(time.Hour)
	h.now = func() time.Time { return now }
	for _, csrf := range []string{"wrong", initial.csrf} {
		w := request(h, "POST", "/admin/activity", "example.test", url.Values{"csrf": {csrf}}, c)
		if csrf == "wrong" {
			if w.Code != 403 || !h.sessions[key(c.Value)].expires.Equal(initial.expires) {
				t.Fatal("invalid CSRF renewed session")
			}
		} else if w.Code != 204 || !h.sessions[key(c.Value)].expires.Equal(now.Add(h.sessionDuration())) {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	r := httptest.NewRequest("POST", "https://example.test/admin/activity", strings.NewReader(url.Values{"csrf": {initial.csrf}}.Encode()))
	r.AddCookie(c)
	r.TLS = &tls.ConnectionState{}
	r.Header.Set("Origin", "https://other.test")
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusForbidden {
		t.Fatal(w.Code)
	}
}

func TestSessionCapacityAndExpiredPruning(t *testing.T) {
	h, _ := fixture(t)
	now := time.Now()
	values := map[string]savedSession{}
	for i := 0; i < maxState; i++ {
		token, _ := token()
		values[token] = savedSession{"owner@example.test", "example.test", strings.Repeat("a", 64), now.Add(-48 * time.Hour), now.Add(-time.Hour)}
	}
	data, _ := json.Marshal(values)
	if err := os.WriteFile(h.options.SessionPath, data, 0600); err != nil {
		t.Fatal(err)
	}
	if err := PruneSessions(h.options); err != nil {
		t.Fatal(err)
	}
	data, _ = os.ReadFile(h.options.SessionPath)
	if string(data) != "{}\n" {
		t.Fatal("expired sessions not pruned")
	}
}

func TestConcurrentSessionRenewalAndLogout(t *testing.T) {
	h, mail := fixture(t)
	c := login(t, h, mail)
	csrf := h.sessions[key(c.Value)].csrf
	var group sync.WaitGroup
	for i := 0; i < 12; i++ {
		group.Go(func() { request(h, "POST", "/admin/activity", "example.test", url.Values{"csrf": {csrf}}, c) })
	}
	group.Go(func() { request(h, "POST", "/admin/logout", "example.test", url.Values{"csrf": {csrf}}, c) })
	group.Wait()
	next := New(h.options)
	w := request(next, "GET", "/admin", "example.test", nil, c)
	if strings.Contains(w.Body.String(), "Blog is not enabled") {
		t.Fatal("concurrent renewal restored logged-out session")
	}
}
