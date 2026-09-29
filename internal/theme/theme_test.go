package theme

import (
	"os"
	"strings"
	"testing"
)

func TestStarterAndPreservation(t *testing.T) {
	b, e := os.ReadFile("../../templates/site/css/theme.css")
	if e != nil {
		t.Fatal(e)
	}
	fs, e := Fields(string(b))
	if e != nil || len(fs) < 60 {
		t.Fatal(len(fs), e)
	}
	next, e := Apply(string(b), map[string]string{"--brand-1": "#123456"})
	if e != nil {
		t.Fatal(e)
	}
	if next != strings.Replace(string(b), "#286047", "#123456", 1) {
		t.Fatal("unrelated source changed")
	}
}
func TestRejectInjection(t *testing.T) {
	for _, value := range []string{"red; color:green", "url(https://evil.test)", "r\\65 d", "expression(alert(1))", "red/*x*/", "red}</style>"} {
		if _, e := Apply(":root {--brand-1: red;}", map[string]string{"--brand-1": value}); e == nil {
			t.Fatal(value)
		}
	}
}

func TestFontColorOptions(t *testing.T) {
	css := `:root {--font-color: var(--black); --black: black; --brand-1: blue; --error: red;}`
	for _, value := range []string{"var(--black)", "var(--brand-1)"} {
		if _, err := Apply(css, map[string]string{"--font-color": value}); err != nil {
			t.Fatal(err)
		}
	}
	for _, value := range []string{"red", "var(--error)", "var(--missing)"} {
		if _, err := Apply(css, map[string]string{"--font-color": value}); err == nil {
			t.Fatal("accepted", value)
		}
	}
}

func TestBackgroundColor(t *testing.T) {
	b, err := os.ReadFile("../../templates/site/css/theme.css")
	if err != nil {
		t.Fatal(err)
	}
	fields, err := Fields(string(b))
	if err != nil {
		t.Fatal(err)
	}
	found := 0
	for _, f := range fields {
		if f.Name == "--font-color" || f.Name == "--bg-color" {
			if f.Group != "Page colors" {
				t.Fatal(f)
			}
			found++
		}
	}
	if found != 2 || !strings.Contains(string(b), "--bg-color: var(--lighter)") {
		t.Fatal("page colors missing")
	}
	if _, err := Apply(string(b), map[string]string{"--bg-color": "var(--brand-1)"}); err != nil {
		t.Fatal(err)
	}
	if _, err := Apply(string(b), map[string]string{"--bg-color": "var(--error)"}); err == nil {
		t.Fatal("semantic background accepted")
	}
}
