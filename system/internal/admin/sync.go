package admin

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"itsy/internal/host"
	"itsy/internal/site"
	"itsy/internal/snips"
	"itsy/internal/syncstate"
)

// Snapshot is a privileged SSH-only export. It never includes config, sessions,
// credentials or arbitrary paths. The lock gives content one revision.
func (m Manager) Snapshot(name, branch string) (syncstate.Snapshot, error) {
	unlock, e := m.lock()
	if e != nil {
		return syncstate.Snapshot{}, e
	}
	defer unlock()
	if _, e = os.Lstat(m.pending()); !os.IsNotExist(e) {
		return syncstate.Snapshot{}, fmt.Errorf("deployment recovery pending")
	}
	c, _, e := host.Load(m.ConfigPath, nil)
	if e != nil {
		return syncstate.Snapshot{}, e
	}
	return m.snapshot(c, name, branch)
}
func (m Manager) snapshot(c host.Config, name, branch string) (syncstate.Snapshot, error) {
	s := syncstate.Snapshot{Snips: snips.NewStore()}
	if !regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,63}$`).MatchString(name) {
		return s, fmt.Errorf("invalid sync destination")
	}
	if owner := c.SiteBranches[name]; owner != "" && owner != branch {
		return s, fmt.Errorf("site belongs to another branch")
	}
	parent := filepath.Join(m.DataDir, "sites", name, "releases")
	for domain, root := range c.Sites {
		if filepath.Dir(root) != parent {
			continue
		}
		s.Exists = true
		s.Revision = root
		var content *snips.Store
		var e error
		if path := c.Content[domain]; path != "" {
			content, e = snips.ReadFile(path)
			if e != nil {
				return s, e
			}
		}
		mediaRoot, runtimeRoot := "", ""
		if c.Content[domain] != "" {
			mediaRoot = site.ManagedMediaRoot(c.Content[domain])
			runtimeRoot = site.ManagedRuntimeRoot(c.Content[domain])
		}
		p, e := site.LoadWithMedia(root, content, mediaRoot, runtimeRoot)
		if e != nil {
			return s, e
		}
		s.Snips = p.Content
		return s, nil
	}
	return s, nil
}
