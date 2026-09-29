package site

import (
	"flea/internal/snips"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"strings"
)

func authorMultisnips(file, source string) ([]multiBlock, error) {
	z := html.NewTokenizer(strings.NewReader(source))
	offset := 0
	var stack []string
	var active *multiBlock
	var blocks []multiBlock
	depth := 0
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
		switch kind {
		case html.StartTagToken, html.SelfClosingTagToken:
			name := ""
			count := 0
			for _, a := range token.Attr {
				if a.Key == "multisnip" {
					count++
					name = a.Val
				}
			}
			void := strings.Contains(" area base br col embed hr img input link meta param source track wbr ", " "+token.Data+" ") || kind == html.SelfClosingTagToken
			if count > 0 {
				if count != 1 || active != nil || !snips.ValidFieldName(name) || void || snipScope(file) == "" || strings.HasPrefix(file, "admin/") {
					return nil, fmt.Errorf("%s: invalid, nested or non-page multisnip", file)
				}
				active = &multiBlock{name: name, start: start, bodyStart: offset}
				depth = len(stack)
			}
			if !void {
				stack = append(stack, token.Data)
			}
		case html.EndTagToken:
			at := -1
			for i := len(stack) - 1; i >= 0; i-- {
				if stack[i] == token.Data {
					at = i
					break
				}
			}
			if active != nil && (at < 0 || at != len(stack)-1) {
				return nil, fmt.Errorf("%s: multisnip HTML requires balanced closing tags", file)
			}
			if at >= 0 {
				if active != nil && at == depth {
					active.bodyEnd = start
					active.end = offset
					blocks = append(blocks, *active)
					active = nil
				}
				stack = stack[:at]
			}
		}
	}
	if active != nil {
		return nil, fmt.Errorf("%s: unclosed multisnip", file)
	}
	return blocks, nil
}

func extractMultisnips(root, file, source string, store *snips.Store, used map[string]struct{}) (string, error) {
	blocks, err := authorMultisnips(file, source)
	if err != nil {
		return "", err
	}
	seen := map[string]string{}
	seedNew := map[string]bool{}
	previousName := ""
	var out strings.Builder
	last := 0
	for _, b := range blocks {
		priorTemplate, repeated := seen[b.name]
		if repeated && (previousName != b.name || strings.TrimSpace(source[last:b.start]) != "") {
			return "", fmt.Errorf("%s: seed cards for one multisnip must be adjacent", file)
		}
		outer := source[b.start:b.bodyStart]
		from, to := attributeSpan(outer, "multisnip")
		if from < 0 {
			return "", fmt.Errorf("missing multisnip attribute")
		}
		body := source[b.bodyStart:b.bodyEnd]
		body, photoFields, err := extractMultiPhotos(root, body, store, used)
		if err != nil {
			return "", err
		}
		regions, refs, err := scanRegions(file, body, nil)
		if err != nil {
			return "", err
		}
		if len(refs) > 0 {
			return "", fmt.Errorf("multisnip seeds cannot contain existing snip references")
		}
		fields := photoFields
		var inner strings.Builder
		pos := 0
		for _, r := range regions {
			if !snips.ValidFieldName(r.label) {
				return "", fmt.Errorf("multisnip fields need a name, such as snip=\"name\"")
			}
			if _, ok := fields[r.label]; ok {
				return "", fmt.Errorf("duplicate multisnip field")
			}
			fields[r.label] = r.content
			inner.WriteString(body[pos:r.attributeStart])
			inner.WriteString(body[r.attributeEnd:r.start])
			inner.WriteString("{snip." + r.label + "}")
			pos = r.end
		}
		inner.WriteString(body[pos:])
		templateBody := outer[:from] + outer[to:] + inner.String() + source[b.bodyEnd:b.end]
		if repeated && templateBody != priorTemplate {
			return "", fmt.Errorf("%s: multisnip seed cards must use identical markup outside their fields", file)
		}
		seen[b.name] = templateBody
		def, err := definition(templateBody)
		if err != nil {
			return "", err
		}
		for name, value := range fields {
			if def.Photos[name] {
				continue
			}
			clean, err := snips.CleanField(value, def.Fields[name])
			if err != nil {
				return "", err
			}
			fields[name] = clean
		}
		page := store.Page(snipScope(file))
		if page.Multisnips == nil {
			page.Multisnips = map[string][]snips.MultiItem{}
		}
		if !repeated {
			_, exists := page.Multisnips[b.name]
			seedNew[b.name] = !exists
		}
		if seedNew[b.name] {
			id, err := snips.ReserveID(used)
			if err != nil {
				return "", err
			}
			page.Multisnips[b.name] = append(page.Multisnips[b.name], snips.MultiItem{ID: id, Fields: fields})
		}
		out.WriteString(source[last:b.start])
		if !repeated {
			out.WriteString("{#multisnip " + b.name + "}\n" + templateBody + "\n{/multisnip}")
		}
		last = b.end
		previousName = b.name
	}
	out.WriteString(source[last:])
	return out.String(), nil
}
