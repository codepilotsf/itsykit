package editor

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"flea/internal/mailer"
	"flea/internal/snips"
)

type Document struct {
	Store   *snips.Store
	Routes  []string
	Version string
}
type Update struct{ Version, Route, ID, Content, Title, Description string }
type Backend interface {
	Read() (Document, error)
	Save(Update) error
}
type Options struct {
	Emails  []string
	Name    string
	Sender  mailer.Sender
	Backend Backend
	Local   bool
	Logger  *log.Logger
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
	h := &Handler{options: o, allowed: map[string]bool{}, challenges: map[[32]byte]challenge{}, sessions: map[[32]byte]session{}, limits: map[string]bucket{}, now: time.Now}
	for _, email := range o.Emails {
		h.allowed[email] = true
	}
	return h
}
func emailMessage(email, code, host string) mailer.Message {
	return mailer.Message{To: email, Subject: "Your Flea CMS sign-in code", Text: fmt.Sprintf("Your sign-in code for %s is:\n\n%s\n\nIt expires in 10 minutes and can only be used once.\nIf you did not request it, ignore this email.", host, code)}
}
func IsPath(path string) bool { return path == "/admin" || strings.HasPrefix(path, "/admin/") }

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'self'; form-action 'self'; frame-ancestors 'none'; base-uri 'none'")
	if len(h.allowed) == 0 {
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
		if host != "localhost" && (ip == nil || !ip.IsLoopback()) {
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
		r.Body = http.MaxBytesReader(w, r.Body, 2<<20)
		if err := r.ParseForm(); err != nil {
			http.Error(w, "This form is too large or invalid.", 400)
			return
		}
	}
	if r.URL.Path == "/admin/style.css" && r.Method == "GET" {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		_, _ = w.Write([]byte(styles))
		return
	}
	if h.options.Sender == nil {
		message := "Sign-in emails are not set up for this site.\n\nAsk the person who manages this site's server to configure email delivery."
		if h.options.Local {
			message = "Email login needs valid Resend settings.\n\nAdd these to your site's .env file:\n\nRESEND_API_KEY=\"re_your_api_key\"\nFLEA_EMAIL_FROM=\"login@your-verified-domain.com\"\n\nUse your Resend API key and a sender address from a domain verified in Resend.\nKeep .env out of Git. Flea CMS does not load it automatically.\n\nStop dev, then run these commands from your site's folder:\n\nset -a\nsource .env\nset +a\nflea dev\n\nReturn to /admin to sign in."
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		http.Error(w, message, http.StatusServiceUnavailable)
		return
	}
	s, logged := h.current(r)
	if r.Method == "POST" {
		switch r.URL.Path {
		case "/admin/code":
			h.requestCode(w, r)
			return
		case "/admin/verify":
			h.verifyCode(w, r)
			return
		case "/admin/logout", "/admin/save":
			if !logged || !equal(s.csrf, r.FormValue("csrf")) {
				http.Error(w, "Your session expired. Please sign in again.", 403)
				return
			}
			if r.URL.Path == "/admin/logout" {
				h.mu.Lock()
				delete(h.sessions, key(h.getCookie(r, "session")))
				h.mu.Unlock()
				h.cookie(w, "session", "", -1)
				http.Redirect(w, r, "/admin", 303)
				return
			}
			u := Update{Version: r.FormValue("version"), Route: r.FormValue("route"), ID: r.FormValue("id"), Content: r.FormValue("content"), Title: r.FormValue("title"), Description: r.FormValue("description")}
			if err := h.options.Backend.Save(u); err != nil {
				if h.options.Logger != nil {
					h.options.Logger.Printf("Admin content save failed: %v", err)
				}
				draft := u.Content
				if u.ID == "" {
					draft = "Page title: " + u.Title + "\nDescription: " + u.Description
				}
				w.WriteHeader(409)
				h.page(w, view{Mode: "error", Message: "The page changed or could not be saved. Copy your text, reload the dashboard, and try again.", Draft: draft})
				return
			}
			http.Redirect(w, r, "/admin?saved=1", 303)
			return
		default:
			http.NotFound(w, r)
			return
		}
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
	doc, err := h.options.Backend.Read()
	if err != nil {
		http.Error(w, "Content is temporarily unavailable. Please try again shortly.", 503)
		return
	}
	v := view{Mode: "dashboard", Email: s.email, CSRF: s.csrf, Version: doc.Version}
	if r.URL.Query().Get("saved") == "1" {
		v.Message = "Your changes are live."
	}
	for _, route := range doc.Routes {
		p := doc.Store.Pages[route]
		if p == nil {
			p = &snips.Page{}
		}
		v.Pages = append(v.Pages, pageView{Route: route, Title: p.Title, Description: p.Description, Snips: p.Snips})
	}
	v.Shared = doc.Store.Shared.Snips
	h.page(w, v)
}

//go:embed templates/dashboard.html
var markup string

//go:embed templates/style.css
var styles string
var dashboard = template.Must(template.New("dashboard").Parse(markup))

type pageView struct {
	Route, Title, Description string
	Snips                     map[string]snips.Entry
}
type view struct {
	Mode, Name, Email, CSRF, Version, Message, Draft string
	Pages                                            []pageView
	Shared                                           map[string]snips.Entry
}

func (h *Handler) page(w http.ResponseWriter, v view) {
	v.Name = h.options.Name
	var b bytes.Buffer
	if err := dashboard.Execute(&b, v); err != nil {
		http.Error(w, "Could not show this page.", 500)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(b.Bytes())
}
