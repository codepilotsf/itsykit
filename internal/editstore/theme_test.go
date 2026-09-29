package editstore

import (
	"encoding/json"
	"flea/internal/editor"
	"flea/internal/theme"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestThemeSaveAndStaleVersion(t *testing.T) {
	for _, remote := range []bool{false, true} {
		t.Run(map[bool]string{false: "local", true: "remote"}[remote], func(t *testing.T) {
			s, server := fixture(t, remote)
			const css = ":root {--brand-1: #123456;} /* keep */"
			path := filepath.Join(s.Root, "site/css/theme.css")
			if e := os.MkdirAll(filepath.Dir(path), 0755); e != nil {
				t.Fatal(e)
			}
			if remote {
				path = s.ContentPath + ".theme.css"
				s.ThemePath = path
				raw, e := os.ReadFile(s.ConfigPath)
				if e != nil {
					t.Fatal(e)
				}
				var c map[string]any
				if e = json.Unmarshal(raw, &c); e != nil {
					t.Fatal(e)
				}
				sites := c["sites"].(map[string]any)
				themes := map[string]string{}
				for domain := range sites {
					themes[domain] = path
				}
				c["themes"] = themes
				raw, _ = json.Marshal(c)
				if e = os.WriteFile(s.ConfigPath, raw, 0644); e != nil {
					t.Fatal(e)
				}
			}
			if e := os.WriteFile(path, []byte(css), 0644); e != nil {
				t.Fatal(e)
			}
			doc, e := s.Read()
			if e != nil {
				t.Fatal(e)
			}
			if e = s.Save(editor.Update{Version: doc.Version, Theme: map[string]string{"--brand-1": "url(https://bad.test)"}}); e == nil {
				t.Fatal("unsafe CSS accepted")
			}
			update := editor.Update{Version: doc.Version, Theme: map[string]string{"--brand-1": "#abcdef"}}
			if e = s.Save(update); e != nil {
				t.Fatal(e)
			}
			if e = s.Save(update); e == nil {
				t.Fatal("stale theme accepted")
			}
			b, e := os.ReadFile(path)
			if e != nil || string(b) != strings.Replace(css, "#123456", "#abcdef", 1) {
				t.Fatal(string(b), e)
			}
			w := httptest.NewRecorder()
			server.ServeHTTP(w, httptest.NewRequest("GET", "/css/theme.css", nil))
			if !strings.Contains(w.Body.String(), "#abcdef") {
				t.Fatal("theme not published", w.Body.String())
			}
			starter, err := os.ReadFile("../../templates/site/css/theme.css")
			if err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(path, starter, 0644); err != nil {
				t.Fatal(err)
			}
			doc, err = s.Read()
			if err != nil {
				t.Fatal(err)
			}
			settings := theme.Scales(string(starter))["text"].Settings
			settings.MaxBase = 24
			if err = s.Save(editor.Update{Version: doc.Version, Scales: map[string]theme.ScaleEdit{"text": {Settings: settings}}}); err != nil {
				t.Fatal(err)
			}
			saved, err := os.ReadFile(path)
			if err != nil || theme.Scales(string(saved))["text"].Settings != settings {
				t.Fatal("scale not persisted", err)
			}
			w = httptest.NewRecorder()
			server.ServeHTTP(w, httptest.NewRequest("GET", "/css/theme.css", nil))
			if !strings.Contains(w.Body.String(), theme.Clamp(settings, 0)) {
				t.Fatal("scale not published")
			}
			doc, err = s.Read()
			if err != nil {
				t.Fatal(err)
			}
			if err = s.Save(editor.Update{Version: doc.Version, Theme: map[string]string{"--font-body": "Lora, serif"}}); err != nil {
				t.Fatal(err)
			}
			saved, err = os.ReadFile(path)
			if err != nil || !strings.Contains(string(saved), "family=Lora") {
				t.Fatal("font import not saved", err)
			}
			w = httptest.NewRecorder()
			server.ServeHTTP(w, httptest.NewRequest("GET", "/css/theme.css", nil))
			if !strings.Contains(w.Body.String(), "family=Lora") {
				t.Fatal("font import not published")
			}
		})
	}
}
