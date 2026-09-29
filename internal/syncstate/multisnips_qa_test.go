package syncstate

import (
	"flea/internal/snips"
	"testing"
)

func TestQAMultisnipReorderDeleteConflict(t *testing.T) {
	makeStore := func(ids ...string) *snips.Store {
		s := snips.NewStore()
		items := []snips.MultiItem{}
		for _, id := range ids {
			items = append(items, snips.MultiItem{ID: id, Fields: map[string]string{"name": id}})
		}
		s.Page("/").Multisnips = map[string][]snips.MultiItem{"team": items}
		return s
	}
	base := makeStore("a123456789", "b123456789")
	local := makeStore("b123456789", "a123456789")
	remote := makeStore("a123456789")
	if _, err := MergeSnips(base, local, remote, nil); err == nil {
		t.Fatal("concurrent reorder/deletion silently merged")
	}
	out, err := MergeSnips(base, local, remote, func(c Conflict) (bool, error) { return false, nil })
	if err != nil {
		t.Fatal(err)
	}
	if items := out.Pages["/"].Multisnips["team"]; len(items) != 1 || items[0].ID != "a123456789" {
		t.Fatal("keep live did not retain live collection")
	}
}
