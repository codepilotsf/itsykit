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
	for _, bad := range []string{"style=", "class=", "onerror", "onclick", "javascript", "<svg", "<img", "<iframe"} {
		if strings.Contains(got, bad) {
			t.Fatal(got)
		}
	}
	if !strings.Contains(got, "<em>world</em>") || !strings.Contains(got, `href="https://example.com/?a=1&amp;b=2"`) {
		t.Fatal(got)
	}
}
func TestMergePreservesMovesArchivesAndRestores(t *testing.T) {
	remote := NewStore()
	remote.Page("/old").Title = "Owner title"
	remote.Page("/old").Snips["a7Km2pQ9xB"] = Entry{Label: "old", Content: "Owner edit"}
	remote.Shared.Snips["b8Ln3rR0yC"] = Entry{Content: "Removed owner edit"}
	seed := NewStore()
	seed.Page("/new").Snips["a7Km2pQ9xB"] = Entry{Label: "new", Content: "Local stale"}
	out, err := Merge(seed, remote, map[string]string{"a7Km2pQ9xB": "/new"})
	if err != nil {
		t.Fatal(err)
	}
	if out.Pages["/new"].Snips["a7Km2pQ9xB"].Content != "Owner edit" || out.Archive["b8Ln3rR0yC"].Content != "Removed owner edit" || out.Pages["/old"].Title != "Owner title" {
		t.Fatal(out)
	}
	restored, err := Merge(seed, out, map[string]string{"b8Ln3rR0yC": ""})
	if err != nil || restored.Shared.Snips["b8Ln3rR0yC"].Content != "Removed owner edit" {
		t.Fatal(restored, err)
	}
}
