package host

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeIndexingConfig(t *testing.T, file string, c Config) {
	t.Helper()
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(file, data, 0644); err != nil {
		t.Fatal(err)
	}
}

func TestSEOSharedHostIndexingIsolation(t *testing.T) {
	file := testConfig(t, ":80")
	c, _, err := Load(file, nil)
	if err != nil {
		t.Fatal(err)
	}
	root := c.Sites["example.test"]
	c.Sites["staging.test"] = root
	c.Indexing = map[string]string{"staging.test": "noindex"}
	for name, body := range map[string]string{
		"itsy.config.json":  `{"app":{},"seo":{"url":"https://example.test","exclude":["/receipt"]}}`,
		"site/index.html":   `<html><head><title>Home</title></head><body>Home</body></html>`,
		"site/receipt.html": "Receipt",
	} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	writeIndexingConfig(t, file, c)
	_, handler, err := Load(file, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, method := range []string{"GET", "HEAD"} {
		for _, host := range []string{"example.test", "staging.test", "unknown.test"} {
			for _, route := range []string{"/", "/receipt", "/robots.txt", "/sitemap.xml"} {
				r := httptest.NewRequest(method, "https://"+host+route, nil)
				r.Header.Set("X-Forwarded-Host", "attacker.test")
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, r)
				if host == "unknown.test" {
					if w.Code != 404 || strings.Contains(w.Body.String(), "example.test") {
						t.Fatal("SEO bypassed host registration", w)
					}
					continue
				}
				wantRobots := ""
				if host == "staging.test" {
					wantRobots = "noindex, nofollow"
				} else if route == "/receipt" {
					wantRobots = "noindex"
				}
				if w.Code != 200 || w.Header().Get("X-Robots-Tag") != wantRobots || (method == "HEAD" && w.Body.Len() != 0) {
					t.Fatalf("%s %s%s: %d %s", method, host, route, w.Code, w.Header())
				}
				if method == "GET" && route != "/receipt" {
					body := w.Body.String()
					if !strings.Contains(body, "https://example.test/") || strings.Contains(body, "https://staging.test") || strings.Contains(body, "attacker.test") || strings.Contains(body, "Disallow: /\n") {
						t.Fatal("SEO selected the wrong origin or blocked crawling", body)
					}
				}
			}
		}
	}
}

