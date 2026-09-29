package theme

import (
	"strings"
	"testing"
)

func TestFontImports(t *testing.T) {
	css := "@charset \"UTF-8\";\n@import url(\"/custom.css\");\n:root {--font-body: Arial, sans-serif; --font-heading: Georgia, serif;}"
	next, err := Apply(css, map[string]string{"--font-body": "\"Lora\", serif"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(next, "@charset \"UTF-8\";") || !strings.Contains(next, `https://fonts.googleapis.com/css2?display=swap&family=Lora`) || !strings.Contains(next, `@import url("/custom.css");`) {
		t.Fatal(next)
	}
	again, err := Apply(next, map[string]string{"--font-heading": "Lora, serif"})
	if err != nil || strings.Count(again, "family=Lora") != 1 || strings.Count(again, fontStart) != 1 {
		t.Fatal(again, err)
	}
	again, err = Apply(again, map[string]string{"--font-body": "Arial, sans-serif", "--font-heading": "Georgia, serif"})
	if err != nil || strings.Contains(again, fontStart) || !strings.Contains(again, `@import url("/custom.css");`) {
		t.Fatal(again, err)
	}
	for _, v := range []string{`url(https://evil.test)`, `Lora; color:red`, `"Lora`, "Lora,", `@import url(x)`, `Lora&family=Other`} {
		if _, err := Apply(css, map[string]string{"--font-body": v}); err == nil {
			t.Fatal("unsafe family", v)
		}
	}
}
