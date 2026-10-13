// Package editstore validates and publishes owner edits under the same lock as
// deployment (or local extraction). The HTTP editor never chooses file paths.
package editstore

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"unicode/utf8"

	"flea/internal/editor"
	"flea/internal/site"
	"flea/internal/snips"
	"flea/internal/state"
)

type Store struct {
	Root, ContentPath, ConfigPath string
	Emails                        []string
	Publish                       func(*site.Project)
}

func (s *Store) lock() (func(), error) {
	lock, pending := "", filepath.Join(s.Root, "flea", "runtime", ".snips-pending.json")
	if s.ConfigPath != "" {
		lock = filepath.Join(filepath.Dir(s.ConfigPath), "deploy.lock")
		pending = filepath.Join(filepath.Dir(s.ConfigPath), "deploy.pending.json")
	}
	var unlock func()
	var err error
	if s.ConfigPath == "" {
		unlock = site.LockContent()
	} else {
		unlock, err = state.Lock(lock)
	}
	if err != nil {
		return nil, err
	}
	if _, err = os.Lstat(pending); !os.IsNotExist(err) {
		unlock()
		return nil, fmt.Errorf("content recovery is pending")
	}
	if s.ConfigPath != "" {
		if _, e := os.Lstat(filepath.Join(filepath.Dir(s.ConfigPath), "upgrade.pending.json")); !os.IsNotExist(e) {
			unlock()
			return nil, fmt.Errorf("server upgrade recovery pending")
		}
		data, err := os.ReadFile(s.ConfigPath)
		var c struct{ Sites, Content map[string]string }
		if err == nil {
			err = json.Unmarshal(data, &c)
		}
		active := false
		for domain, root := range c.Sites {
			if filepath.Clean(root) == filepath.Clean(s.Root) && filepath.Clean(c.Content[domain]) == filepath.Clean(s.ContentPath) {
				active = true
			}
		}
		if err != nil || !active {
			unlock()
			return nil, fmt.Errorf("the active release changed")
		}
	}
	return unlock, nil
}
func (s *Store) load() (*site.Project, error) {
	var content *snips.Store
	var err error
	if s.ConfigPath != "" {
		content, err = snips.ReadFile(s.ContentPath)
		if err != nil {
			return nil, err
		}
	}
	p, err := site.LoadWithMedia(s.Root, content, s.mediaRoot())
	if err != nil {
		return nil, err
	}
	if !slices.Equal(p.Admin.Emails, s.Emails) {
		return nil, fmt.Errorf("admin access changed; restart dev")
	}
	return p, nil
}
func (s *Store) Read() (editor.Document, error) {
	unlock, err := s.lock()
	if err != nil {
		return editor.Document{}, err
	}
	defer unlock()
	p, err := s.load()
	if err != nil {
		return editor.Document{}, err
	}
	return editor.Document{BlogMount: p.Apps["blog"].Mount, Store: p.Content, Routes: p.Routes(), Version: p.EditingVersion()}, nil
}
func (s *Store) Save(u editor.Update) error {
	unlock, err := s.lock()
	if err != nil {
		return err
	}
	defer unlock()
	p, err := s.load()
	if err != nil {
		return err
	}
	if u.Page != nil {
		u.Version = u.Page.Version
	}
	if u.Version == "" || u.Version != p.EditingVersion() {
		return fmt.Errorf("content changed since opening this form")
	}
	if u.Blog != nil {
		if err := s.applyBlog(p, *u.Blog); err != nil {
			return err
		}
	} else if u.Page != nil {
		if err := s.applyMedia(p, *u.Page); err != nil {
			return err
		}
		if err := applyPage(p, *u.Page); err != nil {
			return err
		}
	} else if u.ID != "" {
		refs, err := p.SnipReferences()
		if err != nil {
			return err
		}
		if _, ok := refs[u.ID]; !ok {
			return fmt.Errorf("snip is not referenced by this site")
		}
		index, err := p.Content.Index()
		if err != nil {
			return err
		}
		entry, ok := index[u.ID]
		if !ok {
			return fmt.Errorf("unknown snip")
		}
		entry.Content, err = snips.Sanitize(u.Content)
		if err != nil {
			return err
		}
		if _, ok := p.Content.Shared.Snips[u.ID]; ok {
			p.Content.Shared.Snips[u.ID] = entry
		} else {
			for _, page := range p.Content.Pages {
				if _, ok := page.Snips[u.ID]; ok {
					page.Snips[u.ID] = entry
				}
			}
		}
	} else {
		if _, ok := p.Pages[u.Route]; !ok {
			return fmt.Errorf("unknown page")
		}
		if utf8.RuneCountInString(u.Title) > 200 || utf8.RuneCountInString(u.Description) > 500 {
			return fmt.Errorf("page details exceed the length limit")
		}
		page := p.Content.Page(u.Route)
		page.Title = u.Title
		page.Description = u.Description
	}
	// Validate the full candidate before making it durable or visible to readers.
	next, err := site.LoadWithMedia(s.Root, p.Content, s.mediaRoot())
	if err != nil {
		return err
	}
	if s.ConfigPath == "" {
		err = site.SaveContent(s.Root, p.Content)
	} else {
		err = snips.WriteFile(s.ContentPath, p.Content)
	}
	if err != nil {
		return err
	}
	s.Publish(next)
	if u.Blog != nil && u.BlogResult != nil {
		for _, post := range next.Content.Blog {
			if post.ID == u.Blog.Post.ID {
				*u.BlogResult = editor.BlogResult{Version: next.EditingVersion(), Post: post}
				if photo, ok := next.Content.Media[post.Cover]; ok {
					u.BlogResult.Cover = &photo
				}
				break
			}
		}
	}
	if u.Page != nil && u.Result != nil {
		page := next.Content.Page(u.Page.Route)
		result := editor.PageResult{Page: editor.PageEdit{Version: next.EditingVersion(), Route: u.Page.Route, Title: page.Title, Description: page.Description, Changes: []editor.ContentEdit{}, Multisnips: map[string][]snips.MultiItem{}}, Media: map[string]snips.Entry{}}
		index, _ := next.Content.Index() // The candidate has already been validated.
		result.Photos = next.Content.Media
		for _, change := range u.Page.Changes {
			result.Page.Changes = append(result.Page.Changes, editor.ContentEdit{ID: change.ID, Content: index[change.ID].Content})
		}
		for name := range u.Page.Multisnips {
			result.Page.Multisnips[name] = page.Multisnips[name]
		}
		for _, change := range u.Page.Media {
			if !change.Library {
				result.Media[change.ID] = index[change.ID]
			}
		}
		*u.Result = result
	}
	return nil
}
