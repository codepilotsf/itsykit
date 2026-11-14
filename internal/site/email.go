package site

import (
	"bytes"
	"fmt"
	htmltemplate "html/template"
	"itsy/internal/mailer"
	"os"
	"regexp"
	"strings"
	texttemplate "text/template"
)

type emailTemplates struct {
	html *htmltemplate.Template
	text *texttemplate.Template
}

type emailBuffer struct{ bytes.Buffer }

func (b *emailBuffer) WriteString(value string) (int, error) { return b.Write([]byte(value)) }

func (b *emailBuffer) Write(data []byte) (int, error) {
	if len(data) > 256<<10-b.Len() {
		return 0, fmt.Errorf("rendered email exceeds 256 KiB")
	}
	return b.Buffer.Write(data)
}

var emailValue = regexp.MustCompile(`\{((?:app|email)\.[A-Za-z][A-Za-z0-9_.]*)\}`)

func emailSource(source string) (string, error) {
	if len(source) > 128<<10 || strings.Contains(source, "[[MAIL") {
		return "", fmt.Errorf("email template exceeds limits or uses a reserved delimiter")
	}
	return emailValue.ReplaceAllString(source, `[[MAIL lookup . "$1" MAIL]]`), nil
}

func (p *Project) loadEmails(root *os.Root) error {
	p.emails = map[string]emailTemplates{}
	if _, err := root.Lstat("mailers"); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	layout, err := readFile(root, "mailers/main.layout.html")
	if err != nil {
		return err
	}
	if strings.Count(string(layout), "{slot}") != 1 {
		return fmt.Errorf("mailers/main.layout.html needs exactly one {slot}")
	}
	for _, kind := range []string{"sign-in", "contact"} {
		h, err := readFile(root, "mailers/"+kind+".email.html")
		if err != nil {
			return err
		}
		txt, err := readFile(root, "mailers/"+kind+".email.txt")
		if err != nil {
			return err
		}
		hs, err := emailSource(strings.Replace(string(layout), "{slot}", string(h), 1))
		if err != nil {
			return err
		}
		ts, err := emailSource(string(txt))
		if err != nil {
			return err
		}
		ht, err := htmltemplate.New(kind).Delims("[[MAIL", "MAIL]]").Funcs(htmltemplate.FuncMap{"lookup": lookup}).Parse(hs)
		if err != nil {
			return fmt.Errorf("mailers/%s.email.html: %w", kind, err)
		}
		tt, err := texttemplate.New(kind).Delims("[[MAIL", "MAIL]]").Funcs(texttemplate.FuncMap{"lookup": lookup}).Parse(ts)
		if err != nil {
			return fmt.Errorf("mailers/%s.email.txt: %w", kind, err)
		}
		p.emails[kind] = emailTemplates{ht, tt}
		if _, err := p.RenderEmail(kind, map[string]string{}); err != nil {
			return err
		}
	}
	return nil
}

// Templates control bodies only, never recipients, sender, headers or delivery.
func (p *Project) RenderEmail(kind string, values map[string]string) (mailer.Message, error) {
	t, ok := p.emails[kind]
	if !ok {
		return mailer.Message{}, fmt.Errorf("missing mailers/%s templates", kind)
	}
	email := map[string]any{}
	for k, v := range values {
		email[k] = v
	}
	data := map[string]any{"app": p.App, "email": email}
	var h, txt emailBuffer
	if err := t.html.Execute(&h, data); err != nil {
		return mailer.Message{}, err
	}
	if err := t.text.Execute(&txt, data); err != nil {
		return mailer.Message{}, err
	}
	if h.Len() > 256<<10 || txt.Len() > 256<<10 {
		return mailer.Message{}, fmt.Errorf("rendered email exceeds limits")
	}
	return mailer.Message{HTML: h.String(), Text: txt.String()}, nil
}
func (s *Server) RenderEmail(kind string, values map[string]string) (mailer.Message, error) {
	s.mu.RLock()
	p, err := s.project, s.err
	s.mu.RUnlock()
	if err != nil {
		return mailer.Message{}, err
	}
	return p.RenderEmail(kind, values)
}
