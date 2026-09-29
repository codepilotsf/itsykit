package site

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"flea/internal/snips"
	"golang.org/x/net/html"
)

var richRef = regexp.MustCompile(`\{\s*snip\s+([A-Za-z0-9]{10})\s*\}`)

func richMatches(source string) [][]string {
	return richRef.FindAllStringSubmatch(source, -1)
}

type region struct {
	start, end                   int
	attributeStart, attributeEnd int
	label, content, id           string
}
type frame struct {
	tag    string
	region *region
}

// scanRegions uses source offsets rather than serializing HTML, so extraction
// changes only the editable region and preserves the developer's formatting.
func scanRegions(file, source string, legacy map[string]bool) (regions []region, refs []string, err error) {
	bindings, err := scanMediaBindings(file, source)
	if err != nil {
		return nil, nil, err
	}
	mediaStarts := map[int]bool{}
	for _, b := range bindings {
		mediaStarts[b.start] = true
		if b.id != "" {
			refs = append(refs, b.id)
		}
	}
	// Read-only loading also sees authoring seeds before dev extracts them.
	// Image fields belong to their multisnip, not to rich-text regions.
	seeds, err := authorMultisnips(file, source)
	if err != nil {
		return nil, nil, err
	}
	z := html.NewTokenizer(strings.NewReader(source))
	offset := 0
	position := 0
	defer func() {
		if err != nil {
			err = sourceError(file, source, position, err)
		}
	}()
	var stack []frame
	var active *region
	for {
		typ := z.Next()
		raw := append([]byte(nil), z.Raw()...)
		start := offset
		position = start
		offset += len(raw)
		if typ == html.ErrorToken {
			if z.Err() != io.EOF {
				return nil, nil, z.Err()
			}
			break
		}
		switch typ {
		case html.StartTagToken, html.SelfClosingTagToken:
			token := z.Token()
			label := ""
			has := false
			for _, a := range token.Attr {
				if a.Key == "flea" {
					return nil, nil, fmt.Errorf("the flea attribute is unsupported; use snip")
				}
				for _, match := range richMatches(a.Val) {
					if !legacy[match[1]] {
						return nil, nil, fmt.Errorf("rich snips are not allowed in attributes")
					}
				}
				if a.Key == "snip" {
					if has {
						return nil, nil, fmt.Errorf("duplicate snip attribute")
					}
					has = true
					label = a.Val
				}
			}
			void := strings.Contains(" area base br col embed hr img input link meta param source track wbr ", " "+token.Data+" ") || typ == html.SelfClosingTagToken
			if has && token.Data == "img" && snips.ValidFieldName(label) {
				for _, seed := range seeds {
					if start >= seed.bodyStart && start < seed.bodyEnd {
						has = false
						break
					}
				}
			}
			if has {
				if mediaStarts[start] {
					has = false
				}
			}
			if has {
				if active != nil {
					return nil, nil, fmt.Errorf("nested snip regions are not allowed")
				}
				if void || strings.Contains(" script style textarea title select option head html table tr tbody thead tfoot ", " "+token.Data+" ") {
					return nil, nil, fmt.Errorf("snip regions are not supported on <%s>", token.Data)
				}
				from, to := snipAttributeSpan(string(raw))
				if from < 0 {
					return nil, nil, fmt.Errorf("could not locate snip attribute")
				}
				active = &region{start: offset, label: label, attributeStart: start + from, attributeEnd: start + to}
			}
			if !void {
				f := frame{tag: token.Data}
				if has {
					f.region = active
				}
				stack = append(stack, f)
			}
		case html.EndTagToken:
			token := z.Token()
			at := -1
			for i := len(stack) - 1; i >= 0; i-- {
				if stack[i].tag == token.Data {
					at = i
					break
				}
			}
			if at < 0 {
				if active != nil {
					return nil, nil, fmt.Errorf("unmatched closing tag inside snip")
				}
				continue
			}
			if active != nil && at != len(stack)-1 {
				return nil, nil, fmt.Errorf("snip HTML must have explicit, balanced closing tags")
			}
			if stack[at].region != nil {
				active.end = start
				active.content = source[active.start:active.end]
				matches := richMatches(strings.TrimSpace(active.content))
				if len(matches) == 1 && !legacy[matches[0][1]] && strings.TrimSpace(active.content) == matches[0][0] {
					active.id = matches[0][1]
				} else if len(matches) > 0 || strings.ContainsAny(active.content, "{}") {
					return nil, nil, fmt.Errorf("snip content cannot contain template directives; encode literal braces as HTML entities")
				}
				regions = append(regions, *active)
				active = nil
			}
			stack = stack[:at]
		case html.TextToken:
			rawElement := false
			for _, f := range stack {
				if f.tag == "script" || f.tag == "style" {
					rawElement = true
				}
			}
			if rawElement {
				continue
			}
			matches := richMatches(string(raw))
			filtered := matches[:0]
			for _, match := range matches {
				if !legacy[match[1]] {
					filtered = append(filtered, match)
				}
			}
			matches = filtered
			if len(matches) > 0 {
				if len(stack) == 0 {
					return nil, nil, fmt.Errorf("rich snips must be inside an HTML element")
				}
				for _, f := range stack {
					if strings.Contains(" script style textarea title select option head ", " "+f.tag+" ") {
						return nil, nil, fmt.Errorf("rich snips are not allowed in <%s>", f.tag)
					}
				}
				for _, m := range matches {
					refs = append(refs, m[1])
				}
			}
		}
	}
	if active != nil {
		position = active.attributeStart
		return nil, nil, fmt.Errorf("unclosed snip region")
	}
	return regions, refs, nil
}

