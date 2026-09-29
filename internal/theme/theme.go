// Package theme edits only existing custom properties in a single :root block.
// It preserves imports, comments, selectors, and all untouched source bytes.
package theme

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

const MaxBytes = 128 << 10

type Field struct {
	Name       string `json:"name"`
	Value      string `json:"value"`
	Group      string `json:"group"`
	Kind       string `json:"kind"`
	start, end int
}

var property = regexp.MustCompile(`--[a-zA-Z][a-zA-Z0-9-]*`)
var functions = regexp.MustCompile(`([a-zA-Z-]+)\s*\(`)
var safeChars = regexp.MustCompile(`^[a-zA-Z0-9#.,%() /+*'"_-]+$`)

func valid(value string) bool {
	if len(value) == 0 || len(value) > 1000 || !safeChars.MatchString(value) || strings.Contains(value, "/*") || strings.Contains(value, "*/") {
		return false
	}
	depth := 0
	quote := byte(0)
	for i := 0; i < len(value); i++ {
		c := value[i]
		if quote != 0 {
			if c == quote {
				quote = 0
			}
			continue
		}
		if c == '\'' || c == '"' {
			quote = c
			continue
		}
		if c == '(' {
			depth++
		}
		if c == ')' {
			depth--
			if depth < 0 {
				return false
			}
		}
	}
	if depth != 0 || quote != 0 {
		return false
	}
	for _, m := range functions.FindAllStringSubmatch(value, -1) {
		switch strings.ToLower(m[1]) {
		case "var", "calc", "clamp", "min", "max", "rgb", "rgba", "hsl", "hsla", "oklch", "oklab", "lch", "lab", "color", "color-mix":
		default:
			return false
		}
	}
	return true
}

func Fields(css string) ([]Field, error) {
	if len(css) > MaxBytes {
		return nil, fmt.Errorf("theme.css exceeds 128 KiB")
	}
	// Mask comments and strings while locating structural braces. Offsets remain
	// byte-identical, so a comment containing :root cannot select a false block.
	mask := []byte(css)
	quote := byte(0)
	comment := false
	for i := 0; i < len(mask); i++ {
		c := css[i]
		if comment {
			mask[i] = ' '
			if c == '*' && i+1 < len(mask) && css[i+1] == '/' {
				mask[i+1] = ' '
				i++
				comment = false
			}
			continue
		}
		if quote != 0 {
			mask[i] = ' '
			if c == '\\' && i+1 < len(mask) {
				i++
				mask[i] = ' '
				continue
			}
			if c == quote {
				quote = 0
			}
			continue
		}
		if c == '/' && i+1 < len(mask) && css[i+1] == '*' {
			mask[i] = ' '
			i++
			mask[i] = ' '
			comment = true
			continue
		}
		if c == '\'' || c == '"' {
			quote = c
			mask[i] = ' '
		}
	}
	if comment || quote != 0 {
		return nil, fmt.Errorf("unclosed theme comment or string")
	}
	roots := regexp.MustCompile(`:root\s*\{`).FindAllIndex(mask, -1)
	if len(roots) != 1 {
		return nil, fmt.Errorf("theme editing needs one :root block")
	}
	start := roots[0][1]
	end := strings.IndexByte(string(mask[start:]), '}')
	if end < 0 {
		return nil, fmt.Errorf("unclosed :root block")
	}
	end += start
	if strings.ContainsRune(string(mask[start:end]), '{') {
		return nil, fmt.Errorf("nested theme rules are not editable")
	}
	fields := []Field{}
	seen := map[string]bool{}
	for _, loc := range property.FindAllIndex(mask[start:end], -1) {
		a, b := start+loc[0], start+loc[1]
		// References inside var(...) are not declarations.
		j := b
		for j < end && strings.ContainsRune(" \t\r\n", rune(css[j])) {
			j++
		}
		if j >= end || css[j] != ':' {
			continue
		}
		name := css[a:b]
		if seen[name] {
			return nil, fmt.Errorf("duplicate theme property %s", name)
		}
		seen[name] = true
		from := j + 1
		for from < end && strings.ContainsRune(" \t\r\n", rune(css[from])) {
			from++
		}
		to := strings.IndexByte(string(mask[from:end]), ';')
		if to < 0 {
			return nil, fmt.Errorf("theme properties need semicolons")
		}
		to += from
		value := strings.TrimSpace(css[from:to])
		to = from + len(value)
		if !valid(value) {
			continue
		} // Unsupported developer CSS stays intact, never editable.
		kind, group := "text", "Other"
		switch {
		case name == "--font-color" || name == "--bg-color":
			kind, group = "color", "Page colors"
		case strings.HasPrefix(name, "--brand-"):
			kind, group = "color", "Brand colors"
		case strings.Contains(" --black --darker --dark --midtone --light --lighter --white ", " "+name+" "):
			kind, group = "color", "Neutral colors"
		case strings.HasPrefix(name, "--font-"):
			group = "Typography"
			if name == "--font-body" || name == "--font-heading" {
				kind = "font"
			}
			if name == "--font-color" {
				kind = "color"
			}
		case strings.HasPrefix(name, "--text-"):
			group = "Text sizes"
		case strings.HasPrefix(name, "--space-"):
			group = "Spacing"
		case strings.HasPrefix(name, "--rounded") || strings.HasPrefix(name, "--border") || strings.HasPrefix(name, "--shadow"):
			group = "Shape & shadows"
		case strings.HasPrefix(name, "--container") || strings.HasPrefix(name, "--max-w"):
			group = "Layout"
		case strings.HasPrefix(name, "--field-"):
			group = "Forms"
			if name == "--field-bg" || name == "--field-focus-color" {
				kind = "color"
			}
		default:
			kind, group = "color", "Semantic colors"
		}
		fields = append(fields, Field{name, value, group, kind, from, to})
	}
	return fields, nil
}

func Apply(css string, changes map[string]string) (string, error) {
	fields, err := Fields(css)
	if err != nil {
		return "", err
	}
	known := map[string]Field{}
	for _, f := range fields {
		known[f.Name] = f
	}
	edits := []Field{}
	for name, value := range changes {
		if name == "--font-color" || name == "--bg-color" {
			allowed := false
			for _, color := range fields {
				if (color.Group == "Neutral colors" || color.Group == "Brand colors") && color.Kind == "color" && value == "var("+color.Name+")" {
					allowed = true
					break
				}
			}
			if !allowed {
				return "", fmt.Errorf("page color must reference a neutral or brand color")
			}
		}
		f, ok := known[name]
		if !ok || !valid(value) {
			return "", fmt.Errorf("unsupported theme value for %s", name)
		}
		f.Value = value
		edits = append(edits, f)
	}
	sort.Slice(edits, func(i, j int) bool { return edits[i].start > edits[j].start })
	for _, f := range edits {
		css = css[:f.start] + f.Value + css[f.end:]
	}
	css, err = updateFonts(css, changes)
	if err != nil {
		return "", err
	}
	if len(css) > MaxBytes {
		return "", fmt.Errorf("theme too large")
	}
	return css, nil
}
