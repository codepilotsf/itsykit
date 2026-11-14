package snips

import (
	"strings"
	"testing"
)

func TestSanitizeFormattingAndAttacks(t *testing.T) {
	source := `<p class="x" style="color:red">Hello <em>world</em><img src=x onerror=alert(1)><svg><script>bad</script></svg><iframe srcdoc="bad"></iframe><a href="jav&#x61;script:bad" onclick="bad">bad</a><a href="https://example.com/?a=1&amp;b=2" title="OK">safe</a></p>`
	got, err := Sanitize(source)
	if err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{"style=", "onerror", "onclick", "javascript", "<svg", "<img", "<iframe"} {
		if strings.Contains(got, bad) {
			t.Fatal(got)
		}
	}
	if !strings.Contains(got, "<em>world</em>") || !strings.Contains(got, `href="https://example.com/?a=1&amp;b=2"`) {
		t.Fatal(got)
	}
}

func TestEditorFormattingAllowlist(t *testing.T) {
	got, err := Sanitize(`<h1 style="text-align: CENTER">Title</h1><p style="text-align:right;color:red">No CSS</p><a href="/about" target="_blank" rel="opener">About</a><a href="javascript:bad" target="evil">Bad</a><p style="background:url(https://evil.test)">No request</p>`)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`<h1 style="text-align: center;">Title</h1>`, `target="_blank" rel="noopener noreferrer"`} {
		if !strings.Contains(got, want) {
			t.Fatal(got)
		}
	}
	for _, bad := range []string{"color:", "background", "javascript:", `target="evil"`, `rel="opener"`} {
		if strings.Contains(got, bad) {
			t.Fatal(got)
		}
	}
	if again, err := Sanitize(got); err != nil || again != got {
		t.Fatal("not idempotent", again, err)
	}
}

func TestStyledSpansSurvive(t *testing.T) {
	source := `<p class="text-lg">Hello <span class="text-brand-one font-bold">brand <em>name</em></span></p>`
	got, err := Sanitize(source)
	if err != nil || got != source {
		t.Fatalf("styled content changed: %s %v", got, err)
	}
	got, err = Sanitize(`<span class="x&quot; onclick=&quot;bad" onclick="bad" style="position:fixed">Safe</span>`)
	if err != nil || strings.Contains(got, `" onclick="`) || strings.Contains(got, ` style=`) {
		t.Fatalf("unsafe attributes survived: %s %v", got, err)
	}
	if again, err := Sanitize(got); err != nil || again != got {
		t.Fatal("not idempotent", again, err)
	}
}
