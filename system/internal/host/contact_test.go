package host

import (
	"context"
	"encoding/json"
	"itsy/internal/appbundle"
	"itsy/internal/mailer"
	"itsy/templates"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"
)

type contactSender struct {
	mu       sync.Mutex
	messages []mailer.Message
}

func (s *contactSender) Send(_ context.Context, m mailer.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = append(s.messages, m)
	return nil
}
func TestContactSiteIsolation(t *testing.T) {
	if !appbundle.Available() {
		if os.Getenv("FLEA_REQUIRE_APP_BUNDLE") == "1" {
			t.Fatal("Contact bundle missing")
		}
		t.Skip("requires scripts/build.sh")
	}
	dir := t.TempDir()
	sites := map[string]string{}
	for _, name := range []string{"one", "two"} {
		root := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Join(root, "site/contact"), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(root, "mailers"), 0755); err != nil {
			t.Fatal(err)
		}
		for _, file := range []string{"main.layout.html", "sign-in.email.html", "sign-in.email.txt", "contact.email.html", "contact.email.txt"} {
			data, err := templates.Files.ReadFile("mailers/" + file)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, "mailers", file), data, 0644); err != nil {
				t.Fatal(err)
			}
		}
		for path, body := range map[string]string{"site/index.html": "Hello", "site/contact/index.html": `<form>{csrf}</form>{contact.error}`, "site/contact/thank-you.html": "Thanks", "itsy.config.json": `{"app":{"site_name":"test"},"apps":{"contact":{"mount":"/contact","email":{"to":"` + name + `@example.test"}}}}`} {
			if err := os.WriteFile(filepath.Join(root, path), []byte(body), 0644); err != nil {
				t.Fatal(err)
			}
		}
		sites[name+".test"] = root
	}
	sites["alias.test"] = sites["one.test"]
	data, _ := json.Marshal(Config{Listen: "127.0.0.1:8080", Sites: sites})
	file := filepath.Join(dir, "server.json")
	if err := os.WriteFile(file, data, 0644); err != nil {
		t.Fatal(err)
	}
	sender := &contactSender{}
	_, h, err := load(file, nil, sender)
	if err != nil {
		t.Fatal(err)
	}

	runtime := h.(*runtimeHandler)
	if len(runtime.sites) != 2 {
		t.Fatal("aliases must share one site runtime")
	}
	// Loading only constructs handlers; app launch is explicit.
	stop := runtime.start(context.Background())
	defer stop()
	post := func(host, origin, body string) *httptest.ResponseRecorder {
		get := httptest.NewRecorder()
		h.ServeHTTP(get, httptest.NewRequest("GET", "https://"+host+"/contact", nil))
		match := regexp.MustCompile(`name="_csrf" value="([^"]+)"`).FindStringSubmatch(get.Body.String())
		token := ""
		if len(match) == 2 {
			token = match[1]
		}
		r := httptest.NewRequest("POST", "https://"+host+"/contact", strings.NewReader("_csrf="+token+"&"+body))
		for _, c := range get.Result().Cookies() {
			r.AddCookie(c)
		}
		r.Header.Set("Origin", "https://"+origin)
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	for _, host := range []string{"one.test", "two.test"} {
		deadline := time.Now().Add(8 * time.Second)
		for {
			w := post(host, host, "email=invalid")
			if w.Code == 422 {
				break
			}
			if time.Now().After(deadline) {
				t.Fatal("not ready", w.Code, w.Body.String())
			}
			time.Sleep(20 * time.Millisecond)
		}
	}
	for _, tc := range []struct {
		host, origin, extra string
		status              int
	}{
		{"one.test", "one.test", "", 303},
		{"two.test", "two.test", "", 303},
		{"two.test", "one.test", "", 403},
		{"one.test", "one.test", "&to=two@example.test", 422},
		{"unknown.test", "unknown.test", "", 404},
		{"alias.test", "alias.test", "", 303},
		{"one.test", "one.test", "", 303},
		{"alias.test", "alias.test", "", 429},
	} {
		w := post(tc.host, tc.origin, "name=Visitor&email=v%40example.test&message=Hi"+tc.extra)
		if w.Code != tc.status {
			t.Fatalf("%s: %d %s", tc.host, w.Code, w.Body.String())
		}
	}
	sender.mu.Lock()
	defer sender.mu.Unlock()
	if len(sender.messages) != 4 || sender.messages[0].To != "one@example.test" || sender.messages[1].To != "two@example.test" || sender.messages[2].To != "one@example.test" {
		t.Fatal(sender.messages)
	}
}
