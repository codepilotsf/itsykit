package contact

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	contactapp "itsy/apps/contact"
	"itsy/internal/appapi"
	"itsy/internal/mailer"
)

type Email struct {
	To   string `json:"to"`
	From string `json:"from,omitempty"`
}

func (e Email) Validate() error {
	if !mailer.ValidAddress(e.To) {
		return fmt.Errorf("apps.contact.email.to must be one email address")
	}
	if e.From != "" && !mailer.ValidAddress(e.From) {
		return fmt.Errorf("apps.contact.email.from must be one email address")
	}
	return nil
}

// Delivery is a site-bound, single-use capability for the currently admitted
// visitor submission. Possessing the app credential alone is not a mail relay.
type Delivery struct {
	Email   Email
	Sender  mailer.Sender
	Budget  *Budget
	Render  func(string, map[string]string) (mailer.Message, error)
	mu      sync.Mutex
	ips     map[string]bucket
	window  bucket
	pending *contactapp.Submission
	parent  context.Context
	used    bool
}

func (d *Delivery) Reserve(ctx context.Context, ip string, s contactapp.Submission) (func(), bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	now := time.Now()
	if d.ips == nil {
		d.ips = map[string]bucket{}
	}
	for k, b := range d.ips {
		if !now.Before(b.until) {
			delete(d.ips, k)
		}
	}
	if !now.Before(d.window.until) {
		d.window = bucket{until: now.Add(time.Hour)}
	}
	b, exists := d.ips[ip]
	if !exists {
		b = bucket{until: now.Add(15 * time.Minute)}
	}
	if d.pending != nil || d.window.count >= 20 || b.count >= 3 || (!exists && len(d.ips) >= 1024) {
		return nil, false
	}
	b.count++
	d.ips[ip] = b
	d.window.count++
	parent, cancel := context.WithCancel(ctx)
	d.pending = &s
	d.parent = parent
	d.used = false
	return func() {
		cancel()
		d.mu.Lock()
		d.pending = nil
		d.parent = nil
		d.used = false
		d.mu.Unlock()
	}, true
}

func (d *Delivery) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	fields, err := appapi.ReadFields(w, r)
	if err != nil || len(fields) != 3 {
		http.Error(w, "Invalid notification", 422)
		return
	}
	s, invalid, spam := contactapp.Validate(fields)
	if len(invalid) != 0 || spam {
		http.Error(w, "Invalid notification", 422)
		return
	}
	d.mu.Lock()
	if d.pending == nil || *d.pending != s || d.used || d.parent.Err() != nil {
		d.mu.Unlock()
		http.Error(w, "Notification not authorized", 403)
		return
	}
	d.used = true
	parent := d.parent
	d.mu.Unlock()
	if d.Email.Validate() != nil || d.Sender == nil || d.Render == nil || d.Budget == nil {
		http.Error(w, "Delivery unavailable", 503)
		return
	}
	if !d.Budget.acquire(time.Now()) {
		http.Error(w, "Delivery unavailable", 503)
		return
	}
	defer d.Budget.release()
	ctx, cancel := context.WithTimeout(parent, 10*time.Second)
	defer cancel()
	message, err := d.Render("contact", map[string]string{"name": s.Name, "address": s.Email, "message": s.Message})
	if err == nil {
		message.To = d.Email.To
		message.From = d.Email.From
		message.ReplyTo = s.Email
		message.Subject = "New website contact message"
		err = d.Sender.Send(ctx, message)
	}
	if err != nil {
		http.Error(w, "Could not confirm delivery", 503)
		return
	}
	w.WriteHeader(204)
}
