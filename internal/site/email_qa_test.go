package site

import (
	"strings"
	"testing"
)

func TestQAEmailOutputBoundDuringRendering(t *testing.T) {
	root := fixture(t)
	put(t, root, "mailers/main.layout.html", "<html><body>{slot}</body></html>")
	for _, kind := range []string{"sign-in", "contact"} {
		put(t, root, "mailers/"+kind+".email.html", strings.Repeat("{email.message}", 100))
		put(t, root, "mailers/"+kind+".email.txt", "{email.message}")
	}
	p := mustLoad(t, root)
	if _, err := p.RenderEmail("contact", map[string]string{"message": strings.Repeat("x", 10000)}); err == nil {
		t.Fatal("oversized output accepted")
	}
	var buf emailBuffer
	if _, err := buf.WriteString(strings.Repeat("x", 256<<10)); err != nil {
		t.Fatal(err)
	}
	if _, err := buf.WriteString("x"); err == nil {
		t.Fatal("WriteString bypassed bound")
	}
	if buf.Len() != 256<<10 {
		t.Fatal("overflow write mutated buffer")
	}
}
