package site

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Redirects are compiled into the immutable Project, separately from public
// route validation: old .html and trailing-slash URLs may redirect, never serve.
func (p *Project) loadRedirects(raw json.RawMessage) error {
	if raw == nil {
		return nil
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	if token, err := d.Token(); err != nil || token != json.Delim('{') {
		return fmt.Errorf(".flea.config.json: redirects must be an object of old paths to destination URLs")
	}
	p.redirects = map[string]url.URL{}
	for d.More() {
		token, err := d.Token()
		if err != nil {
			return fmt.Errorf(".flea.config.json: invalid redirects: %w", err)
		}
		from := token.(string)
		source, err := redirectURL(from)
		if err != nil || source.RawQuery != "" || source.ForceQuery || strings.Contains(from, "#") || !redirectSourcePath(source.Path) {
			return fmt.Errorf(".flea.config.json: redirect source %q must be a public path without a query or fragment", from)
		}
		if _, exists := p.redirects[source.Path]; exists {
			return fmt.Errorf(".flea.config.json: duplicate redirect source %q", from)
		}
		var destination string
		if err := d.Decode(&destination); err != nil {
			return fmt.Errorf(".flea.config.json: redirect %q needs a destination string", from)
		}
		target, err := redirectURL(destination)
		if err != nil || !safeURL(target.Path) || reservedRedirectPath(target.Path) {
			return fmt.Errorf(".flea.config.json: redirect %q needs a public destination URL on this site", from)
		}
		p.redirects[source.Path] = target
	}
	for source, target := range p.redirects {
		if _, exists := p.redirects[target.Path]; exists {
			return fmt.Errorf(".flea.config.json: redirect %q points to another redirect; use the final destination", source)
		}
		if p.Pages[source] != nil || p.Assets[source] != "" {
			return fmt.Errorf(".flea.config.json: redirect %q conflicts with an existing page or asset", source)
		}
		for name, app := range p.Apps {
			// Blog's former slugs may redirect; its published routes were checked
			// above. Other apps own dispatch for their entire subtree.
			if name != "blog" && (source == app.Mount || strings.HasPrefix(source, app.Mount+"/")) {
				return fmt.Errorf(".flea.config.json: redirect %q conflicts with app %q", source, name)
			}
		}
		if p.Pages[target.Path] == nil && p.Assets[target.Path] == "" {
			return fmt.Errorf(".flea.config.json: redirect %q destination %q is not a public page or asset", source, target.Path)
		}
	}
	return nil
}

func redirectURL(value string) (url.URL, error) {
	if !utf8.ValidString(value) || strings.ContainsFunc(value, unicode.IsControl) || !strings.HasPrefix(value, "/") || strings.HasPrefix(value, "//") {
		return url.URL{}, fmt.Errorf("expected a URL starting with one slash")
	}
	u, err := url.Parse(value)
	if err != nil || u.IsAbs() || u.Host != "" || u.User != nil || strings.ContainsFunc(u.Path, unicode.IsControl) || strings.ContainsAny(u.Path, "\\%") {
		return url.URL{}, fmt.Errorf("invalid site URL")
	}
	return *u, nil
}

func reservedRedirectPath(value string) bool {
	return value == "/admin" || strings.HasPrefix(value, "/admin/") || value == "/releases" || strings.HasPrefix(value, "/releases/")
}

func redirectSourcePath(value string) bool {
	trimmed := value
	if value != "/" {
		trimmed = strings.TrimSuffix(value, "/")
	}
	if trimmed == "" || path.Clean(trimmed) != trimmed || reservedRedirectPath(value) {
		return false
	}
	for _, part := range strings.Split(value, "/") {
		if strings.HasPrefix(part, "_") || strings.HasPrefix(part, ".") || strings.HasSuffix(strings.ToLower(part), ".layout.html") {
			return false
		}
	}
	return true
}

func (p *Project) serveRedirect(w http.ResponseWriter, r *http.Request) bool {
	target, ok := p.redirects[r.URL.Path]
	if !ok {
		return false
	}
	// A configured query replaces the incoming one; otherwise retain tracking
	// parameters. Copy the URL value so requests never mutate the snapshot.
	if target.RawQuery == "" && !target.ForceQuery {
		target.RawQuery, target.ForceQuery = r.URL.RawQuery, r.URL.ForceQuery
	}
	http.Redirect(w, r, target.String(), http.StatusMovedPermanently)
	return true
}
