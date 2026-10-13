package site

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"flea/internal/sitefiles"
	"flea/internal/snips"
)

// Project is a validated serving snapshot. After publishing it to a Server,
// callers must not mutate its maps or Content; reload builds a new Project.
type Project struct {
	syncOnly           bool
	videos             map[string]videoAsset
	MediaRoot          string
	emails             map[string]emailTemplates
	layouts            map[string]bool
	adminTemplates     map[string]*template.Template
	Admin              AdminAccess
	Dev                DevSettings
	Apps               map[string]MiniApp
	Content            *snips.Store
	rich               map[string]template.HTML
	contentIndex       map[string]snips.Entry
	Root               string
	App                map[string]any
	SiteNameConfigured bool
	Snips              map[string]map[string]any
	Partials           map[string]string
	Pages              map[string]*template.Template
	Assets             map[string]string
	AssetData          map[string][]byte
	assetFiles         map[string]fileAsset
	Errors             map[int]*template.Template
	files              map[string]string
	signature          string
	icons              map[string]string
	Warnings           []string
	SnipsUpdated       bool // A dev extraction changed templates or the active content index.
	missingIcons       map[string]string
	csrfTemplates      map[*template.Template]bool
}

func Load(root string) (*Project, error) {
	return LoadWithSnips(root, nil)
}

// LoadWithSnips loads a code release with a separately managed content snapshot.
func LoadWithSnips(root string, content *snips.Store) (*Project, error) {
	return loadProject(root, content, "", false)
}

// LoadSyncSource validates metadata/templates for pull before ignored video files
// are available. Its result is deliberately not an HTTP serving snapshot.
func LoadSyncSource(root string) (*Project, error) { return loadProject(root, nil, "", true) }

