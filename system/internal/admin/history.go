package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"itsy/internal/host"
	"itsy/internal/release"
	"itsy/internal/site"
)

var releaseID = regexp.MustCompile(`^[0-9]{8}-[0-9]{6}-[0-9a-f]{24}$`)

type ReleaseRecord struct {
	ID       string `json:"id"`
	Previous string `json:"previous,omitempty"`
	Branch   string `json:"branch"`
	Commit   string `json:"commit"`
}

func (m Manager) writeReceipt(name string, record ReleaseRecord) error {
	if err := realDirs(m.DataDir, []string{"sites", name, "history"}); err != nil {
		return err
	}
	b, err := json.Marshal(record)
	if err != nil {
		return err
	}
	return atomicWrite(filepath.Join(m.DataDir, "sites", name, "history", record.ID+".json"), b, 0600)
}

func (m Manager) history(c host.Config, name, branch string) ([]ReleaseRecord, error) {
	if !regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,63}$`).MatchString(name) || c.SiteBranches[name] != branch {
		return nil, fmt.Errorf("site does not belong to this branch")
	}
	parent := filepath.Join(m.DataDir, "sites", name, "releases")
	current := ""
	for _, root := range c.Sites {
		if filepath.Dir(root) == parent {
			current = filepath.Base(root)
			break
		}
	}
	if current == "" {
		return nil, fmt.Errorf("this site has not been deployed")
	}
	var records []ReleaseRecord
	seen := map[string]bool{}
	for current != "" {
		if len(records) >= 1000 {
			break
		}
		if !releaseID.MatchString(current) || seen[current] {
			return nil, fmt.Errorf("invalid or oversized release history")
		}
		seen[current] = true
		b, err := site.ReadManagedFile(filepath.Join(m.DataDir, "sites", name, "history"), current+".json")
		if err != nil {
			return nil, fmt.Errorf("release history unavailable: %w", err)
		}
		var record ReleaseRecord
		if err = json.Unmarshal(b, &record); err != nil {
			return nil, err
		}
		if record.ID != current || record.Branch != branch {
			return nil, fmt.Errorf("release history identity mismatch")
		}
		records = append(records, record)
		current = record.Previous
	}
	return records, nil
}

func (m Manager) History(name, branch string) ([]ReleaseRecord, error) {
	unlock, err := m.lock()
	if err != nil {
		return nil, err
	}
	defer unlock()
	if _, err = os.Lstat(m.pending()); !os.IsNotExist(err) {
		return nil, fmt.Errorf("deployment recovery pending")
	}
	c, _, err := host.Load(m.ConfigPath, nil)
	if err != nil {
		return nil, err
	}
	return m.history(c, name, branch)
}

// Rollback copies retained code into a new release while preserving live state.
// Target must be reachable from the current successful activation chain.
func (m Manager) Rollback(ctx context.Context, name, branch, target, expected string) (string, error) {
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
	records, err := m.history(c, name, branch)
	if err != nil {
		return "", err
	}
	if records[0].ID != expected {
		return "", fmt.Errorf("active release changed; review rollback again")
	}
	var selected *ReleaseRecord
	for i := 1; i < len(records); i++ {
		if records[i].ID == target {
			selected = &records[i]
			break
		}
	}
	if selected == nil {
		return "", fmt.Errorf("choose a previous release from this site's history")
	}
	parent := filepath.Join(m.DataDir, "sites", name, "releases")
	info, err := os.Lstat(filepath.Join(parent, target))
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("release is not a real directory")
	}
	domains := []string{}
	secure := true
	for domain, root := range c.Sites {
		if filepath.Dir(root) == parent {
			domains = append(domains, domain)
			found := false
			if c.TLS != nil {
				for _, d := range c.TLS.Domains {
					if d == domain {
						found = true
					}
				}
			}
			secure = secure && found
		}
	}
	sort.Strings(domains)
	stage, err := os.CreateTemp("", "flea-rollback-*.tar.gz")
	if err != nil {
		return "", err
	}
	defer os.Remove(stage.Name())
	defer stage.Close()
	if err = release.Pack(filepath.Join(parent, target), stage); err != nil {
		return "", err
	}
	if err = stage.Close(); err != nil {
		return "", err
	}
	o := Options{Archive: stage.Name(), Name: name, Branch: branch, Commit: selected.Commit, Domains: domains, HTTPS: secure, keepRouting: true}
	if c.TLS != nil {
		o.Email = c.TLS.Email
		o.AgreeTOS = c.TLS.AgreeTOS
	}
	if err = o.Validate(); err != nil {
		return "", err
	}
	return m.applyLocked(ctx, o)
}