// Locate the authoring attribute without reserializing the tag. Quoted values
// may contain spaces, >, or the word snip; they are never treated as attributes.
func snipAttributeSpan(tag string) (int, int) {
	return attributeSpan(tag, "snip")
}
func attributeSpan(tag, wanted string) (int, int) {
	space := func(c byte) bool { return c == ' ' || c == '\t' || c == '\r' || c == '\n' || c == '\f' }
	i := 1
	for i < len(tag) && !space(tag[i]) && tag[i] != '>' && tag[i] != '/' {
		i++
	}
	for i < len(tag) {
		before := i
		for i < len(tag) && space(tag[i]) {
			i++
		}
		if i == len(tag) || tag[i] == '>' || tag[i] == '/' {
			break
		}
		start := i
		for i < len(tag) && !space(tag[i]) && tag[i] != '=' && tag[i] != '>' && tag[i] != '/' {
			i++
		}
		if start == i {
			return -1, -1
		}
		name := tag[start:i]
		end := i
		for i < len(tag) && space(tag[i]) {
			i++
		}
		if i < len(tag) && tag[i] == '=' {
			i++
			for i < len(tag) && space(tag[i]) {
				i++
			}
			if i < len(tag) && (tag[i] == '\'' || tag[i] == '"') {
				quote := tag[i]
				i++
				for i < len(tag) && tag[i] != quote {
					i++
				}
				if i < len(tag) {
					i++
				}
			} else {
				for i < len(tag) && !space(tag[i]) && tag[i] != '>' {
					i++
				}
			}
			end = i
		}
		if strings.EqualFold(name, wanted) {
			return before, end
		}
		// Leave whitespace for the next attribute's removable separator.
		i = end
	}
	return -1, -1
}
func snipScope(file string) string {
	if strings.HasSuffix(file, ".layout.html") {
		return ""
	}
	for _, part := range strings.Split(file, "/") {
		if strings.HasPrefix(part, "_") {
			return ""
		}
	}
	return pageRoute(file)
}
func rejectOldContentPaths(root *os.Root) error {
	for _, name := range []string{"flea/fleas.json", "flea/fleas", "flea/.fleas-pending.json"} {
		if _, err := root.Lstat(name); err == nil {
			return fmt.Errorf("unsupported old content path %s; update the site to the snip format before continuing", name)
		} else if !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}
func readSnipStore(root *os.Root) (*snips.Store, error) {
	if err := rejectOldContentPaths(root); err != nil {
		return nil, err
	}
	data, err := readFile(root, "flea/snips.json")
	if os.IsNotExist(err) {
		return snips.NewStore(), nil
	}
	if err != nil {
		return nil, err
	}
	return snips.Decode(data)
}
func (p *Project) SnipReferences() (map[string]string, error) {
	refs := map[string]string{}
	legacy := p.legacySnipIDs()
	for file, source := range p.files {
		_, ids, err := scanRegions(file, source, legacy)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", file, err)
		}
		for _, id := range ids {
			scope := snipScope(file)
			if old, ok := refs[id]; ok && old != scope {
				scope = ""
			}
			refs[id] = scope
		}
	}
	return refs, nil
}

func (p *Project) CheckExtractedSnips() error {
	legacy := p.legacySnipIDs()
	for file, source := range p.files {
		if blocks, err := authorMultisnips(file, source); err != nil {
			return err
		} else if len(blocks) > 0 {
			return fmt.Errorf("%s: run flea dev to extract multisnips before deploying", file)
		}
		regions, _, err := scanRegions(file, source, legacy)
		if err != nil {
			return err
		}
		if len(regions) > 0 {
			return fmt.Errorf("%s: snip attribute needs extraction. Run snip dev, then commit the generated files before deploying.", file)
		}
		bindings, e := scanMediaBindings(file, source)
		if e != nil {
			return e
		}
		for _, b := range bindings {
			if b.author {
				return fmt.Errorf("%s: run flea dev to extract media snips before deploying", file)
			}
		}
	}
	return nil
}

func (p *Project) legacySnipIDs() map[string]bool {
	legacy := map[string]bool{}
	for name := range p.Snips {
		if _, rich := p.rich[name]; !rich {
			legacy[name] = true
		}
	}
	return legacy
}
