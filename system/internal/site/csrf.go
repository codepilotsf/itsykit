package site

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"

	"golang.org/x/net/html"
)

const csrfAction = `[[NW csrfField $ NW]]`

// Check the expanded source, including partials, before handing it to Go's
// contextual escaper. Returning trusted markup is allowed only as HTML content.
func checkCSRFContexts(source string) error {
	if !strings.Contains(source, csrfAction) && !strings.Contains(source, blogBodyAction) {
		return nil
	}
	z := html.NewTokenizer(strings.NewReader(source))
	var blocked []string
	for {
		kind := z.Next()
		raw := string(z.Raw())
		if (strings.Contains(raw, csrfAction) || strings.Contains(raw, blogBodyAction)) && (kind != html.TextToken || len(blocked) != 0) {
			return fmt.Errorf("rich blog body and {csrf} require ordinary HTML element content, not an attribute or raw-text context")
		}
		switch kind {
		case html.ErrorToken:
			if z.Err() != io.EOF {
				return z.Err()
			}
			return nil
		case html.StartTagToken:
			name, _ := z.TagName()
			tag := string(name)
			if strings.Contains(" script style textarea title xmp iframe noembed noframes plaintext noscript svg math select option ", " "+tag+" ") {
				blocked = append(blocked, tag)
			}
		case html.EndTagToken:
			name, _ := z.TagName()
			if len(blocked) > 0 && blocked[len(blocked)-1] == string(name) {
				blocked = blocked[:len(blocked)-1]
			}
		}
	}
}

type renderInputs struct {
	csrf    func() (string, error)
	appName string
	appData map[string]any
}

const appRenderLimit = 2 << 20

type appRenderBuffer struct{ bytes.Buffer }

func (b *appRenderBuffer) Write(data []byte) (int, error) {
	if len(data) > appRenderLimit-b.Len() {
		return 0, fmt.Errorf("app page exceeds the 2 MiB rendering limit")
	}
	return b.Buffer.Write(data)
}

func (b *appRenderBuffer) WriteString(value string) (int, error) { return b.Write([]byte(value)) }

func csrfField(data map[string]any) (template.HTML, error) {
	inputs, _ := data["__request"].(*renderInputs)
	// Read-only validation and non-public renderers have no visitor context.
	if inputs == nil || inputs.csrf == nil {
		return "", nil
	}
	token, err := inputs.csrf()
	if err != nil {
		return "", err
	}
	return template.HTML(`<input type="hidden" name="_csrf" value="` + template.HTMLEscapeString(token) + `">`), nil
}

// One lazy token per render, even if multiple forms use the directive.
func (s *Server) formRender(r *http.Request) (*renderInputs, func(http.ResponseWriter)) {
	var used bool
	var token string
	var cookie *http.Cookie
	var err error
	inputs := &renderInputs{csrf: func() (string, error) {
		if !used {
			used = true
			token, cookie, err = s.forms.Token(r)
		}
		return token, err
	}}
	return inputs, func(w http.ResponseWriter) {
		if used && err == nil {
			w.Header().Set("Cache-Control", "private, no-store")
			if cookie != nil {
				http.SetCookie(w, cookie)
			}
		}
	}
}
