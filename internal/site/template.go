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

// compile translates the deliberately small FleaCMS language to html/template.
// Go's contextual escaping also protects URL and attribute contexts.
func (p *Project) compile(source, file string, stack []string, aliases map[string]string) (string, error) {
	mediaBindings, mediaErr := scanMediaBindings(file, source)
	if mediaErr != nil {
		return "", mediaErr
	}
	mediaInsert := map[int]string{}
	mediaExpressions := map[string]bool{}
	index := p.contentIndex
	for _, b := range mediaBindings {
		if b.id == "" {
			continue
		}
		entry, ok := index[b.id]
		expected := "image"
		if b.tag == "video" {
			expected = "video"
		}
		if !ok || entry.Type != expected {
			return "", sourceError(file, source, b.start, fmt.Errorf("missing or incorrect %s snip %s", expected, b.id))
		}
		at := b.start + strings.IndexAny(source[b.start:b.end], " \t\r\n/>")
		mediaInsert[at] = b.id
		for _, value := range b.attrs {
			if m := mediaRef.FindStringSubmatch(value); m != nil {
				mediaExpressions[value] = true
			}
		}
	}
	ms, err := multisnipBlocks(file, source)
	if err != nil {
		return "", err
	}
	for _, b := range ms {
		if _, err := definition(source[b.bodyStart:b.bodyEnd]); err != nil {
			return "", fmt.Errorf("%s: %w", file, err)
		}
	}
	var out strings.Builder
	type block struct {
		kind, alias, previous string
		elseSeen              bool
		start                 int
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
		if original, ok := p.files[file]; ok && strings.HasSuffix(original, source) {
			return "", sourceError(file, original, len(original)-len(source)+pos, err)
		}
		return "", sourceError(file, source, pos, err)
	}
	lexer := newBraceLexer(source)
	lexer.aliases = aliases
	for i := 0; i < len(source); {
		if id, ok := mediaInsert[i]; ok {
			out.WriteString(`[[NW if (lookup $ "__ownerPreview") NW]] data-snip-media="` + id + `"[[NW end NW]]`)
		}
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
				if strings.Contains(source[i:end], "{blog.data.post.body}") {
					return fail(i, fmt.Errorf("blog body requires ordinary HTML element content"))
				}
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
		next, syntax, err := lexer.group(i)
		if err != nil {
			return fail(i, err)
		}
		if !syntax {
			out.WriteString(source[i:next])
			i = next
			continue
		}
		end := next - 1
		token := strings.TrimSpace(source[i+1 : end])
		var rendered string
		switch {
		case token == "blog.data.post.body":
			rendered = blogBodyAction
		case token == "csrf":
			rendered = csrfAction
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
			if file != "main.layout.html" && !p.layouts[file] {
				return fail(i, fmt.Errorf("{slot} is only allowed in a layout"))
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
			if m := mediaRef.FindStringSubmatch("{" + token + "}"); m != nil {
				if !mediaExpressions["{"+token+"}"] {
					return fail(i, fmt.Errorf("media snip references require img/video attribute bindings"))
				}
				rendered = "[[NW snipField " + strconv.Quote(m[1]) + " " + strconv.Quote(m[2]) + " NW]]"
				break
			}
			if !namePattern.MatchString(name) {
				return fail(i, fmt.Errorf("use an unquoted snip ID, for example {snip a7Km2pQ9xB}"))
			}
			if _, ok := p.rich[name]; ok {
				rendered = "[[NW richFlea " + strconv.Quote(name) + " (lookup $ \"__ownerPreview\") NW]]"
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
		case strings.HasPrefix(token, "#if ") || strings.HasPrefix(token, "#unless "):
			kind, path, _ := strings.Cut(token[1:], " ")
			expr, err := lookup(strings.TrimSpace(path))
			if err != nil {
				return fail(i, err)
			}
			blocks = append(blocks, block{kind: kind, start: i})
			if kind == "unless" {
				expr = "not " + expr
			}
			rendered = "[[NW if " + expr + " NW]]"
		case strings.HasPrefix(token, "#multisnip "):
			for _, b := range blocks {
				if b.kind == "each" {
					return fail(i, fmt.Errorf("multisnips cannot be nested in a loop"))
				}
			}
			name := strings.TrimSpace(strings.TrimPrefix(token, "#multisnip "))
			if !namePattern.MatchString(name) || aliases["snip"] != "" {
				return fail(i, fmt.Errorf("invalid or nested multisnip"))
			}
			variable := fmt.Sprintf("$multi%d", len(blocks))
			blocks = append(blocks, block{kind: "multisnip", alias: "snip", previous: name, start: i})
			aliases["snip"] = variable
			rendered = "[[NW multiBoundary " + strconv.Quote(name) + " false (lookup $ \"__ownerPreview\") NW]][[NW range " + variable + " := multiRows $ " + strconv.Quote(name) + " NW]][[NW multiItemBoundary " + variable + " false NW]]"
		case token == "/multisnip":
			if len(blocks) == 0 || blocks[len(blocks)-1].kind != "multisnip" {
				return fail(i, fmt.Errorf("unmatched /multisnip"))
			}
			b := blocks[len(blocks)-1]
			blocks = blocks[:len(blocks)-1]
			rendered = "[[NW multiItemBoundary " + aliases["snip"] + " true NW]][[NW end NW]][[NW multiBoundary " + strconv.Quote(b.previous) + " true (lookup $ \"__ownerPreview\") NW]]"
			delete(aliases, "snip")
		case strings.HasPrefix(token, "#each "):
			parts := strings.Split(token[6:], " as ")
			if len(parts) != 2 || !aliasPattern.MatchString(parts[1]) {
				return fail(i, fmt.Errorf("expected {#each dotted.path as alias}"))
			}
			if parts[1] == "app" || parts[1] == "page" || parts[1] == "snip" || parts[1] == "data" {
				return fail(i, fmt.Errorf("loop alias %q is reserved", parts[1]))
			}
			expr, err := lookup(strings.TrimSpace(parts[0]))
			if err != nil {
				return fail(i, err)
			}
			// Include the partial depth so an included loop cannot shadow a caller's variable.
			variable := fmt.Sprintf("$item%d_%d", len(stack), len(blocks))
			blocks = append(blocks, block{kind: "each", alias: parts[1], previous: aliases[parts[1]], start: i})
			aliases[parts[1]] = variable
			rendered = "[[NW range " + variable + " := collection " + expr + " NW]]"
		case token == ":else" || strings.HasPrefix(token, ":else if "):
			if len(blocks) == 0 {
				return fail(i, fmt.Errorf("{%s} requires an open if, unless or each block", token))
			}
			b := &blocks[len(blocks)-1]
			if b.elseSeen {
				return fail(i, fmt.Errorf("{%s} cannot follow {:else}", token))
			}
			if token != ":else" {
				if b.kind != "if" && b.kind != "unless" {
					return fail(i, fmt.Errorf("{:else if} is only allowed in an if or unless block"))
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
		case token == "/if" || token == "/unless" || token == "/each":
			if len(blocks) == 0 || blocks[len(blocks)-1].kind != token[1:] {
				return fail(i, fmt.Errorf("unmatched {%s}", token))
			}
			b := blocks[len(blocks)-1]
			blocks = blocks[:len(blocks)-1]
			restoreAlias(b.alias, b.previous)
			rendered = "[[NW end NW]]"
		default:
			if strings.HasPrefix(token, "snip.") && aliases["snip"] != "" {
				if parts := strings.Split(strings.TrimPrefix(token, "snip."), "."); len(parts) == 2 {
					rendered = "[[NW multiPhoto " + aliases["snip"] + " " + strconv.Quote(parts[0]) + " " + strconv.Quote(parts[1]) + " NW]]"
					break
				}
				rendered = "[[NW multiValue " + aliases["snip"] + " " + strconv.Quote(strings.TrimPrefix(token, "snip.")) + " (lookup $ \"__ownerPreview\") NW]]"
				break
			}
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
		b := blocks[len(blocks)-1]
		return fail(b.start, fmt.Errorf("unclosed %s block; expected {/%s}", b.kind, b.kind))
	}
	compiled := out.String()
	if err := checkCSRFContexts(compiled); err != nil {
		return fail(0, err)
	}
	return compiled, nil
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
