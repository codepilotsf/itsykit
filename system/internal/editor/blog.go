package editor

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"itsy/internal/media"
	"itsy/internal/snips"
)

type BlogEdit struct {
	Version string         `json:"version"`
	Post    snips.BlogPost `json:"post"`
	Action  string         `json:"action"`
	Upload  *CoverUpload   `json:"upload,omitempty"`
}
type BlogResult struct {
	Version string         `json:"version"`
	Post    snips.BlogPost `json:"post"`
	Cover   *media.Photo   `json:"cover,omitempty"`
}

func (h *Handler) showBlog(w http.ResponseWriter, r *http.Request, s session) {
	d, err := h.options.Backend.Read()
	if err != nil {
		http.Error(w, "Could not read blog posts.", 503)
		return
	}
	if d.BlogMount == "" {
		http.NotFound(w, r)
		return
	}
	mode := "blog"
	post := snips.BlogPost{Status: "draft"}
	if r.URL.Path == "/admin/blog/edit" {
		mode = "blog-edit"
		id := r.URL.Query().Get("id")
		if id == "" {
			used := map[string]struct{}{}
			for _, v := range d.Store.Blog {
				used[v.ID] = struct{}{}
			}
			post.ID, err = snips.ReserveID(used)
			if err != nil {
				http.Error(w, "Please try again.", 503)
				return
			}
		}
		if id != "" {
			found := false
			for _, v := range d.Store.Blog {
				if v.ID == id {
					post = v
					found = true
					break
				}
			}
			if !found {
				http.NotFound(w, r)
				return
			}
		}
	}
	posts := []any{}
	sorted := append([]snips.BlogPost(nil), d.Store.Blog...)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].PublishedAt > sorted[j].PublishedAt })
	for _, v := range sorted {
		posts = append(posts, map[string]any{"id": v.ID, "title": v.Title, "status": v.Status, "url": d.BlogMount + "/" + v.Slug})
	}
	photos := []any{}
	ids := []string{}
	for id := range d.Store.Media {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		v := d.Store.Media[id]
		photos = append(photos, map[string]any{"id": id, "src": "/media/" + v.File, "alt": v.Alt, "name": v.Name})
	}
	data := map[string]any{"email": s.email, "csrf": s.csrf, "version": d.Version, "posts": posts, "photos": photos, "post": map[string]any{"id": post.ID, "title": post.Title, "slug": post.Slug, "excerpt": post.Excerpt, "body": post.Body, "status": post.Status, "published_at": post.PublishedAt, "cover": post.Cover}}
	rendered, err := h.options.Render(mode, data)
	if err != nil {
		http.Error(w, "Could not show this page. Copy the blog admin templates from a fresh ItsyKit site.", 500)
		return
	}
	// Script policy is scoped to the bundled editor, not arbitrary site scripts.
	scheme := "https"
	if h.options.Local {
		scheme = "http"
	}
	assetURL := (&url.URL{Scheme: scheme, Host: r.Host, Path: "/admin/assets/blog.js"}).String()
	w.Header().Set("Content-Security-Policy", strings.Replace(w.Header().Get("Content-Security-Policy"), "img-src 'self' https: data:", "img-src 'self' https: data: blob:", 1)+"; script-src "+assetURL+"; connect-src 'self'")
	if h.options.Local && h.options.ReloadScript != nil {
		if script := h.options.ReloadScript(); script != "" {
			rendered = appendTrustedScript(w, rendered, script)
		}
	}
	w.Write(rendered)
}

func (h *Handler) saveBlog(w http.ResponseWriter, r *http.Request) {
	var edit BlogEdit
	raw := r.PostForm.Get("edit")
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.DisallowUnknownFields()
	if len(raw) > 1<<20 || decoder.Decode(&edit) != nil || decoder.Decode(new(any)) != io.EOF || edit.Version == "" || !snips.ValidID(edit.Post.ID) {
		http.Error(w, "Invalid blog changes.", 400)
		return
	}
	if edit.Action != "save" && edit.Action != "publish" && edit.Action != "unpublish" {
		http.Error(w, "Choose Save, Publish or Unpublish.", 400)
		return
	}
	if r.MultipartForm != nil {
		for name, files := range r.MultipartForm.File {
			if name != "cover" || edit.Upload == nil || len(files) != 1 || files[0].Size <= 0 || files[0].Size > media.MaxFile {
				http.Error(w, "Choose one JPEG or PNG up to 8 MiB.", 400)
				return
			}
			f, err := files[0].Open()
			if err != nil {
				http.Error(w, "Could not read cover photo.", 400)
				return
			}
			edit.Upload.Data, err = io.ReadAll(io.LimitReader(f, media.MaxFile+1))
			f.Close()
			if err != nil || len(edit.Upload.Data) > media.MaxFile {
				http.Error(w, "Could not read cover photo.", 400)
				return
			}
		}
	}
	if edit.Upload != nil && (len(edit.Upload.Data) == 0 || !snips.ValidID(edit.Upload.ID) || edit.Upload.ID != edit.Post.Cover) {
		http.Error(w, "Invalid cover photo.", 400)
		return
	}
	var result BlogResult
	if err := h.options.Backend.Save(Update{Version: edit.Version, Blog: &edit, BlogResult: &result}); err != nil {
		if h.options.Logger != nil {
			h.options.Logger.Printf("Blog save failed: %q", err.Error())
		}
		var validation *MediaValidationError
		if errors.As(err, &validation) {
			http.Error(w, validation.Message, 400)
			return
		}
		http.Error(w, "Could not save. Content may have changed. Your draft is still here; copy it before reloading.", 409)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// Publication always starts as a draft and is explicit. Existing published
// records keep their status on Save, so edits take effect immediately.
func ApplyBlogAction(post *snips.BlogPost, action string, existing *snips.BlogPost) {
	post.Status = "draft"
	if existing != nil {
		post.Status = existing.Status
	}
	switch action {
	case "publish":
		post.Status = "published"
		if post.PublishedAt == "" {
			post.PublishedAt = time.Now().UTC().Format(time.RFC3339)
		}
	case "unpublish":
		post.Status = "draft"
	}
}
