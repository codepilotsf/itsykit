package site

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"flea/internal/fleas"
	"golang.org/x/net/html"
)

var richRef = regexp.MustCompile(`\{\s*flea\s+([A-Za-z0-9]{10})\s*\}`)

func richMatches(source string) [][]string {
	return richRef.FindAllStringSubmatch(source, -1)
}

type region struct {
	start, end         int
	label, content, id string
}
type frame struct {
	tag    string
	region *region
}

// scanRegions uses source offsets rather than serializing HTML, so extraction
// changes only the editable region and preserves the developer's formatting.
func scanRegions(source string, legacy map[string]bool) ([]region, []string, error) {
	z := html.NewTokenizer(strings.NewReader(source))
	offset := 0
	var stack []frame
	var active *region
	var regions []region
	var refs []string
	for {
		typ := z.Next()
		raw := append([]byte(nil), z.Raw()...)
		start := offset
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
				if a.Key == "snip" {
					return nil, nil, fmt.Errorf("the snip attribute is unsupported; use flea")
				}
				for _, match := range richMatches(a.Val) {
					if !legacy[match[1]] {
						return nil, nil, fmt.Errorf("rich fleas are not allowed in attributes")
					}
				}
				if a.Key == "flea" {
					if has {
						return nil, nil, fmt.Errorf("duplicate flea attribute")
					}
					has = true
					label = a.Val
				}
			}
			void := strings.Contains(" area base br col embed hr img input link meta param source track wbr ", " "+token.Data+" ") || typ == html.SelfClosingTagToken
			if has {
				if active != nil {
					return nil, nil, fmt.Errorf("nested flea regions are not allowed")
				}
				if void || strings.Contains(" script style textarea title select option head html table tr tbody thead tfoot ", " "+token.Data+" ") {
					return nil, nil, fmt.Errorf("flea regions are not supported on <%s>", token.Data)
				}
				active = &region{start: offset, label: label}
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
					return nil, nil, fmt.Errorf("unmatched closing tag inside flea")
				}
				continue
			}
			if active != nil && at != len(stack)-1 {
				return nil, nil, fmt.Errorf("flea HTML must have explicit, balanced closing tags")
			}
			if stack[at].region != nil {
				active.end = start
				active.content = source[active.start:active.end]
				matches := richMatches(strings.TrimSpace(active.content))
				if len(matches) == 1 && !legacy[matches[0][1]] && strings.TrimSpace(active.content) == matches[0][0] {
					active.id = matches[0][1]
				} else if len(matches) > 0 || strings.ContainsAny(active.content, "{}") {
					return nil, nil, fmt.Errorf("flea content cannot contain template directives; encode literal braces as HTML entities")
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
					return nil, nil, fmt.Errorf("rich fleas must be inside an HTML element")
				}
				for _, f := range stack {
					if strings.Contains(" script style textarea title select option head ", " "+f.tag+" ") {
						return nil, nil, fmt.Errorf("rich fleas are not allowed in <%s>", f.tag)
					}
				}
				for _, m := range matches {
					refs = append(refs, m[1])
				}
			}
		}
	}
	if active != nil {
		return nil, nil, fmt.Errorf("unclosed flea region")
	}
	return regions, refs, nil
}
func fleaScope(file string) string {
	for _, part := range strings.Split(file, "/") {
		if strings.HasPrefix(part, "_") {
			return ""
		}
	}
	return pageRoute(file)
}
func rejectOldContentPaths(root *os.Root) error {
	for _, name := range []string{"_/snips.json", "_/snips", "_/.snips-pending.json"} {
		if _, err := root.Lstat(name); err == nil {
			return fmt.Errorf("unsupported old content path %s; update the site to the flea format before continuing", name)
		} else if !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}
func readFleaStore(root *os.Root) (*fleas.Store, error) {
	if err := rejectOldContentPaths(root); err != nil {
		return nil, err
	}
	data, err := readFile(root, "_/fleas.json")
	if os.IsNotExist(err) {
		return fleas.NewStore(), nil
	}
	if err != nil {
		return nil, err
	}
	return fleas.Decode(data)
}
func (p *Project) FleaReferences() (map[string]string, error) {
	refs := map[string]string{}
	legacy := p.legacyFleaIDs()
	for file, source := range p.files {
		_, ids, err := scanRegions(source, legacy)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", file, err)
		}
		for _, id := range ids {
			scope := fleaScope(file)
			if old, ok := refs[id]; ok && old != scope {
				scope = ""
			}
			refs[id] = scope
		}
	}
	return refs, nil
}

func (p *Project) CheckExtractedFleas() error {
	legacy := p.legacyFleaIDs()
	for file, source := range p.files {
		regions, _, err := scanRegions(source, legacy)
		if err != nil {
			return err
		}
		for _, r := range regions {
			if r.id == "" {
				return fmt.Errorf("%s: new flea region %q needs extraction. Run flea dev, then commit the generated files before deploying.", file, r.label)
			}
		}
	}
	return nil
}

func (p *Project) legacyFleaIDs() map[string]bool {
	legacy := map[string]bool{}
	for name := range p.Fleas {
		if _, rich := p.rich[name]; !rich {
			legacy[name] = true
		}
	}
	return legacy
}