func loadProject(root string, content *snips.Store, mediaRoot string, syncOnly bool) (*Project, error) {
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
	loaded := false
	defer func() {
		if !loaded {
			p.closeAssetFiles()
		}
	}()
	p.assetFiles = map[string]fileAsset{}
	p.MediaRoot = mediaRoot
	p.syncOnly = syncOnly
	if p.MediaRoot == "" {
		p.MediaRoot = filepath.Join(root, "flea", "media")
	}
	p.signature = fingerprint(root)
	var config struct {
		Dev     DevSettings     `json:"dev"`
		App     map[string]any  `json:"app"`
		Admin   AdminAccess     `json:"admin"`
		Contact json.RawMessage `json:"contact"`
		Apps    json.RawMessage `json:"apps"`
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
	if err = readJSON(boundary, ".flea.config.json", &config); err != nil {
		return nil, err
	}
	p.App = config.App
	p.Apps, err = decodeApps(config.Apps)
	if err != nil {
		return nil, err
	}
	if _, exists := p.App["current_admin"]; exists {
		return nil, fmt.Errorf("app.current_admin is reserved for the authenticated request; remove it from config")
	}
	if err := config.Admin.validate(); err != nil {
		return nil, err
	}
	p.Admin = config.Admin
	p.Dev = config.Dev
	if config.Contact != nil {
		return nil, fmt.Errorf("contact settings now belong in apps.contact; see the Contact app documentation")
	}
	if p.App == nil {
		return nil, fmt.Errorf(".flea.config.json: app must be a JSON object")
	}
	if value, exists := p.App["site_name"]; exists {
		name, ok := value.(string)
		if !ok || strings.TrimSpace(name) == "" {
			return nil, fmt.Errorf(".flea.config.json: app.site_name must be a nonempty string")
		}
		p.SiteNameConfigured = true
	} else {
		p.App["site_name"] = DefaultSiteName(root)
	}
	if err := p.loadEmails(boundary); err != nil {
		return nil, err
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
	p.contentIndex = contentIndex
	for id, entry := range contentIndex {
		if entry.Type != "" {
			continue
		}
		clean, e := snips.Sanitize(entry.Content)
		if e != nil {
			return nil, fmt.Errorf("snip %s: %w", id, e)
		}
		p.rich[id] = template.HTML(clean)
	}
	p.AssetData = map[string][]byte{}
	views, err := boundary.OpenRoot("site")
	if err != nil {
		return nil, fmt.Errorf("open site directory: %w", err)
	}
	defer views.Close()
	// Legacy snips and structured data intentionally share a namespace.
	for _, source := range []struct{ directory, kind string }{
		{"flea/snips", "snip"},
		{"flea/data", "data"},
	} {
		if err := loadDataDirectory(boundary, source.directory, source.kind, p.Snips); err != nil {
			return nil, err
		}
	}
	err = fs.WalkDir(views.FS(), ".", func(path string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if path != "." && !sitefiles.Included("site/"+path) {
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
			data, e := readFile(views, path)
			if e != nil {
				return e
			}
			if bytes.Contains(data, []byte("[[NW")) {
				return fmt.Errorf("%s: reserved template delimiter [[NW", rel)
			}
			p.files[rel] = string(data)
		}
		if strings.HasPrefix(d.Name(), "_") && strings.HasSuffix(rel, ".html") && rel != "main.layout.html" && rel != "_404.html" && rel != "_500.html" {
			p.Partials[rel] = p.files[rel]
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	// Compile all partials, even unused ones, so broken includes fail validation.
	if err := p.collectLayouts(); err != nil {
		return nil, err
	}
	for name := range p.layouts {
		if source, ok := p.files[name]; ok {
			if _, err := p.compile(source, name, []string{name}, map[string]string{}); err != nil {
				return nil, err
			}
		}
	}
	if _, err = p.SnipReferences(); err != nil {
		return nil, err
	}
	if err = p.checkContentStructure(); err != nil {
		return nil, err
	}
	for name, text := range p.Partials {
		if _, err = p.compile(text, name, []string{name}, map[string]string{}); err != nil {
			return nil, err
		}
	}
	routes := map[string]string{}
	err = fs.WalkDir(views.FS(), ".", func(path string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if path == "." {
			return nil
		}
		if !sitefiles.Included("site/" + path) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		rel := path
		if strings.HasSuffix(rel, ".layout.html") {
			return nil
		}
		if strings.HasPrefix(d.Name(), "_") || strings.HasPrefix(d.Name(), ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if rel == "admin" && d.IsDir() {
			return filepath.SkipDir
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
			p.Assets[route] = "site/" + path
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
		if e := p.loadAsset(boundary, route, path); e != nil {
			return nil, e
		}
	}
	if err = p.loadMedia(); err != nil {
		return nil, err
	}
	if !syncOnly {
		if err = p.loadVideos(); err != nil {
			return nil, err
		}
	}
	if err = p.validateApps(); err != nil {
		return nil, err
	}
	if err = p.loadBlog(); err != nil {
		return nil, err
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
	if err := p.buildAdminLayout(); err != nil {
		return nil, err
	}
	names := make([]string, 0, len(p.missingIcons))
	for name := range p.missingIcons {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		p.Warnings = append(p.Warnings, p.missingIcons[name])
	}
	loaded = true
	return p, nil
}

func (p *Project) buildTemplate(file string, layout bool) (*template.Template, error) {
	text, selected, err := pageSettings(file, p.files[file])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}
	if selected == "" {
		selected = "main.layout.html"
	}
	body, err := p.compile(text, file, nil, map[string]string{})
	if err != nil {
		return nil, err
	}
	source := body
	if wrapper, ok := p.files[selected]; layout && ok {
		compiled, e := p.compile(wrapper, selected, nil, map[string]string{})
		if e != nil {
			return nil, e
		}
		if e := checkCSRFContexts(strings.ReplaceAll(compiled, `[[NW template "body" $ NW]]`, body)); e != nil {
			return nil, fmt.Errorf("%s: %w", selected, e)
		}
		source = `[[NW define "body" NW]]` + body + `[[NW end NW]]` + compiled
	}
	if err := checkCSRFContexts(source); err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}
	t, err := template.New(file).Delims("[[NW", "NW]]").Funcs(template.FuncMap{"csrfField": csrfField, "blogBody": blogBody, "snipField": p.snipField, "multiPhoto": multiPhoto, "multiRows": p.multiRows, "multiValue": multiValue, "multiBoundary": multiBoundary, "multiItemBoundary": multiItemBoundary, "lookup": lookup, "collection": collection, "richFlea": func(id string, owner bool) template.HTML {
		if owner {
			return template.HTML("<!--snip-start:"+id+"-->") + p.rich[id] + template.HTML("<!--snip-end:"+id+"-->")
		}
		return p.rich[id]
	}}).Parse(source)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}
	if strings.Contains(source, csrfAction) {
		if p.csrfTemplates == nil {
			p.csrfTemplates = map[*template.Template]bool{}
		}
		p.csrfTemplates[t] = true
	}
	return t, nil
}

func (p *Project) render(t *template.Template, route string) ([]byte, error) {
	return p.renderWithBody(t, route, nil)
}

func (p *Project) renderWithBody(t *template.Template, route string, admin map[string]any) ([]byte, error) {
	return p.renderPage(t, route, admin, false)
}

func (p *Project) renderPage(t *template.Template, route string, admin map[string]any, owner bool) ([]byte, error) {
	return p.renderRequest(t, route, admin, owner, nil)
}

func (p *Project) renderRequest(t *template.Template, route string, admin map[string]any, owner bool, inputs *renderInputs) ([]byte, error) {
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
	blog := p.blogData(route)
	if current, ok := blog["post"].(map[string]any); ok {
		title = current["title"].(string)
		description = current["excerpt"].(string)
	}
	if page := p.Content.Pages[route]; page != nil {
		if page.Title != "" {
			title = page.Title
		}
		if page.Description != "" {
			description = page.Description
		}
	}
	app := make(map[string]any, len(p.App)+1)
	for key, value := range p.App {
		app[key] = value
	}
	app["current_admin"] = nil
	if owner && admin != nil {
		copy := map[string]any{}
		for k, v := range admin {
			copy[k] = v
		}
		if _, enabled := p.Apps["blog"]; enabled {
			copy["blog_enabled"] = true
		}
		app["current_admin"] = copy
	}
	data := map[string]any{"app": app, "page": map[string]any{"title": title, "description": description, "path": route}, "snip": snips, "data": snips}
	photos := map[string]any{}
	for id, photo := range p.Content.Media {
		photos[id] = map[string]any{"src": "/media/" + photo.File, "alt": photo.Alt, "width": photo.Width, "height": photo.Height, "name": photo.Name}
	}
	data["media"] = photos
	data["blog"] = map[string]any{"data": blog}
	videos := map[string]any{}
	for id, v := range p.Content.Videos {
		poster := ""
		if photo, ok := p.Content.Media[v.Poster]; ok {
			poster = "/media/" + photo.File
		}
		videos[id] = map[string]any{"src": "/media/" + v.File, "poster": poster, "width": v.Width, "height": v.Height, "name": v.Name}
	}
	data["videos"] = videos
	data["admin"] = admin
	data["__ownerPreview"] = owner
	data["__request"] = inputs
	if inputs != nil && inputs.appName != "" {
		if _, exists := data[inputs.appName]; exists {
			return nil, fmt.Errorf("app name conflicts with a template namespace")
		}
		data[inputs.appName] = inputs.appData
		var out appRenderBuffer
		err := t.Execute(&out, data)
		return out.Bytes(), err
	}
	if cfg, ok := p.Apps["blog"]; ok && (route == cfg.Mount || strings.HasPrefix(route, cfg.Mount+"/")) {
		var out appRenderBuffer
		err := t.Execute(&out, data)
		return out.Bytes(), err
	}
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
