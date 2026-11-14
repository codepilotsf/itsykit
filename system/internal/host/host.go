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
	"itsy/internal/sitefiles"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"itsy/internal/contact"
	"itsy/internal/editor"
	"itsy/internal/editstore"
	"itsy/internal/mailer"
	"itsy/internal/site"
	"itsy/internal/snips"
)

const HealthPath = "/_flea/health"

type Config struct {
	Indexing     map[string]string `json:"indexing,omitempty"`
	Downloads    map[string]string `json:"downloads,omitempty"`
	Content      map[string]string `json:"content,omitempty"`
	Listen       string            `json:"listen"`
	Sites        map[string]string `json:"sites"`
	Revision     string            `json:"revision,omitempty"`
	TLS          *TLSConfig        `json:"tls,omitempty"`
	Redirects    map[string]string `json:"redirects,omitempty"`
	SiteBranches map[string]string `json:"site_branches,omitempty"`
	Commits      map[string]string `json:"commits,omitempty"`

	indexingWarnings map[string][]string
}

func Load(configPath string, logger *log.Logger) (Config, http.Handler, error) {
	return load(configPath, logger, mailer.Env())
}

func load(configPath string, logger *log.Logger, sender mailer.Sender) (Config, http.Handler, error) {
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
	if err := validateIndexing(config); err != nil {
		return config, nil, err
	}
	robotsWarnings := map[string]string{}
	handlers := map[string]http.Handler{}
	downloadHandlers := map[string]*downloads{}
	downloadRoots := map[string]string{}
	downloadStores := map[string]*downloads{}
	for hostname, directory := range config.Downloads {
		if config.Sites[hostname] == "" || directory == "" {
			return config, nil, fmt.Errorf("downloads require a configured site and directory")
		}
		if downloadStores[directory] == nil {
			store, e := loadDownloads(directory)
			if e != nil {
				return config, nil, fmt.Errorf("downloads for %s: %w", hostname, e)
			}
			downloadStores[directory] = store
		}
		downloadHandlers[hostname] = downloadStores[directory]
	}
	// Aliases share one immutable snapshot. Different content paths for the same
	// release would make deployment preservation depend on map iteration order.
	byRoot := map[string]http.Handler{}
	var appSites []*site.Server
	contentPaths := map[string]string{}
	sessionOwners := map[string]string{}
	contactBudget := &contact.Budget{}
	type owner struct {
		root string
		file os.FileInfo
	}
	var contentOwners []owner
	for host, root := range config.Sites {
		if !ValidHost(host) {
			return config, nil, fmt.Errorf("invalid site hostname %q", host)
		}
		if !filepath.IsAbs(root) {
			return config, nil, fmt.Errorf("site %s: root must be absolute", host)
		}
		key := filepath.Clean(root)
		if previous, ok := downloadRoots[key]; ok && previous != config.Downloads[host] {
			return config, nil, fmt.Errorf("aliases must use the same download store")
		}
		downloadRoots[key] = config.Downloads[host]
		contentPath := config.Content[host]
		if contentPath != "" {
			if !filepath.IsAbs(contentPath) {
				return config, nil, fmt.Errorf("content path must be absolute")
			}
			contentPath = filepath.Clean(contentPath)
			sessionPath := managedSessionPath(contentPath)
			if previous, ok := sessionOwners[sessionPath]; ok && previous != key {
				return config, nil, fmt.Errorf("different sites must not share session storage")
			}
			sessionOwners[sessionPath] = key
			info, e := os.Stat(contentPath)
			if e != nil {
				return config, nil, e
			}
			for _, previous := range contentOwners {
				if previous.root != key && os.SameFile(previous.file, info) {
					return config, nil, fmt.Errorf("different sites must not share a content file")
				}
			}
			contentOwners = append(contentOwners, owner{key, info})
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
		mediaRoot, runtimeRoot := "", ""
		if contentPath != "" {
			mediaRoot = site.ManagedMediaRoot(contentPath)
			// Custom-app storage is stable runtime state beside admin sessions,
			// never inside the immutable release directory.
			runtimeRoot = site.ManagedRuntimeRoot(contentPath)
		}
		project, err := site.LoadWithMedia(root, content, mediaRoot, runtimeRoot)
		if err != nil {
			return config, nil, fmt.Errorf("site %s: %w", host, err)
		}
		if contentPath == "" {
			for name := range project.Apps {
				if name != "contact" && name != "blog" {
					return config, nil, fmt.Errorf("site %s: custom apps require a managed content snapshot; deploy this site first", host)
				}
			}
		}
		if logger != nil {
			for _, warning := range project.Warnings {
				logger.Printf("Warning: %s: %s", host, warning)
			}
		}
		robotsWarnings[key] = robotsCrawlWarning(project)
		public := site.NewProductionServer(project, logger)
		appSites = append(appSites, public)
		handlers[host] = public
		if len(project.Admin.Emails) > 0 {
			if contentPath == "" {
				return config, nil, fmt.Errorf("site %s: admin requires a managed content snapshot; deploy this site first", host)
			}
			store := &editstore.Store{Root: root, ContentPath: contentPath, ConfigPath: configPath, Emails: project.Admin.Emails, Publish: public.Publish}
			dashboard := editor.New(editor.Options{Emails: project.Admin.Emails, SessionPath: managedSessionPath(contentPath), SessionIdleDays: project.Admin.SessionIdleDays, Name: fmt.Sprint(project.App["site_name"]), Sender: sender, Backend: store, Logger: logger, Render: public.RenderAdmin, RenderEmail: public.RenderEmail})
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
	config.indexingWarnings = map[string][]string{}
	for domain, mode := range config.Indexing {
		if warning := robotsWarnings[filepath.Clean(config.Sites[domain])]; mode == "noindex" && warning != "" {
			config.indexingWarnings[domain] = []string{warning}
			if logger != nil {
				logger.Printf("Warning: %s: %s", domain, warning)
			}
		}
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == HealthPath {
			w.Header().Set("Cache-Control", "no-store")
			w.Header().Set("X-Flea-Revision", config.Revision)
			if _, ok := handlers[requestHost(r)]; ok {
				w.Header().Set("X-Flea-Site", requestHost(r))
			}
			w.Header().Set("Content-Type", "application/json")
			if r.Method != "GET" && r.Method != "HEAD" {
				w.Header().Set("Allow", "GET, HEAD")
				w.WriteHeader(405)
				return
			}
			if r.Method == "GET" {
				_, _ = io.WriteString(w, `{"service":"flea","status":"ready"}`)
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
			if store := downloadHandlers[host]; store != nil &&
				(r.URL.Path == "/releases" || strings.HasPrefix(r.URL.Path, "/releases/")) && r.URL.Path != "/releases/latest.txt" {
				store.ServeHTTP(w, r)
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
	return config, &runtimeHandler{Handler: indexingHandler(config, handler), sites: appSites, sender: sender, budget: contactBudget}, nil
}

type runtimeHandler struct {
	http.Handler
	sites  []*site.Server
	sender mailer.Sender
	budget *contact.Budget
}

func (h *runtimeHandler) start(ctx context.Context) func() {
	var stops []func()
	for _, s := range h.sites {
		stops = append(stops, s.StartApps(ctx, h.sender, h.budget))
	}
	return func() {
		for _, stop := range stops {
			stop()
		}
	}
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
	return serveWithListeners(ctx, config, handler, logger, testTLS, nil)
}

func serveWithListeners(ctx context.Context, config Config, handler http.Handler, logger *log.Logger, testTLS *tls.Config, bound *listeners) error {
	if err := initializeSessions(config); err != nil {
		return err
	}
	httpHandler, tlsConfig, err := transports(config, handler)
	if err != nil {
		return err
	}
	if testTLS != nil {
		tlsConfig = testTLS
	}
	if bound == nil {
		bound = &listeners{}
		bound.http, err = net.Listen("tcp", config.Listen)
		if err != nil {
			return err
		}
		if tlsConfig != nil {
			bound.https, err = net.Listen("tcp", config.TLS.Listen)
			if err != nil {
				bound.close()
				return fmt.Errorf("HTTPS listen: %w", err)
			}
		}
	}
	defer bound.close()
	if runtime, ok := handler.(*runtimeHandler); ok {
		stop := runtime.start(ctx)
		defer stop()
	}
	newServer := func(h http.Handler) *http.Server {
		return &http.Server{Handler: h, ErrorLog: logger, ReadHeaderTimeout: 10 * time.Second, ReadTimeout: 30 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	}
	server := newServer(httpHandler)
	servers := []*http.Server{server}
	done := make(chan error, 2)
	if bound.https != nil {
		if tlsConfig == nil {
			// Systemd retains :443 even with no HTTPS domains. Reject handshakes
			// promptly instead of leaving clients queued or serving plain HTTP.
			tlsConfig = &tls.Config{MinVersion: tls.VersionTLS12, GetCertificate: func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
				return nil, fmt.Errorf("HTTPS is not configured")
			}}
		}
		s := newServer(tlsHandler(config, handler))
		s.TLSConfig = tlsConfig
		servers = append(servers, s)
		go func() { done <- s.ServeTLS(bound.https, "", "") }()
	}
	go func() { done <- server.Serve(bound.http) }()
	if logger != nil {
		logger.Printf("ItsyWeb server listening on %s", bound.http.Addr())
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

// Managed content is <site>/content/<revision>.json; sessions belong beside
// content, not within either the revision or its immutable release directory.
func managedSessionPath(contentPath string) string {
	return sitefiles.Path(filepath.Dir(filepath.Dir(contentPath)), filepath.Join("itsy", "runtime", "sessions.json"))
}

func initializeSessions(config Config) error {
	seen := map[string]bool{}
	for hostname, root := range config.Sites {
		content := config.Content[hostname]
		if content == "" || seen[root] {
			continue
		}
		seen[root] = true
		project, err := site.Load(root)
		if err != nil {
			return err
		}
		if err := editor.PruneSessions(editor.Options{SessionPath: managedSessionPath(content), Emails: project.Admin.Emails, SessionIdleDays: project.Admin.SessionIdleDays}); err != nil {
			return fmt.Errorf("site %s: session storage: %w", hostname, err)
		}
	}
	return nil
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
	if response.StatusCode != 200 || string(body) != `{"service":"flea","status":"ready"}` {
		return fmt.Errorf("ItsyWeb server readiness check failed")
	}
	if response.Header.Get("X-Flea-Revision") != config.Revision || (domain != "" && response.Header.Get("X-Flea-Site") != domain) {
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
