package site

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Only explicitly referenced SVGs are downloaded. No icon library is embedded.
const LucideVersion = "1.47.0"
const iconCacheDir = "_/icons/lucide-" + LucideVersion
const iconSource = "https://unpkg.com/lucide-static@" + LucideVersion + "/"
const maxIconBytes = 64 << 10

var iconNamePattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
var iconAttributePattern = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

type missingIconError struct{ name string }

type invalidIconError struct{ err error }

func (e *invalidIconError) Error() string { return e.err.Error() }
func (e *invalidIconError) Unwrap() error { return e.err }

func (e *missingIconError) Error() string {
	return fmt.Sprintf("icon %q is not cached; run flea dev to download it into %s/", e.name, iconCacheDir)
}

// LoadDev resolves missing icons before publishing a project. Load (used by
// status) stays read-only. Downloads never happen in an HTTP request handler.
func LoadDev(ctx context.Context, root string) (*Project, error) {
	client := &http.Client{Timeout: 15 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error {
		return errors.New("icon source redirects are not allowed")
	}}
	return loadWithIcons(ctx, root, func(ctx context.Context, name string) error {
		return downloadIcon(ctx, root, name, client)
	})
}

func loadWithIcons(ctx context.Context, root string, fetch func(context.Context, string) error) (*Project, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	p, err := Load(root)
	if err != nil || len(p.missingIcons) == 0 {
		return p, err
	}
	names := make([]string, 0, len(p.missingIcons))
	for name := range p.missingIcons {
		names = append(names, name)
	}
	sort.Strings(names)
	failures := map[string]string{}
	for i, name := range names {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if i >= 256 {
			failures[name] = "download limit reached; restart dev to retry"
			continue
		}
		if err := fetch(ctx, name); err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			var invalid *invalidIconError
			if errors.As(err, &invalid) {
				return nil, fmt.Errorf("%s: %w", p.missingIcons[name], err)
			}
			failures[name] = err.Error()
		}
	}
	p, err = Load(root)
	if err != nil {
		return nil, err
	}
	p.Warnings = nil
	names = names[:0]
	for name := range p.missingIcons {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		warning := p.missingIcons[name]
		if failure := failures[name]; failure != "" {
			warning += " (" + failure + ")"
		}
		p.Warnings = append(p.Warnings, warning)
	}
	return p, nil
}

func (p *Project) iconBody(name string) (string, error) {
	if body, ok := p.icons[name]; ok {
		return body, nil
	}
	root, err := os.OpenRoot(p.Root)
	if err != nil {
		return "", err
	}
	defer root.Close()
	data, err := readFile(root, iconCacheDir+"/"+name+".svg")
	if os.IsNotExist(err) {
		return "", &missingIconError{name}
	}
	if err != nil {
		return "", err
	}
	body, err := sanitizeIcon(data)
	if err != nil {
		return "", fmt.Errorf("cached icon %q: %w", name, err)
	}
	if p.icons == nil {
		p.icons = map[string]string{}
	}
	p.icons[name] = body
	return body, nil
}

