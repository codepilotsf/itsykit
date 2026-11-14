// Package editstore validates and publishes owner edits under the same lock as
// deployment (or local dev reload). The HTTP editor never chooses file paths.
package editstore

import (
	"encoding/json"
	"fmt"
	"itsy/internal/sitefiles"
	"os"
	"path/filepath"
	"slices"

	"itsy/internal/editor"
	"itsy/internal/site"
	"itsy/internal/snips"
	"itsy/internal/state"
)

type Store struct {
	Root, ContentPath, ConfigPath string
	Emails                        []string
	Publish                       func(*site.Project)
}

func (s *Store) lock() (func(), error) {
	lock, pending := "", sitefiles.Path(s.Root, filepath.Join("itsy", "runtime", ".snips-pending.json"))
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
	p, err := site.LoadWithMedia(s.Root, content, s.mediaRoot(), s.runtimeRoot())
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
	return editor.Document{BlogMount: p.Apps["blog"].Mount, Store: p.Content, Version: p.EditingVersion()}, nil
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
	if u.Version == "" || u.Version != p.EditingVersion() {
		return fmt.Errorf("content changed since opening this form")
	}
	if u.Blog != nil {
		if err := s.applyBlog(p, *u.Blog); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("only Blog edits are supported")
	}
	// Validate the full candidate before making it durable or visible to readers.
	next, err := site.LoadWithMedia(s.Root, p.Content, s.mediaRoot(), s.runtimeRoot())
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
	return nil
}
