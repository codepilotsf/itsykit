package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"flea/internal/editor"
	"flea/internal/editstore"
	"flea/internal/host"
)

func qaUpgrader(t *testing.T) (Upgrader, string) {
	t.Helper()
	m, _ := fixture(t)
	binary := filepath.Join(filepath.Dir(m.ConfigPath), "flea-server")
	candidate := filepath.Join(t.TempDir(), "candidate")
	if e := os.WriteFile(binary, []byte("old executable"), 0755); e != nil {
		t.Fatal(e)
	}
	if e := os.WriteFile(candidate, []byte("new executable"), 0755); e != nil {
		t.Fatal(e)
	}
	return Upgrader{Manager: m, Binary: binary, Validate: func(context.Context, string) error { return nil }}, candidate
}
func TestQAUpgradeValidatesBeforeMutation(t *testing.T) {
	for _, kind := range []string{"checksum", "candidate validation", "installed config"} {
		t.Run(kind, func(t *testing.T) {
			u, candidate := qaUpgrader(t)
			before, _ := os.ReadFile(u.Binary)
			sum := binaryHash([]byte("new executable"))
			switch kind {
			case "checksum":
				sum = "bad"
			case "candidate validation":
				u.Validate = func(context.Context, string) error { return errors.New("invalid sites") }
			case "installed config":
				if e := os.WriteFile(u.Manager.ConfigPath, []byte("invalid"), 0644); e != nil {
					t.Fatal(e)
				}
			}
			if e := u.Upgrade(context.Background(), candidate, sum); e == nil {
				t.Fatal("accepted invalid upgrade")
			}
			after, e := os.ReadFile(u.Binary)
			if e != nil || !bytes.Equal(before, after) {
				t.Fatal("binary mutated", e)
			}
			for _, p := range []string{u.previous(), upgradePending(u.Manager.ConfigPath)} {
				if _, e := os.Stat(p); !os.IsNotExist(e) {
					t.Fatal("created recovery state before validation", p, e)
				}
			}
		})
	}
}
func TestQAUpgradePinsBytesPreservesConfigAndRecoversProbe(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{false: "success", true: "failed probe"}[fail], func(t *testing.T) {
			u, candidate := qaUpgrader(t)
			config, _ := os.ReadFile(u.Manager.ConfigPath)
			calls := 0
			u.Validate = func(_ context.Context, pinned string) error {
				b, e := os.ReadFile(pinned)
				if e != nil || string(b) != "new executable" || pinned == candidate {
					t.Fatal("not pinned", e)
				}
				return os.WriteFile(candidate, []byte("changed independently"), 0755)
			}
			u.Manager.Verify = func(context.Context, host.Config, []string) error {
				calls++
				if fail && calls == 1 {
					return errors.New("failed readiness")
				}
				return nil
			}
			e := u.Upgrade(context.Background(), candidate, binaryHash([]byte("new executable")))
			if (e != nil) != fail {
				t.Fatal(e)
			}
			b, e := os.ReadFile(u.Binary)
			want := "new executable"
			if fail {
				want = "old executable"
			}
			if e != nil || string(b) != want {
				t.Fatal("incorrect active binary", string(b), e)
			}
			after, e := os.ReadFile(u.Manager.ConfigPath)
			if e != nil || !bytes.Equal(after, config) {
				t.Fatal("config changed", e)
			}
			if _, e = os.Stat(upgradePending(u.Manager.ConfigPath)); !os.IsNotExist(e) {
				t.Fatal("journal remains", e)
			}
		})
	}
}
func TestQAUpgradeRecoveryRejectsTamperingAndBlocksOtherWrites(t *testing.T) {
	u, candidate := qaUpgrader(t)
	ctx := context.Background()
	u.Manager.Verify = func(context.Context, host.Config, []string) error { return errors.New("offline") }
	if e := u.Upgrade(ctx, candidate, binaryHash([]byte("new executable"))); e == nil {
		t.Fatal("expected failure")
	}
	if _, e := os.Stat(upgradePending(u.Manager.ConfigPath)); e != nil {
		t.Fatal("missing recovery journal", e)
	}
	before, _ := os.ReadFile(u.Binary)
	m, o := fixture(t)
	o.Branch = "main"
	o.Commit = strings.Repeat("a", 40)
	// An upgrade journal blocks all Manager entry points through its shared lock.
	journal, _ := json.Marshal(upgradeJournal{OldSHA: binaryHash(before), NewSHA: binaryHash([]byte("new executable"))})
	if e := os.WriteFile(upgradePending(m.ConfigPath), journal, 0600); e != nil {
		t.Fatal(e)
	}
	if _, e := m.Apply(ctx, o); e == nil {
		t.Fatal("deploy bypassed upgrade journal")
	}
	if e := m.Recover(ctx); e == nil || !strings.Contains(e.Error(), "upgrade") {
		t.Fatal("deploy recovery bypassed upgrade journal", e)
	}
	store := editstore.Store{ConfigPath: m.ConfigPath, Root: t.TempDir()}
	if e := store.Save(editor.Update{}); e == nil || !strings.Contains(e.Error(), "upgrade") {
		t.Fatal("save bypassed upgrade journal", e)
	}
	if e := os.WriteFile(u.Binary, []byte("tampered"), 0755); e != nil {
		t.Fatal(e)
	}
	if e := u.Recover(ctx); e == nil || !strings.Contains(e.Error(), "independently") {
		t.Fatal("accepted changed executable", e)
	}
	if e := os.WriteFile(u.Binary, before, 0755); e != nil {
		t.Fatal(e)
	}
	if e := os.WriteFile(u.previous(), []byte("tampered backup"), 0700); e != nil {
		t.Fatal(e)
	}
	if e := u.Recover(ctx); e == nil || !strings.Contains(e.Error(), "checksum") {
		t.Fatal("accepted changed previous executable", e)
	}
}
