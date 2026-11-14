package site

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSourceExcerptWindow(t *testing.T) {
	var lines []string
	for i := 1; i <= 20; i++ {
		lines = append(lines, fmt.Sprintf("line %d", i))
	}
	source := strings.Join(lines, "\n")
	for _, tc := range []struct{ line, first, last int }{{1, 1, 6}, {10, 5, 15}, {20, 15, 20}} {
		pos := strings.Index(source, fmt.Sprintf("line %d", tc.line))
		var diagnostic *SourceError
		if !errors.As(sourceError("page.html", source, pos, errors.New("test")), &diagnostic) {
			t.Fatal("missing source context")
		}
		if len(diagnostic.Lines) != tc.last-tc.first+1 {
			t.Fatal("wrong excerpt size", diagnostic.Lines)
		}
		for i, line := range diagnostic.Lines {
			if line.Number != tc.first+i || line.Highlight != (line.Number == tc.line) || line.Text != lines[line.Number-1] {
				t.Fatal("incorrect source line", line)
			}
		}
	}
}

func TestDiagnosticSourceLocations(t *testing.T) {
	for _, tc := range []struct{ file, source, location string }{
		{"index.html", "<section snip>\n<p snip>Nested</p>\n</section>", "index.html:2"},
		{"index.html", "<!-- flea\nlayout: /main.layout.html\nunknown: value\n-->\n<p>Hi</p>", "index.html:3"},
		{"_copy.html", "<p>Copy</p>\n{#if bad expression}", "_copy.html:2"},
	} {
		t.Run(tc.location, func(t *testing.T) {
			root := fixture(t)
			s := NewServer(mustLoad(t, root))
			put(t, root, tc.file, tc.source)
			if err := s.Reload(root); err == nil {
				t.Fatal("expected source error")
			}
			body := localRequest(s, "GET", "/").Body.String()
			if !strings.Contains(body, tc.location) || !strings.Contains(body, `class="line marked"`) {
				t.Fatal("missing source location", body)
			}
		})
	}
}

func TestUnclosedConditionalDiagnostic(t *testing.T) {
	for _, tc := range []struct{ source, want string }{
		{"<main>\n{#if app.site_name}\n<p>Hello</p>\n</main>", "unclosed if block; expected {/if}"},
		{"<main>\n{#unless app.site_name}\n<p>Hello</p>\n</main>", "unclosed unless block; expected {/unless}"},
		{"<main>\n{#if app.site_name\n<p>{app.site_name}</p>\n</main>", "unclosed template directive"},
	} {
		t.Run(tc.want, func(t *testing.T) {
			root := fixture(t)
			put(t, root, "index.html", tc.source)
			_, err := Load(root)
			if err == nil || !strings.Contains(err.Error(), tc.want) || !strings.Contains(err.Error(), "index.html:2") {
				t.Fatal(err)
			}
		})
	}
}

func localRequest(s *Server, method, path string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, "http://localhost:3000"+path, nil)
	r.RemoteAddr = "127.0.0.1:54321"
	w := httptest.NewRecorder()
	s.ServeHTTP(w, r)
	return w
}

func TestDevelopmentDiagnosticAndRecovery(t *testing.T) {
	root := fixture(t)
	var logs bytes.Buffer
	s := NewServer(mustLoad(t, root), log.New(&logs, "", 0))
	put(t, root, "index.html", "<main>\n<h1>Hi</h1>\n{#if bad expression}\n<script>alert('not executed')</script>\n</main>")
	err := s.Reload(root)
	if err == nil {
		t.Fatal("expected invalid template")
	}
	w := localRequest(s, "GET", "/?token=never-show-this")
	for _, want := range []string{"Application error", "How to fix it", "index.html:3", "unsupported expression", "&lt;script&gt;", "class=\"line marked\"", reloadPath} {
		if !strings.Contains(w.Body.String(), want) {
			t.Fatalf("missing %q", want)
		}
	}
	for _, removed := range []string{"Development diagnostic", "Try again", "Go to homepage", "<footer", "eyebrow"} {
		if strings.Contains(w.Body.String(), removed) {
			t.Fatal("removed diagnostic UI remains", removed)
		}
	}
	if w.Code != 500 || strings.Contains(w.Body.String(), "never-show-this") || strings.Contains(w.Body.String(), "<script>alert") {
		t.Fatal("unsafe diagnostic", w.Code)
	}
	if !strings.Contains(logs.String(), "unsupported expression") || strings.Contains(logs.String(), "never-show-this") {
		t.Fatal("missing/unsafe log", logs.String())
	}
	if w.Header().Get("Cache-Control") != "private, no-store" {
		t.Fatal(w.Header())
	}
	script := reloadScript(1)
	hash := sha256.Sum256(script[len("<script>") : len(script)-len("</script>")])
	if !strings.Contains(w.Header().Get("Content-Security-Policy"), base64.StdEncoding.EncodeToString(hash[:])) {
		t.Fatal("reload script blocked")
	}
	head := localRequest(s, "HEAD", "/")
	if head.Code != 500 || head.Body.Len() != 0 || head.Header().Get("Content-Length") != w.Header().Get("Content-Length") {
		t.Fatal("HEAD error", head.Header())
	}
	put(t, root, "index.html", "<h1>Fixed</h1>")
	if err := s.Reload(root); err != nil {
		t.Fatal(err)
	}
	if got := localRequest(s, "GET", "/"); got.Code != 200 || !strings.Contains(got.Body.String(), "Fixed") {
		t.Fatal("recovery failed")
	}
}

