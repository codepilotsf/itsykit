package setup

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"itsy/internal/admin"
	"itsy/internal/backup"
	"itsy/internal/release"
	"itsy/internal/snips"
	"itsy/internal/syncstate"
)

func qaRecoverySite(t *testing.T) string {
	t.Helper()
	root := repoSite(t)
	s := DeploySettings{Target: "root@server", Name: "demo", Domain: "example.test", Branch: "main"}
	if e := saveDeploymentConfig(root, DeploymentConfig{Branches: map[string]DeploySettings{"main": s}}); e != nil {
		t.Fatal(e)
	}
	return root
}

func TestQARecoveryConfirmationAndExpectedState(t *testing.T) {
	root := qaRecoverySite(t)
	ctx := context.Background()
	current := "20260926-120000-" + strings.Repeat("a", 24)
	previous := "20260925-120000-" + strings.Repeat("b", 24)
	calls := []string{}
	state := syncstate.Snapshot{Exists: true, Revision: "active", Snips: snips.NewStore()}
	run := func(_ context.Context, c command) error {
		for _, a := range c.args {
			if a == "-O" {
				return nil
			}
		}
		cmd := c.args[len(c.args)-1]
		calls = append(calls, cmd)
		switch {
		case strings.Contains(cmd, "-protocol"):
			fmt.Fprintln(c.out, release.Protocol)
		case strings.Contains(cmd, "-history"):
			return json.NewEncoder(c.out).Encode([]admin.ReleaseRecord{{ID: current, Previous: previous}, {ID: previous}})
		case strings.Contains(cmd, "-export-state"):
			return json.NewEncoder(c.out).Encode(state)
		case strings.Contains(cmd, "mktemp"):
			fmt.Fprintln(c.out, "/tmp/flea-restore.abcdefghij")
		case strings.HasPrefix(cmd, "cat >"):
			_, e := io.Copy(io.Discard, c.in)
			return e
		case strings.Contains(cmd, "-rollback"), strings.Contains(cmd, "-restore"), strings.HasPrefix(cmd, "rm -f"):
			return nil
		default:
			t.Fatal("unexpected command", cmd)
		}
		return nil
	}
	if e := rollback(ctx, root, "", false, strings.NewReader("n\n"), io.Discard, io.Discard, run); e == nil {
		t.Fatal("rollback cancellation accepted")
	}
	for _, c := range calls {
		if strings.Contains(c, "-rollback") {
			t.Fatal("rollback ran before confirmation")
		}
	}
	calls = nil
	if e := rollback(ctx, root, "", false, strings.NewReader("y\n"), io.Discard, io.Discard, run); e != nil {
		t.Fatal(e)
	}
	if !strings.Contains(strings.Join(calls, "\n"), "-expected-release "+shellQuote(current)) {
		t.Fatal("rollback omitted reviewed release", calls)
	}
	var buf bytes.Buffer
	if e := backup.Write(root, backup.Manifest{Name: "demo", Branch: "main", Commit: strings.Repeat("a", 40)}, &buf); e != nil {
		t.Fatal(e)
	}
	file := filepath.Join(t.TempDir(), "site.backup")
	if e := os.WriteFile(file, buf.Bytes(), 0600); e != nil {
		t.Fatal(e)
	}
	calls = nil
	if e := restoreSite(ctx, root, file, strings.NewReader("yes\n"), io.Discard, io.Discard, run); e == nil {
		t.Fatal("restore accepted nonliteral confirmation")
	}
	for _, c := range calls {
		if strings.Contains(c, "mktemp") || strings.Contains(c, "-restore") || strings.HasPrefix(c, "cat >") {
			t.Fatal("restore staged before RESTORE", c)
		}
	}
	calls = nil
	if e := restoreSite(ctx, root, file, strings.NewReader("RESTORE\n"), io.Discard, io.Discard, run); e != nil {
		t.Fatal(e)
	}
	if !strings.Contains(strings.Join(calls, "\n"), "-expected-state "+shellQuote(state.Version())) {
		t.Fatal("restore omitted reviewed state", calls)
	}
}

func TestQABackupNeverOverwritesExistingLocalFile(t *testing.T) {
	file := filepath.Join(t.TempDir(), "existing.backup")
	if e := os.WriteFile(file, []byte("keep"), 0600); e != nil {
		t.Fatal(e)
	}
	run := func(context.Context, command) error {
		t.Fatal("contacted server before local overwrite refusal")
		return nil
	}
	if e := backupSite(context.Background(), t.TempDir(), file, strings.NewReader(""), io.Discard, io.Discard, run); e == nil {
		t.Fatal("overwrote existing backup")
	}
	b, e := os.ReadFile(file)
	if e != nil || string(b) != "keep" {
		t.Fatal("existing backup changed", e)
	}
}
