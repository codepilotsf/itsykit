package snips

import (
	"fmt"
	"regexp"
	"time"
	"unicode/utf8"
)

// BlogPost is private editable content. Public views are built by site, never
// by handing the complete content store to a template.
type BlogPost struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Slug        string `json:"slug"`
	Excerpt     string `json:"excerpt"`
	Body        string `json:"body"`
	Status      string `json:"status"`
	PublishedAt string `json:"published_at,omitempty"`
	Cover       string `json:"cover,omitempty"`
}

var blogSlug = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

const MaxBlogPosts = 200

func (s *Store) ValidateBlog() error {
	if len(s.Blog) > MaxBlogPosts {
		return fmt.Errorf("blog exceeds %d posts", MaxBlogPosts)
	}
	ids, slugs := map[string]bool{}, map[string]bool{}
	for _, p := range s.Blog {
		if !ValidID(p.ID) || ids[p.ID] {
			return fmt.Errorf("invalid or duplicate blog post ID")
		}
		ids[p.ID] = true
		if !blogSlug.MatchString(p.Slug) || len(p.Slug) > 100 || slugs[p.Slug] {
			return fmt.Errorf("invalid or duplicate blog slug")
		}
		slugs[p.Slug] = true
		if (!utf8.ValidString(p.Title) || !utf8.ValidString(p.Excerpt) || !utf8.ValidString(p.Body)) || p.Title == "" || utf8.RuneCountInString(p.Title) > 200 || utf8.RuneCountInString(p.Excerpt) > 500 {
			return fmt.Errorf("invalid blog title or excerpt")
		}
		if _, err := Sanitize(p.Body); err != nil {
			return err
		}
		if p.Status != "draft" && p.Status != "published" {
			return fmt.Errorf("blog status must be draft or published")
		}
		if p.PublishedAt != "" {
			if _, err := time.Parse(time.RFC3339, p.PublishedAt); err != nil {
				return fmt.Errorf("blog publication date must be RFC3339")
			}
		}
		if p.Status == "published" && p.PublishedAt == "" {
			return fmt.Errorf("published blog post needs a publication date")
		}
		if p.Cover != "" {
			if _, ok := s.Media[p.Cover]; !ok {
				return fmt.Errorf("unknown blog cover photo")
			}
		}
	}
	return nil
}

// PreserveBlog keeps remote values and retains removed posts. A source-only ID
// seeds a new post. Explicit synchronization handles updates to existing IDs.
func PreserveBlog(local, remote []BlogPost) []BlogPost {
	out := append([]BlogPost(nil), remote...)
	seen := map[string]bool{}
	for _, p := range remote {
		seen[p.ID] = true
	}
	for _, p := range local {
		if !seen[p.ID] {
			out = append(out, p)
			seen[p.ID] = true
		}
	}
	return out
}
