package admin

import (
	"fmt"
	"itsy/internal/sitefiles"
	"os"
	"os/user"
	"path/filepath"
	"strconv"

	"itsy/internal/host"
	"itsy/internal/site"
	"itsy/internal/snips"
	"itsy/internal/syncstate"
)

// prepareContentPlan stages a complete Blog snapshot without modifying the active one.
// The caller holds the deployment lock through candidate validation and activation.
func (m Manager) prepareContentPlan(c host.Config, name, id, dest string, project *site.Project, plan *syncstate.Plan) (string, *snips.Store, error) {
	var e error
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
	activeContent, err := snips.Merge(project.Content, selected)
	if err != nil {
		return "", nil, err
	}
	if err = realDirs(m.DataDir, []string{"sites", name, "content"}); err != nil {
		return "", nil, err
	}
	if err = realDirs(m.DataDir, []string{"sites", name, sitefiles.ResolveAt(filepath.Join(m.DataDir, "sites", name), "itsy"), "runtime"}); err != nil {
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
		runtime := sitefiles.Path(filepath.Join(m.DataDir, "sites", name), "itsy/runtime")
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
	// The complete candidate remains private until config activation.
	candidateBytes, e := activeContent.Bytes()
	if e != nil {
		return "", nil, e
	}
	if err = atomicWrite(contentPath, candidateBytes, 0644); err != nil {
		return "", nil, err
	}
	if _, err = site.LoadWithMedia(dest, activeContent, site.ManagedMediaRoot(contentPath), site.ManagedRuntimeRoot(contentPath)); err != nil {
		return "", nil, fmt.Errorf("templates and preserved content: %w", err)
	}
	return contentPath, activeContent, nil
}