func TestIndexingSharedRootAndResponsePaths(t *testing.T) {
	file := testConfig(t, ":80")
	c, _, err := Load(file, nil)
	if err != nil {
		t.Fatal(err)
	}
	root := c.Sites["example.test"]
	c.Sites["staging.test"] = root
	c.Sites["www.staging.test"] = root
	c.Sites["unconfigured.test"] = root
	c.Redirects = map[string]string{"www.staging.test": "staging.test"}
	c.Indexing = map[string]string{"staging.test": "noindex", "www.staging.test": "noindex", "example.test": "allow"}
	store := downloadFixture(t)
	if err := os.Chmod(store, 0755); err != nil {
		t.Fatal(err)
	}
	c.Downloads = map[string]string{}
	for d := range c.Sites {
		c.Downloads[d] = store
	}
	robots := "User-agent: *\nDisallow:\n"
	for name, body := range map[string]string{"robots.txt": robots, "asset.txt": "asset", "large.bin": strings.Repeat("a", 128<<10)} {
		if err := os.WriteFile(filepath.Join(root, "site", name), []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	writeIndexingConfig(t, file, c)
	loaded, handler, err := Load(file, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(handler.(*runtimeHandler).sites) != 1 {
		t.Fatal("aliases did not share a site handler")
	}
	if len(loaded.IndexingWarnings("staging.test")) != 0 {
		t.Fatal("crawlable robots.txt diagnosed as blocked")
	}
	for round := 0; round < 3; round++ {
		for _, method := range []string{"GET", "HEAD"} {
			for _, host := range []string{"staging.test", "example.test", "STAGING.TEST.:8080", "unconfigured.test", "unknown.test", "staging.test.evil", "staging.test:bad:port"} {
				for _, tc := range []struct {
					path   string
					status int
				}{
					{"/", 200}, {"/asset.txt", 200}, {"/large.bin", 206}, {"/robots.txt", 200}, {"/missing", 404}, {"/releases/0.7.0/flea-linux-amd64", 200}, {"/releases/missing", 404}, {HealthPath, 200},
				} {
					r := httptest.NewRequest(method, tc.path, nil)
					r.Host = host
					r.Header.Set("X-Forwarded-Host", "staging.test")
					r.Header.Set("Forwarded", "host=staging.test;proto=https")
					if tc.path == "/large.bin" {
						r.Header.Set("Range", "bytes=2-4")
					}
					w := httptest.NewRecorder()
					handler.ServeHTTP(w, r)
					response := w.Result()
					registered := c.Sites[requestHost(r)] != ""
					wantStatus := tc.status
					if !registered && tc.path != HealthPath {
						wantStatus = 404
					}
					want := ""
					if c.Indexing[requestHost(r)] == "noindex" && tc.path != HealthPath {
						want = "noindex, nofollow"
					}
					if response.StatusCode != wantStatus || response.Header.Get("X-Robots-Tag") != want {
						t.Fatalf("%s %s %s: status %d, header %q; want %d %q", method, host, tc.path, response.StatusCode, response.Header.Get("X-Robots-Tag"), wantStatus, want)
					}
					if method == "HEAD" && w.Body.Len() != 0 && registered && wantStatus < 400 {
						t.Fatalf("HEAD body: %q", w.Body.String())
					}
					if tc.path == "/robots.txt" && method == "GET" && registered && w.Body.String() != robots {
						t.Fatal("robots.txt changed")
					}
					if tc.path == "/large.bin" && method == "GET" && registered && (w.Body.String() != "aaa" || response.Header.Get("Content-Range") == "") {
						t.Fatal("range changed")
					}
					if tc.path == HealthPath && (response.Header.Get("Cache-Control") != "no-store" || (method == "GET" && w.Body.String() != `{"service":"flea","status":"ready"}`)) {
						t.Fatal("health contract changed")
					}
				}
			}
		}
	}
	// Forwarding headers must not remove a staging restriction either.
	r := httptest.NewRequest("GET", "/", nil)
	r.Host = "staging.test"
	r.Header.Set("X-Forwarded-Host", "example.test")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Result().Header.Get("X-Robots-Tag") != "noindex, nofollow" {
		t.Fatal("forwarded host bypassed restriction")
	}
	// Both redirect boundaries preserve status, escaped path and query.
	for _, method := range []string{"GET", "HEAD"} {
		for _, secure := range []bool{false, true} {
			r := httptest.NewRequest(method, "/hello%2Fworld?q=1", nil)
			r.Host = "www.staging.test"
			h := handler
			scheme := "http"
			if secure {
				c.TLS = &TLSConfig{Domains: []string{"www.staging.test", "staging.test"}}
				h = redirectHTTPS(c, handler)
				scheme = "https"
			}
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			if w.Code != 308 || w.Result().Header.Get("Location") != scheme+"://staging.test/hello%2Fworld?q=1" || w.Result().Header.Get("X-Robots-Tag") != "noindex, nofollow" {
				t.Fatal(w.Code, w.Result().Header)
			}
		}
	}
	// Host/SNI mismatch keeps the existing rejection boundary.
	r = httptest.NewRequest("GET", "/", nil)
	r.Host = "staging.test"
	r.TLS = &tls.ConnectionState{ServerName: "example.test"}
	w = httptest.NewRecorder()
	tlsHandler(c, handler).ServeHTTP(w, r)
	if w.Code != 421 {
		t.Fatal("SNI mismatch accepted")
	}
}

func TestIndexingCannotBeClearedAndStreams(t *testing.T) {
	c := Config{Sites: map[string]string{"stage.test": "/site", "live.test": "/site"}, Indexing: map[string]string{"stage.test": "noindex", "live.test": "allow"}}
	for _, mode := range []string{"write", "header", "flush", "controller", "empty"} {
		t.Run(mode, func(t *testing.T) {
			h := indexingHandler(c, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Del("X-Robots-Tag")
				w.Header()["x-robots-tag"] = []string{"index, follow"}
				w.Header().Set("X-Robots-Tag", "noarchive")
				w.Header().Set("Cache-Control", "public, max-age=60")
				w.Header().Set("Content-Type", "text/plain")
				switch mode {
				case "header":
					w.WriteHeader(500)
				case "flush":
					w.(http.Flusher).Flush()
				case "controller":
					if err := http.NewResponseController(w).Flush(); err != nil {
						t.Error(err)
					}
				case "empty":
					return
				}
				_, _ = io.WriteString(w, "first")
			}))
			for _, domain := range []string{"stage.test", "live.test"} {
				r := httptest.NewRequest("GET", "/", nil)
				r.Host = domain
				w := httptest.NewRecorder()
				h.ServeHTTP(w, r)
				header := w.Result().Header
				want := "noarchive"
				if domain == "stage.test" {
					want = "noindex, nofollow"
				}
				if header.Get("X-Robots-Tag") != want || header.Get("Cache-Control") != "public, max-age=60" || header.Get("Content-Type") != "text/plain" {
					t.Fatal(header)
				}
			}
		})
	}
	// Real HTTP streaming: first bytes must arrive while the handler is blocked.
	release := make(chan struct{})
	h := indexingHandler(c, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Del("X-Robots-Tag")
		w.(http.Flusher).Flush()
		_, _ = io.WriteString(w, "first\n")
		_ = http.NewResponseController(w).Flush()
		<-release
		_, _ = io.WriteString(w, "last\n")
	}))
	server := httptest.NewServer(h)
	defer server.Close()
	defer close(release)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	r, _ := http.NewRequestWithContext(ctx, "GET", server.URL, nil)
	r.Host = "stage.test"
	response, err := server.Client().Do(r)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	first := make([]byte, 6)
	if _, err = io.ReadFull(response.Body, first); err != nil || string(first) != "first\n" || response.Header.Get("X-Robots-Tag") != "noindex, nofollow" {
		t.Fatal(string(first), err, response.Header)
	}
}