func TestDiagnosticPrivacyBoundary(t *testing.T) {
	root := fixture(t)
	p := mustLoad(t, root)
	for _, production := range []bool{false, true} {
		var logs bytes.Buffer
		s := NewServer(p, log.New(&logs, "", 0))
		s.production = production
		put(t, root, "index.html", "{#if private diagnostic}")
		if err := s.Reload(root); err == nil {
			t.Fatal("expected invalid template")
		}
		for _, tc := range []struct {
			host, peer string
			local      bool
		}{
			{"localhost:3000", "127.0.0.1:1", true}, {"[::1]:3000", "[::1]:1", true},
			{"external.test", "127.0.0.1:1", false}, {"localhost:3000", "192.0.2.1:1", false},
		} {
			r := httptest.NewRequest("GET", "http://"+tc.host+"/", nil)
			r.RemoteAddr = tc.peer
			w := httptest.NewRecorder()
			s.ServeHTTP(w, r)
			visible := strings.Contains(w.Body.String(), "private diagnostic")
			if visible != (!production && tc.local) {
				t.Fatalf("production=%v, host=%s, leaked=%v", production, tc.host, visible)
			}
			if !visible && !strings.Contains(w.Body.String(), "Try later") {
				t.Fatal("custom production error lost")
			}
		}
		if !strings.Contains(logs.String(), "private diagnostic") {
			t.Fatal("diagnostic missing from logs")
		}
	}
}

func TestDevMissingRouteAndMethod(t *testing.T) {
	s := NewServer(mustLoad(t, fixture(t)))
	w := localRequest(s, "GET", "/missing?secret=hidden")
	if w.Code != 404 || !strings.Contains(w.Body.String(), "Not here") || strings.Contains(w.Body.String(), "Development diagnostic") || !strings.Contains(w.Body.String(), reloadPath) {
		t.Fatal("dev must use the ordinary 404 template", w.Body.String())
	}
	if strings.Contains(w.Body.String(), "secret=hidden") {
		t.Fatal("query leaked")
	}
	w = localRequest(s, "POST", "/about")
	if w.Code != 405 || w.Header().Get("Allow") != "GET, HEAD" || !strings.Contains(w.Body.String(), "dedicated handler") {
		t.Fatal("method diagnostic")
	}
}

func TestJSONDiagnosticDoesNotExposeSourceValues(t *testing.T) {
	root := fixture(t)
	s := NewServer(mustLoad(t, root))
	put(t, root, "itsy.config.json", "{\n\"private\":\"secret-value\",\n}")
	if err := s.Reload(root); err == nil {
		t.Fatal("expected invalid JSON")
	}
	body := localRequest(s, "GET", "/").Body.String()
	if !strings.Contains(body, "itsy.config.json:3") || strings.Contains(body, "secret-value") {
		t.Fatal("JSON diagnostic lacks safe location")
	}
}

func TestProductionRenderAndErrorTemplateFailuresAreLogged(t *testing.T) {
	root := fixture(t)
	put(t, root, "index.html", `{#each snip.team.members as member}{member.name}{/each}`)
	p := mustLoad(t, root)
	// Simulate an unexpected runtime failure, including a broken custom fallback.
	p.Snips["team"]["members"] = "wrong type"
	p.Errors[500] = p.Pages["/"]
	var logs bytes.Buffer
	s := NewProductionServer(p, log.New(&logs, "", 0))
	w := localRequest(s, "GET", "/?token=private")
	if w.Code != 500 || strings.Contains(w.Body.String(), "expects a JSON array") || strings.Contains(w.Body.String(), "Development diagnostic") {
		t.Fatal("unsafe production error response")
	}
	for _, want := range []string{"Page error: GET /:", "expects a JSON array", "Error page 500 failed:"} {
		if !strings.Contains(logs.String(), want) {
			t.Fatal("missing error log", want)
		}
	}
	if strings.Contains(logs.String(), "token=private") {
		t.Fatal("query leaked to logs")
	}
}

func TestDiagnosticLineIncludesLeadingPageSettings(t *testing.T) {
	root := fixture(t)
	s := NewServer(mustLoad(t, root))
	put(t, root, "custom.layout.html", `<main>{slot}</main>`)
	put(t, root, "index.html", "<!-- flea\nlayout: /custom.layout.html\n-->\n{#if bad expression}")
	if err := s.Reload(root); err == nil {
		t.Fatal("expected invalid template")
	}
	if !strings.Contains(localRequest(s, "GET", "/").Body.String(), "index.html:4") {
		t.Fatal("location was relative to the stripped page body")
	}
}

func TestNavigationStylesOnRealPages(t *testing.T) {
	p := mustLoad(t, fixture(t))
	for _, s := range []*Server{NewServer(p), NewProductionServer(p, nil)} {
		for _, route := range []string{"/", "/about"} {
			body := request(s, "GET", route).Body.String()
			for _, forbidden := range []string{"data-flea-navigation", "@view-transition", "itsy-page-out"} {
				if strings.Contains(body, forbidden) {
					t.Fatalf("%s received injected %q", route, forbidden)
				}
			}
		}
		if strings.Contains(request(s, "GET", "/css/main.css").Body.String(), "data-flea-navigation") {
			t.Fatal("asset modified")
		}
	}
}
