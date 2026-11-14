package host

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func testConfig(t *testing.T, listen string) string {
	t.Helper()
	dir := t.TempDir()
	root := filepath.Join(dir, "site")
	for path, body := range map[string]string{"itsy.config.json": `{"app":{"site_name":"Host test"},"dev":{"email_preview":true}}`, "site/index.html": "<h1>{app.site_name}</h1>", "site/_404.html": "Custom missing"} {
		file := filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	data, err := json.Marshal(Config{Listen: listen, Sites: map[string]string{"example.test": root}})
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(dir, "server.json")
	if err := os.WriteFile(file, data, 0644); err != nil {
		t.Fatal(err)
	}
	return file
}

func TestHostRoutesWithoutDevelopmentFeatures(t *testing.T) {
	file := testConfig(t, "127.0.0.1:8080")
	_, handler, err := Load(file, log.New(io.Discard, "", 0))
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		host, path string
		status     int
		body       string
	}{
		{"example.test", "/", 200, "Host test"}, {"EXAMPLE.TEST:8080", "/", 200, "Host test"},
		{"example.test", "/missing", 404, "Custom missing"}, {"example.test", "/_flea/reload", 404, "Custom missing"},
		{"example.test", "/_flea/email-preview/123456", 404, "Custom missing"},
		{"other.test", "/", 404, "404 page not found"}, {"localhost", HealthPath, 200, `"status":"ready"`},
	} {
		r := httptest.NewRequest("GET", test.path, nil)
		r.Host = test.host
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		if w.Code != test.status || !strings.Contains(w.Body.String(), test.body) || strings.Contains(w.Body.String(), "setInterval") {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	for _, body := range []string{`{"listen":"example.com:80"}`, `{"listen":":0"}`, `{"listen":":65536"}`, `{"listen":":http"}`, `{"listen":"80"}`, `{"listen":"127.0.0.1:8080","unknown":1}`, `{"listen":"127.0.0.1:8080"} {}`} {
		if err := os.WriteFile(file, []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
		if _, _, err := Load(file, nil); err == nil {
			t.Fatal("accepted invalid config", body)
		}
	}
}

func TestHostReadinessAndShutdown(t *testing.T) {
	for _, bind := range []string{"127.0.0.1", "", "0.0.0.0"} {
		t.Run("bind="+bind, func(t *testing.T) { testReadinessAndShutdown(t, bind) })
	}
}

func TestPublicListeners(t *testing.T) {
	for listen, probe := range map[string]string{":80": "127.0.0.1:80", "0.0.0.0:80": "127.0.0.1:80", "[::]:80": "[::1]:80", "192.0.2.1:80": "192.0.2.1:80", "[::1]:8080": "[::1]:8080"} {
		if _, _, err := Load(testConfig(t, listen), nil); err != nil {
			t.Fatalf("%s: %v", listen, err)
		}
		if got := checkAddress(listen); got != probe {
			t.Fatalf("probe %s: got %s, want %s", listen, got, probe)
		}
	}
}

func testReadinessAndShutdown(t *testing.T, bind string) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	listener.Close()
	_, port, _ := net.SplitHostPort(address)
	file := testConfig(t, net.JoinHostPort(bind, port))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- Run(ctx, file, nil) }()
	deadline := time.Now().Add(5 * time.Second)
	for {
		err = Check(ctx, file)
		if err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal(err)
		}
		time.Sleep(20 * time.Millisecond)
	}
	response, err := http.Get("http://" + address + HealthPath)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(6 * time.Second):
		t.Fatal("server did not stop")
	}
}

func TestRemovedThemeMappingIsRejectedWithoutRewritingConfig(t *testing.T) {
	file := testConfig(t, "127.0.0.1:8080")
	original, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var config map[string]any
	if err = json.Unmarshal(original, &config); err != nil {
		t.Fatal(err)
	}
	config["themes"] = map[string]string{"example.test": "/private/saved-theme.css"}
	data, err := json.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(file, data, 0644); err != nil {
		t.Fatal(err)
	}
	if _, _, err = Load(file, nil); err == nil || !strings.Contains(err.Error(), "themes") {
		t.Fatal("removed theme mapping accepted", err)
	}
	after, err := os.ReadFile(file)
	if err != nil || string(after) != string(data) {
		t.Fatal("rejected config was modified", err)
	}
}
