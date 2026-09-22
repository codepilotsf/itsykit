package admin

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"flea/internal/host"
	"flea/internal/release"
)

func fixture(t *testing.T) (Manager, Options) {
	t.Helper()
	dir := t.TempDir()
	data := filepath.Join(dir, "data")
	os.Mkdir(data, 0755)
	config := filepath.Join(dir, "server.json")
	os.WriteFile(config, []byte(`{"listen":"127.0.0.1:8080","sites":{}}`), 0644)
	m := Manager{ConfigPath: config, DataDir: data, Restart: func(context.Context) error { return nil }, Verify: func(context.Context, host.Config, []string) error { return nil }}
	return m, Options{Archive: archive(t, "one"), Name: "demo", Domains: []string{"example.test"}}
}

func archive(t *testing.T, title string) string {
	t.Helper()
	root := t.TempDir()
	for name, body := range map[string]string{"_/flea.config": `{"site":{"name":"` + title + `"}}`, "index.html": "<h1>{site.name}</h1>", "style.css": "body{}"} {
		file := filepath.Join(root, name)
		os.MkdirAll(filepath.Dir(file), 0755)
		os.WriteFile(file, []byte(body), 0644)
	}
	file := filepath.Join(t.TempDir(), "site.tar.gz")
	f, err := os.Create(file)
	if err != nil {
		t.Fatal(err)
	}
	if err := release.Pack(root, f); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return file
}

