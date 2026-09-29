package snips

import "testing"

func TestQAMergePreservesExplicitEmptyMultisnip(t *testing.T) {
	seed, remote := NewStore(), NewStore()
	seed.Page("/").Multisnips = map[string][]MultiItem{"team": {{ID: "a123456789", Fields: map[string]string{"name": "Joe"}}}}
	remote.Page("/").Multisnips = map[string][]MultiItem{"team": {}}
	out, err := Merge(seed, remote, map[string]string{})
	if err != nil {
		t.Fatal(err)
	}
	if items, ok := out.Pages["/"].Multisnips["team"]; !ok || items == nil || len(items) != 0 {
		t.Fatal("owner deletion was reseeded")
	}
}
