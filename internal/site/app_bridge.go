package site

import (
	"fmt"
	"net"
	"net/http"

	contactapp "flea/apps/contact"
	"flea/internal/appapi"
	"flea/internal/contact"
	"flea/internal/viewer"
)

func (s *Server) serveInstalledApp(w http.ResponseWriter, r *http.Request, mount appapi.Mount) {
	readonly := r.Method == "GET" || r.Method == "HEAD"
	if r.URL.Path != mount.Path && !(readonly && r.URL.Path == mount.Path+"/thanks") {
		http.NotFound(w, r)
		return
	}
	s.serveApp(w, r, mount, nil, readonly, true)
}

// Internal integration seam for independent protocol services.
func (s *Server) serveAppRequest(w http.ResponseWriter, r *http.Request, mount appapi.Mount, client *appapi.Client) {
	s.serveApp(w, r, mount, client, false, false)
}

func (s *Server) serveApp(w http.ResponseWriter, r *http.Request, mount appapi.Mount, client *appapi.Client, readonly, installed bool) {
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "same-origin")
	s.mu.RLock()
	p, loadErr, revision := s.project, s.err, s.revision
	s.mu.RUnlock()
	if loadErr != nil || p == nil || p.syncOnly {
		http.Error(w, "This page is unavailable.", 503)
		return
	}
	// Resolve the running config against the same immutable snapshot rendered
	// below. A reload between dispatch and this point cannot use an old recipient.
	var delivery *contact.Delivery
	if installed {
		current, running, ok := s.appFor(p, r.URL.Path)
		if !ok || current != mount {
			http.Error(w, "This app is temporarily unavailable. Please reload the page.", 503)
			return
		}
		if running != nil {
			delivery = running.delivery
			if running.manager != nil {
				client = running.manager.Client()
			}
		}
	}
	if _, err := mount.ServicePath(r.URL.Path); err != nil || r.URL.RawPath != "" {
		http.NotFound(w, r)
		return
	}
	t, ok := p.Pages[r.URL.Path]
	if !ok {
		http.NotFound(w, r)
		return
	}
	if r.Method != "GET" && r.Method != "HEAD" && r.Method != "POST" {
		w.Header().Set("Allow", "GET, HEAD, POST")
		http.Error(w, "Method not allowed.", 405)
		return
	}
	var fields map[string]any
	if r.Method == "POST" {
		if r.URL.RawQuery != "" {
			http.Error(w, "Form settings cannot be supplied in the URL.", 400)
			return
		}
		var err error
		fields, err = appapi.DecodeForm(w, r, s.forms)
		if err != nil {
			failure := err.(*appapi.RequestError)
			http.Error(w, failure.Message, failure.Status)
			return
		}
	}
	envelope, status := appapi.Envelope{}, 200
	limited := false
	if delivery != nil && r.Method == "POST" {
		submission, invalid, spam := contactapp.Validate(fields)
		if len(invalid) == 0 && !spam {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				http.Error(w, "Could not identify the connection.", 400)
				return
			}
			release, allowed := delivery.Reserve(r.Context(), ip, submission)
			if allowed {
				defer release()
			} else {
				limited = true
				status = 429
				envelope.Error = "Too many submissions. Please try again later."
				w.Header().Set("Retry-After", "900")
			}
		}
	}
	if !readonly && !limited {
		var err error
		method := r.Method
		if method == "HEAD" {
			method = "GET"
		}
		if client == nil {
			err = appapi.ErrUnavailable
		} else {
			envelope, status, err = client.Request(r.Context(), mount, method, r.URL.Path, r.URL.RawQuery, fields)
		}
		if err != nil {
			envelope = appapi.Envelope{Error: appapi.ErrUnavailable.Error()}
			status = 503
		}
	}
	if status >= 400 && mount.Name == "contact" && fields != nil {
		envelope.Data = contactapp.Redisplay(fields)
	}
	if envelope.Redirect != "" {
		destination, err := mount.Redirect(envelope.Redirect)
		if err != nil || p.Pages[destination] == nil {
			http.Error(w, "The app returned an invalid destination.", 503)
			return
		}
		s.storeFlash(w, r, mount.Name, destination, envelope.Flash)
		http.Redirect(w, r, destination, 303)
		return
	}
	flash, finish := s.takeFlash(r, mount.Name)
	finished := false
	defer func() {
		if !finished {
			finish(w, false)
		}
	}()
	if flash != nil {
		envelope.Flash = flash
	}
	inputs, commit := s.formRender(r)
	inputs.appName, inputs.appData = mount.Name, envelope.Values()
	decorate, owner := r.Context().Value(ownerRenderKey{}).(ownerTransform)
	var admin map[string]any
	if identity, authenticated := viewer.From(r.Context()); owner && authenticated {
		admin = map[string]any{"email": identity.Email, "csrf_token": identity.CSRF, "version": p.EditingVersion(), "revision": ""}
		if !s.production {
			admin["revision"] = fmt.Sprint(revision)
		}
	}
	body, err := p.renderRequest(t, r.URL.Path, admin, owner, inputs)
	if err != nil {
		http.Error(w, "This page could not be rendered.", 500)
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
	commit(w)
	finish(w, true)
	finished = true
	writeHTML(w, r, status, body)
}
