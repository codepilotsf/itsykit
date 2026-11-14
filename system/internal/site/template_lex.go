package site

import (
	"fmt"
	"html"
	"strings"

	xhtml "golang.org/x/net/html"
)

// braceLexer pairs literal groups once, without parsing JSON or decoding source.
// HTML markup ends a group. Unpaired openings consume only their own byte.
// Quotes (including HTML-encoded quotes) and backslash escapes hide nested braces.
type braceLexer struct {
	source  string
	pairs   map[int]int
	aliases map[string]string
}

func newBraceLexer(source string) braceLexer {
	l := braceLexer{source: source, pairs: map[int]int{}}
	var stack []int
	var quote byte
	for i := 0; i < len(source); i++ {
		c, size := lexicalChar(source, i)
		if source[i] == '<' && quote == 0 {
			stack = stack[:0]
			quote = 0
			continue
		}
		if len(stack) == 0 {
			if c == '{' && size == 1 {
				stack = append(stack, i)
			}
			continue
		}
		if quote != 0 {
			if c == '\\' {
				_, n := lexicalChar(source, i+size)
				i += size + n - 1
				continue
			}
			if c == quote {
				quote = 0
			}
		} else {
			switch c {
			case '\'', '"':
				quote = c
			case '{':
				if size == 1 {
					stack = append(stack, i)
				}
			case '}':
				if size == 1 {
					start := stack[len(stack)-1]
					stack = stack[:len(stack)-1]
					l.pairs[start] = i + 1
				}
			}
		}
		i += size - 1
	}
	return l
}

func lexicalChar(s string, i int) (byte, int) {
	if i >= len(s) {
		return 0, 0
	}
	if s[i] == '&' {
		if end := strings.IndexByte(s[i:min(len(s), i+12)], ';'); end >= 0 {
			entity := s[i : i+end+1]
			decoded := html.UnescapeString(entity)
			if len(decoded) == 1 {
				return decoded[0], len(entity)
			}
		}
	}
	return s[i], 1
}

func commandStart(s string) bool {
	end := 0
	for end < len(s) && !strings.ContainsRune(" \t\r\n\f{}<", rune(s[end])) {
		end++
	}
	switch s[:end] {
	case "flea", "snip", "icon", "render", "slot", "csrf", "#if", "#unless", "#each", "#multisnip", ":else", "/if", "/unless", "/each", "/multisnip":
		return true
	}
	return false
}

// group returns an exclusive end, whether this is ItsyWeb syntax, and a lexical
// error only for recognizable syntax. Literal groups are never recursively read.
func (l braceLexer) group(start int) (int, bool, error) {
	s := l.source
	command := commandStart(strings.TrimLeft(s[start+1:], " \t\r\n\f"))
	end := start + 1
	var quote byte
	for ; end < len(s); end++ {
		c := s[end]
		if quote != 0 {
			if c == '\\' {
				end++
				continue
			}
			if c == quote {
				quote = 0
			}
			continue
		}
		if c == '\'' || c == '"' {
			quote = c
			continue
		}
		if c == '{' || c == '}' || c == '<' {
			break
		}
	}
	token := strings.TrimSpace(s[start+1 : min(end, len(s))])
	// Whitespace-padded bare words are prose, e.g. "Use { and }".
	// Dotted paths remain recognizable when padded or unclosed.
	rawToken := s[start+1 : min(end, len(s))]
	expression := pathPattern.MatchString(token) && (strings.Contains(token, ".") || l.aliases[token] != "" || (end < len(s) && s[end] == '}' && rawToken == token))
	syntax := command || expression
	if syntax {
		if end >= len(s) || s[end] != '}' {
			return start + 1, true, fmt.Errorf("unclosed template directive; expected } before the next directive or HTML element")
		}
		return end + 1, true, nil
	}
	if end, ok := l.pairs[start]; ok {
		return end, false, nil
	}
	return start + 1, false, nil
}

// directiveView keeps byte offsets and HTML structure while hiding literal
// braces from reference/field scanners. Entities are never executable syntax.
// This view is for validation only, never rendering or extraction writes.
func directiveView(source string) string {
	l := newBraceLexer(source)
	out := []byte(source)
	for i := 0; i < len(source); {
		if source[i] == '{' {
			end, syntax, _ := l.group(i)
			if !syntax {
				for j := i; j < end; {
					c, n := lexicalChar(source, j)
					if c == '{' || c == '}' {
						out[j] = ' '
					}
					j += n
				}
			}
			i = end
			continue
		}
		c, n := lexicalChar(source, i)
		if n > 1 && (c == '{' || c == '}') {
			out[i] = ' '
		}
		i += n
	}
	return string(out)
}

func directiveAttributes(raw string) map[string]string {
	z := xhtml.NewTokenizer(strings.NewReader(directiveView(raw)))
	z.Next()
	attrs := map[string]string{}
	for _, a := range z.Token().Attr {
		attrs[a.Key] = a.Val
	}
	return attrs
}
