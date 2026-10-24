package admin

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"

	"flea/internal/host"
)

// SetIndexing changes only the named registered host. Omitted settings on code
// deployment never call this operation. The reviewed site digest includes all
// aliases and their policies, preventing stale removal or policy confirmation.
func (m Manager) SetIndexing(ctx context.Context, domain, mode, expected string) error {
	if !host.ValidHost(domain) || !host.ValidIndexingMode(mode) {
		return fmt.Errorf("indexing requires an exact hostname and mode noindex or allow")
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
		return fmt.Errorf("unfinished host configuration change; run sudo flea-server -recover first")
	}
	old, err := os.ReadFile(m.ConfigPath)
	if err != nil {
		return err
	}
	c, _, err := host.Load(m.ConfigPath, nil)
	if err != nil {
		return err
	}
	for _, g := range groupSites(c, m.DataDir) {
		for _, d := range g.domains {
			if d != domain {
				continue
			}
			if !g.managed {
				return fmt.Errorf("hostname %s is not managed by this server", domain)
			}
			if stateFor(c, g).digest() != expected {
				return fmt.Errorf("site %s changed since you reviewed it; review and confirm again", g.name)
			}
			if c.Indexing == nil {
				c.Indexing = map[string]string{}
			}
			c.Indexing[domain] = mode
			// Readiness must distinguish the restarted process from the prior
			// process even though this operation leaves every release unchanged.
			revision := make([]byte, 12)
			if _, err := rand.Read(revision); err != nil {
				return err
			}
			c.Revision = "indexing-" + hex.EncodeToString(revision)
			return m.publishHostConfig(ctx, c, old, "Updating indexing policy and checking site health...")
		}
	}
	return fmt.Errorf("hostname %s is not registered on this server", domain)
}
