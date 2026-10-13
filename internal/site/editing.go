package site

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"flea/internal/snips"
	"golang.org/x/net/html"
)

// EditingVersion binds drafts to both content and developer-owned templates.
func (p *Project) EditingVersion() string {
	data, _ := json.Marshal(struct {
		Content *snips.Store
		Files   map[string]string
		App     map[string]any
		Data    map[string]map[string]any
		Root    string
		CSS     []byte
	}{p.Content, p.files, p.App, p.Snips, p.Root, p.AssetData["/css/theme.css"]})
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// EditableRegions returns only complete rendered wrappers. Navigation, mixed
// template content and structural/media containers are not owner-editable.
// The bool is true for containers that can hold block-level rich text.
func (p *Project) EditableRegions(route string) (map[string]bool, error) {
	t, ok := p.Pages[route]
	if !ok {
		return nil, fmt.Errorf("unknown page")
	}
	body, err := p.renderPage(t, route, nil, true)
	if err != nil {
		return nil, err
	}
	root, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	result := map[string]bool{}
	var walk func(*html.Node, bool)
	walk = func(n *html.Node, blocked bool) {
		if n.Type == html.ElementNode && strings.Contains(" nav button a form script style head ", " "+n.Data+" ") {
			blocked = true
		}
		if !blocked && n.Type == html.ElementNode && n.Data != "html" && n.Data != "body" {
			first, last := n.FirstChild, n.LastChild
			for first != nil && first.Type == html.TextNode && strings.TrimSpace(first.Data) == "" {
				first = first.NextSibling
			}
			for last != nil && last.Type == html.TextNode && strings.TrimSpace(last.Data) == "" {
				last = last.PrevSibling
			}
			if first != nil && last != nil && first.Type == html.CommentNode && last.Type == html.CommentNode && strings.HasPrefix(first.Data, "snip-start:") {
				id := strings.TrimPrefix(first.Data, "snip-start:")
				unsafe := false
				var check func(*html.Node)
				check = func(c *html.Node) {
					if c.Type == html.ElementNode && strings.Contains(" img video form input button iframe table ", " "+c.Data+" ") {
						unsafe = true
					}
					for child := c.FirstChild; child != nil; child = child.NextSibling {
						check(child)
					}
				}
				check(n)
				if snips.ValidID(id) && last.Data == "snip-end:"+id && !unsafe {
					rich := strings.Contains(" div section article main aside header footer ", " "+n.Data+" ")
					if prior, exists := result[id]; exists {
						rich = rich && prior
					}
					result[id] = rich
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c, blocked)
		}
	}
	walk(root, false)
	return result, nil
}
