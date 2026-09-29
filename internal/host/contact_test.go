package host

import (
	"context"
	"encoding/json"
	"flea/internal/mailer"
	"flea/templates"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type contactSender struct{ messages []mailer.Message }

func (s *contactSender) Send(_ context.Context, m mailer.Message) error {
	s.messages = append(s.messages, m)
	return nil
}
func TestContactSiteIsolation(t *testing.T) {
	dir := t.TempDir()
	sites := map[string]string{}
	for _, name := range []string{"one", "two"} {
		root := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Join(root, "site"), 0755); err != nil {
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
		for path, body := range map[string]string{"site/index.html": "Hello", ".flea.config.json": `{"app":{"site_name":"test"},"contact":{"to":"` + name + `@example.test"}}`} {
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
	for _, tc := range []struct {
		host, origin, extra string
		status              int
	}{{"one.test", "one.test", "", 200}, {"two.test", "two.test", "", 200}, {"two.test", "one.test", "", 403}, {"one.test", "one.test", "&to=two@example.test", 400}, {"unknown.test", "unknown.test", "", 404}, {"alias.test", "alias.test", "", 200}, {"one.test", "one.test", "", 200}, {"alias.test", "alias.test", "", 429}} {
		r := httptest.NewRequest("POST", "https://"+tc.host+"/_flea/contact", strings.NewReader("name=Visitor&email=v%40example.test&message=Hi"+tc.extra))
		r.Header.Set("Origin", "https://"+tc.origin)
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != tc.status {
			t.Fatalf("%s: %d %s", tc.host, w.Code, w.Body.String())
		}
	}
	if len(sender.messages) != 4 || sender.messages[0].To != "one@example.test" || sender.messages[1].To != "two@example.test" || sender.messages[2].To != "one@example.test" {
		t.Fatal(sender.messages)
	}
}
