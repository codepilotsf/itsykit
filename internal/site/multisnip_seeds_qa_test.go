package site

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"flea/internal/snips"
)

const qaJoe = `<article multisnip="team"><h2 snip="name">Joe</h2><p snip="role">Founder</p></article>`
const qaJane = `<article multisnip="team"><h2 snip="name">Jane</h2><p snip="role">Designer</p></article>`

func TestQAAdjacentMultisnipSeedsOrderAndIdempotence(t *testing.T) {
	root := fixture(t)
	put(t, root, "index.html", `<section>`+qaJoe+"\n"+qaJane+`</section>`)
	if e := ExtractSnips(root); e != nil {
		t.Fatal(e)
	}
	p := mustLoad(t, root)
	items := p.Content.Pages["/"].Multisnips["team"]
	if len(items) != 2 || items[0].Fields["name"] != "Joe" || items[1].Fields["name"] != "Jane" || items[1].Fields["role"] != "Designer" || items[0].ID == items[1].ID || len(items[0].ID) != 10 || len(items[1].ID) != 10 {
		t.Fatal(items)
	}
	source, _ := os.ReadFile(filepath.Join(root, "site", "index.html"))
	if strings.Count(string(source), "{#multisnip team}") != 1 {
		t.Fatal("multiple generated loops")
	}
	before, _ := os.ReadFile(filepath.Join(root, "flea/snips.json"))
	if e := ExtractSnips(root); e != nil {
		t.Fatal(e)
	}
	after, _ := os.ReadFile(filepath.Join(root, "flea/snips.json"))
	if !bytes.Equal(before, after) {
		t.Fatal("not idempotent")
	}
}
func TestQAMultisnipSeedMismatchLeavesSourceAndContent(t *testing.T) {
	for _, source := range []string{qaJoe + `<hr>` + qaJane, qaJoe + strings.Replace(qaJane, "<article ", `<article class="different" `, 1)} {
		root := fixture(t)
		put(t, root, "index.html", source)
		before, e := os.ReadFile(filepath.Join(root, "flea/snips.json"))
		missing := os.IsNotExist(e)
		if e := ExtractSnips(root); e == nil {
			t.Fatal("accepted incompatible repeated seed")
		}
		after, _ := os.ReadFile(filepath.Join(root, "site", "index.html"))
		if string(after) != source {
			t.Fatal("rewrote rejected source")
		}
		content, e := os.ReadFile(filepath.Join(root, "flea/snips.json"))
		if missing {
			if !os.IsNotExist(e) {
				t.Fatal("created content on failed extraction")
			}
		} else if !bytes.Equal(before, content) {
			t.Fatal("modified content on failed extraction")
		}
	}
}
func TestQARepeatedSeedsPreserveExistingArray(t *testing.T) {
	for _, items := range [][]snips.MultiItem{{}, {{ID: "a123456789", Fields: map[string]string{"name": "Owner", "role": "Editor"}}}} {
		root := fixture(t)
		put(t, root, "index.html", qaJoe+"\n"+qaJane)
		s := snips.NewStore()
		s.Page("/").Multisnips = map[string][]snips.MultiItem{"team": items}
		if e := SaveContent(root, s); e != nil {
			t.Fatal(e)
		}
		if e := ExtractSnips(root); e != nil {
			t.Fatal(e)
		}
		p := mustLoad(t, root)
		if !reflect.DeepEqual(items, p.Content.Pages["/"].Multisnips["team"]) {
			t.Fatal("reseeded existing array")
		}
	}
}
