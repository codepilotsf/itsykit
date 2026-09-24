package admin

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"

	"flea/internal/fleas"
	"flea/internal/host"
	"flea/internal/site"
)

// prepareContent stages a new snapshot without modifying the active one. The
// returned store is the post-readiness version: removed IDs are archived only
// after activation succeeds. The caller holds the deployment lock throughout.
func (m Manager) prepareContent(c host.Config, name, id, dest string, project *site.Project) (string, *fleas.Store, error) {
	refs, e := project.FleaReferences()
	if e != nil {
		return "", nil, e
	}
	remote := fleas.NewStore()
	parent := filepath.Dir(dest)
	for domain, previous := range c.Sites {
		if filepath.Dir(previous) == parent {
			if file := c.Content[domain]; file != "" {
				remote, e = fleas.ReadFile(file)
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
	activeContent, err := fleas.Merge(project.Content, remote, refs)
	if err != nil {
		return "", nil, err
	}
	if err = realDirs(m.DataDir, []string{"sites", name, "content"}); err != nil {
		return "", nil, err
	}
	if err = realDirs(m.DataDir, []string{"sites", name, "_"}); err != nil {
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
		runtime := filepath.Join(m.DataDir, "sites", name, "_")
		if e = os.Chown(runtime, uid, gid); e != nil {
			return "", nil, e
		}
		if e = os.Chmod(runtime, 0700); e != nil {
			return "", nil, e
		}
	}
	contentPath := filepath.Join(m.DataDir, "sites", name, "content", id+".json")
	// Candidate is private until config activation. Keep orphans active until
	// readiness succeeds, then Apply writes the pruned store after verification.
	candidateBytes, e := activeContent.Bytes()
	if e != nil {
		return "", nil, e
	}
	candidateStore, e := fleas.Decode(candidateBytes)
	if e != nil {
		return "", nil, e
	}
	oldIndex, e := remote.Index()
	if e != nil {
		return "", nil, e
	}
	for key, entry := range oldIndex {
		if _, ok := refs[key]; !ok {
			candidateStore.Shared.Fleas[key] = entry
		}
	}
	candidateBytes, e = candidateStore.Bytes()
	if e != nil {
		return "", nil, e
	}
	if err = atomicWrite(contentPath, candidateBytes, 0644); err != nil {
		return "", nil, err
	}
	if _, err = site.LoadWithFleas(dest, activeContent); err != nil {
		return "", nil, fmt.Errorf("templates and preserved content: %w", err)
	}
	return contentPath, activeContent, nil
}
