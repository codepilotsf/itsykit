package admin

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"

	"flea/internal/host"
	"flea/internal/site"
	"flea/internal/snips"
	"flea/internal/syncstate"
)

// prepareContent stages a new snapshot without modifying the active one. The
// returned store is the post-readiness version: removed IDs are archived only
// after activation succeeds. The caller holds the deployment lock throughout.
func (m Manager) prepareContent(c host.Config, name, id, dest string, project *site.Project) (string, *snips.Store, error) {
	return m.prepareContentPlan(c, name, id, dest, project, nil)
}
func (m Manager) prepareContentPlan(c host.Config, name, id, dest string, project *site.Project, plan *syncstate.Plan) (string, *snips.Store, error) {
	refs, e := project.SnipReferences()
	if e != nil {
		return "", nil, e
	}
	remote := snips.NewStore()
	parent := filepath.Dir(dest)
	for domain, previous := range c.Sites {
		if filepath.Dir(previous) == parent {
			if file := c.Content[domain]; file != "" {
				remote, e = snips.ReadFile(file)
			} else {
				p, e := site.Load(previous)
				if e != nil {
					return "", nil, e
				}
				remote = p.Content
			}
			if e != nil {
				return "", nil, e
			}
			break
		}
	}
	selected := remote
	if plan != nil && plan.Snips != nil {
		selected = plan.Snips
	}
	activeContent, err := snips.Merge(project.Content, selected, refs)
	if err != nil {
		return "", nil, err
	}
	if err = realDirs(m.DataDir, []string{"sites", name, "content"}); err != nil {
		return "", nil, err
	}
	if err = realDirs(m.DataDir, []string{"sites", name, "flea", "runtime"}); err != nil {
		return "", nil, err
	}
	// Only content and private runtime directories are writable. Code releases,
	// site parents and host configuration remain administrator-owned.
	if m.ConfigPath == "/etc/flea/server.json" {
		account, e := user.Lookup("flea")
		if e != nil {
			return "", nil, e
		}
		uid, e := strconv.Atoi(account.Uid)
		if e != nil {
			return "", nil, e
		}
		gid, e := strconv.Atoi(account.Gid)
		if e != nil {
			return "", nil, e
		}
		directory := filepath.Join(m.DataDir, "sites", name, "content")
		if e = os.Chown(directory, uid, gid); e != nil {
			return "", nil, e
		}
		if e = os.Chmod(directory, 0750); e != nil {
			return "", nil, e
		}
		runtime := filepath.Join(m.DataDir, "sites", name, "flea", "runtime")
		if e = os.Chown(runtime, uid, gid); e != nil {
			return "", nil, e
		}
		if e = os.Chmod(runtime, 0700); e != nil {
			return "", nil, e
		}
	}
	contentPath := filepath.Join(m.DataDir, "sites", name, "content", id+".json")
	if err = m.prepareMedia(dest, contentPath, activeContent); err != nil {
		return "", nil, err
	}
	// Candidate is private until config activation. Keep orphans active until
	// readiness succeeds, then Apply writes the pruned store after verification.
	candidateBytes, e := activeContent.Bytes()
	if e != nil {
		return "", nil, e
	}
	candidateStore, e := snips.Decode(candidateBytes)
	if e != nil {
		return "", nil, e
	}
	oldIndex, e := remote.Index()
	if e != nil {
		return "", nil, e
	}
	for key, entry := range oldIndex {
		if _, ok := refs[key]; !ok {
			candidateStore.Shared.Snips[key] = entry
		}
	}
	candidateBytes, e = candidateStore.Bytes()
	if e != nil {
		return "", nil, e
	}
	if err = atomicWrite(contentPath, candidateBytes, 0644); err != nil {
		return "", nil, err
	}
	if _, err = site.LoadWithMedia(dest, activeContent, site.ManagedMediaRoot(contentPath)); err != nil {
		return "", nil, fmt.Errorf("templates and preserved content: %w", err)
	}
	return contentPath, activeContent, nil
}
