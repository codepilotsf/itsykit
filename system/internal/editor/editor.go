package editor

import (
	"fmt"
	"html"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"itsy/internal/devaccess"
	"itsy/internal/mailer"
	"itsy/internal/snips"
)

type Document struct {
	Store     *snips.Store
	Version   string
	BlogMount string
}
type Update struct {
	Version    string
	Blog       *BlogEdit
	BlogResult *BlogResult
	// BlogResult is filled under the publication lock.
}
type CoverUpload struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Alt  string `json:"alt"`
	Data []byte `json:"-"`
}
type Backend interface {
	Read() (Document, error)
	Save(Update) error
}
type Options struct {
	RenderEmail     func(string, map[string]string) (mailer.Message, error)
	SessionPath     string
	SessionIdleDays int
	Render          func(string, map[string]any) ([]byte, error)
	ReloadScript    func() string
	Emails          []string
	Name            string
	Sender          mailer.Sender
	Backend         Backend
	Local           bool
	Logger          *log.Logger
}
type Handler struct {
	options    Options
	allowed    map[string]bool
	mu         sync.Mutex
	challenges map[[32]byte]challenge
	sessions   map[[32]byte]session
	limits     map[string]bucket
	now        func() time.Time
}

func New(o Options) *Handler {
	if o.SessionIdleDays == 0 {
		o.SessionIdleDays = 30
	}
	h := &Handler{options: o, allowed: map[string]bool{}, challenges: map[[32]byte]challenge{}, sessions: map[[32]byte]session{}, limits: map[string]bucket{}, now: time.Now}
	for _, email := range o.Emails {
		h.allowed[email] = true
	}
	return h
}
func emailMessage(email, code, host string) mailer.Message {
	return mailer.Message{To: email, Subject: "Your ItsyWeb sign-in code", Text: fmt.Sprintf("Your sign-in code for %s is:\n\n%s\n\nIt expires in 10 minutes and can only be used once.\nIf you did not request it, ignore this email.", host, code), HTML: fmt.Sprintf(`<div style="font-family:Arial,sans-serif;font-size:16px;line-height:1.5"><p>Your sign-in code for %s is:</p><p style="font-family:monospace;font-size:36px;font-weight:700;letter-spacing:6px">%s</p><p>It expires in 10 minutes and can only be used once.</p><p>If you did not request it, ignore this email.</p></div>`, html.EscapeString(host), html.EscapeString(code))}
}
func IsPath(path string) bool { return path == "/admin" || strings.HasPrefix(path, "/admin/") }

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Referrer-Policy", "same-origin")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if serveEditorAsset(w, r) {
		return
	}
	switch r.URL.Path {
	case "/admin", "/admin/", "/admin/code", "/admin/verify", "/admin/logout", "/admin/activity", "/admin/blog", "/admin/blog/edit", "/admin/blog/save":
	default:
		http.NotFound(w, r)
		return
	}
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'self' https: 'unsafe-inline'; font-src 'self' https: data:; img-src 'self' https: data:; form-action 'self'; frame-ancestors 'none'; base-uri 'none'")
	if len(h.allowed) == 0 && !h.options.Local {
		http.NotFound(w, r)
		return
	}
	if !h.options.Local && r.TLS == nil {
		http.Error(w, "Admin access requires HTTPS.", 403)
		return
	}
	if h.options.Local {
		host := r.Host
		if name, _, e := net.SplitHostPort(host); e == nil {
			host = name
		}
		ip := net.ParseIP(host)
		if !devaccess.LAN(r) && host != "localhost" && (ip == nil || !ip.IsLoopback()) {
			http.Error(w, "Use the localhost address shown in your terminal.", 403)
			return
		}
	}
	if r.URL.RawPath != "" {
		http.NotFound(w, r)
		return
	}
	if r.Method != "GET" && r.Method != "POST" {
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "Method not allowed", 405)
		return
	}
	if r.Method == "POST" {
		pageUpload := r.URL.Path == "/admin/blog/save" && strings.HasPrefix(strings.ToLower(r.Header.Get("Content-Type")), "multipart/form-data")
		scheme := "https"
		if h.options.Local {
			scheme = "http"
		}
		// Require an exact browser origin, including login and logout. Never trust
		// forwarded headers to grant TLS or same-origin privileges.
		if r.Header.Get("Origin") != scheme+"://"+r.Host {
			http.Error(w, "Please submit this form from the admin page.", 403)
			return
		}
		limit := int64(2 << 20)
		if pageUpload {
			photoSession, authenticated, e := h.current(r)
			if e != nil {
				h.sessionError(w, e)
				return
			}
			if !authenticated || !equal(photoSession.csrf, r.Header.Get("X-CSRF-Token")) {
				http.Error(w, "Your session expired. Please sign in again.", 403)
				return
			}
			select {
			case mediaSlot <- struct{}{}:
				defer func() { <-mediaSlot }()
			default:
				http.Error(w, "Media processing is busy. Try again shortly.", 429)
				return
			}
			limit = 40 << 20
			// Only authenticated, CSRF-checked uploads get a longer bounded window.
			controller := http.NewResponseController(w)
			_ = controller.SetReadDeadline(time.Now().Add(3 * time.Minute))
			_ = controller.SetWriteDeadline(time.Now().Add(4 * time.Minute))
		}
		r.Body = http.MaxBytesReader(w, r.Body, limit)
		if pageUpload {
			err := r.ParseMultipartForm(1 << 20)
			if r.MultipartForm != nil {
				defer r.MultipartForm.RemoveAll()
			}
			if err != nil {
				http.Error(w, "This upload is invalid or exceeds 40 MiB total. Choose a JPEG or PNG up to 8 MiB.", 400)
				return
			}
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, "This form is too large or invalid.", 400)
			return
		}
	}
	if len(h.allowed) == 0 {
		w.WriteHeader(http.StatusServiceUnavailable)
		h.page(w, view{Mode: "setup", Instructions: "Add the first admin from your site's folder:\n\nitsy admin admin@example.com\n\nThe command sets up email if needed. Then return here to sign in."})
		return
	}
	s, logged, err := h.current(r)
	if err != nil {
		h.sessionError(w, err)
		return
	}
	if h.options.Sender == nil && (!logged || r.URL.Path == "/admin/code" || r.URL.Path == "/admin/verify") {
		message := "Sign-in emails are not set up for this site.\n\nAsk the person who manages this site's server to configure email delivery."
		if h.options.Local {
			message = "Email is not configured yet.\n\nRun this from your site's folder:\n\nitsy email\n\nThen return to /admin."
		}
		w.WriteHeader(http.StatusServiceUnavailable)
		h.page(w, view{Mode: "setup", Instructions: message})
		return
	}
	if r.Method == "POST" {
		switch r.URL.Path {
		case "/admin/code":
			h.requestCode(w, r)
			return
		case "/admin/verify":
			h.verifyCode(w, r)
			return
		case "/admin/logout", "/admin/activity", "/admin/blog/save":
			if !logged || !equal(s.csrf, r.FormValue("csrf")) {
				http.Error(w, "Your session expired. Please sign in again.", 403)
				return
			}
			if r.URL.Path == "/admin/logout" {
				h.mu.Lock()
				err := h.sessionTransaction(func(values map[[32]byte]session) error {
					delete(values, key(h.getCookie(r, "session")))
					return nil
				})
				h.mu.Unlock()
				if err != nil {
					h.sessionError(w, err)
					return
				}
				h.cookie(w, "session", "", -1)
				http.Redirect(w, r, "/admin", 303)
				return
			}
			if !h.renew(w, r) {
				return
			}
			if r.URL.Path == "/admin/activity" {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			if r.URL.Path == "/admin/blog/save" {
				h.saveBlog(w, r)
				return
			}
		default:
			http.NotFound(w, r)
			return
		}
	}
	if r.URL.Path == "/admin/blog" || r.URL.Path == "/admin/blog/edit" {
		if !logged {
			http.Redirect(w, r, "/admin", 303)
			return
		}
		if !h.renew(w, r) {
			return
		}
		h.showBlog(w, r, s)
		return
	}
	if r.URL.Path != "/admin" && r.URL.Path != "/admin/" {
		http.NotFound(w, r)
		return
	}
	if !logged {
		csrf, err := token()
		if err != nil {
			http.Error(w, "Please try again later.", 503)
			return
		}
		h.cookie(w, "login", csrf, int(codeLife.Seconds()))
		h.page(w, view{Mode: "login", CSRF: csrf})
		return
	}
	if !h.renew(w, r) {
		return
	}
	d, err := h.options.Backend.Read()
	if err != nil {
		http.Error(w, "Could not read site content.", 503)
		return
	}
	if d.BlogMount != "" {
		http.Redirect(w, r, "/admin/blog", http.StatusSeeOther)
		return
	}
	h.page(w, view{Mode: "setup", Message: "Blog is not enabled", Instructions: "Blog is not enabled. Website pages are edited in HTML.", Email: s.email, CSRF: s.csrf})
}

type view struct {
	Instructions               string
	Mode, Email, CSRF, Message string
}

func (h *Handler) page(w http.ResponseWriter, v view) {
	if h.options.Render == nil {
		http.Error(w, "Admin pages are not configured.", 500)
		return
	}
	data, err := h.options.Render(v.Mode, map[string]any{"name": h.options.Name, "email": v.Email, "csrf": v.CSRF, "message": v.Message, "instructions": v.Instructions})
	if err != nil {
		if h.options.Logger != nil {
			h.options.Logger.Printf("Admin page could not be rendered: %v", err)
		}
		http.Error(w, "Could not show this page. Ask your developer to check the admin templates.", 500)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if v.Mode == "verify" {
		data = appendCodeInput(w, data)
	}
	if h.options.Local && h.options.ReloadScript != nil {
		if script := h.options.ReloadScript(); script != "" {
			data = appendTrustedScript(w, data, script)
			w.Header().Set("Content-Security-Policy", w.Header().Get("Content-Security-Policy")+"; connect-src 'self'")
		}
	}
	_, _ = w.Write(data)
}
