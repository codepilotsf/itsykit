package site

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"flea/internal/contact"
	"flea/internal/sitefiles"
	"flea/internal/snips"
)

// Project is a validated serving snapshot. After publishing it to a Server,
// callers must not mutate its maps or Content; reload builds a new Project.
type Project struct {
	Admin        AdminAccess
	Contact      contact.Config
	Content      *snips.Store
	rich         map[string]template.HTML
	Root         string
	Site         map[string]any
	Snips        map[string]map[string]any
	Partials     map[string]string
	Pages        map[string]*template.Template
	Assets       map[string]string
	AssetData    map[string][]byte
	Errors       map[int]*template.Template
	files        map[string]string
	signature    string
	icons        map[string]string
	Warnings     []string
	SnipsUpdated bool // A dev extraction changed templates or the active content index.
	missingIcons map[string]string
}

func Load(root string) (*Project, error) {
	return LoadWithSnips(root, nil)
}

// LoadWithSnips loads a code release with a separately managed content snapshot.
func LoadWithSnips(root string, content *snips.Store) (*Project, error) {
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
		Site    map[string]any `json:"site"`
		Admin   AdminAccess    `json:"admin"`
		Contact contact.Config `json:"contact"`
	}
	// Reject symbolic links throughout the site to keep both public and private
	// reads within the project. Deployment can reuse this validation boundary.
	err = fs.WalkDir(boundary.FS(), ".", func(path string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if !sitefiles.Included(path) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
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
	// Validate the tree before reading configuration or any content.
	if err = readJSON(boundary, "_/flea.config", &config); err != nil {
		return nil, err
	}
	p.Site = config.Site
	if err := config.Admin.validate(); err != nil {
		return nil, err
	}
	p.Admin = config.Admin
	if err := config.Contact.Validate(); err != nil {
		return nil, err
	}
	p.Contact = config.Contact
	if name, ok := p.Site["name"].(string); !ok || strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("_/flea.config: site.name must be a nonempty string")
	}
	if content == nil {
		content, err = readSnipStore(boundary)
		if err != nil {
			return nil, err
		}
	}
	p.Content = content
	p.rich = map[string]template.HTML{}
	contentIndex, err := content.Index()
	if err != nil {
		return nil, err
	}
	for id, entry := range contentIndex {
		clean, e := snips.Sanitize(entry.Content)
		if e != nil {
			return nil, fmt.Errorf("snip %s: %w", id, e)
		}
		p.rich[id] = template.HTML(clean)
	}
	p.AssetData = map[string][]byte{}
	// Legacy snips and structured data intentionally share a namespace.
	for _, source := range []struct{ directory, kind string }{
		{"_/snips", "snip"},
		{"_/data", "data"},
	} {
		if err := loadDataDirectory(boundary, source.directory, source.kind, p.Snips); err != nil {
			return nil, err
		}
	}
	err = fs.WalkDir(boundary.FS(), ".", func(path string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if !sitefiles.Included(path) || path == "_" {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}
		rel := path
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
	if _, err = p.SnipReferences(); err != nil {
		return nil, err
	}
	for name, text := range p.Partials {
		if _, err = p.compile(text, name, []string{name}, map[string]string{}); err != nil {
			return nil, err
		}
	}
	routes := map[string]string{}
	err = fs.WalkDir(boundary.FS(), ".", func(path string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if path == "." {
			return nil
		}
		if !sitefiles.Included(path) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		rel := path
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
			route = pageRoute(rel)
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
			t, e := p.buildTemplate(name, code == 404)
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
	names := make([]string, 0, len(p.missingIcons))
	for name := range p.missingIcons {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		p.Warnings = append(p.Warnings, p.missingIcons[name])
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
	t, err := template.New(file).Delims("[[NW", "NW]]").Funcs(template.FuncMap{"lookup": lookup, "collection": collection, "richSnip": func(id string) template.HTML { return p.rich[id] }}).Parse(source)
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
	description := ""
	if page := p.Content.Pages[route]; page != nil {
		if page.Title != "" {
			title = page.Title
		}
		description = page.Description
	}
	data := map[string]any{"site": p.Site, "page": map[string]any{"title": title, "description": description, "path": route}, "snip": snips, "data": snips}
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
