package site

import (
	"flea/templates"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestEmailTemplates(t *testing.T) {
	root := t.TempDir()
	if err := os.CopyFS(root, templates.Files); err != nil {
		t.Fatal(err)
	}
	if err := SetSiteName(root, "My & Company"); err != nil {
		t.Fatal(err)
	}
	p := mustLoad(t, root)
	mail, err := p.RenderEmail("contact", map[string]string{"name": "<script>bad</script>", "address": "visitor@example.com", "message": "First\nSecond {app.site_name}"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(mail.HTML, "<script>") || !strings.Contains(mail.HTML, "&lt;script&gt;") || !strings.Contains(mail.HTML, "My &amp; Company") || !strings.Contains(mail.Text, "First\nSecond {app.site_name}") {
		t.Fatal(mail)
	}
	if mail.To != "" || mail.ReplyTo != "" || mail.Subject != "" {
		t.Fatal("templates choose headers")
	}
	server := NewServer(p)
	before := fingerprint(root)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, httptest.NewRequest("GET", "/mailers/sign-in.email.html", nil))
	if w.Code != 404 {
		t.Fatal("email template public")
	}
	put(t, root, "mailers/sign-in.email.html", `<h1>Updated {email.code}</h1>`)
	if fingerprint(root) == before {
		t.Fatal("mailer edit invisible to watcher")
	}
	if err := server.Reload(root); err != nil {
		t.Fatal(err)
	}
	mail, err = server.RenderEmail("sign-in", map[string]string{"code": "123456"})
	if err != nil || !strings.Contains(mail.HTML, "Updated 123456") {
		t.Fatal(mail, err)
	}
	if _, err := p.RenderEmail("../../etc/passwd", nil); err == nil {
		t.Fatal("unknown template accepted")
	}
	put(t, root, "mailers/sign-in.email.html", `[[MAIL . MAIL]]`)
	if _, err := Load(root); err == nil {
		t.Fatal("reserved delimiter accepted")
	}
}
