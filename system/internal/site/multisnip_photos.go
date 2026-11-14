package site

import (
	"fmt"
	"golang.org/x/net/html"
	"io"
	"itsy/internal/media"
	"itsy/internal/sitefiles"
	"itsy/internal/snips"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var existingPhoto = regexp.MustCompile(`^\{media\.([A-Za-z][A-Za-z0-9]{9})\.src\}$`)

func extractMultiPhotos(root, body string, store *snips.Store, used map[string]struct{}) (string, map[string]string, error) {
	fields := map[string]string{}
	z := html.NewTokenizer(strings.NewReader(body))
	var out strings.Builder
	for {
		kind := z.Next()
		raw := string(z.Raw())
		if kind == html.ErrorToken {
			if z.Err() != io.EOF {
				return "", nil, z.Err()
			}
			break
		}
		token := z.Token()
		field := ""
		attrs := map[string]string{}
		if (kind == html.StartTagToken || kind == html.SelfClosingTagToken) && token.Data == "img" {
			for _, a := range token.Attr {
				if _, ok := attrs[a.Key]; ok {
					return "", nil, fmt.Errorf("duplicate image attribute")
				}
				attrs[a.Key] = a.Val
				if a.Key == "snip" {
					field = a.Val
				}
			}
		}
		if field == "" {
			out.WriteString(raw)
			continue
		}
		if !snips.ValidFieldName(field) {
			return "", nil, fmt.Errorf("photo fields need a short name")
		}
		if _, exists := fields[field]; exists {
			return "", nil, fmt.Errorf("duplicate photo field")
		}
		id := ""
		src := attrs["src"]
		if match := existingPhoto.FindStringSubmatch(src); match != nil {
			id = match[1]
			if _, ok := store.Media[id]; !ok {
				return "", nil, fmt.Errorf("unknown photo in seed")
			}
		} else if src != "" {
			var err error
			id, err = importPhoto(root, src, attrs["alt"], store, used)
			if err != nil {
				return "", nil, err
			}
		}
		fields[field] = id
		for _, attr := range []string{"snip", "src", "alt", "data-snip-photo", "data-photo-id"} {
			from, to := attributeSpan(raw, attr)
			if from >= 0 {
				raw = raw[:from] + raw[to:]
			}
		}
		at := strings.LastIndex(raw, ">")
		if at < 0 {
			return "", nil, fmt.Errorf("invalid image field")
		}
		if at > 0 && raw[at-1] == '/' {
			at--
		}
		out.WriteString(raw[:at] + ` data-snip-photo="` + field + `" data-photo-id="{snip.` + field + `.id}" src="{snip.` + field + `.src}" alt="{snip.` + field + `.alt}"` + raw[at:])
	}
	return out.String(), fields, nil
}
func importPhoto(root, src, alt string, store *snips.Store, used map[string]struct{}) (string, error) {
	if strings.ContainsAny(src, ":?#\\") {
		return "", fmt.Errorf("photo seed must use a local JPEG/PNG path")
	}
	source := strings.TrimPrefix(src, "/")
	r, err := os.OpenRoot(root)
	if err != nil {
		return "", err
	}
	defer r.Close()
	data, err := readFileBounded(r, "site/"+source, media.MaxFile)
	if err != nil {
		return "", err
	}
	photo, data, err := media.Process(data, filepath.Base(source), alt)
	if err != nil {
		return "", err
	}
	if store.Media == nil {
		store.Media = map[string]media.Photo{}
	}
	for id, p := range store.Media {
		used[id] = struct{}{}
		if p.File == photo.File && p.Alt == photo.Alt {
			return id, nil
		}
	}
	if len(store.Media) >= media.MaxPhotos {
		return "", fmt.Errorf("photo library is full")
	}
	id := ""
	for attempt := 0; attempt < 128; attempt++ {
		id, err = snips.ReserveID(used)
		if err != nil {
			return "", err
		}
		if id[0] >= 'A' && id[0] <= 'Z' || id[0] >= 'a' && id[0] <= 'z' {
			break
		}
		id = ""
	}
	if id == "" {
		return "", fmt.Errorf("could not allocate photo ID")
	}
	if err = r.Mkdir(sitefiles.Resolve(r, "itsy/media"), 0755); err != nil && !os.IsExist(err) {
		return "", err
	}
	parent, err := pinnedDir(r, sitefiles.Resolve(r, "itsy/media"))
	if err != nil {
		return "", err
	}
	defer parent.Close()
	var total int64
	entries, err := os.ReadDir(sitefiles.Path(root, filepath.Join("itsy", "media")))
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		info, e := entry.Info()
		if e != nil {
			return "", e
		}
		if filepath.Ext(entry.Name()) != ".mp4" {
			total += info.Size()
		}
	}
	if total+int64(len(data)) > media.MaxTotal {
		return "", fmt.Errorf("photo storage exceeds 64 MiB")
	}
	if existing, e := readFileBounded(parent, photo.File, media.MaxFile); e == nil {
		if e = media.CheckBytes(photo, existing); e != nil {
			return "", e
		}
	} else if !os.IsNotExist(e) {
		return "", e
	}
	if err = writeSiteFile(r, "itsy/media/"+photo.File, data); err != nil {
		return "", err
	}
	store.Media[id] = photo
	return id, nil
}
