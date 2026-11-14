package editor

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"itsy/internal/site"
	"itsy/internal/snips"
	"itsy/templates"
)

//go:embed inline/blog.js
var blogJS []byte

//go:embed inline/owner.css
var editorCSS []byte

//go:embed icons/*.svg icons/LICENSE
var collectionIcons embed.FS

func collectionIconTemplates() string {
	var out strings.Builder
	for _, icon := range []struct{ id, name, size string }{{"grip", "grip-vertical", "18"}, {"remove", "x", "18"}, {"add", "plus", "24"}} {
		data, err := collectionIcons.ReadFile("icons/" + icon.name + ".svg")
		if err != nil {
			panic(err)
		}
		markup, err := site.CompileIconMarkup(`{icon `+icon.name+` size="`+icon.size+`"}`, data)
		if err != nil {
			panic(err)
		}
		out.WriteString(`<template id="collection-` + icon.id + `">` + markup + `</template>`)
	}
	return out.String()
}

// Bundled assets contain no request data and may be cached across sessions.
var editorAssets = func() map[string][]byte {
	css, err := templates.Files.ReadFile("site/css/ui.css")
	if err != nil {
		panic(err)
	}
	css = append(css, editorCSS...)
	js := append([]byte(nil), inlinePurify...)
	js = append(js, '\n', ';', '\n')
	js = append(js, inlineSquire...)
	js = append(js, []byte(`
;(async()=>{
 const host=document.getElementById('flea-owner-root'); if(!host)return;
 const csrf=document.querySelector('meta[name="csrf-token"]')?.content;
 if(!csrf)throw Error('Missing csrf-token meta tag in the layout.');
 if(!host.shadowRoot)throw Error('Update site/admin/_toolbar.html to include the declarative toolbar shell.');
 const edit=host.shadowRoot.querySelector('#edit');
 try{edit.checked=sessionStorage.getItem('flea-editing:'+host.dataset.email)!=='off';}catch{}
 const [response,markup]=await Promise.all([
   fetch('/admin/editor-context?route='+encodeURIComponent(location.pathname),{cache:'no-store',headers:{'X-CSRF-Token':csrf},signal:AbortSignal.timeout(15000)}),
   fetch('/admin/assets/toolbar.html',{signal:AbortSignal.timeout(15000)})
 ]);
 if(!response.ok)throw Error('Could not load editor. Refresh the page and sign in again.');
 const data=await response.json();
 if(!host.dataset.version || host.dataset.version!==data.page.version)throw Error('The page changed or the toolbar version is missing. Refresh before editing.');
 if(!markup.ok)throw Error('Could not load the toolbar.');
 const template=document.createElement('template');template.innerHTML=await markup.text();host.shadowRoot.append(template.content);
 host.dataset.csrf=csrf;host.dataset.email=data.email;host.dataset.shared=JSON.stringify(data.shared);host.dataset.page=JSON.stringify(data.page);host.dataset.photos=JSON.stringify(data.photos);host.dataset.media=JSON.stringify(data.media);host.dataset.photoPlaceholder=data.photoPlaceholder;
`)...)
	for _, script := range [][]byte{inlineOwner, inlineCollections, inlineMedia} {
		js = append(js, '\n', ';', '\n')
		js = append(js, script...)
	}
	js = append(js, []byte(`
})().catch(error=>{const host=document.getElementById('flea-owner-root');if(host){const status=host.shadowRoot?.querySelector('#status')||host;status.textContent=error.message;status.setAttribute('role','alert');}console.error(error);});`)...)
	blog := append([]byte(nil), inlinePurify...)
	blog = append(blog, '\n', ';', '\n')
	blog = append(blog, inlineSquire...)
	blog = append(blog, '\n', ';', '\n')
	blog = append(blog, blogJS...)
	return map[string][]byte{"blog.js": blog, "editor.js": js, "editor.css": css, "toolbar.html": []byte(inlineUI + collectionIconTemplates())}
}()

func serveEditorAsset(w http.ResponseWriter, r *http.Request) bool {
	if !strings.HasPrefix(r.URL.Path, "/admin/assets/") {
		return false
	}
	data, ok := editorAssets[strings.TrimPrefix(r.URL.Path, "/admin/assets/")]
	if !ok || r.URL.RawPath != "" {
		http.NotFound(w, r)
		return true
	}
	if r.Method != "GET" && r.Method != "HEAD" {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "Method not allowed", 405)
		return true
	}
	kind := "text/html; charset=utf-8"
	if strings.HasSuffix(r.URL.Path, ".js") {
		kind = "text/javascript; charset=utf-8"
	}
	if strings.HasSuffix(r.URL.Path, ".css") {
		kind = "text/css; charset=utf-8"
	}
	w.Header().Set("Content-Type", kind)
	w.Header().Set("Cache-Control", "public, max-age=0, must-revalidate")
	w.Header().Set("ETag", fmt.Sprintf(`"%x"`, sha256.Sum256(data)))
	http.ServeContent(w, r, r.URL.Path, time.Time{}, bytes.NewReader(data))
	return true
}

func (h *Handler) readEditorContext(w http.ResponseWriter, r *http.Request, s session) {
	if r.Header.Get("X-CSRF-Token") != s.csrf {
		http.Error(w, "Refresh the page before editing.", 403)
		return
	}
	d, err := h.options.Backend.Read()
	route := r.URL.Query().Get("route")
	if err != nil || d.Store == nil {
		http.Error(w, "Could not read site content.", 503)
		return
	}
	if !slices.Contains(d.Routes, route) {
		http.NotFound(w, r)
		return
	}
	page := PageEdit{Route: route, Version: d.Version}
	if p := d.Store.Pages[route]; p != nil {
		page.Title = p.Title
		page.Description = p.Description
		page.Multisnips = p.Multisnips
	}
	shared := []string{}
	mediaSnips := map[string]snips.Entry{}
	collectMedia := func(entries map[string]snips.Entry) {
		for id, entry := range entries {
			if entry.Type == "image" || entry.Type == "video" {
				mediaSnips[id] = entry
			}
		}
	}
	collectMedia(d.Store.Shared.Snips)
	if p := d.Store.Pages[route]; p != nil {
		collectMedia(p.Snips)
	}
	for id := range d.Store.Shared.Snips {
		shared = append(shared, id)
	}
	slices.Sort(shared)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "private, no-store")
	_ = json.NewEncoder(w).Encode(map[string]any{"email": s.email, "page": page, "shared": shared, "photos": d.Store.Media, "media": mediaSnips, "photoPlaceholder": site.EmptyPhotoSource})
}
