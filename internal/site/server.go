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

	"nanoweb/internal/sitefiles"
)

// Server publishes complete validated snapshots. Failed edits produce a 500
// rather than silently serving an apparently up-to-date site.
type Server struct {
	// Serialize writers without holding the request lock during disk/network I/O.
	reloadMu   sync.Mutex
	mu         sync.RWMutex
	project    *Project
	err        error
	revision   uint64
	logger     *log.Logger
	production bool
}

func NewServer(p *Project, loggers ...*log.Logger) *Server {
	s := &Server{project: p}
	if len(loggers) > 0 {
		s.logger = loggers[0]
	}
	return s
}

func NewProductionServer(p *Project, logger *log.Logger) *Server {
	s := NewServer(p, logger)
	s.production = true
	return s
}

// Publish replaces a fully validated snapshot after a content save.
func (s *Server) Publish(p *Project) {
	s.reloadMu.Lock()
	defer s.reloadMu.Unlock()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.project, s.err = p, nil
	s.revision++
}

func (s *Server) Reload(root string) error {
	return s.reload(context.Background(), root, false)
}

func (s *Server) reload(ctx context.Context, root string, resolveIcons bool) error {
	if resolveIcons {
		unlock := LockContent()
		defer unlock()
	}
	s.reloadMu.Lock()
	defer s.reloadMu.Unlock()
	var p *Project
	var err error
	if resolveIcons {
		p, err = loadEditing(ctx, root)
	} else {
		p, err = Load(root)
	}
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
			err := s.reload(ctx, root, true)
			if err == nil {
				s.mu.RLock()
				project := s.project
				last = project.signature
				s.mu.RUnlock()
				if logger != nil {
					if project.SnipsUpdated {
						logger.Print("Snips updated in _/snips.json.")
					}
					for _, warning := range project.Warnings {
						logger.Printf("Warning: %s", warning)
					}
				}
			}
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
	err = fs.WalkDir(r.FS(), ".", func(name string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if !sitefiles.Included(name) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
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
	if s.logger != nil {
		recorded := &statusWriter{ResponseWriter: w}
		w = recorded
		defer func() {
			status := recorded.status
			if status == 0 {
				status = http.StatusOK
			}
			// Successful browser reload polls are internal housekeeping.
			if r.URL.Path == reloadPath && r.URL.RawPath == "" && status == http.StatusOK {
				return
			}
			route := r.URL.EscapedPath() // No query strings or control characters in logs.
			if status == http.StatusOK {
				s.logger.Printf("%s %s", r.Method, route)
			} else {
				s.logger.Printf("%d %s %s", status, r.Method, route)
			}
		}()
	}
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
	if !s.production && r.URL.Path == reloadPath && r.URL.RawPath == "" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if r.Method != "HEAD" {
			fmt.Fprint(w, revision)
		}
		return
	}
	if !safeURL(r.URL.Path) {
		serveReloadError(w, r, p, 404, revision, !s.production)
		return
	}
	if loadErr != nil {
		serveReloadError(w, r, p, 500, revision, !s.production)
		return
	}
	if t, ok := p.Pages[r.URL.Path]; ok {
		body, err := p.render(t, r.URL.Path)
		if err != nil {
			serveReloadError(w, r, p, 500, revision, !s.production)
			return
		}
		if !s.production {
			body = withReload(body, revision)
		}
		writeHTML(w, r, 200, body)
		return
	}
	if data, ok := p.AssetData[r.URL.Path]; ok {
		http.ServeContent(w, r, path.Base(r.URL.Path), time.Time{}, bytes.NewReader(data))
		return
	}
	serveReloadError(w, r, p, 404, revision, !s.production)
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	if status >= 100 && status < 200 {
		w.ResponseWriter.WriteHeader(status)
		return
	}
	if w.status != 0 {
		return
	}
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(body)
}

func (w *statusWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

func serveReloadError(w http.ResponseWriter, r *http.Request, p *Project, code int, revision uint64, reload bool) {
	body := []byte("<!doctype html><title>" + http.StatusText(code) + "</title><h1>" + http.StatusText(code) + "</h1>")
	if p != nil {
		if t, ok := p.Errors[code]; ok {
			if b, e := p.render(t, fmt.Sprintf("/%d", code)); e == nil {
				body = b
			}
		}
	}
	if reload {
		body = withReload(body, revision)
	}
	writeHTML(w, r, code, body)
}

func writeHTML(w http.ResponseWriter, r *http.Request, code int, body []byte) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Length", fmt.Sprint(len(body)))
	w.WriteHeader(code)
	if r.Method != "HEAD" {
		_, _ = w.Write(body)
	}
}
