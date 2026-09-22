// Package admin implements privileged, serialized release activation. It is
// invoked through SSH/sudo, never exposed as an HTTP endpoint.
package admin

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"nanoweb/internal/host"
	"nanoweb/internal/release"
	"nanoweb/internal/site"
	"nanoweb/internal/snips"
	"nanoweb/internal/state"
)

type Options struct {
	Archive   string
	Name      string
	Domains   []string
	HTTPS     bool
	Email     string
	AgreeTOS  bool
	Canonical string
	Branch    string
	Commit    string
}

func (o Options) Validate() error {
	if o.Branch != "" && !regexp.MustCompile(`^[0-9a-f]{40}([0-9a-f]{24})?$`).MatchString(o.Commit) {
		return fmt.Errorf("deployment requires a full Git commit ID")
	}
	if o.Archive != "" && !regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,63}$`).MatchString(o.Name) {
		return fmt.Errorf("site name must be 1-64 letters, digits, hyphens or underscores")
	}
	if len(o.Domains) == 0 {
		return fmt.Errorf("at least one domain or IPv4 address is required")
	}
	seen := map[string]bool{}
	for _, domain := range o.Domains {
		if !host.ValidHost(domain) || seen[domain] {
			return fmt.Errorf("invalid or duplicate hostname %q", domain)
		}
		if o.HTTPS && !host.ValidDomain(domain) {
			return fmt.Errorf("HTTPS requires a DNS domain, not an IP or wildcard: %s", domain)
		}
		seen[domain] = true
	}
	if o.HTTPS && (!o.AgreeTOS || o.Email == "") {
		return fmt.Errorf("HTTPS requires -email and explicit -agree-tos (https://letsencrypt.org/repository/)")
	}
	if o.HTTPS {
		address, err := mail.ParseAddress(o.Email)
		if err != nil || address.Address != o.Email {
			return fmt.Errorf("HTTPS requires a valid contact email")
		}
	}
	if o.Archive == "" && !o.HTTPS {
		return fmt.Errorf("an archive or HTTPS action is required")
	}
	if o.Canonical != "" {
		if len(o.Domains) != 2 || o.Domains[0] != o.Canonical || o.Domains[1] != "www."+o.Canonical {
			return fmt.Errorf("canonical redirect must pair the primary domain with its www alias")
		}
	}
	return nil
}

type Manager struct {
	ConfigPath string
	DataDir    string
	Restart    func(context.Context) error
	Verify     func(context.Context, host.Config, []string) error
}

func System() Manager {
	return Manager{ConfigPath: "/etc/nanoweb/server.json", DataDir: "/var/lib/nanoweb", Restart: func(ctx context.Context) error {
		output, err := exec.CommandContext(ctx, "systemctl", "restart", "nanoweb.service").CombinedOutput()
		if err != nil {
			return fmt.Errorf("restart: %w: %s", err, output)
		}
		return nil
	}, Verify: verify}
}

func verify(ctx context.Context, c host.Config, domains []string) error {
	wait, stop := context.WithTimeout(ctx, 120*time.Second)
	defer stop()
	if len(domains) == 0 {
		domains = []string{""}
	}
	for _, domain := range domains {
		secure := false
		if c.TLS != nil {
			for _, name := range c.TLS.Domains {
				if domain == name {
					secure = true
				}
			}
		}
		for {
			err := host.Probe(wait, c, domain, secure)
			if err == nil {
				break
			}
			select {
			case <-wait.Done():
				return fmt.Errorf("readiness/certificate for %s: %w (%v)", domain, wait.Err(), err)
			case <-time.After(time.Second):
			}
		}
	}
	return nil
}

func (m Manager) lock() (func(), error) {
	return state.Lock(filepath.Join(filepath.Dir(m.ConfigPath), "deploy.lock"))
}

func (m Manager) pending() string {
	return filepath.Join(filepath.Dir(m.ConfigPath), "deploy.pending.json")
}

func atomicWrite(name string, data []byte, mode os.FileMode) error {
	f, err := os.CreateTemp(filepath.Dir(name), ".nanoweb-")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	defer f.Close()
	if err = f.Chmod(mode); err != nil {
		return err
	}
	if _, err = f.Write(data); err != nil {
		return err
	}
	if err = f.Sync(); err != nil {
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	if err = os.Rename(f.Name(), name); err != nil {
		return err
	}
	dir, err := os.Open(filepath.Dir(name))
	if err != nil {
		return err
	}
	defer dir.Close()
	return dir.Sync()
}

func (m Manager) Apply(ctx context.Context, o Options) (string, error) {
	if err := o.Validate(); err != nil {
		return "", err
	}
	unlock, err := m.lock()
	if err != nil {
		return "", err
	}
	defer unlock()
	if _, err := os.Lstat(m.pending()); !os.IsNotExist(err) {
		return "", fmt.Errorf("unfinished deployment; run sudo nanoweb-server -recover before retrying")
	}
	old, err := os.ReadFile(m.ConfigPath)
	if err != nil {
		return "", err
	}
	c, _, err := host.Load(m.ConfigPath, nil)
	if err != nil {
		return "", err
	}
	if c.Sites == nil {
		c.Sites = map[string]string{}
	}
	if c.SiteBranches == nil {
		c.SiteBranches = map[string]string{}
	}
	if c.Commits == nil {
		c.Commits = map[string]string{}
	}
	if c.Redirects == nil {
		c.Redirects = map[string]string{}
	}
	if previous := c.SiteBranches[o.Name]; previous != "" && previous != o.Branch {
		return "", fmt.Errorf("site %s belongs to branch %s, not %s", o.Name, previous, o.Branch)
	}
	// Also enforce branch ownership for HTTPS-only changes, which carry no archive.
	for _, domain := range o.Domains {
		if root := c.Sites[domain]; root != "" {
			name := filepath.Base(filepath.Dir(filepath.Dir(root)))
			if owner := c.SiteBranches[name]; owner != "" && owner != o.Branch {
				return "", fmt.Errorf("hostname %s belongs to branch %s", domain, owner)
			}
		}
	}
	idBytes := make([]byte, 12)
	if _, err = rand.Read(idBytes); err != nil {
		return "", err
	}
	id := time.Now().UTC().Format("20060102-150405") + "-" + hex.EncodeToString(idBytes)
	var contentPath string
	var activeContent *snips.Store
	if o.Archive != "" {
		parent := filepath.Join(m.DataDir, "sites", o.Name, "releases")
		for _, domain := range o.Domains {
			if existing := c.Sites[domain]; existing != "" && filepath.Dir(existing) != parent {
				return "", fmt.Errorf("hostname %s belongs to a different site; refusing to replace it", domain)
			}
		}
		// The data tree is administrator-owned. Never follow a preexisting symlink
		// while creating a release, even if an old installation had writable data.
		if err = realDirs(m.DataDir, []string{"sites", o.Name, "releases"}); err != nil {
			return "", err
		}
		dest := filepath.Join(parent, id)
		if err = os.Mkdir(dest, 0755); err != nil {
			return "", err
		}
		f, err := os.Open(o.Archive)
		if err != nil {
			return "", err
		}
		err = release.Extract(f, dest)
		f.Close()
		if err != nil {
			return "", fmt.Errorf("extract release (retained at %s): %w", dest, err)
		}
		project, loadErr := site.Load(dest)
		if loadErr != nil {
			return "", fmt.Errorf("invalid release %s: %w", dest, loadErr)
		}
		if err = project.CheckExtractedSnips(); err != nil {
			return "", err
		}
		if len(project.Admin.Emails) > 0 && !o.HTTPS {
			return "", fmt.Errorf("sites with admin access require HTTPS")
		}
		contentPath, activeContent, err = m.prepareContent(c, o.Name, id, dest, project)
		if err != nil {
			return "", err
		}
		if c.Content == nil {
			c.Content = map[string]string{}
		}
		// All existing aliases for this site move together; unrelated sites survive.
		for domain, root := range c.Sites {
			if filepath.Dir(root) == parent {
				c.Sites[domain] = dest
				c.Content[domain] = contentPath
			}
		}
		for _, domain := range o.Domains {
			c.Sites[domain] = dest
			c.Content[domain] = contentPath
		}
		if o.Branch != "" {
			c.SiteBranches[o.Name] = o.Branch
			c.Commits[o.Name] = o.Commit
		}
	}
	if o.Canonical != "" {
		c.Redirects["www."+o.Canonical] = o.Canonical
	} else {
		for _, domain := range o.Domains {
			delete(c.Redirects, domain)
			delete(c.Redirects, "www."+domain)
		}
	}
	if o.HTTPS {
		if c.TLS == nil {
			c.TLS = &host.TLSConfig{Listen: ":443", CacheDir: filepath.Join(m.DataDir, "certs")}
		}
		c.TLS.Email = o.Email
		c.TLS.AgreeTOS = o.AgreeTOS
		for _, domain := range o.Domains {
			if c.Sites[domain] == "" {
				return "", fmt.Errorf("deploy %s before enabling HTTPS", domain)
			}
			found := false
			for _, existing := range c.TLS.Domains {
				if existing == domain {
					found = true
				}
			}
			if !found {
				c.TLS.Domains = append(c.TLS.Domains, domain)
			}
		}
	}
	if !o.HTTPS && o.Archive != "" && c.TLS != nil {
		remaining := c.TLS.Domains[:0]
		for _, domain := range c.TLS.Domains {
			if filepath.Dir(c.Sites[domain]) != filepath.Join(m.DataDir, "sites", o.Name, "releases") {
				remaining = append(remaining, domain)
			}
		}
		c.TLS.Domains = remaining
		if len(remaining) == 0 {
			c.TLS = nil
		}
	}
	c.Revision = id
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return "", err
	}
	data = append(data, '\n')
	candidate, err := os.CreateTemp(filepath.Dir(m.ConfigPath), ".candidate-")
	if err != nil {
		return "", err
	}
	defer os.Remove(candidate.Name())
	if _, err = candidate.Write(data); err != nil {
		candidate.Close()
		return "", err
	}
	candidate.Close()
	if _, _, err = host.Load(candidate.Name(), nil); err != nil {
		return "", fmt.Errorf("candidate config: %w", err)
	}
	// Persist recovery before replacing config. A killed SSH command is recoverable.
	if err = atomicWrite(m.pending(), old, 0600); err != nil {
		return "", err
	}
	if err = atomicWrite(m.ConfigPath, data, 0644); err == nil {
		err = m.Restart(ctx)
		if err == nil {
			err = m.Verify(ctx, c, o.Domains)
		}
	}
	if err != nil {
		recovery, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()
		if restoreErr := m.restore(recovery); restoreErr != nil {
			return "", fmt.Errorf("activation failed: %w; recovery also failed: %v; run sudo nanoweb-server -recover", err, restoreErr)
		}
		return "", fmt.Errorf("activation failed; previous configuration restored: %w", err)
	}
	// Keep the previous configuration and every release for manual recovery.
	if activeContent != nil {
		data, e := activeContent.Bytes()
		if e == nil {
			e = atomicWrite(contentPath, data, 0644)
		}
		if e != nil {
			return "", fmt.Errorf("release active, but snip cleanup failed; recovery journal retained: %w", e)
		}
	}
	if err = atomicWrite(filepath.Join(filepath.Dir(m.ConfigPath), "server.previous.json"), old, 0600); err != nil {
		return "", fmt.Errorf("release active but backup failed; recovery journal retained: %w", err)
	}
	if err = os.Remove(m.pending()); err != nil {
		return "", err
	}
	return id, nil
}

func realDirs(base string, parts []string) error {
	if info, err := os.Lstat(base); err != nil || !info.IsDir() {
		return fmt.Errorf("data directory is missing or not a real directory: %s", base)
	}
	for _, part := range parts {
		base = filepath.Join(base, part)
		if err := os.Mkdir(base, 0755); err != nil && !errors.Is(err, os.ErrExist) {
			return err
		}
		info, err := os.Lstat(base)
		if err != nil || !info.IsDir() {
			return fmt.Errorf("not a real directory: %s", base)
		}
	}
	return nil
}

func (m Manager) Recover(ctx context.Context) error {
	unlock, err := m.lock()
	if err != nil {
		return err
	}
	defer unlock()
	return m.restore(ctx)
}

func (m Manager) restore(ctx context.Context) error {
	old, err := os.ReadFile(m.pending())
	if err != nil {
		return err
	}
	if _, _, err = host.Load(m.pending(), nil); err != nil {
		return fmt.Errorf("invalid recovery config: %w", err)
	}
	if err = atomicWrite(m.ConfigPath, old, 0644); err != nil {
		return err
	}
	if err = m.Restart(ctx); err != nil {
		return err
	}
	var c host.Config
	if err = json.Unmarshal(old, &c); err != nil {
		return err
	}
	if err = m.Verify(ctx, c, nil); err != nil {
		return err
	}
	return os.Remove(m.pending())
}

func Domains(value string) []string {
	parts := strings.Split(value, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(strings.ToLower(parts[i]))
	}
	return parts
}
