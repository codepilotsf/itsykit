package editor

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"flea/internal/fleas"
	"flea/internal/mailer"
)

type Document struct {
	Store   *fleas.Store
	Routes  []string
	Version string
}
type Update struct{ Version, Route, ID, Content, Title, Description string }
type Backend interface {
	Read() (Document, error)
	Save(Update) error
}
type Options struct {
	SessionPath     string
	SessionIdleDays int
	Render          func(string, map[string]any) ([]byte, error)
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
	return mailer.Message{To: email, Subject: "Your FleaCMS sign-in code", Text: fmt.Sprintf("Your sign-in code for %s is:\n\n%s\n\nIt expires in 10 minutes and can only be used once.\nIf you did not request it, ignore this email.", host, code)}
}
func IsPath(path string) bool { return path == "/admin" || strings.HasPrefix(path, "/admin/") }

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Referrer-Policy", "same-origin")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'self' https: 'unsafe-inline'; font-src 'self' https: data:; img-src 'self' https: data:; form-action 'self'; frame-ancestors 'none'; base-uri 'none'"+activityPolicy())
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
	if len(h.allowed) == 0 {
		w.WriteHeader(http.StatusServiceUnavailable)
		h.page(w, view{Mode: "setup", Instructions: "Add the first site owner from your site's folder:\n\nflea admin owner@example.com\n\nThe command sets up email if needed. Then return here to sign in."})
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
			message = "Email is not configured yet.\n\nRun this from your site's folder:\n\nflea email\n\nThen return to /admin."
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
		case "/admin/logout", "/admin/save", "/admin/activity":
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
			target := "/admin?saved=1&page=" + url.QueryEscape(r.FormValue("page"))
			if r.FormValue("shared") == "1" {
				target = "/admin?saved=1&shared=1"
			}
			http.Redirect(w, r, target, 303)
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
	v.Selected = r.URL.Query().Get("page")
	v.SharedSelected = r.URL.Query().Get("shared") == "1"
	if v.Selected == "" && len(doc.Routes) > 0 {
		v.Selected = doc.Routes[0]
	}
	found := false
	if r.URL.Query().Get("saved") == "1" {
		v.Message = "Your changes are live."
	}
	for _, route := range doc.Routes {
		if route == v.Selected {
			found = true
		}
		p := doc.Store.Pages[route]
		if p == nil {
			p = &fleas.Page{}
		}
		v.Pages = append(v.Pages, pageView{Route: route, Title: p.Title, Description: p.Description, Fleas: p.Fleas})
	}
	if !v.SharedSelected && !found && v.Selected != "" {
		http.NotFound(w, r)
		return
	}
	if !h.renew(w, r) {
		return
	}
	v.Shared = doc.Store.Shared.Fleas
	h.page(w, v)
}

type pageView struct {
	Route, Title, Description string
	Fleas                     map[string]fleas.Entry
}
type view struct {
	Selected                                         string
	SharedSelected                                   bool
	Instructions                                     string
	Mode, Name, Email, CSRF, Version, Message, Draft string
	Pages                                            []pageView
	Shared                                           map[string]fleas.Entry
}

func (h *Handler) page(w http.ResponseWriter, v view) {
	if h.options.Render == nil {
		http.Error(w, "Admin pages are not configured.", 500)
		return
	}
	entries := func(fleas map[string]fleas.Entry) []any {
		keys := make([]string, 0, len(fleas))
		for id := range fleas {
			keys = append(keys, id)
		}
		sort.Strings(keys)
		var result []any
		for _, id := range keys {
			entry := fleas[id]
			label := entry.Label
			if label == "" {
				label = "Content"
			}
			result = append(result, map[string]any{"id": id, "label": label, "content": entry.Content})
		}
		return result
	}
	var pages []any
	for _, page := range v.Pages {
		label := page.Route
		if label == "/" {
			label = "Homepage"
		}
		pages = append(pages, map[string]any{"route": page.Route, "label": label, "selected": !v.SharedSelected && page.Route == v.Selected, "title": page.Title, "description": page.Description, "fleas": entries(page.Fleas)})
	}
	data, err := h.options.Render(v.Mode, map[string]any{"name": h.options.Name, "email": v.Email, "csrf": v.CSRF, "version": v.Version, "message": v.Message, "instructions": v.Instructions, "draft": v.Draft, "pages": pages, "shared": entries(v.Shared), "sharedSelected": v.SharedSelected})
	if err != nil {
		if h.options.Logger != nil {
			h.options.Logger.Printf("Admin page could not be rendered: %v", err)
		}
		http.Error(w, "Could not show this page. Ask your developer to check the admin templates.", 500)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if v.Mode == "dashboard" {
		data = withActivity(data)
	}
	_, _ = w.Write(data)
}
