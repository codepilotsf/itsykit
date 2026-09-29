package editor

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"html/template"
	"net/http"
	"regexp"
	"strings"

	"flea/internal/site"
	"flea/internal/snips"
	"flea/internal/viewer"
	"flea/templates"
)

// Isolated toolbar behavior tests supply their own explicit test layout and
// data, rather than depending on the bootstrap endpoint. The complete external
// bootstrap is covered separately with a real site and backend.
func browserDecorate(w http.ResponseWriter, r *http.Request, decorate func([]byte, *snips.Store, string) ([]byte, bool)) func([]byte, *snips.Store, string) ([]byte, bool) {
	return func(body []byte, store *snips.Store, version string) ([]byte, bool) {
		body, ok := decorate(body, store, version)
		if !ok {
			return body, false
		}
		identity, _ := viewer.From(r.Context())
		shared := []string{}
		for id := range store.Shared.Snips {
			shared = append(shared, id)
		}
		page := PageEdit{Route: r.URL.Path, Version: version}
		if p := store.Pages[r.URL.Path]; p != nil {
			page.Title = p.Title
			page.Description = p.Description
			page.Multisnips = p.Multisnips
		}
		sharedJSON, _ := json.Marshal(shared)
		pageJSON, _ := json.Marshal(page)
		photosJSON, _ := json.Marshal(store.Media)
		mediaEntries := map[string]snips.Entry{}
		for id, entry := range store.Shared.Snips {
			if entry.Type != "" {
				mediaEntries[id] = entry
			}
		}
		if p := store.Pages[r.URL.Path]; p != nil {
			for id, entry := range p.Snips {
				if entry.Type != "" {
					mediaEntries[id] = entry
				}
			}
		}
		mediaJSON, _ := json.Marshal(mediaEntries)
		var out bytes.Buffer
		markup, _ := templates.Files.ReadFile("site/admin/_toolbar.html")
		section := string(markup)
		section = section[strings.Index(section, "<section") : strings.Index(section, "</section>")+len("</section>")]
		section = strings.ReplaceAll(section, "{app.current_admin.email}", template.HTMLEscapeString(identity.Email))
		section = regexp.MustCompile(`\{icon ([a-z-]+) size="18"\}`).ReplaceAllStringFunc(section, func(directive string) string {
			name := strings.Fields(directive)[1]
			data, err := templates.Files.ReadFile("flea/icons/" + name + ".svg")
			if err != nil {
				panic(err)
			}
			markup, err := site.CompileIconMarkup(directive, data)
			if err != nil {
				panic(err)
			}
			return markup
		})
		shell := template.Must(template.New("fixture").Parse(`<div id="flea-owner-root" data-revision="" data-email="{{.Email}}" data-csrf="{{.CSRF}}" data-shared="{{.Shared}}" data-page="{{.Page}}" data-photos="{{.Photos}}" data-media="{{.Media}}" data-photo-placeholder="{{.PhotoPlaceholder}}"><template shadowrootmode="open">{{.UI}}</template></div>`))
		_ = shell.Execute(&out, map[string]any{"Email": identity.Email, "CSRF": identity.CSRF, "Shared": string(sharedJSON), "Page": string(pageJSON), "Photos": string(photosJSON), "Media": string(mediaJSON), "PhotoPlaceholder": site.EmptyPhotoSource, "UI": template.HTML("<style>" + string(editorAssets["editor.css"]) + "</style>" + section + inlineUI + collectionIconTemplates())})
		policy := w.Header().Get("Content-Security-Policy")
		shellMarkup := bytes.Clone(out.Bytes())
		out.Reset()
		for _, script := range [][]byte{inlinePurify, inlineSquire, inlineOwner, inlineCollections, inlineTheme, inlineMedia} {
			hash := sha256.Sum256(script)
			policy += " 'sha256-" + base64.StdEncoding.EncodeToString(hash[:]) + "'"
			out.WriteString("<script>")
			out.Write(script)
			out.WriteString("</script>")
		}
		w.Header().Set("Content-Security-Policy", policy)
		if start := bytes.Index(body, []byte("<body>")); start >= 0 {
			start += len("<body>")
			body = append(append(append([]byte(nil), body[:start]...), shellMarkup...), body[start:]...)
		} else {
			body = append(body, shellMarkup...)
		}
		return append(body, out.Bytes()...), true
	}
}
