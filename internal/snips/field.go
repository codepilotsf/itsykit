package snips

import (
	"fmt"
	"golang.org/x/net/html"
	"strings"
)

func CleanField(value string, rich bool) (string, error) {
	if len(value) > 64<<10 {
		return "", fmt.Errorf("multisnip field exceeds 64 KiB")
	}
	clean, err := Sanitize(value)
	if err != nil {
		return "", err
	}
	if !rich {
		z := html.NewTokenizer(strings.NewReader(clean))
		for kind := z.Next(); kind != html.ErrorToken; kind = z.Next() {
			if kind == html.TextToken {
				continue
			}
			if kind == html.StartTagToken || kind == html.EndTagToken || kind == html.SelfClosingTagToken {
				if strings.Contains(" br strong b em i u s del mark small sub sup code a ", " "+z.Token().Data+" ") {
					continue
				}
			}
			return "", fmt.Errorf("this multisnip field accepts inline formatting only")
		}
	}
	return clean, nil
}
