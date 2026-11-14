package admin

import (
	"context"
	"fmt"
	"io"
	"itsy/internal/sitefiles"
	"os"
	"path/filepath"
	"sort"

	"itsy/internal/backup"
	"itsy/internal/host"
	"itsy/internal/release"
	"itsy/internal/site"
	"itsy/internal/syncstate"
)

func copyRelease(root, dest string) error {
	info, err := os.Lstat(root)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("release is not a real directory")
	}
	f, err := os.CreateTemp("", "flea-release-copy-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	defer f.Close()
	if err = release.Pack(root, f); err != nil {
		return err
	}
	if _, err = f.Seek(0, io.SeekStart); err != nil {
		return err
	}
	return release.Extract(f, dest)
}

func (m Manager) Backup(name, branch string, out io.Writer) error {
	unlock, err := m.lock()
	if err != nil {
		return err
	}
	defer unlock()
	if _, err = os.Lstat(m.pending()); !os.IsNotExist(err) {
		return fmt.Errorf("deployment recovery pending")
	}
	c, _, err := host.Load(m.ConfigPath, nil)
	if err != nil {
		return err
	}
	live, err := m.snapshot(c, name, branch)
	if err != nil {
		return err
	}
	if !live.Exists {
		return fmt.Errorf("this site has not been deployed")
	}
	if c.SiteBranches[name] != branch || branch == "" {
		return fmt.Errorf("site does not belong to this branch")
	}
	stage, err := os.MkdirTemp("", "flea-backup-site-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)
	if err = copyRelease(live.Revision, stage); err != nil {
		return err
	}
	if err = site.SaveContent(stage, live.Snips); err != nil {
		return err
	}
	if len(live.Snips.ImageAssets()) > 0 {
		if err = os.MkdirAll(sitefiles.Path(stage, filepath.Join("itsy", "media")), 0755); err != nil {
			return err
		}
		for _, photo := range live.Snips.ImageAssets() {
			data, e := site.ReadPhoto(filepath.Join(m.DataDir, "sites", name, "content", "media"), photo.File)
			if e != nil {
				return e
			}
			if e = site.WriteManagedFile(stage, "itsy/media/"+photo.File, data); e != nil {
				return e
			}
		}
	}
	if _, err = site.Load(stage); err != nil {
		return err
	}
	return backup.Write(stage, backup.Manifest{Name: name, Branch: branch, Commit: c.Commits[name], StateVersion: live.Version()}, out)
}

func recoveryOptions(c host.Config, dataDir, name, branch, commit, archive string) Options {
	parent := filepath.Join(dataDir, "sites", name, "releases")
	o := Options{Name: name, Branch: branch, Commit: commit, Archive: archive, HTTPS: true, keepRouting: true}
	for domain, root := range c.Sites {
		if filepath.Dir(root) == parent {
			o.Domains = append(o.Domains, domain)
			found := false
			if c.TLS != nil {
				for _, d := range c.TLS.Domains {
					if d == domain {
						found = true
					}
				}
			}
			o.HTTPS = o.HTTPS && found
		}
	}
	sort.Strings(o.Domains)
	if c.TLS != nil {
		o.Email = c.TLS.Email
		o.AgreeTOS = c.TLS.AgreeTOS
	}
	return o
}

// Restore intentionally restores content as well as code, only after review of
// the current state. Host routing, certificates, credentials and sessions remain.
func (m Manager) Restore(ctx context.Context, name, branch, file, expected string) (string, error) {
	stage, err := os.MkdirTemp("", "flea-restore-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(stage)
	input, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer input.Close()
	manifest, err := backup.Extract(input, stage)
	if err != nil {
		return "", err
	}
	if manifest.Name != name || manifest.Branch != branch {
		return "", fmt.Errorf("backup belongs to another site or branch")
	}
	p, err := site.Load(stage)
	if err != nil {
		return "", err
	}
	archive, err := os.CreateTemp("", "flea-restore-release-*")
	if err != nil {
		return "", err
	}
	defer os.Remove(archive.Name())
	defer archive.Close()
	if err = release.Pack(stage, archive); err != nil {
		return "", err
	}
	if err = archive.Close(); err != nil {
		return "", err
	}
	unlock, err := m.lock()
	if err != nil {
		return "", err
	}
	defer unlock()
	if _, err = os.Lstat(m.pending()); !os.IsNotExist(err) {
		return "", fmt.Errorf("deployment recovery pending")
	}
	c, _, err := host.Load(m.ConfigPath, nil)
	if err != nil {
		return "", err
	}
	live, err := m.snapshot(c, name, branch)
	if err != nil {
		return "", err
	}
	if !live.Exists || live.Version() != expected {
		return "", fmt.Errorf("live state changed; review restore again")
	}
	o := recoveryOptions(c, m.DataDir, name, branch, manifest.Commit, archive.Name())
	o.Sync = &syncstate.Plan{Expected: expected, Snips: p.Content}
	if err = o.Validate(); err != nil {
		return "", err
	}
	return m.applyLocked(ctx, o)
}
