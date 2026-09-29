package site

import (
	"fmt"
	"path"
	"strings"
)

// pageSettings is deliberately a small key:value format, not YAML. Only a
// leading flea comment is metadata; ordinary HTML comments remain inert.
func pageSettings(file, source string) (body, layout string, err error) {
	body = source
	trimmed := strings.TrimLeft(source, " \t\r\n\ufeff")
	position := len(source) - len(trimmed)
	defer func() {
		if err != nil {
			err = sourceError(file, source, position, err)
		}
	}()
	if !strings.HasPrefix(trimmed, "<!-- flea\n") && !strings.HasPrefix(trimmed, "<!-- flea\r\n") {
		return
	}
	end := strings.Index(trimmed, "-->")
	if end < 0 {
		return "", "", fmt.Errorf("unclosed flea settings comment")
	}
	settings := strings.TrimPrefix(trimmed[:end], "<!-- flea")
	position += len("<!-- flea")
	for _, line := range strings.Split(settings, "\n") {
		lineStart := position
		position += len(line) + 1
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Keep the diagnostic on this setting if validation fails.
		next := position
		position = lineStart
		key, value, ok := strings.Cut(line, ":")
		if !ok || strings.TrimSpace(key) != "layout" {
			return "", "", fmt.Errorf("unknown page setting %q; supported setting: layout", line)
		}
		if layout != "" {
			return "", "", fmt.Errorf("duplicate layout setting")
		}
		value = strings.TrimSpace(value)
		if !strings.HasPrefix(value, "/") || !renderPathPattern.MatchString(value) || path.Clean(value) != value || !strings.HasSuffix(value, ".layout.html") || strings.Contains(value, "..") {
			return "", "", fmt.Errorf("layout requires a root-relative private HTML path, for example /admin.layout.html")
		}
		layout = strings.TrimPrefix(value, "/")
		if layout == "_404.html" || layout == "_500.html" {
			return "", "", fmt.Errorf("error pages cannot be used as layouts")
		}
		position = next
	}
	body = trimmed[end+3:]
	rest := strings.TrimSpace(body)
	if strings.HasPrefix(rest, "<!-- flea\n") || strings.HasPrefix(rest, "<!-- flea\r\n") {
		position = len(source) - len(rest)
		return "", "", fmt.Errorf("use one flea settings comment per page")
	}
	return
}

func (p *Project) collectLayouts() error {
	p.layouts = map[string]bool{"main.layout.html": true}
	for file, source := range p.files {
		if strings.HasSuffix(file, ".layout.html") {
			p.layouts[file] = true
		}
		_, layout, err := pageSettings(file, source)
		if err != nil {
			return fmt.Errorf("%s: %w", file, err)
		}
		if layout != "" {
			if file == "_500.html" {
				return fmt.Errorf("_500.html must remain standalone")
			}
			if _, ok := p.files[layout]; !ok {
				return fmt.Errorf("%s: missing layout %q", file, layout)
			}
			p.layouts[layout] = true
		}
	}
	for file := range p.layouts {
		delete(p.Partials, file)
		if source, ok := p.files[file]; ok {
			_, parent, err := pageSettings(file, source)
			if err != nil {
				return err
			}
			if parent != "" {
				return fmt.Errorf("%s: layouts cannot select another layout", file)
			}
		}
	}
	for file, source := range p.Partials {
		_, layout, _ := pageSettings(file, source)
		if layout != "" {
			return fmt.Errorf("%s: partials cannot select a layout", file)
		}
	}
	return nil
}
