package admin

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"flea/internal/host"
	"flea/internal/state"
)

// Upgrader swaps only the executable, not service policy or site data. The new
// executable must validate the complete installed configuration before mutation.
type Upgrader struct {
	Manager  Manager
	Binary   string
	Validate func(context.Context, string) error
}
type upgradeJournal struct {
	OldSHA string `json:"old_sha256"`
	NewSHA string `json:"new_sha256"`
}

func SystemUpgrader() Upgrader {
	return Upgrader{Manager: System(), Binary: "/usr/local/bin/flea-server", Validate: func(ctx context.Context, path string) error {
		output, err := exec.CommandContext(ctx, path, "-validate", "-config", "/etc/flea/server.json").CombinedOutput()
		if err != nil {
			return fmt.Errorf("new server cannot load installed sites: %w: %s", err, output)
		}
		return nil
	}}
}
func upgradePending(config string) string {
	return filepath.Join(filepath.Dir(config), "upgrade.pending.json")
}
func (u Upgrader) previous() string {
	return filepath.Join(filepath.Dir(u.Manager.ConfigPath), "server.binary.previous")
}
func binaryBytes(path string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > 100<<20 {
		return nil, fmt.Errorf("server executable is not a bounded regular file")
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	actual, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !os.SameFile(info, actual) {
		return nil, fmt.Errorf("server executable changed")
	}
	b, err := io.ReadAll(io.LimitReader(f, (100<<20)+1))
	if err != nil {
		return nil, err
	}
	if len(b) == 0 || len(b) > 100<<20 {
		return nil, fmt.Errorf("invalid executable size")
	}
	return b, nil
}
func binaryHash(b []byte) string { sum := sha256.Sum256(b); return hex.EncodeToString(sum[:]) }

func (u Upgrader) Upgrade(ctx context.Context, candidate, expected string) error {
	unlock, err := u.Manager.lock()
	if err != nil {
		return err
	}
	defer unlock()
	if _, err = os.Lstat(u.Manager.pending()); !os.IsNotExist(err) {
		return fmt.Errorf("deployment recovery pending")
	}
	next, err := binaryBytes(candidate)
	if err != nil {
		return err
	}
	if expected == "" || binaryHash(next) != expected {
		return fmt.Errorf("uploaded server checksum mismatch")
	}
	// Validate a private pinned copy, then install those exact bytes.
	pinned, err := os.CreateTemp(filepath.Dir(u.Manager.ConfigPath), ".upgrade-check-")
	if err != nil {
		return err
	}
	defer os.Remove(pinned.Name())
	defer pinned.Close()
	if err = pinned.Chmod(0700); err != nil {
		return err
	}
	if _, err = pinned.Write(next); err != nil {
		return err
	}
	if err = pinned.Close(); err != nil {
		return err
	}
	if err = u.Validate(ctx, pinned.Name()); err != nil {
		return err
	}
	old, err := binaryBytes(u.Binary)
	if err != nil {
		return err
	}
	c, _, err := host.Load(u.Manager.ConfigPath, nil)
	if err != nil {
		return err
	}
	if err = atomicWrite(u.previous(), old, 0700); err != nil {
		return err
	}
	journal, _ := json.Marshal(upgradeJournal{binaryHash(old), binaryHash(next)})
	if err = atomicWrite(upgradePending(u.Manager.ConfigPath), journal, 0600); err != nil {
		return err
	}
	if err = atomicWrite(u.Binary, next, 0755); err == nil {
		err = u.Manager.Restart(ctx)
		if err == nil {
			err = u.Manager.Verify(ctx, c, allDomains(c))
		}
	}
	if err != nil {
		recovery, stop := context.WithTimeout(context.Background(), 120*time.Second)
		defer stop()
		if restoreErr := u.recoverLocked(recovery); restoreErr != nil {
			return fmt.Errorf("upgrade failed: %w; recovery failed: %v; run flea update-server --recover", err, restoreErr)
		}
		return fmt.Errorf("upgrade failed; previous server restored: %w", err)
	}
	if err = os.Remove(upgradePending(u.Manager.ConfigPath)); err != nil {
		return fmt.Errorf("server upgraded but recovery journal retained: %w", err)
	}
	return nil
}
func allDomains(c host.Config) []string {
	domains := make([]string, 0, len(c.Sites))
	for domain := range c.Sites {
		domains = append(domains, domain)
	}
	return domains
}
func (u Upgrader) Recover(ctx context.Context) error {
	unlock, err := state.Lock(filepath.Join(filepath.Dir(u.Manager.ConfigPath), "deploy.lock"))
	if err != nil {
		return err
	}
	defer unlock()
	return u.recoverLocked(ctx)
}
func (u Upgrader) recoverLocked(ctx context.Context) error {
	if _, err := os.Lstat(u.Manager.pending()); !os.IsNotExist(err) {
		return fmt.Errorf("deployment recovery pending")
	}
	raw, err := os.ReadFile(upgradePending(u.Manager.ConfigPath))
	if err != nil {
		return err
	}
	var journal upgradeJournal
	if err = json.Unmarshal(raw, &journal); err != nil {
		return err
	}
	previous, err := binaryBytes(u.previous())
	if err != nil {
		return err
	}
	if binaryHash(previous) != journal.OldSHA {
		return fmt.Errorf("previous server checksum mismatch; recovery stopped")
	}
	current, err := binaryBytes(u.Binary)
	if err != nil {
		return err
	}
	hash := binaryHash(current)
	if hash != journal.OldSHA && hash != journal.NewSHA {
		return fmt.Errorf("server changed independently; recovery stopped")
	}
	if err = atomicWrite(u.Binary, previous, 0755); err != nil {
		return err
	}
	if err = u.Manager.Restart(ctx); err != nil {
		return err
	}
	c, _, err := host.Load(u.Manager.ConfigPath, nil)
	if err != nil {
		return err
	}
	if err = u.Manager.Verify(ctx, c, allDomains(c)); err != nil {
		return err
	}
	return os.Remove(upgradePending(u.Manager.ConfigPath))
}
