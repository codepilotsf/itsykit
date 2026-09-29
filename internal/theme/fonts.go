package theme

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

const fontStart = "/* Flea Google Fonts */"
const fontEnd = "/* End Flea Google Fonts */"

var familyName = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9 -]{0,99}$`)
var systemFamilies = strings.Split("serif,sans-serif,monospace,cursive,fantasy,system-ui,ui-serif,ui-sans-serif,ui-monospace,ui-rounded,emoji,math,fangsong,arial,verdana,georgia,times new roman,courier new,helvetica,tahoma,trebuchet ms,impact,comic sans ms,segoe ui,apple color emoji,inherit,initial,unset,revert,revert-layer", ",")

// FontFamily accepts a family stack, never CSS declarations, imports or URLs.
// Generic and common system fonts do not need a remote stylesheet.
func FontFamily(value string) (string, error) {
	parts := strings.Split(value, ",")
	if len(parts) > 10 {
		return "", fmt.Errorf("too many font fallbacks")
	}
	first := ""
	for i, part := range parts {
		name := strings.TrimSpace(part)
		if len(name) >= 2 && (name[0] == '\'' || name[0] == '"') && name[len(name)-1] == name[0] {
			name = name[1 : len(name)-1]
		}
		if !familyName.MatchString(name) {
			return "", fmt.Errorf("enter a font family, such as \"Lora\", serif")
		}
		if i == 0 {
			first = name
		}
	}
	for _, name := range systemFamilies {
		if strings.EqualFold(first, name) {
			return "", nil
		}
	}
	return first, nil
}

func FontURL(values map[string]string) string {
	query := url.Values{}
	seen := map[string]bool{}
	for _, key := range []string{"--font-body", "--font-heading"} {
		family, err := FontFamily(values[key])
		if err != nil || family == "" || seen[family] {
			continue
		}
		seen[family] = true
		query.Add("family", family)
	}
	if len(seen) == 0 {
		return ""
	}
	query.Set("display", "swap")
	return "https://fonts.googleapis.com/css2?" + query.Encode()
}

func updateFonts(css string, changes map[string]string) (string, error) {
	changed := false
	for _, key := range []string{"--font-body", "--font-heading"} {
		if value, ok := changes[key]; ok {
			if _, err := FontFamily(value); err != nil {
				return "", err
			}
			changed = true
		}
	}
	if !changed {
		return css, nil
	}
	fields, err := Fields(css)
	if err != nil {
		return "", err
	}
	values := map[string]string{}
	for _, f := range fields {
		values[f.Name] = f.Value
	}
	// Only our own import block is replaced. Developer imports and local font
	// declarations stay byte-for-byte intact, even if currently unused.
	a, b := strings.Index(css, fontStart), strings.Index(css, fontEnd)
	if a >= 0 || b >= 0 {
		if a < 0 || b < a || strings.Count(css, fontStart) != 1 || strings.Count(css, fontEnd) != 1 {
			return "", fmt.Errorf("invalid managed font import block")
		}
		css = css[:a] + css[b+len(fontEnd):]
	}
	if address := FontURL(values); address != "" {
		block := fontStart + "\n@import url(\"" + address + "\");\n" + fontEnd
		if a >= 0 {
			css = css[:a] + block + css[a:]
		} else {
			// @charset must remain the first statement; imports precede rules.
			pos := 0
			if strings.HasPrefix(css, "\ufeff") {
				pos = len("\ufeff")
			}
			if strings.HasPrefix(css[pos:], "@charset ") {
				end := strings.IndexByte(css[pos:], ';')
				if end < 0 {
					return "", fmt.Errorf("invalid charset declaration")
				}
				pos += end + 1
			}
			css = css[:pos] + "\n" + block + "\n" + css[pos:]
		}
	}
	return css, nil
}
