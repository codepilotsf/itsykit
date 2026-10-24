package admin

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	"flea/internal/host"
)

// ValidSiteName reports whether name is a stable managed site identifier.
func ValidSiteName(name string) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,63}$`).MatchString(name)
}

// SiteSummary is a deterministic, read-only inventory entry for one managed
// site. State digests exactly the routing and ancillary entries that removal
// would change, so a stale review cannot remove a concurrently edited site.
type SiteSummary struct {
	Indexing  map[string]string `json:"indexing"`
	Warnings  []string          `json:"warnings,omitempty"`
	Name      string            `json:"name"`
	Branch    string            `json:"branch,omitempty"`
	Domains   []string          `json:"domains"`
	Release   string            `json:"release,omitempty"`
	Redirects []string          `json:"redirects,omitempty"`
	HTTPS     []string          `json:"https,omitempty"`
	Retained  string            `json:"retained,omitempty"`
	State     string            `json:"state"`
	Managed   bool              `json:"managed"`
}

// siteGroup is the routing currently owned by one release root.
type siteGroup struct {
	name    string
	root    string
	domains []string
	managed bool
}

// siteState is the canonical form the removal digest covers. Map keys marshal
// in sorted order, so the digest is stable across processes.
type siteState struct {
	Indexing  map[string]string `json:"indexing,omitempty"`
	Name      string            `json:"name"`
	Branch    string            `json:"branch,omitempty"`
	Commit    string            `json:"commit,omitempty"`
	Release   string            `json:"release,omitempty"`
	Domains   []string          `json:"domains"`
	Content   map[string]string `json:"content,omitempty"`
	Downloads map[string]string `json:"downloads,omitempty"`
	Redirects map[string]string `json:"redirects,omitempty"`
	TLS       []string          `json:"tls,omitempty"`
}

func (s siteState) digest() string {
	b, _ := json.Marshal(s)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

// Sites lists every routing group on the host. It never mutates state and does
// not take the deployment lock, so it can inspect a server with a pending
// recovery journal.
func (m Manager) Sites() ([]SiteSummary, error) {
	c, _, err := host.Load(m.ConfigPath, nil)
	if err != nil {
		return nil, err
	}
	return siteSummaries(c, m.DataDir), nil
}

// SiteInfo returns the reviewed state for one managed site.
func (m Manager) SiteInfo(name string) (SiteSummary, error) {
	if !ValidSiteName(name) {
		return SiteSummary{}, fmt.Errorf("invalid site name %q", name)
	}
	c, _, err := host.Load(m.ConfigPath, nil)
	if err != nil {
		return SiteSummary{}, err
	}
	g, err := findSite(c, m.DataDir, name)
	if err != nil {
		return SiteSummary{}, err
	}
	return summarize(c, m.DataDir, g), nil
}

// RemoveSite unregisters one site's routing and ancillary host entries after
// checking the reviewed digest. It reuses the deployment lock, recovery journal,
// restart/probe and rollback paths. Files, releases, content, media, sessions and
// certificates are never deleted.
func (m Manager) RemoveSite(ctx context.Context, name, expected string) error {
	if !ValidSiteName(name) {
		return fmt.Errorf("invalid site name %q", name)
	}
	if expected == "" {
		return fmt.Errorf("missing reviewed state; review the site again")
	}
	unlock, err := m.lock()
	if err != nil {
		return err
	}
	defer unlock()
	if _, err = os.Lstat(m.pending()); !os.IsNotExist(err) {
		return fmt.Errorf("unfinished deployment or removal; run sudo flea-server -recover first")
	}
	old, err := os.ReadFile(m.ConfigPath)
	if err != nil {
		return err
	}
	c, _, err := host.Load(m.ConfigPath, nil)
	if err != nil {
		return err
	}
	g, err := findSite(c, m.DataDir, name)
	if err != nil {
		return err
	}
	if stateFor(c, g).digest() != expected {
		return fmt.Errorf("site %s changed since you reviewed it; review and confirm again", name)
	}
	if err = removeSiteEntries(&c, g); err != nil {
		return err
	}
	return m.publishHostConfig(ctx, c, old, fmt.Sprintf("Removing site %s from routing (files retained)...", name))
}

// publishHostConfig requires the deployment lock and a checked recovery journal.
// Both removal and indexing changes use the same validated, recoverable transaction.
func (m Manager) publishHostConfig(ctx context.Context, c host.Config, old []byte, progress string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	candidate, err := os.CreateTemp(filepath.Dir(m.ConfigPath), ".candidate-")
	if err != nil {
		return err
	}
	defer os.Remove(candidate.Name())
	if _, err = candidate.Write(data); err != nil {
		candidate.Close()
		return err
	}
	if err = candidate.Close(); err != nil {
		return err
	}
	if _, _, err = host.Load(candidate.Name(), nil); err != nil {
		return fmt.Errorf("candidate config: %w", err)
	}
	// Persist recovery before replacing config. A killed SSH command is recoverable.
	if err = atomicWrite(m.pending(), old, 0600); err != nil {
		return err
	}
	if m.Progress != nil {
		fmt.Fprintln(m.Progress, progress)
	}
	if err = atomicWrite(m.ConfigPath, data, 0644); err == nil {
		err = m.Restart(ctx)
		if err == nil {
			err = m.Verify(ctx, c, remainingDomains(c))
		}
	}
	if err != nil {
		recovery, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()
		if restoreErr := m.restore(recovery); restoreErr != nil {
			return fmt.Errorf("host configuration change failed: %w; recovery also failed: %v; run sudo flea-server -recover", err, restoreErr)
		}
		return fmt.Errorf("host configuration change failed; previous configuration restored: %w", err)
	}
	if err = atomicWrite(filepath.Join(filepath.Dir(m.ConfigPath), "server.previous.json"), old, 0600); err != nil {
		return fmt.Errorf("host configuration active but backup failed; recovery journal retained: %w", err)
	}
	if err = os.Remove(m.pending()); err != nil {
		return err
	}
	return nil
}

// removeSiteEntries mutates only the target site's routing and target-owned
// ancillary entries. Shared values and other sites are preserved. A redirect
// that touches the target at only one end is ambiguous and fails closed.
func removeSiteEntries(c *host.Config, g siteGroup) error {
	owned := map[string]bool{}
	for _, d := range g.domains {
		owned[d] = true
	}
	for _, d := range g.domains {
		delete(c.Sites, d)
		delete(c.Content, d)
		delete(c.Downloads, d)
		delete(c.Indexing, d)
	}
	for from, to := range c.Redirects {
		if !owned[from] && !owned[to] {
			continue
		}
		if owned[from] && (to == "" || owned[to]) {
			delete(c.Redirects, from)
			continue
		}
		return fmt.Errorf("redirect %s -> %s crosses site ownership; refusing to guess", from, to)
	}
	if c.TLS != nil {
		remaining := c.TLS.Domains[:0]
		for _, d := range c.TLS.Domains {
			if !owned[d] {
				remaining = append(remaining, d)
			}
		}
		c.TLS.Domains = remaining
		if len(remaining) == 0 {
			c.TLS = nil
		}
	}
	delete(c.SiteBranches, g.name)
	delete(c.Commits, g.name)
	return nil
}

func remainingDomains(c host.Config) []string {
	domains := make([]string, 0, len(c.Sites))
	for domain := range c.Sites {
		domains = append(domains, domain)
	}
	sort.Strings(domains)
	return domains
}

func siteSummaries(c host.Config, dataDir string) []SiteSummary {
	groups := groupSites(c, dataDir)
	out := make([]SiteSummary, 0, len(groups))
	for _, g := range groups {
		out = append(out, summarize(c, dataDir, g))
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Name != out[j].Name {
			return out[i].Name < out[j].Name
		}
		return out[i].State < out[j].State
	})
	return out
}

func summarize(c host.Config, dataDir string, g siteGroup) SiteSummary {
	st := stateFor(c, g)
	s := SiteSummary{
		Name:      g.name,
		Branch:    c.SiteBranches[g.name],
		Domains:   append([]string{}, st.Domains...),
		Release:   st.Release,
		State:     st.digest(),
		Managed:   g.managed,
		HTTPS:     append([]string{}, st.TLS...),
		Redirects: make([]string, 0, len(st.Redirects)),
	}
	s.Indexing = map[string]string{}
	for _, domain := range s.Domains {
		mode := c.Indexing[domain]
		if mode == "" {
			mode = "allow"
		}
		s.Indexing[domain] = mode
		for _, warning := range c.IndexingWarnings(domain) {
			s.Warnings = append(s.Warnings, domain+": "+warning)
		}
	}
	if g.managed {
		s.Retained = filepath.Join(dataDir, "sites", g.name)
	}
	for from, to := range st.Redirects {
		s.Redirects = append(s.Redirects, from+" -> "+to)
	}
	sort.Strings(s.Redirects)
	if len(s.Redirects) == 0 {
		s.Redirects = nil
	}
	return s
}

func stateFor(c host.Config, g siteGroup) siteState {
	domains := append([]string{}, g.domains...)
	sort.Strings(domains)
	owned := map[string]bool{}
	for _, d := range domains {
		owned[d] = true
	}
	st := siteState{Name: g.name, Branch: c.SiteBranches[g.name], Commit: c.Commits[g.name], Domains: domains}
	if g.root != "" {
		st.Release = filepath.Base(g.root)
	}
	for _, d := range domains {
		if v := c.Indexing[d]; v != "" {
			if st.Indexing == nil {
				st.Indexing = map[string]string{}
			}
			st.Indexing[d] = v
		}
		if v := c.Content[d]; v != "" {
			if st.Content == nil {
				st.Content = map[string]string{}
			}
			st.Content[d] = v
		}
		if v := c.Downloads[d]; v != "" {
			if st.Downloads == nil {
				st.Downloads = map[string]string{}
			}
			st.Downloads[d] = v
		}
	}
	for from, to := range c.Redirects {
		if !owned[from] && !owned[to] {
			continue
		}
		if st.Redirects == nil {
			st.Redirects = map[string]string{}
		}
		st.Redirects[from] = to
	}
	if c.TLS != nil {
		for _, d := range c.TLS.Domains {
			if owned[d] {
				st.TLS = append(st.TLS, d)
			}
		}
		sort.Strings(st.TLS)
	}
	return st
}

// groupSites groups hostnames by release root, which is how aliases are defined.
func groupSites(c host.Config, dataDir string) []siteGroup {
	index := map[string]int{}
	var groups []siteGroup
	for domain, root := range c.Sites {
		root = filepath.Clean(root)
		i, ok := index[root]
		if !ok {
			name, managed := managedSiteName(dataDir, root)
			groups = append(groups, siteGroup{name: name, root: root, managed: managed})
			i = len(groups) - 1
			index[root] = i
		}
		groups[i].domains = append(groups[i].domains, domain)
	}
	for i := range groups {
		sort.Strings(groups[i].domains)
	}
	sort.Slice(groups, func(i, j int) bool {
		if groups[i].name != groups[j].name {
			return groups[i].name < groups[j].name
		}
		return groups[i].root < groups[j].root
	})
	return groups
}

func managedSiteName(dataDir, root string) (string, bool) {
	releases := filepath.Dir(root)
	name := filepath.Base(filepath.Dir(releases))
	if ValidSiteName(name) && releases == filepath.Join(dataDir, "sites", name, "releases") {
		return name, true
	}
	if !ValidSiteName(name) {
		name = filepath.Base(root)
	}
	return name, false
}

func findSite(c host.Config, dataDir, name string) (siteGroup, error) {
	var matches []siteGroup
	for _, g := range groupSites(c, dataDir) {
		if g.name == name {
			matches = append(matches, g)
		}
	}
	switch {
	case len(matches) == 0:
		return siteGroup{}, fmt.Errorf("no site named %q is registered on this server", name)
	case len(matches) > 1:
		return siteGroup{}, fmt.Errorf("site name %q is ambiguous; refusing to guess", name)
	case !matches[0].managed:
		return siteGroup{}, fmt.Errorf("site %q is not managed by this server; refusing to change it", name)
	}
	return matches[0], nil
}
