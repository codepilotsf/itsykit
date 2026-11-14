package site

import (
	"fmt"
	"path"
	"strings"
)

type pageOptions struct {
	Layout      string
	Title       string
	Description string
}

// pageSettings is deliberately a small key:value format, not YAML. Only a
// leading itsy (or legacy flea) comment is metadata; ordinary comments remain inert.
func pageSettings(file, source string) (body string, options pageOptions, err error) {
	body = source
	trimmed := strings.TrimLeft(source, " \t\r\n\ufeff")
	position := len(source) - len(trimmed)
	defer func() {
		if err != nil {
			err = sourceError(file, source, position, err)
		}
	}()
	prefix := settingsCommentPrefix(trimmed)
	if prefix == "" {
		return
	}
	end := strings.Index(trimmed, "-->")
	if end < 0 {
		return "", options, fmt.Errorf("unclosed itsy settings comment")
	}
	settings := strings.TrimPrefix(trimmed[:end], prefix)
	position += len(prefix)
	seen := map[string]bool{}
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
		key = strings.TrimSpace(key)
		if !ok || (key != "layout" && key != "title" && key != "description") {
			return "", options, fmt.Errorf("unknown page setting %q; supported settings: layout, title, description", line)
		}
		if seen[key] {
			return "", options, fmt.Errorf("duplicate %s setting", key)
		}
		seen[key] = true
		value = strings.TrimSpace(value)
		switch key {
		case "title":
			options.Title = value
		case "description":
			options.Description = value
		case "layout":
			if !strings.HasPrefix(value, "/") || !renderPathPattern.MatchString(value) || path.Clean(value) != value || !strings.HasSuffix(value, ".layout.html") || strings.Contains(value, "..") {
				return "", options, fmt.Errorf("layout requires a root-relative private HTML path, for example /admin.layout.html")
			}
			options.Layout = strings.TrimPrefix(value, "/")
			if options.Layout == "_404.html" || options.Layout == "_500.html" {
				return "", options, fmt.Errorf("error pages cannot be used as layouts")
			}
		}
		position = next
	}
	body = trimmed[end+3:]
	rest := strings.TrimSpace(body)
	if settingsCommentPrefix(rest) != "" {
		position = len(source) - len(rest)
		return "", options, fmt.Errorf("use one itsy settings comment per page")
	}
	return
}

func settingsCommentPrefix(source string) string {
	for _, prefix := range []string{"<!-- itsy", "<!-- flea"} {
		if strings.HasPrefix(source, prefix+"\n") || strings.HasPrefix(source, prefix+"\r\n") {
			return prefix
		}
	}
	return ""
}

func (p *Project) collectLayouts() error {
	p.layouts = map[string]bool{"main.layout.html": true}
	p.pageOptions = map[string]pageOptions{}
	for file, source := range p.files {
		if strings.HasSuffix(file, ".layout.html") {
			p.layouts[file] = true
		}
		_, options, err := pageSettings(file, source)
		if err != nil {
			return fmt.Errorf("%s: %w", file, err)
		}
		p.pageOptions[file] = options
		layout := options.Layout
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
			_, options, err := pageSettings(file, source)
			if err != nil {
				return err
			}
			if options.Layout != "" {
				return fmt.Errorf("%s: layouts cannot select another layout", file)
			}
			if options.Title != "" || options.Description != "" {
				return fmt.Errorf("%s: page titles and descriptions belong in pages, not layouts", file)
			}
		}
	}
	for file, source := range p.Partials {
		_, options, _ := pageSettings(file, source)
		if options != (pageOptions{}) {
			return fmt.Errorf("%s: partials cannot have page settings", file)
		}
	}
	return nil
}
