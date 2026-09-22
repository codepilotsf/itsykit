package snips

import (
	"bytes"
	"fmt"
	"html"
	"net/url"
	"strings"

	xhtml "golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// Sanitize reconstructs a small formatting vocabulary. No source tag or
// attribute is trusted, and content is never reparsed as Flea CMS template code.
func Sanitize(source string) (string, error) {
	if len(source) > 1<<20 {
		return "", fmt.Errorf("snip content exceeds 1 MiB")
	}
	nodes, err := xhtml.ParseFragment(strings.NewReader(source), &xhtml.Node{Type: xhtml.ElementNode, Data: "div", DataAtom: atom.Div})
	if err != nil {
		return "", err
	}
	var out bytes.Buffer
	var walk func(*xhtml.Node)
	walk = func(n *xhtml.Node) {
		if n.Type == xhtml.TextNode {
			out.WriteString(html.EscapeString(n.Data))
			return
		}
		if n.Type != xhtml.ElementNode {
			return
		}
		switch n.Data {
		case "script", "style", "iframe", "object", "embed", "svg", "math", "template", "form", "input", "textarea", "select", "button":
			return
		}
		allowed := false
		switch n.Data {
		case "p", "br", "hr", "strong", "b", "em", "i", "u", "s", "del", "mark", "small", "sub", "sup", "h2", "h3", "h4", "h5", "h6", "ul", "ol", "li", "blockquote", "pre", "code", "a":
			allowed = true
		}
		if allowed {
			out.WriteString("<" + n.Data)
			if n.Data == "a" {
				for _, a := range n.Attr {
					if a.Namespace != "" {
						continue
					}
					if a.Key == "title" {
						out.WriteString(` title="` + html.EscapeString(a.Val) + `"`)
					}
					if a.Key == "href" && safeLink(a.Val) {
						out.WriteString(` href="` + html.EscapeString(a.Val) + `"`)
					}
				}
			}
			out.WriteByte('>')
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
		if allowed && n.Data != "br" && n.Data != "hr" {
			out.WriteString("</" + n.Data + ">")
		}
	}
	for _, n := range nodes {
		walk(n)
	}
	return out.String(), nil
}
func safeLink(value string) bool {
	if strings.ContainsAny(value, "\\\x00\r\n\t") || strings.HasPrefix(strings.TrimSpace(value), "//") {
		return false
	}
	u, err := url.Parse(strings.TrimSpace(value))
	if err != nil {
		return false
	}
	switch strings.ToLower(u.Scheme) {
	case "", "http", "https", "mailto", "tel":
		return true
	}
	return false
}
