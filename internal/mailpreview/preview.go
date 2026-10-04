// Package mailpreview provides a bounded development-only mail sender.
// Only the dev CLI wires it; the production host always uses real delivery.
package mailpreview

import (
	"bytes"
	"context"
	"crypto/rand"
	_ "embed"
	"encoding/hex"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"flea/internal/devaccess"
	"flea/internal/mailer"
)

const Path = "/_flea/email-preview/"
const lifetime = 10 * time.Minute
const maxMessages = 20

//go:embed preview.html
var page string

var view = template.Must(template.New("email").Parse(page))

type message struct {
	mailer.Message
	ID, FromName, FromAddress, Date, DateTime string
	Created                                   time.Time
}

type Preview struct {
	mu                          sync.Mutex
	messages                    []message
	origin, host, name, address string
	logger                      *log.Logger
	open                        func(context.Context, string) error
	now                         func() time.Time
}

func New(origin, name, address string, logger *log.Logger) (*Preview, error) {
	u, err := url.Parse(origin)
	if err != nil || u.Scheme != "http" || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" || !net.ParseIP(u.Hostname()).IsLoopback() || u.Port() == "" {
		return nil, fmt.Errorf("email previews require a loopback HTTP address")
	}
	if name == "" {
		name = "FleaCMS"
	}
	if address == "" {
		address = "preview@localhost"
	}
	return &Preview{origin: origin, host: u.Host, name: name, address: address, logger: logger, open: openBrowser, now: time.Now}, nil
}

func openBrowser(ctx context.Context, address string) error {
	command := "xdg-open"
	if runtime.GOOS == "darwin" {
		command = "open"
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	return exec.CommandContext(ctx, command, address).Run()
}

// prune is called under mu. Retention is bounded even if nobody visits previews.
func (p *Preview) prune(now time.Time) {
	kept := p.messages[:0]
	for _, m := range p.messages {
		if now.Sub(m.Created) < lifetime {
			kept = append(kept, m)
		}
	}
	clear(p.messages[len(kept):])
	p.messages = kept
}

func (p *Preview) Send(ctx context.Context, m mailer.Message) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if m.From != "" && !mailer.ValidAddress(m.From) {
		return fmt.Errorf("the configured sender email address is invalid")
	}
	if len(m.HTML) > 256<<10 || len(m.Text) > 256<<10 || len(m.To)+len(m.Subject)+len(m.ReplyTo) > 8192 {
		return fmt.Errorf("email preview exceeds size limits")
	}
	var token [32]byte
	if _, err := rand.Read(token[:]); err != nil {
		return fmt.Errorf("could not create email preview")
	}
	id := hex.EncodeToString(token[:])
	p.mu.Lock()
	now := p.now()
	p.prune(now)
	if len(p.messages) == maxMessages {
		copy(p.messages, p.messages[1:])
		p.messages[len(p.messages)-1] = message{}
		p.messages = p.messages[:len(p.messages)-1]
	}
	from := p.address
	if m.From != "" {
		from = m.From
	}
	p.messages = append(p.messages, message{Message: m, ID: id, FromName: p.name, FromAddress: from, Created: now, Date: now.Format("Monday, January 2, 2006 at 3:04 PM MST"), DateTime: now.Format(time.RFC3339)})
	p.mu.Unlock()
	address := p.origin + Path + id
	if err := p.open(ctx, address); err != nil && p.logger != nil {
		// Deliberately show only a local capability link, never message data or
		// process/provider errors. Failure to launch must not discard the code.
		p.logger.Printf("Open this local email preview: %s", address)
	}
	return nil
}

func (p *Preview) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; img-src data:; frame-src 'self'; frame-ancestors 'none'; base-uri 'none'; form-action 'none'")
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	allowed := err == nil && net.ParseIP(ip).IsLoopback() && r.Host == p.host
	origin := p.origin
	if devaccess.LAN(r) {
		allowed = true
		origin = "http://" + r.Host
	}
	if !allowed || (r.Header.Get("Origin") != "" && r.Header.Get("Origin") != origin) || r.Header.Get("Sec-Fetch-Site") == "cross-site" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, Path)
	if !strings.HasPrefix(r.URL.Path, Path) || len(id) != 64 {
		http.NotFound(w, r)
		return
	}
	p.mu.Lock()
	p.prune(p.now())
	var found message
	for _, m := range p.messages {
		if m.ID == id {
			found = m
			break
		}
	}
	p.mu.Unlock()
	if found.ID == "" {
		http.Error(w, "This email preview has expired or is no longer available.", http.StatusNotFound)
		return
	}
	var body bytes.Buffer
	if err := view.Execute(&body, found); err != nil {
		http.Error(w, "Could not display email preview.", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if r.Method == http.MethodGet {
		_, _ = w.Write(body.Bytes())
	}
}
