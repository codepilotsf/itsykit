package snips

import "testing"

func TestBlogValidationAndPreservation(t *testing.T) {
	valid := BlogPost{ID: "Abcdef1234", Title: "Post", Slug: "post", Status: "draft"}
	for _, mutate := range []func(*BlogPost){func(p *BlogPost) { p.Slug = "../secret" }, func(p *BlogPost) { p.Status = "scheduled" }, func(p *BlogPost) { p.Status = "published" }, func(p *BlogPost) { p.PublishedAt = "yesterday" }, func(p *BlogPost) { p.Cover = "missing" }, func(p *BlogPost) { p.ID = "bad" }} {
		s := NewStore()
		p := valid
		mutate(&p)
		s.Blog = []BlogPost{p}
		if _, err := s.Index(); err == nil {
			t.Fatal(p)
		}
	}
	seed, remote := NewStore(), NewStore()
	seed.Blog = []BlogPost{valid}
	changed := valid
	changed.Title = "Admin edit"
	remote.Blog = []BlogPost{changed}
	merged, err := Merge(seed, remote, map[string]string{})
	if err != nil || merged.Blog[0].Title != "Admin edit" {
		t.Fatal(merged, err)
	}
	seed.Blog = nil
	merged, err = Merge(seed, remote, map[string]string{})
	if err != nil || len(merged.Blog) != 1 {
		t.Fatal("removed source must retain live posts", err)
	}
}
