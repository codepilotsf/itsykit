package site

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"regexp"
	"strings"

	"flea/internal/media"
	"flea/internal/snips"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// EmptyPhotoSource is a fixed, non-executable portrait placeholder shared by
// server rendering and the browser's unsaved collection drafts.
var EmptyPhotoSource = "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 160 160"><defs><clipPath id="portrait"><circle cx="80" cy="80" r="80"/></clipPath></defs><circle cx="80" cy="80" r="80" fill="#e2e8f0"/><circle cx="80" cy="59" r="27" fill="#94a3b8"/><path clip-path="url(#portrait)" d="M24 170v-19c0-37 23-61 56-61s56 24 56 61v19Z" fill="#94a3b8"/></svg>`))

type MultiDefinition struct {
	Fields map[string]bool
	Photos map[string]bool
}

var multiPhotoField = regexp.MustCompile(`^\{snip\.([a-z][a-z0-9_]*)\.(src|alt|width|height|id)\}$`)

func (p *Project) VisibleMultisnips(route string) (map[string]MultiDefinition, error) {
	defs, err := p.MultisnipDefinitions(route)
	if err != nil {
		return nil, err
	}
	page := p.Pages[route]
	if page == nil {
		return nil, fmt.Errorf("unknown page")
	}
	body, err := p.renderPage(page, route, nil, true)
	if err != nil {
		return nil, err
	}
	for name := range defs {
		if !strings.Contains(string(body), "<!--multisnip-start:"+name+"-->") {
			delete(defs, name)
		}
	}
	return defs, nil
}

type multiBlock struct {
	name                           string
	start, bodyStart, bodyEnd, end int
}

var multiToken = regexp.MustCompile(`\{(?:#multisnip\s+([a-z][a-z0-9_]{0,39})|/multisnip)\}`)
var multiField = regexp.MustCompile(`^\{snip\.([a-z][a-z0-9_]{0,39})\}$`)

func multisnipBlocks(file, source string) ([]multiBlock, error) {
	z := html.NewTokenizer(strings.NewReader(directiveView(source)))
	offset := 0
	rawTag := ""
	var ancestors []string
	openingDepth := 0
	var blocks []multiBlock
	var active *multiBlock
	seen := map[string]bool{}
	for {
		kind := z.Next()
		raw := string(z.Raw())
		start := offset
		offset += len(raw)
		if kind == html.ErrorToken {
			if z.Err() != io.EOF {
				return nil, z.Err()
			}
			break
		}
		token := z.Token()
		if kind == html.StartTagToken && !strings.Contains(" area base br col embed hr img input link meta param source track wbr ", " "+token.Data+" ") {
			ancestors = append(ancestors, token.Data)
		}
		if kind == html.EndTagToken {
			for i := len(ancestors) - 1; i >= 0; i-- {
				if ancestors[i] == token.Data {
					ancestors = ancestors[:i]
					break
				}
			}
		}
		if kind == html.StartTagToken && (token.Data == "script" || token.Data == "style") {
			rawTag = token.Data
		}
		if kind == html.EndTagToken && token.Data == rawTag {
			rawTag = ""
		}
		if kind != html.TextToken || rawTag != "" {
			continue
		}
		for _, m := range multiToken.FindAllStringSubmatchIndex(raw, -1) {
			if m[2] >= 0 {
				for _, tag := range ancestors {
					if strings.Contains(" nav button a form textarea title select option head script style table tr tbody thead tfoot template noscript ", " "+tag+" ") {
						return nil, fmt.Errorf("multisnips are not supported inside <%s>", tag)
					}
				}
				openingDepth = len(ancestors)
				name := raw[m[2]:m[3]]
				if active != nil || seen[name] || snipScope(file) == "" || strings.HasPrefix(file, "admin/") {
					return nil, fmt.Errorf("%s: multisnips must be unique, nonnested and page-local", file)
				}
				seen[name] = true
				active = &multiBlock{name: name, start: start + m[0], bodyStart: start + m[1]}
			} else {
				if active == nil {
					return nil, fmt.Errorf("%s: unmatched /multisnip", file)
				}
				if len(ancestors) != openingDepth {
					return nil, fmt.Errorf("multisnip must contain complete balanced elements")
				}
				active.bodyEnd = start + m[0]
				active.end = start + m[1]
				blocks = append(blocks, *active)
				active = nil
			}
		}
	}
	if active != nil {
		return nil, fmt.Errorf("%s: unclosed multisnip", file)
	}
	return blocks, nil
}

func definition(body string) (MultiDefinition, error) {
	def := MultiDefinition{Fields: map[string]bool{}, Photos: map[string]bool{}}
	nodes, err := html.ParseFragment(strings.NewReader(directiveView(body)), &html.Node{Type: html.ElementNode, Data: "div", DataAtom: atom.Div})
	if err != nil {
		return def, err
	}
	roots := 0
	var walk func(*html.Node) error
	walk = func(n *html.Node) error {
		photoField := ""
		if n.Type == html.ElementNode && n.Data == "img" {
			for _, a := range n.Attr {
				if a.Key == "src" {
					if match := multiPhotoField.FindStringSubmatch(a.Val); match != nil && match[2] == "src" {
						photoField = match[1]
					}
				}
			}
			if photoField != "" {
				for ancestor := n.Parent; ancestor != nil; ancestor = ancestor.Parent {
					if strings.Contains(" nav button a form head script style template noscript ", " "+ancestor.Data+" ") {
						return fmt.Errorf("photo field cannot appear inside <%s>", ancestor.Data)
					}
				}
				if _, exists := def.Fields[photoField]; exists {
					return fmt.Errorf("duplicate multisnip field")
				}
				def.Fields[photoField] = false
				def.Photos[photoField] = true
				attrs := map[string]string{}
				for _, a := range n.Attr {
					if _, ok := attrs[a.Key]; ok {
						return fmt.Errorf("duplicate photo attribute")
					}
					attrs[a.Key] = a.Val
				}
				if attrs["data-snip-photo"] != photoField || attrs["alt"] != "{snip."+photoField+".alt}" || attrs["data-photo-id"] != "{snip."+photoField+".id}" {
					return fmt.Errorf("photo field requires matching alt, data-snip-photo and data-photo-id attributes")
				}
			}
		}
		for _, a := range n.Attr {
			if strings.ContainsAny(a.Val, "{}") {
				if match := multiPhotoField.FindStringSubmatch(a.Val); match != nil && match[1] == photoField {
					property := a.Key
					if property == "data-photo-id" {
						property = "id"
					}
					if property == match[2] {
						continue
					}
				}
				return fmt.Errorf("multisnip fields cannot appear in attributes")
			}
		}
		if n.Type == html.TextNode && strings.ContainsAny(n.Data, "{}") {
			for ancestor := n.Parent; ancestor != nil; ancestor = ancestor.Parent {
				if strings.Contains(" nav button a form textarea title select option head script style template noscript ", " "+ancestor.Data+" ") {
					return fmt.Errorf("multisnip fields cannot edit <%s>", ancestor.Data)
				}
			}
			match := multiField.FindStringSubmatch(strings.TrimSpace(n.Data))
			if match == nil || n.Parent == nil || n.PrevSibling != nil || n.NextSibling != nil {
				return fmt.Errorf("multisnip fields must fill a complete text element")
			}
			tag := n.Parent.Data
			if !strings.Contains(" h1 h2 h3 h4 h5 h6 p span div section article blockquote li ", " "+tag+" ") {
				return fmt.Errorf("unsupported multisnip field wrapper %s", tag)
			}
			if _, exists := def.Fields[match[1]]; exists {
				return fmt.Errorf("duplicate multisnip field %s", match[1])
			}
			def.Fields[match[1]] = strings.Contains(" div section article ", " "+tag+" ")
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err := walk(c); err != nil {
				return err
			}
		}
		return nil
	}
	for _, n := range nodes {
		if n.Type == html.TextNode && strings.TrimSpace(n.Data) == "" {
			continue
		}
		if n.Type != html.ElementNode {
			return def, fmt.Errorf("multisnip requires one root element")
		}
		roots++
		if err := walk(n); err != nil {
			return def, err
		}
	}
	if roots != 1 || len(def.Fields) == 0 || len(def.Fields) > 30 || len(def.Fields) == len(def.Photos) {
		return def, fmt.Errorf("multisnip needs one root element and 1–30 named fields")
	}
	return def, nil
}

func (p *Project) MultisnipDefinitions(route string) (map[string]MultiDefinition, error) {
	out := map[string]MultiDefinition{}
	for file, source := range p.files {
		if !strings.HasSuffix(file, ".html") || snipScope(file) != route {
			continue
		}
		blocks, err := multisnipBlocks(file, source)
		if err != nil {
			return nil, err
		}
		for _, b := range blocks {
			def, err := definition(source[b.bodyStart:b.bodyEnd])
			if err != nil {
				return nil, fmt.Errorf("%s: %w", file, err)
			}
			out[b.name] = def
		}
	}
	return out, nil
}

type multiRow struct {
	Name, ID    string
	Fields      map[string]string
	Rich        map[string]bool
	Prototype   bool
	PhotoFields map[string]bool
	Photos      map[string]media.Photo
}

func (p *Project) multiRows(data map[string]any, name string) ([]multiRow, error) {
	route, _ := lookup(data, "page.path").(string)
	defs, err := p.MultisnipDefinitions(route)
	if err != nil {
		return nil, err
	}
	def, ok := defs[name]
	if !ok {
		return nil, fmt.Errorf("unknown multisnip %s", name)
	}
	rows := []multiRow{}
	if data["__ownerPreview"] == true {
		rows = append(rows, multiRow{Name: name, ID: "prototype", Fields: map[string]string{}, Rich: def.Fields, Prototype: true, PhotoFields: def.Photos, Photos: p.Content.Media})
	}
	if page := p.Content.Pages[route]; page != nil {
		for _, item := range page.Multisnips[name] {
			rows = append(rows, multiRow{Name: name, ID: item.ID, Fields: item.Fields, Rich: def.Fields, PhotoFields: def.Photos, Photos: p.Content.Media})
		}
	}
	return rows, nil
}
func multiValue(row multiRow, key string, owner bool) (template.HTML, error) {
	if row.PhotoFields[key] {
		return "", fmt.Errorf("photo fields require .src or .alt")
	}
	rich, ok := row.Rich[key]
	if !ok {
		return "", fmt.Errorf("unknown multisnip field")
	}
	clean, err := snips.CleanField(row.Fields[key], rich)
	if err != nil {
		return "", err
	}
	if owner {
		id := row.Name + "/" + row.ID + "/" + key
		return template.HTML("<!--snip-start:" + id + "-->" + clean + "<!--snip-end:" + id + "-->"), nil
	}
	return template.HTML(clean), nil
}
func multiPhoto(row multiRow, key, property string) (any, error) {
	if !row.PhotoFields[key] {
		return "", fmt.Errorf("unknown photo field")
	}
	id := row.Fields[key]
	photo, ok := row.Photos[id]
	if !ok && id != "" {
		return "", fmt.Errorf("unknown photo in multisnip")
	}
	switch property {
	case "id":
		return id, nil
	case "src":
		if id == "" {
			return template.URL(EmptyPhotoSource), nil
		}
		return "/media/" + photo.File, nil
	case "alt":
		return photo.Alt, nil
	case "width":
		return max(1, photo.Width), nil
	case "height":
		return max(1, photo.Height), nil
	}
	return "", fmt.Errorf("unknown photo property")
}
func multiBoundary(name string, end, owner bool) template.HTML {
	if !owner {
		return ""
	}
	edge := "start"
	if end {
		edge = "end"
	}
	return template.HTML("<!--multisnip-" + edge + ":" + name + "-->")
}
func multiItemBoundary(row multiRow, end bool) template.HTML {
	if !row.Prototype {
		return ""
	}
	if end {
		return template.HTML("</template>")
	}
	return template.HTML(`<template data-multisnip="` + row.Name + `">`)
}
