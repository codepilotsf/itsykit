// Package host runs the persistent server, without development reload scripts,
// filesystem watching, or network access to resolve icons.
package host

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"nanoweb/internal/editor"
	"nanoweb/internal/editstore"
	"nanoweb/internal/mailer"
	"nanoweb/internal/site"
	"nanoweb/internal/snips"
)

const HealthPath = "/_nanoweb/health"

type Config struct {
	Content      map[string]string `json:"content,omitempty"`
	Listen       string            `json:"listen"`
	Sites        map[string]string `json:"sites"`
	Revision     string            `json:"revision,omitempty"`
	TLS          *TLSConfig        `json:"tls,omitempty"`
	Redirects    map[string]string `json:"redirects,omitempty"`
	SiteBranches map[string]string `json:"site_branches,omitempty"`
	Commits      map[string]string `json:"commits,omitempty"`
}

func Load(configPath string, logger *log.Logger) (Config, http.Handler, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, nil, err
	}
	var config Config
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&config); err != nil {
		return config, nil, err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return config, nil, fmt.Errorf("server config must contain one JSON object")
	}
	address, port, err := net.SplitHostPort(config.Listen)
	portNumber, portErr := strconv.Atoi(port)
	if err != nil || portErr != nil || portNumber < 1 || portNumber > 65535 || (address != "" && net.ParseIP(address) == nil) {
		return config, nil, fmt.Errorf("server listen must be an IP address (or empty wildcard) and numeric port 1-65535")
	}
	if err := validateTLS(config); err != nil {
		return config, nil, err
	}
	for from, to := range config.Redirects {
		if from != "www."+to || !ValidDomain(to) || config.Sites[from] == "" || config.Sites[from] != config.Sites[to] || config.Redirects[to] != "" {
			return config, nil, fmt.Errorf("invalid WWW redirect %s to %s", from, to)
		}
	}
	handlers := map[string]http.Handler{}
	// Aliases share one immutable snapshot. Different content paths for the same
	// release would make deployment preservation depend on map iteration order.
	byRoot := map[string]http.Handler{}
	contentPaths := map[string]string{}
	for host, root := range config.Sites {
		if !ValidHost(host) {
			return config, nil, fmt.Errorf("invalid site hostname %q", host)
		}
		if !filepath.IsAbs(root) {
			return config, nil, fmt.Errorf("site %s: root must be absolute", host)
		}
		key := filepath.Clean(root)
		contentPath := config.Content[host]
		if contentPath != "" {
			contentPath = filepath.Clean(contentPath)
		}
		if handler, ok := byRoot[key]; ok {
			if contentPaths[key] != contentPath {
				return config, nil, fmt.Errorf("aliases for site %s must use the same content snapshot", root)
			}
			handlers[host] = handler
			continue
		}
		var content *snips.Store
		if name := config.Content[host]; name != "" {
			if !filepath.IsAbs(name) {
				return config, nil, fmt.Errorf("content path must be absolute")
			}
			content, err = snips.ReadFile(name)
			if err != nil {
				return config, nil, err
			}
		}
		project, err := site.LoadWithSnips(root, content)
		if err != nil {
			return config, nil, fmt.Errorf("site %s: %w", host, err)
		}
		if logger != nil {
			for _, warning := range project.Warnings {
				logger.Printf("Warning: %s: %s", host, warning)
			}
		}
		public := site.NewProductionServer(project, logger)
		handlers[host] = public
		if len(project.Admin.Emails) > 0 {
			if contentPath == "" {
				return config, nil, fmt.Errorf("site %s: admin requires a managed content snapshot; deploy this site first", host)
			}
			store := &editstore.Store{Root: root, ContentPath: contentPath, ConfigPath: configPath, Emails: project.Admin.Emails, Publish: public.Publish}
			dashboard := editor.New(editor.Options{Emails: project.Admin.Emails, Name: fmt.Sprint(project.Site["name"]), Sender: mailer.Env(), Backend: store, Logger: logger})
			handlers[host] = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if editor.IsPath(r.URL.Path) {
					dashboard.ServeHTTP(w, r)
				} else {
					public.ServeHTTP(w, r)
				}
			})
		}
		byRoot[key] = handlers[host]
		contentPaths[key] = contentPath
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == HealthPath {
			w.Header().Set("Cache-Control", "no-store")
			w.Header().Set("X-Nanoweb-Revision", config.Revision)
			if _, ok := handlers[requestHost(r)]; ok {
				w.Header().Set("X-Nanoweb-Site", requestHost(r))
			}
			w.Header().Set("Content-Type", "application/json")
			if r.Method != "GET" && r.Method != "HEAD" {
				w.Header().Set("Allow", "GET, HEAD")
				w.WriteHeader(405)
				return
			}
			if r.Method == "GET" {
				_, _ = io.WriteString(w, `{"service":"nanoweb","status":"ready"}`)
			}
			return
		}
		host := requestHost(r)
		if handler, ok := handlers[host]; ok {
			if target := config.Redirects[host]; target != "" {
				scheme := "http"
				if hasTLS(config, target) {
					scheme = "https"
				}
				u := url.URL{Scheme: scheme, Host: target, Path: r.URL.Path, RawPath: r.URL.RawPath, RawQuery: r.URL.RawQuery}
				http.Redirect(w, r, u.String(), http.StatusPermanentRedirect)
				return
			}
			handler.ServeHTTP(w, r)
			return
		}
		if logger != nil {
			logger.Printf("404 %s %s", r.Method, r.URL.EscapedPath())
		}
		http.NotFound(w, r)
	})
	return config, handler, nil
}

