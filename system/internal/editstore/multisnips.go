package editstore

import (
	"fmt"
	"itsy/internal/editor"
	"itsy/internal/site"
	"itsy/internal/snips"
	"strings"
)

func applyMultisnips(p *site.Project, edit editor.PageEdit) error {
	defs, err := p.VisibleMultisnips(edit.Route)
	if err != nil {
		return err
	}
	if err := snips.ValidateMultisnips(edit.Multisnips); err != nil {
		return err
	}
	page := p.Content.Page(edit.Route)
	groups := snips.CloneMultisnips(page.Multisnips)
	if groups == nil {
		groups = map[string][]snips.MultiItem{}
	}
	for name, items := range snips.CloneMultisnips(edit.Multisnips) {
		if _, ok := defs[name]; !ok {
			return fmt.Errorf("unknown multisnip on this page")
		}
		for i := range items {
			for _, old := range groups[name] {
				if old.ID == items[i].ID {
					for key, value := range old.Fields {
						if _, visible := defs[name].Fields[key]; !visible {
							items[i].Fields[key] = value
						}
					}
					break
				}
			}
		}
		groups[name] = items
	}
	seen := map[string]bool{}
	for _, change := range edit.Changes {
		if !strings.Contains(change.ID, "/") {
			continue
		}
		parts := strings.Split(change.ID, "/")
		if len(parts) != 3 || seen[change.ID] {
			return fmt.Errorf("invalid multisnip field change")
		}
		seen[change.ID] = true
		def, ok := defs[parts[0]]
		if !ok {
			return fmt.Errorf("unknown multisnip")
		}
		if _, ok := def.Fields[parts[2]]; !ok {
			return fmt.Errorf("unknown multisnip field")
		}
		found := false
		for i := range groups[parts[0]] {
			item := &groups[parts[0]][i]
			if item.ID == parts[1] {
				item.Fields[parts[2]] = change.Content
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("unknown multisnip item")
		}
	}
	if err := snips.ValidateMultisnips(groups); err != nil {
		return err
	}
	for name, items := range groups {
		def, active := defs[name]
		if !active {
			continue
		} // Retain data if the developer temporarily removes its template.
		for i := range items {
			for field, value := range items[i].Fields {
				if def.Photos[field] {
					if value != "" {
						if _, ok := p.Content.Media[value]; !ok {
							return fmt.Errorf("unknown photo selection")
						}
					}
					continue
				}
				rich, ok := def.Fields[field]
				if !ok {
					continue
				} // Retain developer-removed fields, but do not accept new hidden fields.
				clean, err := snips.CleanField(value, rich)
				if err != nil {
					return err
				}
				items[i].Fields[field] = clean
			}
		}
	}
	// Browser-supplied rows must contain exactly the current template's fields.
	for name, items := range edit.Multisnips {
		for _, item := range items {
			if len(item.Fields) != len(defs[name].Fields) {
				return fmt.Errorf("multisnip fields changed; reload before editing")
			}
			for field := range item.Fields {
				if _, ok := defs[name].Fields[field]; !ok {
					return fmt.Errorf("unknown multisnip field")
				}
			}
		}
	}
	page.Multisnips = groups
	return nil
}