func TestIndexingValidationAndRobotsDiagnostic(t *testing.T) {
	file := testConfig(t, ":80")
	c, _, err := Load(file, nil)
	if err != nil {
		t.Fatal(err)
	}
	for domain, mode := range map[string]string{"unknown.test": "noindex", "EXAMPLE.TEST": "noindex", "example.test:80": "noindex", "*.example.test": "noindex", "example.test.": "noindex", "example.test": "index"} {
		invalid := c
		invalid.Indexing = map[string]string{domain: mode}
		writeIndexingConfig(t, file, invalid)
		if _, _, err = Load(file, nil); err == nil {
			t.Fatalf("accepted %s=%s", domain, mode)
		}
	}
	c.Indexing = map[string]string{"example.test": "noindex"}
	robots := "User-Agent: *\nDisallow: / # private staging\n"
	if err = os.WriteFile(filepath.Join(c.Sites["example.test"], "site", "robots.txt"), []byte(robots), 0644); err != nil {
		t.Fatal(err)
	}
	writeIndexingConfig(t, file, c)
	var logs bytes.Buffer
	loaded, handler, err := Load(file, log.New(&logs, "", 0))
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.IndexingWarnings("example.test")) != 1 || !strings.Contains(logs.String(), "cannot discover the noindex header") {
		t.Fatal(logs.String())
	}
	w := hostGet(handler, "example.test", "/robots.txt")
	if w.Body.String() != robots {
		t.Fatal("robots was rewritten")
	}
}

func TestIndexingCustomAppAndError(t *testing.T) {
	if testing.Short() {
		t.Skip("builds and launches sample app")
	}
	binary := buildSample(t)
	dir := t.TempDir()
	root, content := sampleRelease(t, filepath.Join(dir, "sites", "example"), "first", binary, `<p>{session.data.authenticated}</p>`)
	file := hostConfigFile(t, dir, map[string]string{"demo.test": root, "live.test": root}, map[string]string{"demo.test": content, "live.test": content})
	c, _, err := Load(file, nil)
	if err != nil {
		t.Fatal(err)
	}
	c.Indexing = map[string]string{"demo.test": "noindex"}
	writeIndexingConfig(t, file, c)
	_, handler, err := Load(file, nil)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	halt := handler.(*runtimeHandler).start(ctx)
	defer halt()
	waitAppReady(t, handler, "demo.test")
	for _, domain := range []string{"demo.test", "live.test"} {
		for _, method := range []string{"GET", "HEAD"} {
			r := httptest.NewRequest(method, "https://"+domain+"/session", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)
			want := ""
			if domain == "demo.test" {
				want = "noindex, nofollow"
			}
			if w.Code != 200 || w.Result().Header.Get("X-Robots-Tag") != want {
				t.Fatal(w.Code, w.Header())
			}
		}
	}
	// Stop the child and exercise the actual app-error response through the host.
	halt()
	w := hostGet(handler, "demo.test", "/session")
	if w.Code != 503 || w.Result().Header.Get("X-Robots-Tag") != "noindex, nofollow" {
		t.Fatal(w.Code, w.Header(), w.Body.String())
	}
}

func TestIndexingInformationalResponse(t *testing.T) {
	c := Config{Sites: map[string]string{"stage.test": "/site"}, Indexing: map[string]string{"stage.test": "noindex"}}
	server := httptest.NewServer(indexingHandler(c, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Del("X-Robots-Tag")
		w.WriteHeader(103)
		w.Header().Set("X-Robots-Tag", "index")
		w.WriteHeader(500)
	})))
	defer server.Close()
	r, _ := http.NewRequest("GET", server.URL, nil)
	r.Host = "stage.test"
	response, err := server.Client().Do(r)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != 500 || response.Header.Get("X-Robots-Tag") != "noindex, nofollow" {
		t.Fatal(response.StatusCode, response.Header)
	}
}
