package site

import (
	"fmt"
	"golang.org/x/net/html"
	"itsy/internal/sitefiles"
	"itsy/internal/snips"
	"os"
	"strings"
)

func readContentStore(root *os.Root) (*snips.Store, error) {
	// Refuse unfinished old extraction rather than discarding recovery state.
	for _, name := range []string{"itsy/runtime/.snips-pending.json", "itsy/snips"} {
		if _, err := root.Lstat(sitefiles.Resolve(root, name)); err == nil {
			return nil, fmt.Errorf("%s: page CMS content is no longer supported; retain a backup before converting to HTML/JSON", name)
		} else if !os.IsNotExist(err) {
			return nil, err
		}
	}
	data, err := readFile(root, "itsy/snips.json")
	if os.IsNotExist(err) {
		return snips.NewStore(), nil
	}
	if err != nil {
		return nil, err
	}
	return snips.Decode(data)
}
func checkRemovedPageMarkup(file, source string) error {
	z := html.NewTokenizer(strings.NewReader(source))
	position := 0
	for k := z.Next(); k != html.ErrorToken; k = z.Next() {
		start := position
		position += len(z.Raw())
		if k != html.StartTagToken && k != html.SelfClosingTagToken {
			continue
		}
		for _, a := range z.Token().Attr {
			if a.Key == "snip" || a.Key == "multisnip" || a.Key == "flea" {
				return sourceError(file, source, start, fmt.Errorf("page CMS attribute %q is no longer supported; use ordinary HTML/JSON", a.Key))
			}
		}
	}
	return nil
}
