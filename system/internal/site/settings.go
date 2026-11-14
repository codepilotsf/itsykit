package site

import (
	"fmt"
	"path"
	"strings"
)

// pageSettings is deliberately a small key:value format, not YAML. Only a
// leading itsy (or legacy flea) comment is metadata; ordinary comments remain inert.
type pageConfig struct{ Layout, Title, Description string }

func pageSettings(file, source string) (body string, settingsValue pageConfig, err error) {
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
		return "", pageConfig{}, fmt.Errorf("unclosed itsy settings comment")
	}
	seen := map[string]bool{}
	settings := strings.TrimPrefix(trimmed[:end], prefix)
	position += len(prefix)
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
			return "", pageConfig{}, fmt.Errorf("unknown page setting %q; supported settings: layout, title, description", line)
		}
		if seen[key] {
			return "", pageConfig{}, fmt.Errorf("duplicate %s setting", key)
		}
		seen[key] = true
		value = strings.TrimSpace(value)
		if key == "title" {
			settingsValue.Title = value
			position = next
			continue
		}
		if key == "description" {
			settingsValue.Description = value
			position = next
			continue
		}
		if !strings.HasPrefix(value, "/") || !renderPathPattern.MatchString(value) || path.Clean(value) != value || !strings.HasSuffix(value, ".layout.html") || strings.Contains(value, "..") {
			return "", pageConfig{}, fmt.Errorf("layout requires a root-relative private HTML path, for example /admin.layout.html")
		}
		settingsValue.Layout = strings.TrimPrefix(value, "/")
		if settingsValue.Layout == "_404.html" || settingsValue.Layout == "_500.html" {
			return "", pageConfig{}, fmt.Errorf("error pages cannot be used as layouts")
		}
		position = next
	}
	body = trimmed[end+3:]
	rest := strings.TrimSpace(body)
	if settingsCommentPrefix(rest) != "" {
		position = len(source) - len(rest)
		return "", pageConfig{}, fmt.Errorf("use one itsy settings comment per page")
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
	p.pageSettings = map[string]pageConfig{}
	p.layouts = map[string]bool{"main.layout.html": true}
	for file, source := range p.files {
		if strings.HasSuffix(file, ".layout.html") {
			p.layouts[file] = true
		}
		_, layout, err := pageSettings(file, source)
		if err != nil {
			return fmt.Errorf("%s: %w", file, err)
		}
		p.pageSettings[file] = layout
		if layout.Layout != "" {
			if file == "_500.html" {
				return fmt.Errorf("_500.html must remain standalone")
			}
			if _, ok := p.files[layout.Layout]; !ok {
				return fmt.Errorf("%s: missing layout %q", file, layout.Layout)
			}
			p.layouts[layout.Layout] = true
		}
	}
	for file := range p.layouts {
		delete(p.Partials, file)
		if source, ok := p.files[file]; ok {
			_, parent, err := pageSettings(file, source)
			if err != nil {
				return err
			}
			if parent.Layout != "" {
				return fmt.Errorf("%s: layouts cannot select another layout", file)
			}
		}
	}
	for file, source := range p.Partials {
		_, layout, _ := pageSettings(file, source)
		if layout.Layout != "" {
			return fmt.Errorf("%s: partials cannot select a layout", file)
		}
	}
	return nil
}