// Reconstruct SVG from a small geometry-only vocabulary. Scripts, external
// references, namespaces, CSS, and template delimiters cannot enter the renderer.
func sanitizeIcon(data []byte) (string, error) {
	if len(data) > maxIconBytes {
		return "", fmt.Errorf("SVG exceeds %d bytes", maxIconBytes)
	}
	decoder := xml.NewDecoder(bytes.NewReader(data))
	var out strings.Builder
	depth, roots, shapes := 0, 0, 0
	geometry := map[string]bool{"path": true, "circle": true, "ellipse": true, "rect": true, "line": true, "polyline": true, "polygon": true}
	attrs := map[string]bool{"d": true, "cx": true, "cy": true, "r": true, "x": true, "y": true, "width": true, "height": true, "rx": true, "ry": true, "x1": true, "x2": true, "y1": true, "y2": true, "points": true, "fill": true}
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("invalid SVG: %w", err)
		}
		switch node := token.(type) {
		case xml.StartElement:
			if node.Name.Space != "" && node.Name.Space != "http://www.w3.org/2000/svg" {
				return "", fmt.Errorf("unsupported SVG namespace")
			}
			if depth == 0 {
				roots++
				if roots != 1 || node.Name.Local != "svg" {
					return "", fmt.Errorf("expected one SVG root")
				}
				viewBox := ""
				for _, a := range node.Attr {
					if a.Name.Local == "viewBox" {
						viewBox = a.Value
					}
				}
				if viewBox != "0 0 24 24" {
					return "", fmt.Errorf("expected Lucide viewBox 0 0 24 24")
				}
			} else {
				if depth != 1 || !geometry[node.Name.Local] {
					return "", fmt.Errorf("unsupported SVG element %q", node.Name.Local)
				}
				shapes++
				out.WriteString("<" + node.Name.Local)
				seen := map[string]bool{}
				for _, a := range node.Attr {
					if a.Name.Space != "" || !attrs[a.Name.Local] || seen[a.Name.Local] || strings.ContainsAny(a.Value, "{}") || strings.Contains(a.Value, "[[NW") {
						return "", fmt.Errorf("unsupported SVG attribute %q", a.Name.Local)
					}
					if a.Name.Local == "fill" && a.Value != "none" && a.Value != "currentColor" {
						return "", fmt.Errorf("unsupported SVG fill")
					}
					seen[a.Name.Local] = true
					fmt.Fprintf(&out, ` %s="%s"`, a.Name.Local, html.EscapeString(a.Value))
				}
				out.WriteString(">")
			}
			depth++
		case xml.EndElement:
			depth--
			if depth > 0 {
				out.WriteString("</" + node.Name.Local + ">")
			}
		case xml.CharData:
			if strings.TrimSpace(string(node)) != "" {
				return "", fmt.Errorf("SVG text is unsupported")
			}
		case xml.Comment:
			// License comments are not markup; the full license is cached separately.
		default:
			return "", fmt.Errorf("SVG declarations and processing instructions are unsupported")
		}
	}
	if roots != 1 || shapes == 0 || depth != 0 {
		return "", fmt.Errorf("empty or incomplete SVG")
	}
	return out.String(), nil
}

// Open each private directory through a pinned handle, rejecting symlinks.
func openIconCache(rootPath string) (*os.Root, error) {
	root, err := os.OpenRoot(rootPath)
	if err != nil {
		return nil, err
	}
	for _, component := range strings.Split(iconCacheDir, "/") {
		if err := root.Mkdir(component, 0755); err != nil && !os.IsExist(err) {
			root.Close()
			return nil, err
		}
		info, err := root.Lstat(component)
		if err != nil {
			root.Close()
			return nil, err
		}
		if !info.IsDir() {
			root.Close()
			return nil, fmt.Errorf("icon cache directory must not be a symlink: %s", component)
		}
		next, err := root.OpenRoot(component)
		root.Close()
		if err != nil {
			return nil, err
		}
		actual, err := next.Stat(".")
		if err != nil || !os.SameFile(info, actual) {
			next.Close()
			return nil, fmt.Errorf("icon cache directory changed during open")
		}
		root = next
	}
	return root, nil
}

func downloadIcon(ctx context.Context, rootPath, name string, client *http.Client) error {
	if len(name) > 80 || !iconNamePattern.MatchString(name) {
		return fmt.Errorf("invalid icon name %q", name)
	}
	fetch := func(file string, limit int64) ([]byte, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, iconSource+file, nil)
		if err != nil {
			return nil, err
		}
		response, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("Lucide %s: HTTP %d for %s", LucideVersion, response.StatusCode, file)
		}
		data, err := io.ReadAll(io.LimitReader(response.Body, limit+1))
		if err != nil {
			return nil, err
		}
		if int64(len(data)) > limit {
			return nil, fmt.Errorf("download exceeds %d bytes", limit)
		}
		return data, nil
	}
	data, err := fetch("icons/"+name+".svg", maxIconBytes)
	if err != nil {
		return err
	}
	if _, err := sanitizeIcon(data); err != nil {
		return &invalidIconError{err}
	}
	cache, err := openIconCache(rootPath)
	if err != nil {
		return err
	}
	defer cache.Close()
	if _, err := readFile(cache, "LICENSE"); os.IsNotExist(err) {
		license, err := fetch("LICENSE", 128<<10)
		if err != nil {
			return err
		}
		if err := writeSiteFile(cache, "LICENSE", license); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return writeSiteFile(cache, name+".svg", data)
}

