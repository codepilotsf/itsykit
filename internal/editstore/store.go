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
	"flea/internal/theme"
)

type Store struct {
	ThemePath                     string
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
		var c struct{ Sites, Content, Themes map[string]string }
		if err == nil {
			err = json.Unmarshal(data, &c)
		}
		active := false
		for domain, root := range c.Sites {
			if filepath.Clean(root) == filepath.Clean(s.Root) && filepath.Clean(c.Content[domain]) == filepath.Clean(s.ContentPath) && c.Themes[domain] == s.ThemePath {
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
	var css *string
	if s.ThemePath != "" {
		b, e := site.ReadManagedFile(filepath.Dir(s.ThemePath), filepath.Base(s.ThemePath))
		if e != nil {
			return nil, e
		}
		v := string(b)
		css = &v
	}
	p, err := site.LoadWithMedia(s.Root, content, css, s.mediaRoot())
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
	return editor.Document{Store: p.Content, Theme: p.ThemeCSS(), Routes: p.Routes(), Version: p.EditingVersion()}, nil
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
	if u.Theme != nil || len(u.Scales) > 0 {
		css := p.ThemeCSS()
		if css == nil {
			return fmt.Errorf("no theme.css")
		}
		nextCSS, e := theme.ApplyScales(*css, u.Theme, u.Scales)
		if e != nil {
			return e
		}
		next, e := site.LoadWithMedia(s.Root, p.Content, &nextCSS, s.mediaRoot())
		if e != nil {
			return e
		}
		if s.ConfigPath == "" {
			e = site.WriteManagedFile(s.Root, "site/css/theme.css", []byte(nextCSS))
		} else {
			if s.ThemePath == "" {
				return fmt.Errorf("no managed theme; redeploy first")
			}
			e = site.WriteManagedFile(filepath.Dir(s.ThemePath), filepath.Base(s.ThemePath), []byte(nextCSS))
		}
		if e != nil {
			return e
		}
		s.Publish(next)
		return nil
	}
	if u.Page != nil {
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
	next, err := site.LoadWithMedia(s.Root, p.Content, p.ThemeCSS(), s.mediaRoot())
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
	return nil
}
