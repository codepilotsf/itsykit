package snips

import "testing"

func TestBlogValidationAndPreservation(t *testing.T) {
	valid := BlogPost{ID: "Abcdef1234", Title: "Post", Slug: "post", Status: "draft"}
	for _, mutate := range []func(*BlogPost){func(p *BlogPost) { p.Slug = "../secret" }, func(p *BlogPost) { p.Status = "scheduled" }, func(p *BlogPost) { p.Status = "published" }, func(p *BlogPost) { p.PublishedAt = "yesterday" }, func(p *BlogPost) { p.Cover = "missing" }, func(p *BlogPost) { p.ID = "bad" }} {
		s := NewStore()
		p := valid
		mutate(&p)
		s.Blog = []BlogPost{p}
		if err := s.Validate(); err == nil {
			t.Fatal(p)
		}
	}
	seed, remote := NewStore(), NewStore()
	seed.Blog = []BlogPost{valid}
	changed := valid
	changed.Title = "Admin edit"
	remote.Blog = []BlogPost{changed}
	merged, err := Merge(seed, remote)
	if err != nil || merged.Blog[0].Title != "Admin edit" {
		t.Fatal(merged, err)
	}
	seed.Blog = nil
	merged, err = Merge(seed, remote)
	if err != nil || len(merged.Blog) != 1 {
		t.Fatal("removed source must retain live posts", err)
	}
}

func TestRemovedPageContentFieldsFailWithoutDiscardingBlog(t *testing.T) {
	for _, field := range []string{"pages", "shared", "archive", "videos"} {
		data := []byte(`{"blog":[{"id":"Abcdef1234","title":"Keep","slug":"keep","status":"draft","body":"Body","excerpt":""}],"` + field + `":{}}`)
		if _, err := Decode(data); err == nil {
			t.Fatal("old field silently discarded", field)
		}
	}
	for _, data := range []string{`{"blog":[],"blog":[]}`, `{} {}`, `null`} {
		if _, err := Decode([]byte(data)); err == nil {
			t.Fatal("invalid content accepted", data)
		}
	}
}
