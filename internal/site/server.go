package site

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

// Server publishes complete validated snapshots. Failed edits produce a 500
// rather than silently serving an apparently up-to-date site.
type Server struct {
	reloadMu sync.Mutex
	mu       sync.RWMutex
	project  *Project
	err      error
	revision uint64
}

func NewServer(p *Project) *Server { return &Server{project: p} }

func (s *Server) Reload(root string) error {
	s.reloadMu.Lock()
	defer s.reloadMu.Unlock()
	p, err := Load(root)
	s.mu.Lock()
	defer s.mu.Unlock()
	if err == nil {
		s.project = p
	}
	s.err = err
	s.revision++
	return err
}

func (s *Server) Watch(ctx context.Context, root string, logger *log.Logger) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	previous := ""
	s.mu.RLock()
	last := s.project.signature
	s.mu.RUnlock()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			next := fingerprint(root)
			if next == last {
				continue
			}
			last = next
			err := s.Reload(root)
			message := ""
			if err != nil {
				message = err.Error()
			}
			if message != previous && logger != nil {
				if message == "" {
					logger.Print("Site valid again; changes loaded")
				} else {
					logger.Printf("Site invalid: %s", message)
				}
			}
			previous = message
		}
	}
}

// Hash bytes, not just timestamps, so quick edits and atomic saves are noticed.
func fingerprint(root string) string {
	r, err := os.OpenRoot(root)
	if err != nil {
		return err.Error()
	}
	defer r.Close()
	h := sha256.New()
	for _, dir := range []string{"_", "html"} {
		err = fs.WalkDir(r.FS(), dir, func(name string, d fs.DirEntry, e error) error {
			if e != nil {
				return e
			}
			fmt.Fprintf(h, "%q:%v;", name, d.Type())
			if d.Type()&os.ModeSymlink != 0 {
				return fmt.Errorf("symlink: %s", name)
			}
			if d.IsDir() {
				return nil
			}
			if !d.Type().IsRegular() {
				return fmt.Errorf("unsupported file: %s", name)
			}
			b, e := readFile(r, name)
			if e != nil {
				return e
			}
			fmt.Fprintf(h, "%d:", len(b))
			h.Write(b)
			return nil
		})
		if err != nil {
			fmt.Fprint(h, err)
		}
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

const reloadPath = "/_nanoweb/reload"

func reloadScript(revision uint64) []byte {
	return []byte(fmt.Sprintf(`<script>(()=>{const version=%q;setInterval(()=>fetch('/_nanoweb/reload',{cache:'no-store'}).then(r=>r.text()).then(v=>{if(v!==version)location.reload()}).catch(()=>{}),700)})();</script>`, fmt.Sprint(revision)))
}

func withReload(body []byte, revision uint64) []byte {
	script := reloadScript(revision)
	if end := bytes.LastIndex(bytes.ToLower(body), []byte("</body>")); end >= 0 {
		result := make([]byte, 0, len(body)+len(script))
		result = append(result, body[:end]...)
		result = append(result, script...)
		return append(result, body[end:]...)
	}
	return append(body, script...)
}

func safeURL(p string) bool {
	if p == "" || p[0] != '/' || strings.ContainsAny(p, "\\\x00%") || path.Clean(p) != p {
		return false
	}
	for _, part := range strings.Split(p, "/") {
		if strings.HasPrefix(part, "_") || strings.HasPrefix(part, ".") || strings.Contains(strings.ToLower(part), ".html") {
			return false
		}
	}
	return true
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
	w.Header().Set("Cache-Control", "no-store")
	if r.Method != "GET" && r.Method != "HEAD" {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.mu.RLock()
	p, loadErr, revision := s.project, s.err, s.revision
	s.mu.RUnlock()
	if r.URL.Path == reloadPath && r.URL.RawPath == "" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if r.Method != "HEAD" {
			fmt.Fprint(w, revision)
		}
		return
	}
	if !safeURL(r.URL.Path) {
		serveReloadError(w, r, p, 404, revision)
		return
	}
	if loadErr != nil {
		serveReloadError(w, r, p, 500, revision)
		return
	}
	if t, ok := p.Pages[r.URL.Path]; ok {
		body, err := p.render(t, r.URL.Path)
		if err != nil {
			serveReloadError(w, r, p, 500, revision)
			return
		}
		writeHTML(w, r, 200, withReload(body, revision))
		return
	}
	if data, ok := p.AssetData[r.URL.Path]; ok {
		http.ServeContent(w, r, path.Base(r.URL.Path), time.Time{}, bytes.NewReader(data))
		return
	}
	serveReloadError(w, r, p, 404, revision)
}

func serveReloadError(w http.ResponseWriter, r *http.Request, p *Project, code int, revision uint64) {
	body := []byte("<!doctype html><title>" + http.StatusText(code) + "</title><h1>" + http.StatusText(code) + "</h1>")
	if p != nil {
		if t, ok := p.Errors[code]; ok {
			if b, e := p.render(t, fmt.Sprintf("/%d", code)); e == nil {
				body = b
			}
		}
	}
	writeHTML(w, r, code, withReload(body, revision))
}

func writeHTML(w http.ResponseWriter, r *http.Request, code int, body []byte) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Length", fmt.Sprint(len(body)))
	w.WriteHeader(code)
	if r.Method != "HEAD" {
		_, _ = w.Write(body)
	}
}
