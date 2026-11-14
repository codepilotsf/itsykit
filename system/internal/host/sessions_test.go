package host

import (
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"itsy/templates"
)

func TestManagedSessionsSurviveReleaseChange(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "sites", "example")
	write := func(name string, data []byte, mode os.FileMode) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(name), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(name, data, mode); err != nil {
			t.Fatal(err)
		}
	}
	token := strings.Repeat("a", 64)
	digest := sha256.Sum256([]byte(token))
	now := time.Now()
	record, _ := json.Marshal(map[string]any{hex.EncodeToString(digest[:]): map[string]any{
		"email": "owner@example.test", "host": "example.test", "csrf": strings.Repeat("b", 64), "last_active": now, "expires": now.Add(30 * 24 * time.Hour),
	}})
	sessions := filepath.Join(base, "itsy", "runtime", "sessions.json")
	write(sessions, record, 0600)
	var config Config
	var handler http.Handler
	for _, revision := range []string{"first", "second"} {
		root := filepath.Join(base, "releases", revision)
		if err := os.MkdirAll(root, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.CopyFS(root, templates.Files); err != nil {
			t.Fatal(err)
		}
		write(filepath.Join(root, "itsy.config.json"), []byte(`{"app":{"site_name":"Example"},"admin":{"emails":["owner@example.test"],"session_idle_days":30}}`), 0644)
		content := filepath.Join(base, "content", revision+".json")
		seed, err := os.ReadFile(filepath.Join(root, "itsy", "snips.json"))
		if err != nil {
			t.Fatal(err)
		}
		write(content, seed, 0644)
		config = Config{Listen: "127.0.0.1:8080", Sites: map[string]string{"example.test": root, "alias.example.test": root}, Content: map[string]string{"example.test": content, "alias.example.test": content}}
		data, _ := json.Marshal(config)
		file := filepath.Join(dir, "server.json")
		write(file, data, 0644)
		before, _ := os.ReadFile(sessions)
		_, handler, err = load(file, nil, nil) // Existing sessions do not need working email.
		if err != nil {
			t.Fatal(err)
		}
		after, _ := os.ReadFile(sessions)
		if !bytes.Equal(before, after) {
			t.Fatal("config validation modified session state")
		}
		if err := initializeSessions(config); err != nil {
			t.Fatal(err)
		}
		for _, host := range []string{"example.test", "alias.example.test"} {
			r := httptest.NewRequest("GET", "https://"+host+"/admin", nil)
			r.TLS = &tls.ConnectionState{}
			r.AddCookie(&http.Cookie{Name: "__Host-flea-session", Value: token})
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)
			if host == "example.test" {
				if w.Code != 200 || !strings.Contains(w.Body.String(), "Blog is not enabled") {
					t.Fatal(revision, w.Code, w.Body.String())
				}
			} else if strings.Contains(w.Body.String(), "Blog is not enabled") {
				t.Fatal("session crossed aliases")
			}
		}
		if _, err := os.Stat(filepath.Join(root, "itsy", "runtime", "sessions.json")); !os.IsNotExist(err) {
			t.Fatal("session state entered release")
		}
	}
	// Startup with all owners removed permanently prunes their sessions.
	write(filepath.Join(config.Sites["example.test"], "itsy.config.json"), []byte(`{"app":{"site_name":"Example"},"admin":{"emails":[]}}`), 0644)
	if err := initializeSessions(config); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(sessions)
	if string(data) != "{}\n" {
		t.Fatal("removed owners retained", string(data))
	}
}
