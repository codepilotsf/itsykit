package site

import (
	"bytes"
	"flea/internal/snips"
	"flea/internal/video"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var mediaRef = regexp.MustCompile(`^\{snip ([A-Za-z0-9]{10})\.(src|alt|poster)\}$`)
var mediaTextRef = regexp.MustCompile(`\{snip [A-Za-z0-9]{10}\.[^}]*\}`)

type mediaBinding struct {
	start, end int
	tag, id    string
	attrs      map[string]string
	author     bool
}

// scanMediaBindings validates the complete element, never individual attributes in isolation.
func scanMediaBindings(file, source string) ([]mediaBinding, error) {
	z := html.NewTokenizer(strings.NewReader(source))
	pos := 0
	type ancestor struct {
		tag   string
		snip  bool
		video bool
		inert bool
	}
	var stack []ancestor
	var out []mediaBinding
	for {
		kind := z.Next()
		raw := string(z.Raw())
		start := pos
		pos += len(raw)
		if kind == html.ErrorToken {
			if z.Err() != io.EOF {
				return nil, z.Err()
			}
			break
		}
		t := z.Token()
		fail := func(msg string) ([]mediaBinding, error) {
			return nil, sourceError(file, source, start, fmt.Errorf("%s", msg))
		}
		if kind == html.TextToken && strings.Contains(raw, "{snip ") {
			for _, m := range mediaTextRef.FindAllString(raw, -1) {
				if len(stack) == 0 || (stack[len(stack)-1].tag != "script" && stack[len(stack)-1].tag != "style") {
					return fail("media snip fields are only allowed in img/video attributes: " + m)
				}
			}
		}
		if kind == html.EndTagToken {
			for i := len(stack) - 1; i >= 0; i-- {
				if stack[i].tag == t.Data {
					for _, a := range stack[i+1:] {
						if a.video {
							return fail("unclosed video snip")
						}
					}
					stack = stack[:i]
					break
				}
			}
			continue
		}
		if kind != html.StartTagToken && kind != html.SelfClosingTagToken {
			continue
		}
		attrs := map[string]string{}
		duplicate := false
		author := false
		id := ""
		for _, a := range t.Attr {
			if a.Key == "data-snip-media" {
				return fail("data-snip-media is reserved")
			}
			if _, ok := attrs[a.Key]; ok {
				duplicate = true
			}
			attrs[a.Key] = a.Val
			if strings.Contains(a.Val, "{snip ") && strings.Contains(a.Val, ".") && !mediaRef.MatchString(a.Val) {
				return fail("media snip attributes require one complete field binding")
			}
			if a.Key == "snip" {
				author = true
			}
			if m := mediaRef.FindStringSubmatch(a.Val); m != nil {
				if id != "" && id != m[1] {
					return fail("media attributes must reference one snip")
				}
				id = m[1]
				if a.Key != m[2] {
					return fail("media snip field must match its attribute")
				}
			}
		}
		isMedia := t.Data == "img" || t.Data == "video"
		// Named image fields are handled by the multisnip extractor.
		namedPhoto := t.Data == "img" && author && attrs["snip"] != ""
		activeMedia := isMedia && ((author && !namedPhoto) || id != "")
		if activeMedia && duplicate {
			return fail("duplicate media attribute")
		}
		for _, a := range stack {
			if activeMedia && a.inert {
				return fail("media snips are not editable inside inert content")
			}
			if activeMedia && a.tag == "picture" {
				return fail("image snips cannot be inside picture elements with competing sources")
			}
			if activeMedia && strings.Contains(" button form template head script style ", " "+a.tag+" ") {
				return fail("media snips are not editable in this context")
			}
			if (author || id != "") && a.snip {
				return fail("nested snip regions are not allowed")
			}
			if a.video && t.Data == "source" {
				return fail("video snips use their src attribute, not nested source elements")
			}
		}
		if id != "" {
			if !isMedia {
				return fail("media snip references belong on img or video elements")
			}
			if attrs["src"] != "{snip "+id+".src}" {
				return fail("media snips require a src binding")
			}
			field := "alt"
			if t.Data == "video" {
				field = "poster"
			}
			if attrs[field] != "{snip "+id+"."+field+"}" {
				return fail("media snips require a " + field + " binding")
			}
			for key, value := range attrs {
				if strings.Contains(value, "{snip ") && key != "src" && key != field {
					return fail("unsupported media snip attribute")
				}
			}
		}
		if activeMedia {
			if t.Data == "video" && kind == html.SelfClosingTagToken {
				return fail("video snips require an explicit closing video tag")
			}
			if _, ok := attrs["inert"]; ok {
				return fail("media snips cannot be inert")
			}
			if _, ok := attrs["srcset"]; ok {
				return fail("image snips cannot have a competing srcset")
			}
			out = append(out, mediaBinding{start, pos, t.Data, id, attrs, author})
		}
		void := strings.Contains(" area base br col embed hr img input link meta param source track wbr ", " "+t.Data+" ") || kind == html.SelfClosingTagToken
		if !void {
			_, inert := attrs["inert"]
			stack = append(stack, ancestor{t.Data, author || activeMedia, activeMedia && t.Data == "video", inert})
		}
	}
	for _, a := range stack {
		if a.video {
			return nil, sourceError(file, source, len(source), fmt.Errorf("unclosed video snip"))
		}
	}
	return out, nil
}

func extractMediaSnips(root, file, source string, store *snips.Store, used map[string]struct{}) (string, error) {
	bindings, err := scanMediaBindings(file, source)
	if err != nil {
		return "", err
	}
	var out strings.Builder
	last := 0
	for _, b := range bindings {
		if !b.author {
			continue
		}
		if b.id != "" {
			return "", sourceError(file, source, b.start, fmt.Errorf("remove redundant snip attribute from media binding"))
		}
		id, err := snips.ReserveID(used)
		if err != nil {
			return "", err
		}
		entry := snips.Entry{Type: "image"}
		photo := func(src, alt string) error {
			scratch := snips.NewStore()
			pid, e := importPhoto(root, src, alt, scratch, used)
			if e != nil {
				return e
			}
			p := scratch.Media[pid]
			if b.tag == "img" {
				entry.Image = &p
			} else {
				entry.Poster = &p
			}
			return nil
		}
		if b.tag == "img" {
			err = photo(b.attrs["src"], b.attrs["alt"])
		} else {
			entry.Type = "video"
			src := b.attrs["src"]
			if src == "" || strings.ContainsAny(src, ":?#\\") {
				return "", sourceError(file, source, b.start, fmt.Errorf("video seed requires a local optimized MP4"))
			}
			r, e := os.OpenRoot(root)
			if e != nil {
				return "", e
			}
			data, e := readFileBounded(r, "site/"+strings.TrimPrefix(src, "/"), video.MaxFile)
			if e == nil {
				var m video.Metadata
				m, e = video.Inspect(bytes.NewReader(data), int64(len(data)))
				if e == nil {
					v := video.Video{Metadata: m, Name: filepath.Base(src)}
					entry.Video = &v
					if e = r.Mkdir("flea/media", 0755); os.IsExist(e) {
						e = nil
					}
					if e == nil {
						var entries []os.DirEntry
						entries, e = os.ReadDir(filepath.Join(root, "flea", "media"))
						var total int64
						for _, f := range entries {
							if video.Filename.MatchString(f.Name()) && f.Name() != m.File {
								info, err := f.Info()
								if err != nil {
									e = err
									break
								}
								total += info.Size()
							}
						}
						if e == nil && total+int64(len(data)) > video.MaxTotal {
							e = fmt.Errorf("video storage exceeds 64 MiB")
						}
					}
					if e == nil {
						f, openErr := OpenVideo(filepath.Join(root, "flea", "media"), m.File)
						if openErr == nil {
							e = CheckVideo(f, v)
							f.Close()
						} else if os.IsNotExist(openErr) {
							e = writeSiteFile(r, "flea/media/"+m.File, data)
						} else {
							e = openErr
						}
					}
				}
			}
			r.Close()
			err = e
			if err == nil && b.attrs["poster"] != "" {
				err = photo(b.attrs["poster"], "")
			}
		}
		if err != nil {
			return "", sourceError(file, source, b.start, err)
		}
		store.Page(snipScope(file)).Snips[id] = entry
		raw := source[b.start:b.end]
		field := "alt"
		if b.tag == "video" {
			field = "poster"
		}
		for _, a := range []string{"snip", "src", field} {
			f, t := attributeSpan(raw, a)
			if f >= 0 {
				raw = raw[:f] + raw[t:]
			}
		}
		at := strings.LastIndex(raw, ">")
		if at > 0 && raw[at-1] == '/' {
			at--
		}
		out.WriteString(source[last:b.start])
		out.WriteString(raw[:at] + ` src="{snip ` + id + `.src}" ` + field + `="{snip ` + id + `.` + field + `}"` + raw[at:])
		last = b.end
	}
	out.WriteString(source[last:])
	return out.String(), nil
}

func (p *Project) snipField(id, field string) string {
	e := p.contentIndex[id]
	switch field {
	case "src":
		if e.Image != nil {
			return "/media/" + e.Image.File
		}
		if e.Video != nil {
			return "/media/" + e.Video.File
		}
	case "alt":
		if e.Image != nil {
			return e.Image.Alt
		}
	case "poster":
		if e.Poster != nil {
			return "/media/" + e.Poster.File
		}
	}
	return ""
}

func (p *Project) EditableMedia(route string) (map[string]string, error) {
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
	out := map[string]string{}
	var walk func(*html.Node, bool)
	walk = func(n *html.Node, blocked bool) {
		for _, a := range n.Attr {
			if a.Key == "inert" {
				blocked = true
			}
		}
		if n.Type == html.ElementNode && strings.Contains(" script style head template button form ", " "+n.Data+" ") {
			blocked = true
		}
		if !blocked && (n.Data == "img" || n.Data == "video") {
			for _, a := range n.Attr {
				if a.Key == "data-snip-media" {
					kind := "image"
					if n.Data == "video" {
						kind = "video"
					}
					out[a.Val] = kind
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c, blocked)
		}
	}
	walk(root, false)
	return out, nil
}
