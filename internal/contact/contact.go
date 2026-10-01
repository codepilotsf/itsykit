// Package contact serves email-only forms. Each Handler is bound to one site's
// configuration; requests never select recipients, storage paths or site IDs.
package contact

import (
	"context"
	"fmt"
	"log"
	"mime"
	"net"
	"net/http"
	"net/mail"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"flea/internal/devaccess"
	"flea/internal/mailer"
)

const Path = "/_flea/contact"

type Config struct {
	To          string `json:"to"`
	SuccessPath string `json:"success_path,omitempty"`
}

func email(value string) bool {
	a, err := mail.ParseAddress(value)
	return len(value) <= 254 && !strings.ContainsAny(value, "\r\n") && err == nil && a.Address == value
}

func (c *Config) Validate() error {
	if c.SuccessPath != "" && (!strings.HasPrefix(c.SuccessPath, "/") || strings.HasPrefix(c.SuccessPath, "//") || strings.ContainsAny(c.SuccessPath, "\\?#\r\n\x00") || strings.Contains(c.SuccessPath, "..")) {
		return fmt.Errorf("contact.success_path must be a local absolute path")
	}
	c.To = strings.TrimSpace(c.To)
	if c.To != "" && !email(c.To) {
		return fmt.Errorf("contact.to must be one email address without a display name")
	}
	return nil
}

type bucket struct {
	count int
	until time.Time
}

// Budget is shared only by contact handlers, not login mail. No work queue or
// disk spool exists: exhausted capacity returns 429, rather than accumulating work.
type Budget struct {
	mu     sync.Mutex
	window bucket
	active int
}

func (b *Budget) acquire(now time.Time) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !now.Before(b.window.until) {
		b.window = bucket{until: now.Add(time.Hour)}
	}
	if b.active >= 4 || b.window.count >= 200 {
		return false
	}
	b.active++
	b.window.count++
	return true
}
func (b *Budget) release() { b.mu.Lock(); b.active--; b.mu.Unlock() }

type Handler struct {
	RenderEmail func(string, map[string]string) (mailer.Message, error)
	config      Config
	site        string
	sender      mailer.Sender
	budget      *Budget
	local       bool
	logger      *log.Logger
	mu          sync.Mutex
	ips         map[string]bucket
	window      bucket
	active      bool
}

func New(c Config, site string, sender mailer.Sender, budget *Budget, local bool, logger *log.Logger) *Handler {
	return &Handler{config: c, site: site, sender: sender, budget: budget, local: local, logger: logger, ips: map[string]bucket{}}
}

func (h *Handler) allow(ip string, now time.Time) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	for key, b := range h.ips {
		if !now.Before(b.until) {
			delete(h.ips, key)
		}
	}
	if !now.Before(h.window.until) {
		h.window = bucket{until: now.Add(time.Hour)}
	}
	b, exists := h.ips[ip]
	if !exists {
		b = bucket{until: now.Add(15 * time.Minute)}
	}
	if h.active || h.window.count >= 20 || b.count >= 3 || (!exists && len(h.ips) >= 1024) {
		return false
	}
	b.count++
	h.ips[ip] = b
	h.window.count++
	h.active = true
	return true
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "no-referrer")
	if r.URL.Path != Path || r.URL.RawPath != "" || h.config.To == "" {
		http.NotFound(w, r)
		return
	}
	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		http.Error(w, "Please submit the contact form from the site.", 405)
		return
	}
	scheme := "https"
	if h.local {
		scheme = "http"
		host := r.Host
		if v, _, err := net.SplitHostPort(host); err == nil {
			host = v
		}
		if ip := net.ParseIP(host); !devaccess.LAN(r) && host != "localhost" && (ip == nil || !ip.IsLoopback()) {
			http.Error(w, "Use the local address shown by flea dev.", 403)
			return
		}
	} else if r.TLS == nil {
		http.Error(w, "Contact forms require HTTPS.", 403)
		return
	}
	// Browsers supply Origin on form POSTs. Reject absent, null and foreign origins;
	// forwarded headers are never used for authority or rate-limit identity.
	if r.Header.Get("Origin") != scheme+"://"+r.Host {
		http.Error(w, "Please submit the form from this site.", 403)
		return
	}
	if r.URL.RawQuery != "" {
		http.Error(w, "Form settings cannot be supplied in the URL.", 400)
		return
	}
	kind, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || kind != "application/x-www-form-urlencoded" {
		http.Error(w, "Use a standard form without file uploads.", 415)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 32<<10)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "The form is too large or invalid.", 400)
		return
	}
	for key, values := range r.PostForm {
		if (key != "name" && key != "email" && key != "message" && key != "website") || len(values) != 1 {
			http.Error(w, "The form contains unsupported or repeated fields.", 400)
			return
		}
	}
	name, address, message := strings.TrimSpace(r.PostForm.Get("name")), strings.TrimSpace(r.PostForm.Get("email")), strings.TrimSpace(r.PostForm.Get("message"))
	if name == "" || utf8.RuneCountInString(name) > 100 || strings.IndexFunc(name, unicode.IsControl) >= 0 || !email(address) || message == "" || utf8.RuneCountInString(message) > 5000 || !utf8.ValidString(name+message+address) || strings.ContainsRune(message, 0) {
		http.Error(w, "Enter your name (up to 100 characters), email address and message (up to 5,000 characters).", 400)
		return
	}
	if r.PostForm.Get("website") != "" {
		if h.config.SuccessPath != "" {
			http.Redirect(w, r, h.config.SuccessPath, http.StatusSeeOther)
			return
		}
		fmt.Fprintln(w, "Thank you. Your form was received.")
		return
	}
	if h.sender == nil {
		message := "Email delivery is unavailable. Please contact the site admin directly."
		if h.local {
			message = "Email delivery is not configured. Set RESEND_API_KEY and FLEA_EMAIL_FROM to a verified sender address, then restart flea dev."
		}
		http.Error(w, message, 503)
		return
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		http.Error(w, "Could not identify the connection.", 400)
		return
	}
	if !h.allow(ip, time.Now()) {
		w.Header().Set("Retry-After", "900")
		http.Error(w, "Too many submissions. Please try again later.", 429)
		return
	}
	defer func() { h.mu.Lock(); h.active = false; h.mu.Unlock() }()
	if !h.budget.acquire(time.Now()) {
		w.Header().Set("Retry-After", "900")
		http.Error(w, "Contact delivery is busy. Please try again later.", 429)
		return
	}
	defer h.budget.release()
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	mail := mailer.Message{To: h.config.To, ReplyTo: address, Subject: "New website contact message", Text: "Name: " + name + "\nEmail: " + address + "\n\n" + message}
	if h.RenderEmail != nil {
		rendered, renderErr := h.RenderEmail("contact", map[string]string{"name": name, "address": address, "message": message})
		err = renderErr
		if err == nil {
			mail.HTML, mail.Text = rendered.HTML, rendered.Text
		}
	}
	if err == nil {
		err = h.sender.Send(ctx, mail)
	}
	if err != nil {
		if h.logger != nil {
			h.logger.Printf("Contact delivery failed for site %q", h.site)
		}
		http.Error(w, "We couldn't confirm delivery. Please try again later or contact us directly.", 502)
		return
	}
	if h.logger != nil {
		h.logger.Printf("Contact message accepted for site %q", h.site)
	}
	if h.config.SuccessPath != "" {
		http.Redirect(w, r, h.config.SuccessPath, http.StatusSeeOther)
		return
	}
	fmt.Fprintln(w, "Thank you. Your message was accepted for delivery. You can return to the previous page.")
}
