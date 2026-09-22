package site

import (
	"errors"
	"fmt"
	"maps"
	"path"
	"regexp"
	"strconv"
	"strings"
)

var pathPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]*(\.[A-Za-z0-9_-]+)*$`)
var namePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]*$`)
var aliasPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*$`)
var renderPathPattern = regexp.MustCompile(`^/?[A-Za-z0-9_./-]+$`)

func partialPath(name, caller string) (string, error) {
	if !renderPathPattern.MatchString(name) || strings.HasSuffix(name, "/") {
		return "", fmt.Errorf("render requires an unquoted partial path without an extension")
	}
	base := path.Base(name)
	if !namePattern.MatchString(base) {
		return "", fmt.Errorf("render requires a partial name without the leading underscore or extension")
	}
	var resolved string
	if strings.HasPrefix(name, "/") {
		resolved = path.Clean(strings.TrimPrefix(name, "/"))
	} else {
		resolved = path.Join(path.Dir(caller), name)
	}
	if resolved == ".." || strings.HasPrefix(resolved, "../") || strings.HasPrefix(resolved, "/") {
		return "", fmt.Errorf("partial path %q escapes the project root", name)
	}
	return path.Join(path.Dir(resolved), "_"+path.Base(resolved)+".html"), nil
}

var rawElement = regexp.MustCompile(`(?is)^<(script|style)(?:\s+(?:[^>"']|"[^"]*"|'[^']*')*)?/?>`)
var rawClose = map[string]*regexp.Regexp{
	"script": regexp.MustCompile(`(?is)</script(?:\s[^>]*|/?)>`),
	"style":  regexp.MustCompile(`(?is)</style(?:\s[^>]*|/?)>`),
}

// compile translates the deliberately small Flea CMS language to html/template.
// Go's contextual escaping also protects URL and attribute contexts.
func (p *Project) compile(source, file string, stack []string, aliases map[string]string) (string, error) {
	var out strings.Builder
	type block struct {
		kind, alias, previous string
		elseSeen              bool
	}
	var blocks []block
	restoreAlias := func(alias, previous string) {
		if alias == "" {
			return
		}
		if previous == "" {
			delete(aliases, alias)
		} else {
			aliases[alias] = previous
		}
	}
	lookup := func(path string) (string, error) {
		if !pathPattern.MatchString(path) {
			return "", fmt.Errorf("unsupported expression %q", path)
		}
		parts := strings.SplitN(path, ".", 2)
		if variable, ok := aliases[parts[0]]; ok {
			if len(parts) == 1 {
				return variable, nil
			}
			return "(lookup " + variable + " " + strconv.Quote(parts[1]) + ")", nil
		}
		return "(lookup $ " + strconv.Quote(path) + ")", nil
	}
	fail := func(pos int, err error) (string, error) {
		return "", fmt.Errorf("%s:%d: %w", file, strings.Count(source[:pos], "\n")+1, err)
	}
	for i := 0; i < len(source); {
		if strings.HasPrefix(source[i:], "<!--") {
			end := strings.Index(source[i+4:], "-->")
			if end < 0 {
				return fail(i, fmt.Errorf("unclosed HTML comment"))
			}
			end = i + 4 + end + 3
			out.WriteString(source[i:end])
			i = end
			continue
		}
		if source[i] == '<' {
			if match := rawElement.FindStringSubmatch(source[i:]); match != nil {
				close := rawClose[strings.ToLower(match[1])].FindStringIndex(source[i+len(match[0]):])
				if close == nil {
					return fail(i, fmt.Errorf("unclosed <%s> element", match[1]))
				}
				end := i + len(match[0]) + close[1]
				out.WriteString(source[i:end])
				i = end
				continue
			}
		}
		if source[i] != '{' {
			out.WriteByte(source[i])
			i++
			continue
		}
		end := i + 1
		var quote byte
		for ; end < len(source); end++ {
			if quote != 0 {
				if source[end] == quote {
					quote = 0
				}
				continue
			}
			if source[end] == '"' || source[end] == '\'' {
				quote = source[end]
				continue
			}
			if source[end] == '}' {
				break
			}
		}
		if end == len(source) {
			return fail(i, fmt.Errorf("unclosed template directive"))
		}
		token := strings.TrimSpace(source[i+1 : end])
		var rendered string
		switch {
		case token == "icon" || strings.HasPrefix(token, "icon ") || strings.HasPrefix(token, "icon\t") || strings.HasPrefix(token, "icon\n"):
			var err error
			rendered, err = p.compileIcon(token, lookup)
			if err != nil {
				var missing *missingIconError
				if errors.As(err, &missing) {
					if p.missingIcons == nil {
						p.missingIcons = map[string]string{}
					}
					if _, exists := p.missingIcons[missing.name]; !exists {
						_, located := fail(i, err)
						p.missingIcons[missing.name] = located.Error() + "; icon omitted"
					}
					rendered = ""
					break
				}
				return fail(i, err)
			}
		case token == "slot":
			if file != "_layout.html" {
				return fail(i, fmt.Errorf("{slot} is only allowed in _layout.html"))
			}
			rendered = `[[NW template "body" $ NW]]`
		case token == "render" || strings.HasPrefix(token, "render "):
			name, err := partialPath(strings.TrimSpace(strings.TrimPrefix(token, "render")), file)
			if err != nil {
				return fail(i, err)
			}
			for _, ancestor := range stack {
				if ancestor == name {
					return fail(i, fmt.Errorf("recursive partial inclusion: %s", strings.Join(append(stack, name), " -> ")))
				}
			}
			text, ok := p.Partials[name]
			if !ok {
				return fail(i, fmt.Errorf("missing partial %q", name))
			}
			rendered, err = p.compile(text, name, append(stack, name), maps.Clone(aliases))
			if err != nil {
				return "", err
			}
		case token == "snip" || strings.HasPrefix(token, "snip ") || strings.HasPrefix(token, "snip\t") || strings.HasPrefix(token, "snip\n"):
			name := strings.TrimSpace(token[4:])
			if !namePattern.MatchString(name) {
				return fail(i, fmt.Errorf("use an unquoted snip ID, for example {snip a7Km2pQ9xB}"))
			}
			if _, ok := p.rich[name]; ok {
				rendered = "[[NW richSnip " + strconv.Quote(name) + " NW]]"
				break
			}
			data, ok := p.Snips[name]
			if !ok {
				return fail(i, fmt.Errorf("missing snip %q", name))
			}
			if _, ok := data["text"].(string); !ok {
				return fail(i, fmt.Errorf("snip %q requires a string 'text' field for {snip}; use dotted paths for structured data", name))
			}
			rendered = "[[NW lookup $ " + strconv.Quote("snip."+name+".text") + " NW]]"
		case strings.HasPrefix(token, "#if "):
			expr, err := lookup(strings.TrimSpace(token[4:]))
			if err != nil {
				return fail(i, err)
			}
			blocks = append(blocks, block{kind: "if"})
			rendered = "[[NW if " + expr + " NW]]"
		case strings.HasPrefix(token, "#each "):
			parts := strings.Split(token[6:], " as ")
			if len(parts) != 2 || !aliasPattern.MatchString(parts[1]) {
				return fail(i, fmt.Errorf("expected {#each dotted.path as alias}"))
			}
			if parts[1] == "site" || parts[1] == "page" || parts[1] == "snip" || parts[1] == "data" {
				return fail(i, fmt.Errorf("loop alias %q is reserved", parts[1]))
			}
			expr, err := lookup(strings.TrimSpace(parts[0]))
			if err != nil {
				return fail(i, err)
			}
			// Include the partial depth so an included loop cannot shadow a caller's variable.
			variable := fmt.Sprintf("$item%d_%d", len(stack), len(blocks))
			blocks = append(blocks, block{kind: "each", alias: parts[1], previous: aliases[parts[1]]})
			aliases[parts[1]] = variable
			rendered = "[[NW range " + variable + " := collection " + expr + " NW]]"
		case token == ":else" || strings.HasPrefix(token, ":else if "):
			if len(blocks) == 0 {
				return fail(i, fmt.Errorf("{%s} requires an open if or each block", token))
			}
			b := &blocks[len(blocks)-1]
			if b.elseSeen {
				return fail(i, fmt.Errorf("{%s} cannot follow {:else}", token))
			}
			if token != ":else" {
				if b.kind != "if" {
					return fail(i, fmt.Errorf("{:else if} is only allowed in an if block"))
				}
				expr, err := lookup(strings.TrimSpace(strings.TrimPrefix(token, ":else if ")))
				if err != nil {
					return fail(i, err)
				}
				rendered = "[[NW else if " + expr + " NW]]"
			} else {
				b.elseSeen = true
				// An empty loop has no item. Restore any outer alias for its fallback.
				restoreAlias(b.alias, b.previous)
				rendered = "[[NW else NW]]"
			}
		case token == "/if" || token == "/each":
			if len(blocks) == 0 || blocks[len(blocks)-1].kind != token[1:] {
				return fail(i, fmt.Errorf("unmatched {%s}", token))
			}
			b := blocks[len(blocks)-1]
			blocks = blocks[:len(blocks)-1]
			restoreAlias(b.alias, b.previous)
			rendered = "[[NW end NW]]"
		default:
			expr, err := lookup(token)
			if err != nil {
				return fail(i, err)
			}
			rendered = "[[NW " + expr + " NW]]"
		}
		out.WriteString(rendered)
		i = end + 1
	}
	if len(blocks) > 0 {
		return fail(len(source), fmt.Errorf("unclosed %s block", blocks[len(blocks)-1].kind))
	}
	return out.String(), nil
}

func lookup(data any, path string) any {
	for _, key := range strings.Split(path, ".") {
		object, ok := data.(map[string]any)
		if !ok {
			return nil
		}
		data = object[key]
	}
	return data
}

func collection(data any) ([]any, error) {
	if data == nil {
		return nil, nil
	}
	list, ok := data.([]any)
	if !ok {
		return nil, fmt.Errorf("#each expects a JSON array")
	}
	return list, nil
}