func Run(ctx context.Context, configPath string, logger *log.Logger) error {
	config, handler, err := Load(configPath, logger)
	if err != nil {
		return err
	}
	return serve(ctx, config, handler, logger, nil)
}

// A supplied TLS configuration is used only by integration tests with a local CA.
func serve(ctx context.Context, config Config, handler http.Handler, logger *log.Logger, testTLS *tls.Config) error {
	httpHandler, tlsConfig, err := transports(config, handler)
	if err != nil {
		return err
	}
	if testTLS != nil {
		tlsConfig = testTLS
	}
	listener, err := net.Listen("tcp", config.Listen)
	if err != nil {
		return err
	}
	defer listener.Close()
	newServer := func(h http.Handler) *http.Server {
		return &http.Server{Handler: h, ErrorLog: logger, ReadHeaderTimeout: 10 * time.Second, ReadTimeout: 30 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	}
	server := newServer(httpHandler)
	servers := []*http.Server{server}
	done := make(chan error, 2)
	if tlsConfig != nil {
		secure, err := net.Listen("tcp", config.TLS.Listen)
		if err != nil {
			return fmt.Errorf("HTTPS listen: %w", err)
		}
		defer secure.Close()
		s := newServer(tlsHandler(config, handler))
		s.TLSConfig = tlsConfig
		servers = append(servers, s)
		go func() { done <- s.ServeTLS(secure, "", "") }()
	}
	go func() { done <- server.Serve(listener) }()
	if logger != nil {
		logger.Printf("Nanoweb server listening on %s", listener.Addr())
	}
	var result error
	remaining := len(servers)
	select {
	case result = <-done:
		remaining--
	case <-ctx.Done():
	}
	shutdown, stop := context.WithTimeout(context.Background(), 5*time.Second)
	defer stop()
	for _, s := range servers {
		if err := s.Shutdown(shutdown); err != nil {
			_ = s.Close()
		}
	}
	for i := 0; i < remaining; i++ {
		<-done
	}
	if result == http.ErrServerClosed {
		return nil
	}
	return result
}

func Check(ctx context.Context, configPath string) error {
	config, _, err := Load(configPath, nil)
	if err != nil {
		return err
	}
	return Probe(ctx, config, "", false)
}

// Probe verifies the active revision and hostname without trusting a stale process.
// HTTPS probes use normal certificate verification and SNI, dialing locally.
func Probe(ctx context.Context, config Config, domain string, secure bool) error {
	address := checkAddress(config.Listen)
	scheme, authority := "http", address
	transport := &http.Transport{Proxy: nil}
	if secure {
		if config.TLS == nil {
			return fmt.Errorf("HTTPS is not configured")
		}
		address = checkAddress(config.TLS.Listen)
		scheme, authority = "https", domain
		transport.DialContext = func(ctx context.Context, network, _ string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, network, address)
		}
	}
	request, err := http.NewRequestWithContext(ctx, "GET", scheme+"://"+authority+HealthPath, nil)
	if domain != "" && err == nil {
		request.Host = domain
	}
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: 10 * time.Second, Transport: transport, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	defer client.CloseIdleConnections()
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, 1024))
	if err != nil {
		return err
	}
	if response.StatusCode != 200 || string(body) != `{"service":"nanoweb","status":"ready"}` {
		return fmt.Errorf("Nanoweb server readiness check failed")
	}
	if response.Header.Get("X-Nanoweb-Revision") != config.Revision || (domain != "" && response.Header.Get("X-Nanoweb-Site") != domain) {
		return fmt.Errorf("server is not serving the expected release/hostname")
	}
	return nil
}

// Probe wildcard listeners through loopback, never an unspecified destination.
func checkAddress(listen string) string {
	address, port, _ := net.SplitHostPort(listen)
	switch address {
	case "", "0.0.0.0":
		address = "127.0.0.1"
	case "::":
		address = "::1"
	}
	return net.JoinHostPort(address, port)
}
