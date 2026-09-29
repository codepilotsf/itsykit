package snips

import (
	"fmt"
	"maps"
	"regexp"
)

var fieldName = regexp.MustCompile(`^[a-z][a-z0-9_]{0,39}$`)

func ValidFieldName(name string) bool { return fieldName.MatchString(name) }

func ValidateMultisnips(groups map[string][]MultiItem) error {
	if len(groups) > 20 {
		return fmt.Errorf("at most 20 multisnips per page")
	}
	ids := map[string]bool{}
	for name, items := range groups {
		if !ValidFieldName(name) || items == nil || len(items) > 200 {
			return fmt.Errorf("multisnip %q must be an array of at most 200 items", name)
		}
		for _, item := range items {
			if !ValidID(item.ID) || ids[item.ID] {
				return fmt.Errorf("invalid or duplicate multisnip item ID")
			}
			ids[item.ID] = true
			if len(item.Fields) == 0 || len(item.Fields) > 30 {
				return fmt.Errorf("multisnip items need 1–30 fields")
			}
			for key, value := range item.Fields {
				if !ValidFieldName(key) || len(value) > 64<<10 {
					return fmt.Errorf("invalid or oversized multisnip field")
				}
			}
		}
	}
	return nil
}

func CloneMultisnips(groups map[string][]MultiItem) map[string][]MultiItem {
	if groups == nil {
		return nil
	}
	out := map[string][]MultiItem{}
	for name, items := range groups {
		copyItems := make([]MultiItem, 0, len(items))
		for _, item := range items {
			copyItems = append(copyItems, MultiItem{item.ID, maps.Clone(item.Fields)})
		}
		out[name] = copyItems
	}
	return out
}
