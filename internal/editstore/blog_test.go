package editstore

import (
	"encoding/json"
	"flea/internal/editor"
	"flea/internal/site"
	"flea/internal/snips"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBlogLifecycleLocalAndManagedContent(t *testing.T) {
	for _, remote := range []bool{false, true} {
		t.Run(map[bool]string{false: "local", true: "managed"}[remote], func(t *testing.T) {
			s, public := fixture(t, remote)
			config, err := os.ReadFile(filepath.Join(s.Root, ".flea.config.json"))
			if err != nil {
				t.Fatal(err)
			}
			var cfg map[string]any
			json.Unmarshal(config, &cfg)
			cfg["apps"] = map[string]any{"blog": map[string]any{"mount": "/blog"}}
			config, _ = json.Marshal(cfg)
			os.WriteFile(filepath.Join(s.Root, ".flea.config.json"), config, 0644)
			os.Mkdir(filepath.Join(s.Root, "site", "blog"), 0755)
			os.WriteFile(filepath.Join(s.Root, "site", "blog", "_index.html"), []byte(`{#each blog.data.posts as post}{post.title}{/each}`), 0644)
			os.WriteFile(filepath.Join(s.Root, "site", "blog", "_post.html"), []byte(`{blog.data.post.body}`), 0644)
			doc, err := s.Read()
			if err != nil {
				t.Fatal(err)
			}
			oldVersion := doc.Version
			edit := editor.BlogEdit{Version: doc.Version, Post: snips.BlogPost{ID: "Abcdef1234", Title: "New", Slug: "new", Body: `<p>Hello</p><script>bad()</script>`, Status: "published"}, Action: "save"}
			var result editor.BlogResult
			save := func() {
				t.Helper()
				if err := s.Save(editor.Update{Version: edit.Version, Blog: &edit, BlogResult: &result}); err != nil {
					t.Fatal(err)
				}
				edit.Version = result.Version
				edit.Post = result.Post
			}
			save()
			if result.Post.Status != "draft" || strings.Contains(result.Post.Body, "bad()") {
				t.Fatal(result)
			}
			if got := blogRequest(public, "/blog/new"); got != 404 {
				t.Fatal("draft is public", got)
			}
			edit.Action = "publish"
			save()
			if result.Post.PublishedAt == "" || blogRequest(public, "/blog/new") != 200 {
				t.Fatal(result)
			}
			edit.Action = "save"
			edit.Post.Body = "<p>Updated</p>"
			save()
			if result.Post.Status != "published" {
				t.Fatal("save unpublished live post")
			}
			if err := s.Save(editor.Update{Version: oldVersion, Blog: &edit}); err == nil {
				t.Fatal("stale save accepted")
			}
			edit.Action = "unpublish"
			save()
			if blogRequest(public, "/blog/new") != 404 {
				t.Fatal("unpublished route remains")
			}
			persisted, err := s.Read()
			if err != nil || len(persisted.Store.Blog) != 1 || persisted.Store.Blog[0].Status != "draft" {
				t.Fatal(persisted, err)
			}
		})
	}
}

func blogRequest(s *site.Server, path string) int {
	w := httptest.NewRecorder()
	s.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
	return w.Code
}
