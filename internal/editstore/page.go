package editstore

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"flea/internal/editor"
	"flea/internal/site"
	"flea/internal/snips"
	"golang.org/x/net/html"
)

func applyPage(p *site.Project, edit editor.PageEdit) error {
	if utf8.RuneCountInString(edit.Title) > 200 || utf8.RuneCountInString(edit.Description) > 500 || len(edit.Changes) > 200 {
		return fmt.Errorf("page exceeds editing limits")
	}
	if err := applyMultisnips(p, edit); err != nil {
		return err
	}
	allowed, err := p.EditableRegions(edit.Route)
	if err != nil {
		return err
	}
	// A shared ID may have a narrower wrapper on another page. Editing it here
	// must not introduce block markup into that other wrapper.
	for _, route := range p.Routes() {
		if route == edit.Route {
			continue
		}
		other, err := p.EditableRegions(route)
		if err != nil {
			return err
		}
		for id, rich := range other {
			if _, present := allowed[id]; present && !rich {
				allowed[id] = false
			}
		}
	}
	index, err := p.Content.Index()
	if err != nil {
		return err
	}
	seen := map[string]bool{}
	for _, change := range edit.Changes {
		if strings.Contains(change.ID, "/") {
			continue
		} // Validated with the collection draft above.
		rich, ok := allowed[change.ID]
		if !ok || seen[change.ID] {
			return fmt.Errorf("unknown, hidden or duplicate editable region")
		}
		seen[change.ID] = true
		entry, ok := index[change.ID]
		if !ok {
			return fmt.Errorf("unknown content")
		}
		clean, err := snips.Sanitize(change.Content)
		if err != nil {
			return err
		}
		// A heading or paragraph wrapper cannot receive nested block markup.
		if !rich {
			z := html.NewTokenizer(strings.NewReader(clean))
			for token := z.Next(); token != html.ErrorToken; token = z.Next() {
				if token == html.TextToken {
					continue
				}
				if token == html.StartTagToken || token == html.EndTagToken || token == html.SelfClosingTagToken {
					if strings.Contains(" br strong b em i u s del mark small sub sup code a ", " "+z.Token().Data+" ") {
						continue
					}
				}
				return fmt.Errorf("this region accepts inline formatting only")
			}
		}
		entry.Content = clean
		if _, ok := p.Content.Shared.Snips[change.ID]; ok {
			p.Content.Shared.Snips[change.ID] = entry
		} else {
			for _, page := range p.Content.Pages {
				if _, ok := page.Snips[change.ID]; ok {
					page.Snips[change.ID] = entry
				}
			}
		}
	}
	page := p.Content.Page(edit.Route)
	page.Title, page.Description = edit.Title, edit.Description
	return nil
}
