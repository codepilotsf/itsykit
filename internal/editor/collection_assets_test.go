package editor

import (
	"strings"
	"testing"
)

func TestCollectionIconTemplates(t *testing.T) {
	markup := collectionIconTemplates()
	for _, id := range []string{"grip", "remove", "add"} {
		if !strings.Contains(markup, `id="collection-`+id+`"`) {
			t.Fatalf("missing %s icon template", id)
		}
	}
	if strings.Count(markup, `<svg `) != 3 || strings.Contains(markup, "{icon ") {
		t.Fatal("icons were not compiled through the template syntax")
	}
}
