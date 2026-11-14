package syncstate

import (
	"strings"
	"testing"

	"itsy/internal/media"
	"itsy/internal/snips"
)

func TestQAPhotoSyncConflictsAreExplicit(t *testing.T) {
	makeStore := func(name string) *snips.Store {
		s := snips.NewStore()
		s.Media = map[string]media.Photo{"Abcdefghij": {File: strings.Repeat("a", 64) + ".png", Name: name, Width: 4, Height: 2}}
		return s
	}
	base, local, remote := makeStore("Original"), makeStore("Local"), makeStore("Owner")
	if _, e := MergeContent(base, local, remote, nil); e == nil {
		t.Fatal("silently resolved photo metadata conflict")
	}
	for _, takeLocal := range []bool{false, true} {
		called := false
		out, e := MergeContent(base, local, remote, func(c Conflict) (bool, error) {
			called = true
			if c.Key != "photo Abcdefghij" {
				t.Fatal(c.Key)
			}
			return takeLocal, nil
		})
		if e != nil {
			t.Fatal(e)
		}
		want := "Owner"
		if takeLocal {
			want = "Local"
		}
		if !called || out.Media["Abcdefghij"].Name != want {
			t.Fatal("wrong conflict selection")
		}
	}
	out, e := MergeContent(base, base, remote, nil)
	if e != nil || out.Media["Abcdefghij"].Name != "Owner" {
		t.Fatal("unchanged local photo overwrote owner edit", e)
	}
}
