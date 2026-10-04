package site

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"maps"
	"net/http"
	"sync"
	"time"
)

type flashEntry struct {
	binding, app, route string
	values              map[string]string
	expires             time.Time
	busy                bool
}
type flashStore struct {
	sync.Mutex
	entries map[string]*flashEntry
}

func (s *Server) flashCookie(name, value string, age int) *http.Cookie {
	id := sha256.Sum256([]byte(s.forms.SiteID()))
	prefix := "__Host-flea-flash-"
	if !s.production {
		prefix = fmt.Sprintf("flea_dev_flash_%x_", id[:8])
	}
	return &http.Cookie{Name: prefix + name, Value: value, Path: "/", HttpOnly: true, Secure: s.production, SameSite: http.SameSiteLaxMode, MaxAge: age}
}

func (f *flashStore) prune(now time.Time) {
	for id, entry := range f.entries {
		if !now.Before(entry.expires) && !entry.busy {
			delete(f.entries, id)
		}
	}
}

func (s *Server) storeFlash(w http.ResponseWriter, r *http.Request, name, route string, values map[string]string) {
	if len(values) == 0 {
		return
	}
	binding, ok := s.forms.Binding(r)
	if !ok {
		return
	}
	var random [32]byte
	if _, err := rand.Read(random[:]); err != nil {
		return
	}
	id := hex.EncodeToString(random[:])
	f := &s.apps.flashes
	f.Lock()
	defer f.Unlock()
	f.prune(time.Now())
	if len(f.entries) >= 256 {
		return
	} // Flash is optional; never turn accepted mail into a retry.
	if f.entries == nil {
		f.entries = map[string]*flashEntry{}
	}
	f.entries[id] = &flashEntry{binding: binding, app: name, route: route, values: maps.Clone(values), expires: time.Now().Add(5 * time.Minute)}
	http.SetCookie(w, s.flashCookie(name, id, 300))
}

// Reserve until successful rendering. HEAD, assets and other pages never claim
// flash. A concurrent GET cannot consume the same message twice.
func (s *Server) takeFlash(r *http.Request, name string) (map[string]string, func(http.ResponseWriter, bool)) {
	noop := func(http.ResponseWriter, bool) {}
	if r.Method != "GET" {
		return nil, noop
	}
	binding, ok := s.forms.Binding(r)
	if !ok {
		return nil, noop
	}
	wanted := s.flashCookie(name, "", 0).Name
	id := ""
	count := 0
	for _, c := range r.Cookies() {
		if c.Name == wanted {
			id = c.Value
			count++
		}
	}
	if count != 1 || len(id) != 64 {
		return nil, noop
	}
	f := &s.apps.flashes
	f.Lock()
	defer f.Unlock()
	f.prune(time.Now())
	entry := f.entries[id]
	if entry == nil || entry.busy || entry.app != name || entry.binding != binding || entry.route != r.URL.Path {
		return nil, noop
	}
	entry.busy = true
	return entry.values, func(w http.ResponseWriter, success bool) {
		f.Lock()
		defer f.Unlock()
		if success {
			delete(f.entries, id)
			http.SetCookie(w, s.flashCookie(name, "", -1))
		} else {
			entry.busy = false
		}
	}
}
