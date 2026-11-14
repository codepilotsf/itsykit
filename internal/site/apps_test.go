package site

import (
	"testing"
)

func TestAppRegistration(t *testing.T) {
	for _, tc := range []struct {
		name, settings string
		valid          bool
	}{
		{"disabled", `"apps":{}`, true},
		{"contact", `"apps":{"contact":{"mount":"/contact","email":{"to":"owner@example.test"}}}`, true},
		{"old config", `"contact":{"to":"owner@example.test"}`, false},
		{"null", `"apps":null`, false},
		{"missing recipient", `"apps":{"contact":{"mount":"/contact"}}`, false},
		{"unknown option", `"apps":{"contact":{"mount":"/contact","url":"http://127.0.0.1:9999","email":{"to":"owner@example.test"}}}`, false},
		{"blog email settings", `"apps":{"blog":{"mount":"/contact","email":{"to":"owner@example.test"}}}`, false},
		{"core route", `"apps":{"contact":{"mount":"/admin","email":{"to":"owner@example.test"}}}`, false},
		{"no pages", `"apps":{"contact":{"mount":"/missing","email":{"to":"owner@example.test"}}}`, false},
		{"sender injection", `"apps":{"contact":{"mount":"/contact","email":{"to":"owner@example.test","from":"evil\r\n@example.test"}}}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := fixture(t)
			put(t, root, "contact/index.html", `<form>{csrf}</form>`)
			put(t, root, "contact/thank-you.html", `Thanks`)
			put(t, root, ".itsy.config.json", `{"app":{"site_name":"Test"},`+tc.settings+`}`)
			_, err := Load(root)
			if (err == nil) != tc.valid {
				t.Fatal("registration result", err)
			}
		})
	}
	root := fixture(t)
	put(t, root, "contact/index.html", `Form`)
	put(t, root, "contact/thank-you.html", `Thanks`)
	put(t, root, "contact/style.css", `body { color: red; }`)
	put(t, root, ".itsy.config.json", `{"app":{"site_name":"Test"},"apps":{"contact":{"mount":"/contact","email":{"to":"owner@example.test"}}}}`)
	if _, err := Load(root); err == nil {
		t.Fatal("asset inside app mount")
	}
}