func TestActivateRedeployAliasesAndCollision(t *testing.T) {
	m, o := fixture(t)
	ctx := context.Background()
	o.Domains = append(o.Domains, "www.example.test")
	first, err := m.Apply(ctx, o)
	if err != nil {
		t.Fatal(err)
	}
	c, _, err := host.Load(m.ConfigPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	firstRoot := c.Sites["example.test"]
	o.Archive = archive(t, "two")
	o.Domains = o.Domains[:1]
	second, err := m.Apply(ctx, o)
	if err != nil || first == second {
		t.Fatal(second, err)
	}
	c, _, err = host.Load(m.ConfigPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	if c.Sites["example.test"] == firstRoot || c.Sites["www.example.test"] != c.Sites["example.test"] {
		t.Fatal(c)
	}
	if _, err := os.Stat(firstRoot); err != nil {
		t.Fatal("old release lost", err)
	}
	o.Name = "other"
	if _, err = m.Apply(ctx, o); err == nil {
		t.Fatal("replaced another site's hostname")
	}
}

func TestBranchOwnership(t *testing.T) {
	m, o := fixture(t)
	o.Branch = "main"
	o.Commit = strings.Repeat("a", 40)
	if _, err := m.Apply(context.Background(), o); err != nil {
		t.Fatal(err)
	}
	c, _, err := host.Load(m.ConfigPath, nil)
	if err != nil || c.SiteBranches[o.Name] != "main" || c.Commits[o.Name] != o.Commit {
		t.Fatal(c, err)
	}
	o.Branch = "staging"
	if _, err := m.Apply(context.Background(), o); err == nil {
		t.Fatal("staging replaced main")
	}
	o.Name = "staging"
	if _, err := m.Apply(context.Background(), o); err == nil {
		t.Fatal("staging took main hostname")
	}
}

func TestSavedHTTPSPreferenceKeepsOtherSites(t *testing.T) {
	m, o := fixture(t)
	os.WriteFile(m.ConfigPath, []byte(`{"listen":":80","sites":{}}`), 0644)
	o.HTTPS = true
	o.Email = "dev@example.test"
	o.AgreeTOS = true
	o.Domains = []string{"example.test", "www.example.test"}
	o.Canonical = "example.test"
	if _, err := m.Apply(context.Background(), o); err != nil {
		t.Fatal(err)
	}
	other := o
	other.Name = "other"
	other.Domains = []string{"other.test"}
	other.Canonical = ""
	if _, err := m.Apply(context.Background(), other); err != nil {
		t.Fatal(err)
	}
	o.HTTPS = false
	if _, err := m.Apply(context.Background(), o); err != nil {
		t.Fatal(err)
	}
	c, _, err := host.Load(m.ConfigPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	if c.TLS == nil || len(c.TLS.Domains) != 1 || c.TLS.Domains[0] != "other.test" {
		t.Fatal("TLS preference affected other site", c.TLS)
	}
	if c.Redirects["www.example.test"] != "example.test" {
		t.Fatal("WWW preference lost")
	}
}

func TestActivationRollbackAndRecovery(t *testing.T) {
	for _, failure := range []string{"restart", "verify", "recovery"} {
		t.Run(failure, func(t *testing.T) {
			m, o := fixture(t)
			old, _ := os.ReadFile(m.ConfigPath)
			calls := 0
			m.Restart = func(context.Context) error {
				calls++
				if failure == "recovery" || failure == "restart" && calls == 1 {
					return errors.New("restart failure")
				}
				return nil
			}
			m.Verify = func(_ context.Context, c host.Config, _ []string) error {
				if c.Revision != "" {
					return errors.New("probe failure")
				}
				return nil
			}
			if _, err := m.Apply(context.Background(), o); err == nil {
				t.Fatal("failure reported as success")
			}
			got, _ := os.ReadFile(m.ConfigPath)
			if string(got) != string(old) {
				t.Fatal("old config not restored")
			}
			if failure == "recovery" {
				if _, err := os.Stat(m.pending()); err != nil {
					t.Fatal("recovery journal lost")
				}
				if _, err := m.Apply(context.Background(), o); err == nil {
					t.Fatal("ignored pending recovery")
				}
				m.Restart = func(context.Context) error { return nil }
				if err := m.Recover(context.Background()); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := os.Stat(m.pending()); !os.IsNotExist(err) {
				t.Fatal("pending recovery remains", err)
			}
		})
	}
}

func TestConcurrentActivationAndSymlink(t *testing.T) {
	m, o := fixture(t)
	unlock, err := m.lock()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = m.Apply(context.Background(), o); err == nil {
		t.Fatal("ignored deployment lock")
	}
	unlock()
	os.Symlink(t.TempDir(), filepath.Join(m.DataDir, "sites"))
	if _, err = m.Apply(context.Background(), o); err == nil {
		t.Fatal("followed symlink")
	}
}

func TestHTTPSRequiresConsentAndExistingDomain(t *testing.T) {
	m, _ := fixture(t)
	for _, o := range []Options{{HTTPS: true, Domains: []string{"example.test"}}, {HTTPS: true, Domains: []string{"127.0.0.1"}, Email: "a@example.com", AgreeTOS: true}, {HTTPS: true, Domains: []string{"example.test"}, Email: "a@example.com", AgreeTOS: true}} {
		if _, err := m.Apply(context.Background(), o); err == nil {
			t.Fatal("accepted invalid HTTPS request", o)
		}
	}
}

func TestDeploymentHTTPWorkflow(t *testing.T) {
	m, o := fixture(t)
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := l.Addr().String()
	l.Close()
	data, _ := json.Marshal(host.Config{Listen: address, Sites: map[string]string{}})
	os.WriteFile(m.ConfigPath, data, 0644)
	var stop context.CancelFunc
	var done chan error
	m.Restart = func(context.Context) error {
		if stop != nil {
			stop()
			if err := <-done; err != nil {
				return err
			}
		}
		var ctx context.Context
		ctx, stop = context.WithCancel(context.Background())
		done = make(chan error, 1)
		go func() { done <- host.Run(ctx, m.ConfigPath, nil) }()
		return nil
	}
	t.Cleanup(func() {
		if stop != nil {
			stop()
			<-done
		}
	})
	m.Verify = verify
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	for _, title := range []string{"one", "two"} {
		o.Archive = archive(t, title)
		if _, err := m.Apply(ctx, o); err != nil {
			t.Fatal(err)
		}
		for _, tc := range []struct {
			path   string
			status int
			body   string
		}{{"/", 200, title}, {"/style.css", 200, "body{}"}, {"/_/flea.config", 404, ""}, {"/index.html", 404, ""}} {
			req, _ := http.NewRequest("GET", "http://"+address+tc.path, nil)
			req.Host = "example.test"
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			body, _ := io.ReadAll(res.Body)
			res.Body.Close()
			if res.StatusCode != tc.status || !strings.Contains(string(body), tc.body) || strings.Contains(string(body), "setInterval") {
				t.Fatal(res.StatusCode, string(body))
			}
		}
	}
}
