package site

import (
	"fmt"
	"golang.org/x/net/html"
	"strings"
)

// Check the source context, not a browser-repaired DOM. In particular, moving a
// rich snip into a heading must not silently produce invalid nested blocks.
func (p *Project) checkContentStructure() error {
	for name, source := range p.files {
		if !strings.HasSuffix(name, ".html") {
			continue
		}
		z := html.NewTokenizer(strings.NewReader(source))
		stack := []string{}
		offset := 0
		for {
			kind := z.Next()
			raw := z.Raw()
			position := offset
			offset += len(raw)
			if kind == html.ErrorToken {
				break
			}
			token := z.Token()
			switch kind {
			case html.StartTagToken:
				if !strings.Contains(" area base br col embed hr img input link meta param source track wbr ", " "+token.Data+" ") {
					stack = append(stack, token.Data)
				}
			case html.EndTagToken:
				for i := len(stack) - 1; i >= 0; i-- {
					if stack[i] == token.Data {
						stack = stack[:i]
						break
					}
				}
			case html.TextToken:
				constrained, rawText := false, false
				for _, tag := range stack {
					if strings.Contains(" p h1 h2 h3 h4 h5 h6 span a b strong em i u s small ", " "+tag+" ") {
						constrained = true
					}
					if tag == "script" || tag == "style" {
						rawText = true
					}
				}
				if !constrained || rawText {
					continue
				}
				for _, ref := range richMatches(string(raw)) {
					content, ok := p.rich[ref[1]]
					if !ok {
						continue
					}
					inner := html.NewTokenizer(strings.NewReader(string(content)))
					for k := inner.Next(); k != html.ErrorToken; k = inner.Next() {
						if k == html.StartTagToken || k == html.SelfClosingTagToken {
							t := inner.Token()
							if strings.Contains(" p div h1 h2 h3 h4 h5 h6 ul ol li blockquote ", " "+t.Data+" ") {
								return sourceError(name, source, position, fmt.Errorf("snip %s contains block content that cannot fit its template wrapper", ref[1]))
							}
						}
					}
				}
			}
		}
	}
	return nil
}
