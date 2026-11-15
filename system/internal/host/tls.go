package host

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/mail"
	"net/url"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/acme/autocert"
)

type TLSConfig struct {
	Listen   string   `json:"listen"`
	Email    string   `json:"email"`
	AgreeTOS bool     `json:"agree_tos"`
	CacheDir string   `json:"cache_dir"`
	Domains  []string `json:"domains"`
}

func ValidHost(host string) bool {
	if net.ParseIP(host) != nil {
		return !strings.Contains(host, ":")
	} // IPv4 aliases; IPv6 SSH targets are supported separately.
	if host == "" || len(host) > 253 || host != strings.ToLower(host) {
		return false
	}
	for _, label := range strings.Split(host, ".") {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, c := range label {
			if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-') {
				return false
			}
		}
	}
	return true
}

func ValidDomain(host string) bool {
	return ValidHost(host) && net.ParseIP(host) == nil && strings.Contains(host, ".") && !strings.HasSuffix(host, ".localhost")
}

func validateTLS(c Config) error {
	if c.TLS == nil {
		return nil
	}
	t := c.TLS
	email, err := mail.ParseAddress(t.Email)
	if err != nil || email.Address != t.Email || !strings.Contains(t.Email, "@") {
		return fmt.Errorf("TLS requires a valid contact email")
	}
	if !t.AgreeTOS {
		return fmt.Errorf("TLS requires explicit agree_tos: true for Let's Encrypt terms")
	}
	if !filepath.IsAbs(t.CacheDir) {
		return fmt.Errorf("TLS cache_dir must be absolute")
	}
	// Public ACME uses standard ports. Custom listeners are reserved for internal tests.
	if t.Listen != ":443" || c.Listen != ":80" {
		return fmt.Errorf("automatic HTTPS requires listen :80 and tls.listen :443")
	}
	if len(t.Domains) == 0 {
		return fmt.Errorf("TLS requires at least one configured domain")
	}
	seen := map[string]bool{}
	for _, domain := range t.Domains {
		if !ValidDomain(domain) || c.Sites[domain] == "" || seen[domain] {
			return fmt.Errorf("invalid, duplicate, or unconfigured TLS domain %q", domain)
		}
		seen[domain] = true
	}
	return nil
}

func requestHost(r *http.Request) string {
	host := r.Host
	if name, _, err := net.SplitHostPort(host); err == nil {
		host = name
	}
	return strings.TrimSuffix(strings.ToLower(host), ".")
}

func hasTLS(c Config, domain string) bool {
	if c.TLS != nil {
		for _, h := range c.TLS.Domains {
			if h == domain {
				return true
			}
		}
	}
	return false
}

func transports(c Config, handler http.Handler) (http.Handler, *tls.Config, error) {
	if c.TLS == nil {
		return handler, nil, nil
	}
	manager := &autocert.Manager{Prompt: autocert.AcceptTOS, Email: c.TLS.Email, Cache: autocert.DirCache(c.TLS.CacheDir), HostPolicy: autocert.HostWhitelist(c.TLS.Domains...)}
	return managedTransports(c, handler, manager)
}

func managedTransports(c Config, handler http.Handler, manager *autocert.Manager) (http.Handler, *tls.Config, error) {
	tc := manager.TLSConfig()
	tc.MinVersion = tls.VersionTLS12
	// Enforce the allowlist even for cached certificates (autocert's policy only
	// governs acquisition). Unknown SNI must never cause certificate requests.
	getCertificate := tc.GetCertificate
	tc.GetCertificate = func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		if !hasTLS(c, hello.ServerName) {
			return nil, fmt.Errorf("unconfigured TLS hostname")
		}
		return getCertificate(hello)
	}
	return manager.HTTPHandler(redirectHTTPS(c, handler)), tc, nil
}

func redirectHTTPS(c Config, fallback http.Handler) http.Handler {
	return indexingHandler(c, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := requestHost(r)
		if r.URL.Path == HealthPath || !hasTLS(c, host) {
			fallback.ServeHTTP(w, r)
			return
		}
		target := url.URL{Scheme: "https", Host: host, Path: r.URL.Path, RawPath: r.URL.RawPath, RawQuery: r.URL.RawQuery}
		if canonical := c.Redirects[host]; canonical != "" {
			target.Host = canonical
			if !hasTLS(c, canonical) {
				target.Scheme = "http"
			}
		}
		http.Redirect(w, r, target.String(), http.StatusPermanentRedirect)
	}))
}

func tlsHandler(c Config, fallback http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !hasTLS(c, requestHost(r)) || r.TLS == nil || requestHost(r) != r.TLS.ServerName {
			http.Error(w, "Misdirected Request", http.StatusMisdirectedRequest)
			return
		}
		fallback.ServeHTTP(w, r)
	})
}