// Icon attributes use quoted HTML-style values. Only value lookups may appear
// inside them, keeping content contextually escaped through html/template.
func (p *Project) compileIcon(token string, lookup func(string) (string, error)) (string, error) {
	rest := strings.TrimSpace(strings.TrimPrefix(token, "icon"))
	end := strings.IndexAny(rest, " \t\r\n")
	if end < 0 {
		end = len(rest)
	}
	name := rest[:end]
	rest = rest[end:]
	if len(name) > 80 || !iconNamePattern.MatchString(name) {
		return "", fmt.Errorf("icon requires a literal lowercase kebab-case name")
	}
	values := map[string]string{"width": "24", "height": "24", "fill": "none", "stroke": "currentColor", "stroke-width": "2", "stroke-linecap": "round", "stroke-linejoin": "round", "focusable": "false"}
	allowed := map[string]bool{"class": true, "id": true, "size": true, "width": true, "height": true, "fill": true, "stroke": true, "stroke-width": true, "stroke-linecap": true, "stroke-linejoin": true, "color": true, "role": true, "tabindex": true}
	seen := map[string]bool{}
	for strings.TrimSpace(rest) != "" {
		if !strings.ContainsAny(rest[:1], " \t\r\n") {
			return "", fmt.Errorf("icon attributes must be separated by whitespace")
		}
		rest = strings.TrimSpace(rest)
		eq := strings.IndexByte(rest, '=')
		if eq < 0 {
			return "", fmt.Errorf("icon attributes require quoted values")
		}
		key := strings.TrimSpace(rest[:eq])
		rest = strings.TrimSpace(rest[eq+1:])
		if !iconAttributePattern.MatchString(key) || (!allowed[key] && !strings.HasPrefix(key, "aria-") && !strings.HasPrefix(key, "data-")) {
			return "", fmt.Errorf("unsupported icon attribute %q", key)
		}
		if seen[key] {
			return "", fmt.Errorf("duplicate icon attribute %q", key)
		}
		seen[key] = true
		if len(rest) == 0 || (rest[0] != '\'' && rest[0] != '"') {
			return "", fmt.Errorf("icon attribute %q requires a quoted value", key)
		}
		quote := rest[0]
		rest = rest[1:]
		close := strings.IndexByte(rest, quote)
		if close < 0 {
			return "", fmt.Errorf("unclosed icon attribute %q", key)
		}
		value := rest[:close]
		rest = rest[close+1:]
		var compiled strings.Builder
		for len(value) > 0 {
			start := strings.IndexByte(value, '{')
			if start < 0 {
				compiled.WriteString(iconLiteral(value))
				break
			}
			compiled.WriteString(iconLiteral(value[:start]))
			stop := strings.IndexByte(value[start:], '}')
			if stop < 0 {
				return "", fmt.Errorf("unclosed icon attribute expression")
			}
			stop += start
			expr, err := lookup(strings.TrimSpace(value[start+1 : stop]))
			if err != nil {
				return "", err
			}
			compiled.WriteString("[[NW " + expr + " NW]]")
			value = value[stop+1:]
		}
		values[key] = compiled.String()
	}
	if size, ok := values["size"]; ok {
		if !seen["width"] {
			values["width"] = size
		}
		if !seen["height"] {
			values["height"] = size
		}
		delete(values, "size")
	}
	if _, ok := values["aria-label"]; ok {
		if !seen["role"] {
			values["role"] = "img"
		}
	} else if _, ok := values["aria-labelledby"]; ok {
		if !seen["role"] {
			values["role"] = "img"
		}
	} else if !seen["aria-hidden"] {
		values["aria-hidden"] = "true"
	}
	values["class"] = strings.TrimSpace("lucide lucide-" + name + " " + values["class"])
	body, err := p.iconBody(name)
	if err != nil {
		return "", err
	}
	var out strings.Builder
	out.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"`)
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Fprintf(&out, ` %s="%s"`, key, values[key])
	}
	out.WriteString(">" + body + "</svg>")
	return out.String(), nil
}

func iconLiteral(value string) string {
	// Entity-encoded internal delimiters must remain literal HTML, not become
	// executable Go template source after HTML entity decoding.
	return strings.ReplaceAll(html.EscapeString(html.UnescapeString(value)), "[", "&#91;")
}
