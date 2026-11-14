package site

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
)

func formToken(body string) string {
	_, rest, _ := strings.Cut(body, `name="_csrf" value="`)
	token, _, _ := strings.Cut(rest, `"`)
	return token
}

func TestCSRFRenderLifecycle(t *testing.T) {
	root := fixture(t)
	put(t, root, "index.html", `<form>{render form}{csrf}</form>`)
	put(t, root, "_form.html", `{csrf}`)
	p := mustLoad(t, root)
	plain, err := p.render(p.Pages["/"], "/")
	if err != nil || formToken(string(plain)) != "" {
		t.Fatal("offline render generated token", err)
	}
	s := NewProductionServer(p, nil)
	r := httptest.NewRequest("GET", "https://example.test/", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, r)
	token := formToken(w.Body.String())
	cookies := w.Result().Cookies()
	if w.Code != 200 || len(token) != 43 || len(cookies) != 1 || strings.Count(w.Body.String(), token) != 2 || w.Header().Get("Cache-Control") != "private, no-store" {
		t.Fatal(w.Code, w.Body.String(), w.Header())
	}
	q := httptest.NewRequest("POST", "https://example.test/contact", nil)
	q.Header.Set("Origin", "https://example.test")
	q.AddCookie(cookies[0])
	if err := s.forms.Verify(q, url.Values{"_csrf": {token}}); err != nil {
		t.Fatal(err)
	}
	if err := s.Reload(root); err != nil {
		t.Fatal(err)
	}
	if err := s.forms.Verify(q, url.Values{"_csrf": {token}}); err != nil {
		t.Fatal("reload invalidated token", err)
	}
	if NewProductionServer(p, nil).forms.Verify(q, url.Values{"_csrf": {token}}) == nil {
		t.Fatal("restart accepted token")
	}
	for _, route := range []string{"/about", "/css/main.css", "/missing"} {
		w := request(s, "GET", "https://example.test"+route)
		if len(w.Result().Cookies()) != 0 {
			t.Fatal("non-form issued cookie", route)
		}
	}
	if p.App["csrf"] != nil || p.App["current_admin"] != nil {
		t.Fatal("snapshot changed")
	}
	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() {
			r := httptest.NewRequest("GET", "https://example.test/", nil)
			r.AddCookie(cookies[0])
			w := httptest.NewRecorder()
			s.ServeHTTP(w, r)
			if w.Code != 200 || formToken(w.Body.String()) != token || len(w.Result().Cookies()) != 0 {
				t.Error("concurrent render changed visitor")
			}
		})
	}
	wg.Wait()
}

func TestCSRFContexts(t *testing.T) {
	for _, source := range []string{`<input value="{csrf}">`, `<input {csrf}>`, `<textarea>{csrf}</textarea>`, `<title>{csrf}</title>`, `<svg>{csrf}</svg>`, `<select>{csrf}</select>`, `<p title="{render field}">text</p>`} {
		t.Run(source, func(t *testing.T) {
			root := fixture(t)
			put(t, root, "_field.html", `{csrf}`)
			put(t, root, "index.html", source)
			if _, err := Load(root); err == nil {
				t.Fatal("unsafe placement accepted")
			}
		})
	}
	root := fixture(t)
	put(t, root, "index.html", `{csrf}`)
	put(t, root, "main.layout.html", `<div title="{slot}"></div>`)
	if _, err := Load(root); err == nil {
		t.Fatal("unsafe layout slot accepted")
	}
	root = fixture(t)
	put(t, root, "index.html", `<!-- {csrf} --><script>const example="{csrf}";</script><style>/* {csrf} */</style>{#if app.missing}{csrf}{/if}`)
	w := request(NewProductionServer(mustLoad(t, root), nil), "GET", "https://example.test/")
	if w.Code != 200 || len(w.Result().Cookies()) != 0 || strings.Contains(w.Body.String(), `name="_csrf"`) {
		t.Fatal("inert directive executed")
	}
}

func TestCSRFFailedRenderDoesNotSetCookie(t *testing.T) {
	root := fixture(t)
	put(t, root, "index.html", `{csrf}{#each contact.data.bad as item}{item}{/each}`)
	p := mustLoad(t, root)
	s := NewProductionServer(p, nil)
	r := httptest.NewRequest("GET", "https://example.test/", nil)
	inputs, _ := s.formRender(r)
	inputs.appName = "contact"
	inputs.appData = map[string]any{"data": map[string]any{"bad": "not an array"}}
	if _, err := p.renderRequest(p.Pages["/"], "/", nil, inputs); err == nil {
		t.Fatal("expected failure")
	}
	// Invalid origin cannot issue a token, even if forwarded headers claim TLS.
	r = httptest.NewRequest("GET", "http://example.test/", nil)
	r.Header.Set("X-Forwarded-Proto", "https")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, r)
	if w.Code != http.StatusInternalServerError || len(w.Result().Cookies()) != 0 {
		t.Fatal("failed render published cookie")
	}
}

func TestAppRenderBound(t *testing.T) {
	root := fixture(t)
	put(t, root, "index.html", `{#each contact.data.items as item}{item}{/each}`)
	p := mustLoad(t, root)
	items := make([]any, 10)
	for i := range items {
		items[i] = strings.Repeat("x", appRenderLimit/4)
	}
	inputs := &renderInputs{appName: "contact", appData: map[string]any{"data": map[string]any{"items": items}}}
	body, err := p.renderRequest(p.Pages["/"], "/", nil, inputs)
	if err == nil || len(body) > appRenderLimit {
		t.Fatal("unbounded app rendering", len(body), err)
	}
}
