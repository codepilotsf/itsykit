package site

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Project struct {
	Root      string
	Site      map[string]any
	Snips     map[string]map[string]any
	Partials  map[string]string
	Pages     map[string]*template.Template
	Assets    map[string]string
	AssetData map[string][]byte
	Errors    map[int]*template.Template
	files     map[string]string
	signature string
}

// Open each path component through pinned directory handles. Comparing lstat
// with the opened handle rejects symlink swaps during a reload as well as
// symlinks already present at validation time.
func readFile(root *os.Root, name string) ([]byte, error) {
	parts := strings.Split(name, "/")
	current := root
	for _, part := range parts[:len(parts)-1] {
		info, err := current.Lstat(part)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("not a real directory: %s", name)
		}
		next, err := current.OpenRoot(part)
		if err != nil {
			return nil, err
		}
		defer next.Close()
		actual, err := next.Stat(".")
		if err != nil {
			return nil, err
		}
		if !os.SameFile(info, actual) {
			return nil, fmt.Errorf("directory changed while loading: %s", name)
		}
		current = next
	}
	base := parts[len(parts)-1]
	info, err := current.Lstat(base)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("not a regular file (symlinks unsupported): %s", name)
	}
	f, err := current.Open(base)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	actual, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !os.SameFile(info, actual) {
		return nil, fmt.Errorf("file changed while loading: %s", name)
	}
	return io.ReadAll(f)
}

func readJSON(root *os.Root, path string, target any) error {
	data, err := readFile(root, path)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("%s: invalid JSON: %w", path, err)
	}
	return nil
}

