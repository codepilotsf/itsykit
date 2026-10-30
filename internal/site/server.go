package site

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"flea/internal/formguard"
	"flea/internal/sitefiles"
	"flea/internal/viewer"
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
	forms      *formguard.Guard
	apps       appState
}

func NewServer(p *Project, loggers ...*log.Logger) *Server {
	s := &Server{project: p, forms: formguard.New(p.Root, true)}
	if len(loggers) > 0 {
		s.logger = loggers[0]
	}
	return s
}

func NewProductionServer(p *Project, logger *log.Logger) *Server {
	s := NewServer(p, logger)
	s.production = true
	s.forms = formguard.New(p.Root, false)
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
		f, e := openSiteFile(r, name, 100<<20)
		if e != nil {
			return e
		}
		defer f.Close()
		info, e := f.Stat()
		if e != nil {
			return e
		}
		fmt.Fprintf(h, "%d:", info.Size())
		n, e := io.Copy(h, io.NewSectionReader(f, 0, info.Size()))
		if e == nil && (n != info.Size() || !(fileAsset{f, info}).unchanged()) {
			return fmt.Errorf("file changed while hashing: %s", name)
		}
		return e
	})
	if err != nil {
		fmt.Fprint(h, err)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

const reloadPath = "/_flea/reload"

func reloadScript(revision uint64) []byte {
	return []byte(fmt.Sprintf(`<script>(()=>{const version=%q;setInterval(()=>fetch('/_flea/reload',{cache:'no-store'}).then(r=>r.text()).then(v=>{if(v!==version)location.reload()}).catch(()=>{}),700)})();</script>`, fmt.Sprint(revision)))
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
	s.mu.RLock()
	p, loadErr, revision := s.project, s.err, s.revision
	s.mu.RUnlock()
	if p != nil && p.syncOnly {
		http.Error(w, "Site synchronization is incomplete.", http.StatusServiceUnavailable)
		return
	}
	if p != nil && loadErr == nil && safeURL(r.URL.Path) {
		if mount, _, ok := s.appFor(p, r.URL.Path); ok {
			s.serveInstalledApp(w, r, mount)
			return
		}
	}
	if r.Method != "GET" && r.Method != "HEAD" {
		w.Header().Set("Allow", "GET, HEAD")
		s.serveError(w, r, p, http.StatusMethodNotAllowed, revision, nil)
		return
	}
	if !s.production && r.URL.Path == reloadPath && r.URL.RawPath == "" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if r.Method != "HEAD" {
			fmt.Fprint(w, revision)
		}
		return
	}
	if !safeURL(r.URL.Path) {
		s.serveError(w, r, p, 404, revision, nil)
		return
	}
	if loadErr != nil {
		s.serveError(w, r, p, 500, revision, loadErr)
		return
	}
	if p.serveSEO(w, r) {
		return
	}
	if t, ok := p.Pages[r.URL.Path]; ok {
		if p.seoExcluded[r.URL.Path] {
			w.Header().Add("X-Robots-Tag", "noindex")
		}
		decorate, owner := r.Context().Value(ownerRenderKey{}).(ownerTransform)
		var admin map[string]any
		if identity, authenticated := viewer.From(r.Context()); owner && authenticated {
			admin = map[string]any{"email": identity.Email, "csrf_token": identity.CSRF, "version": p.EditingVersion(), "revision": ""}
			if !s.production {
				admin["revision"] = fmt.Sprint(revision)
			}
		}
		var inputs *renderInputs
		var commitForm func(http.ResponseWriter)
		if p.csrfTemplates[t] {
			inputs, commitForm = s.formRender(r)
		}
		body, err := p.renderRequest(t, r.URL.Path, admin, owner, inputs)
		if err != nil {
			s.serveError(w, r, p, 500, revision, err)
			return
		}
		if owner {
			if !s.production {
				w.Header().Set("X-Flea-Dev-Revision", fmt.Sprint(revision))
			}
			var ok bool
			body, ok = decorate(body, p.Content, p.EditingVersion())
			if !ok {
				return
			}
		} else if !s.production {
			body = withReload(body, revision)
		}
		if commitForm != nil {
			commitForm(w)
		}
		body = p.canonicalResponse(w, r.URL.Path, body)
		writeHTML(w, r, 200, body)
		return
	}
	if p.serveVideo(w, r) {
		return
	}
	if data, ok := p.AssetData[r.URL.Path]; ok {
		http.ServeContent(w, r, path.Base(r.URL.Path), time.Time{}, bytes.NewReader(data))
		return
	}
	if p.serveAssetFile(w, r) {
		return
	}
	s.serveError(w, r, p, 404, revision, nil)
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

func (s *Server) serveError(w http.ResponseWriter, r *http.Request, p *Project, code int, revision uint64, cause error) {
	if cause != nil && s.logger != nil {
		s.logger.Printf("Page error: %s %s: %q", r.Method, r.URL.EscapedPath(), cause.Error())
	}
	if !s.production && code != http.StatusNotFound && localDiagnosticRequest(r) {
		serveDevError(w, r, p, code, revision, cause)
		return
	}
	body := []byte("<!doctype html><title>" + http.StatusText(code) + "</title><h1>" + http.StatusText(code) + "</h1>")
	if p != nil {
		if t, ok := p.Errors[code]; ok {
			if b, e := p.render(t, fmt.Sprintf("/%d", code)); e == nil {
				body = b
			} else if s.logger != nil {
				s.logger.Printf("Error page %d failed: %q", code, e.Error())
			}
		}
	}
	if !s.production {
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
