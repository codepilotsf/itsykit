package syncstate

import (
	"fmt"
	"strings"
	"testing"

	"itsy/internal/snips"
)

func typedImageStore(t *testing.T, alt, label string) *snips.Store {
	t.Helper()
	data := fmt.Sprintf(`{"pages":{"/":{"snips":{"Abcdef1234":{"type":"image","label":%q,"image":{"file":%q,"name":"Portrait","alt":%q,"width":4,"height":2}}}}},"shared":{"snips":{}}}`, label, strings.Repeat("a", 64)+".png", alt)
	s, err := snips.Decode([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestTypedMediaSyncWholeValue(t *testing.T) {
	base := typedImageStore(t, "Original", "Original label")
	local := typedImageStore(t, "Original", "Local label")
	remote := typedImageStore(t, "Owner edit", "Remote label")
	merged, err := MergeSnips(base, local, remote, nil)
	if err != nil {
		t.Fatal(err)
	}
	value := merged.Archive["Abcdef1234"]
	if value.Image.Alt != "Owner edit" || value.Label != "Local label" {
		t.Fatalf("owner metadata or developer label lost: %+v", value)
	}
	if local.Page("/").Snips["Abcdef1234"].Image.Alt != "Original" {
		t.Fatal("merge mutated local input")
	}
	local = typedImageStore(t, "Local edit", "Local label")
	if _, err := MergeSnips(base, local, remote, nil); err == nil {
		t.Fatal("competing typed media changes silently resolved")
	}
	for _, takeLocal := range []bool{false, true} {
		merged, err = MergeSnips(base, local, remote, func(c Conflict) (bool, error) {
			if c.Key != "snip Abcdef1234" {
				t.Fatal(c.Key)
			}
			return takeLocal, nil
		})
		if err != nil {
			t.Fatal(err)
		}
		want := "Owner edit"
		if takeLocal {
			want = "Local edit"
		}
		if merged.Archive["Abcdef1234"].Image.Alt != want {
			t.Fatal("wrong conflict resolution")
		}
	}
	if _, err := MergeSnips(snips.NewStore(), local, remote, nil); err == nil {
		t.Fatal("new typed ID collision ignored")
	}
}

func TestTypedMediaSyncTypeReplacement(t *testing.T) {
	base := typedImageStore(t, "Original", "Label")
	local := typedImageStore(t, "Original", "Local label")
	remote := snips.NewStore()
	remote.Page("/").Snips["Abcdef1234"] = snips.Entry{Content: "Now text"}
	merged, err := MergeSnips(base, local, remote, nil)
	if err != nil {
		t.Fatal(err)
	}
	value := merged.Archive["Abcdef1234"]
	if value.Type != "" || value.Image != nil || value.Content != "Now text" || value.Label != "Local label" {
		t.Fatalf("type replacement retained stale fields: %+v", value)
	}
}