func Load(root string) (*Project, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	boundary, err := os.OpenRoot(root)
	if err != nil {
		return nil, err
	}
	defer boundary.Close()
	p := &Project{Root: root, Snips: map[string]map[string]any{}, Partials: map[string]string{}, Pages: map[string]*template.Template{}, Assets: map[string]string{}, Errors: map[int]*template.Template{}, files: map[string]string{}}
	p.signature = fingerprint(root)
	var config struct {
		Site map[string]any `json:"site"`
	}
	// Reject symbolic links throughout the site to keep both public and private
	// reads within the project. Deployment can reuse this validation boundary.
	for _, directory := range []string{"_", "html"} {
		err = fs.WalkDir(boundary.FS(), directory, func(path string, d fs.DirEntry, e error) error {
			if e != nil {
				return e
			}
			if d.Type()&os.ModeSymlink != 0 {
				return fmt.Errorf("symlinks are unsupported: %s", path)
			}
			if !d.IsDir() && !d.Type().IsRegular() {
				return fmt.Errorf("unsupported file: %s", path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	// Validate the tree before reading configuration or any content.
	if err = readJSON(boundary, "_/nanoweb.config", &config); err != nil {
		return nil, err
	}
	p.Site = config.Site
	if name, ok := p.Site["name"].(string); !ok || strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("_/nanoweb.config: site.name must be a nonempty string")
	}
	p.AssetData = map[string][]byte{}
	entries, err := fs.ReadDir(boundary.FS(), "_/snips")
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".json")
		if !namePattern.MatchString(name) {
			return nil, fmt.Errorf("invalid snip name %q", name)
		}
		var data map[string]any
		if err = readJSON(boundary, "_/snips/"+entry.Name(), &data); err != nil {
			return nil, err
		}
		if data == nil {
			return nil, fmt.Errorf("snip %s must contain a JSON object", name)
		}
		p.Snips[name] = data
	}
	err = fs.WalkDir(boundary.FS(), "html", func(path string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.IsDir() {
			return nil
		}
		rel := strings.TrimPrefix(path, "html/")
		if strings.HasSuffix(rel, ".html") {
			data, e := readFile(boundary, path)
			if e != nil {
				return e
			}
			if bytes.Contains(data, []byte("[[NW")) {
				return fmt.Errorf("%s: reserved template delimiter [[NW", rel)
			}
			p.files[rel] = string(data)
		}
		if strings.HasPrefix(d.Name(), "_") && strings.HasSuffix(rel, ".html") && rel != "_layout.html" && rel != "_404.html" && rel != "_500.html" {
			p.Partials[rel] = p.files[rel]
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	// Compile all partials, even unused ones, so broken includes fail validation.
	for name, text := range p.Partials {
		if _, err = p.compile(text, name, []string{name}, map[string]string{}); err != nil {
			return nil, err
		}
	}
	routes := map[string]string{}
	err = fs.WalkDir(boundary.FS(), "html", func(path string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if path == "html" {
			return nil
		}
		rel := strings.TrimPrefix(path, "html/")
		if strings.HasPrefix(d.Name(), "_") || strings.HasPrefix(d.Name(), ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}
		route := "/" + rel
		isPage := strings.HasSuffix(rel, ".html")
		if isPage {
			route = strings.TrimSuffix(route, ".html")
			if route == "/index" {
				route = "/"
			} else {
				route = strings.TrimSuffix(route, "/index")
			}
		}
		if !safeURL(route) {
			return fmt.Errorf("%s: unsupported public URL %q", rel, route)
		}
		if route == "/admin" || strings.HasPrefix(route, "/admin/") {
			return fmt.Errorf("%s: /admin is reserved for the editor", rel)
		}
		if previous, ok := routes[route]; ok {
			return fmt.Errorf("route collision at %s: %s and %s", route, previous, rel)
		}
		routes[route] = rel
		if isPage {
			t, e := p.buildTemplate(rel, true)
			if e != nil {
				return e
			}
			p.Pages[route] = t
		} else {
			p.Assets[route] = path
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	for code, name := range map[int]string{404: "_404.html", 500: "_500.html"} {
		if _, ok := p.files[name]; ok {
			t, e := p.buildTemplate(name, false)
			if e != nil {
				return nil, e
			}
			p.Errors[code] = t
		}
	}
	// Assets are immutable snapshots, never opened from a request-controlled path.
	for route, path := range p.Assets {
		data, e := readFile(boundary, path)
		if e != nil {
			return nil, e
		}
		p.AssetData[route] = data
	}
	for route, t := range p.Pages {
		if _, err = p.render(t, route); err != nil {
			return nil, fmt.Errorf("%s: %w", routes[route], err)
		}
	}
	for code, t := range p.Errors {
		if _, err = p.render(t, fmt.Sprintf("/%d", code)); err != nil {
			return nil, err
		}
	}
	return p, nil
}

func (p *Project) buildTemplate(file string, layout bool) (*template.Template, error) {
	body, err := p.compile(p.files[file], file, nil, map[string]string{})
	if err != nil {
		return nil, err
	}
	source := body
	if wrapper, ok := p.files["_layout.html"]; layout && ok {
		compiled, e := p.compile(wrapper, "_layout.html", nil, map[string]string{})
		if e != nil {
			return nil, e
		}
		source = `[[NW define "body" NW]]` + body + `[[NW end NW]]` + compiled
	}
	t, err := template.New(file).Delims("[[NW", "NW]]").Funcs(template.FuncMap{"lookup": lookup, "collection": collection}).Parse(source)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}
	return t, nil
}

func (p *Project) render(t *template.Template, route string) ([]byte, error) {
	title := "Home"
	if route != "/" {
		parts := strings.Split(strings.Trim(route, "/"), "/")
		title = strings.ReplaceAll(parts[len(parts)-1], "-", " ")
		if len(title) > 0 {
			title = strings.ToUpper(title[:1]) + title[1:]
		}
	}
	snips := map[string]any{}
	for k, v := range p.Snips {
		snips[k] = v
	}
	data := map[string]any{"site": p.Site, "page": map[string]any{"title": title, "path": route}, "snip": snips}
	var out bytes.Buffer
	err := t.Execute(&out, data)
	return out.Bytes(), err
}

func (p *Project) Routes() []string {
	var routes []string
	for route := range p.Pages {
		routes = append(routes, route)
	}
	sort.Strings(routes)
	return routes
}
