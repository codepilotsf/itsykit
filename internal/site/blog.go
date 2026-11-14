package site

import (
	"fmt"
	"html/template"
	"sort"
	"strings"
	"time"

	"itsy/internal/snips"
)

const blogBodyAction = `[[NW blogBody $ NW]]`

// Published routes are selected from validated snapshots, never request paths.
func (p *Project) loadBlog() error {
	cfg, enabled := p.Apps["blog"]
	if !enabled {
		return nil
	}
	dir := strings.TrimPrefix(cfg.Mount, "/")
	index, err := p.buildTemplate(dir+"/_index.html", true)
	if err != nil {
		return fmt.Errorf("blog index: %w", err)
	}
	post, err := p.buildTemplate(dir+"/_post.html", true)
	if err != nil {
		return fmt.Errorf("blog post: %w", err)
	}
	if p.Pages[cfg.Mount] != nil || p.Assets[cfg.Mount] != "" {
		return fmt.Errorf("blog mount conflicts with a site route")
	}
	// Blog owns its subtree, including unpublished and future slugs.
	for route := range p.Pages {
		if strings.HasPrefix(route, cfg.Mount+"/") {
			return fmt.Errorf("blog subtree conflicts with page %s", route)
		}
	}
	for route := range p.Assets {
		if strings.HasPrefix(route, cfg.Mount+"/") {
			return fmt.Errorf("blog subtree conflicts with asset %s", route)
		}
	}
	p.Pages[cfg.Mount] = index
	for _, v := range p.Content.Blog {
		if v.Status == "published" {
			p.Pages[cfg.Mount+"/"+v.Slug] = post
		}
	}
	return nil
}

func (p *Project) blogData(route string) map[string]any {
	posts := []any{}
	var current any
	cfg, enabled := p.Apps["blog"]
	if !enabled {
		return map[string]any{"posts": posts, "post": nil}
	}
	published := []snips.BlogPost{}
	for _, v := range p.Content.Blog {
		if v.Status == "published" {
			published = append(published, v)
		}
	}
	sort.Slice(published, func(i, j int) bool {
		a, _ := time.Parse(time.RFC3339, published[i].PublishedAt)
		b, _ := time.Parse(time.RFC3339, published[j].PublishedAt)
		if a.Equal(b) {
			return published[i].ID < published[j].ID
		}
		return a.After(b)
	})
	for _, v := range published {
		date, _ := time.Parse(time.RFC3339, v.PublishedAt)
		body := ""
		if route == cfg.Mount+"/"+v.Slug {
			body, _ = snips.Sanitize(v.Body)
		} // Validate once; sanitize only the current body.
		view := map[string]any{"title": v.Title, "slug": v.Slug, "excerpt": v.Excerpt, "body": body, "url": cfg.Mount + "/" + v.Slug, "published_at": v.PublishedAt, "published_label": date.Format("January 2, 2006"), "cover": nil}
		if photo, ok := p.Content.Media[v.Cover]; ok {
			view["cover"] = map[string]any{"src": "/media/" + photo.File, "alt": photo.Alt, "width": photo.Width, "height": photo.Height}
		}
		posts = append(posts, view)
		if route == view["url"] {
			current = view
		}
	}
	return map[string]any{"posts": posts, "post": current, "url": cfg.Mount}
}

// This is the only rich blog directive. The compiler checks its HTML context;
// data comes exclusively from the sanitized published view, not an admin value.
func blogBody(data map[string]any) template.HTML {
	body, _ := lookup(data, "blog.data.post.body").(string)
	return template.HTML(body)
}
