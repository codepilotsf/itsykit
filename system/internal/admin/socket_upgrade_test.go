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

	"itsy/internal/host"
	"itsy/internal/serverunit"
)

func socketUpgrader(t *testing.T) (Upgrader, string, *[]string) {
	t.Helper()
	u, candidate := qaUpgrader(t)
	if err := os.WriteFile(u.Manager.ConfigPath, []byte(`{"listen":":80","sites":{}}`), 0644); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "flea.service"), []byte(serverunit.DirectService), 0644); err != nil {
		t.Fatal(err)
	}
	var commands []string
	u.sockets = &socketUpgrade{dir: dir, run: func(_ context.Context, args ...string) (string, error) {
		command := strings.Join(args, " ")
		commands = append(commands, command)
		switch command {
		case "show flea.service --property=FragmentPath --value":
			return filepath.Join(dir, "flea.service"), nil
		case "show flea.service --property=DropInPaths --value":
			return "", nil
		case "show flea.service --property=NeedDaemonReload --value":
			return "no", nil
		case "show flea.socket --property=LoadState --value":
			return "not-found", nil
		}
		return "", nil
	}}
	return u, candidate, &commands
}

func TestUpgradeConvertsStandardServiceOnce(t *testing.T) {
	u, candidate, commands := socketUpgrader(t)
	config, _ := os.ReadFile(u.Manager.ConfigPath)
	u.Manager.Restart = func(context.Context) error {
		for name, expected := range map[string]string{"flea.service": serverunit.Service, "flea.socket": serverunit.Socket} {
			data, err := u.sockets.read(name)
			if err != nil || data != expected {
				t.Fatalf("restart before units installed: %s: %v", name, err)
			}
		}
		if _, err := os.Stat(upgradePending(u.Manager.ConfigPath)); err != nil {
			t.Fatal("conversion not journaled before restart", err)
		}
		return nil
	}
	if err := u.Upgrade(context.Background(), candidate, binaryHash([]byte("new executable"))); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(u.Manager.ConfigPath)
	if err != nil || !bytes.Equal(config, after) {
		t.Fatal("conversion changed site configuration", err)
	}
	operations := strings.Join(*commands, "\n")
	if !strings.Contains(operations, "stop flea.service\ndaemon-reload\nenable --now flea.socket") {
		t.Fatal("incorrect conversion order", operations)
	}
	*commands = nil
	if err := u.Upgrade(context.Background(), candidate, binaryHash([]byte("new executable"))); err != nil {
		t.Fatal(err)
	}
	if len(*commands) != 0 {
		t.Fatal("ordinary update changed service policy", *commands)
	}
}

func TestSocketConversionFailureRestoresExecutableAndUnits(t *testing.T) {
	for _, failure := range []string{"stop", "daemon-reload", "enable", "restart", "probe"} {
		t.Run(failure, func(t *testing.T) {
			u, candidate, _ := socketUpgrader(t)
			originalRun := u.sockets.run
			failed := false
			u.sockets.run = func(ctx context.Context, args ...string) (string, error) {
				if args[0] == failure && !failed {
					failed = true
					return "", errors.New("systemd test failure")
				}
				return originalRun(ctx, args...)
			}
			u.Manager.Restart = func(context.Context) error {
				if failure == "restart" && !failed {
					failed = true
					return errors.New("restart test failure")
				}
				return nil
			}
			u.Manager.Verify = func(context.Context, host.Config, []string) error {
				if failure == "probe" && !failed {
					failed = true
					return errors.New("probe test failure")
				}
				return nil
			}
			if err := u.Upgrade(context.Background(), candidate, binaryHash([]byte("new executable"))); err == nil || !strings.Contains(err.Error(), "previous server restored") {
				t.Fatal("missing automatic recovery", err)
			}
			service, err := u.sockets.read("flea.service")
			if err != nil || service != serverunit.DirectService {
				t.Fatal("old service not restored", err)
			}
			if _, err := os.Lstat(filepath.Join(u.sockets.dir, "flea.socket")); !os.IsNotExist(err) {
				t.Fatal("socket unit retained", err)
			}
			binary, err := os.ReadFile(u.Binary)
			if err != nil || string(binary) != "old executable" {
				t.Fatal("old executable not restored", err)
			}
			if _, err := os.Stat(upgradePending(u.Manager.ConfigPath)); !os.IsNotExist(err) {
				t.Fatal("recovered journal retained", err)
			}
		})
	}
}

func TestSocketConversionRefusesCustomPolicyBeforeMutation(t *testing.T) {
	for _, kind := range []string{"service", "socket", "symlink", "drop-in", "pending changes", "other unit path", "other socket unit", "systemd unavailable", "listener"} {
		t.Run(kind, func(t *testing.T) {
			u, candidate, commands := socketUpgrader(t)
			originalRun := u.sockets.run
			service := filepath.Join(u.sockets.dir, "flea.service")
			switch kind {
			case "service":
				if err := os.WriteFile(service, []byte(serverunit.DirectService+"# customized\n"), 0644); err != nil {
					t.Fatal(err)
				}
			case "socket":
				if err := os.WriteFile(filepath.Join(u.sockets.dir, "flea.socket"), []byte("custom socket"), 0644); err != nil {
					t.Fatal(err)
				}
			case "symlink":
				if err := os.Remove(service); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(candidate, service); err != nil {
					t.Fatal(err)
				}
			case "listener":
				if err := os.WriteFile(u.Manager.ConfigPath, []byte(`{"listen":"127.0.0.1:8080","sites":{}}`), 0644); err != nil {
					t.Fatal(err)
				}
			default:
				u.sockets.run = func(ctx context.Context, args ...string) (string, error) {
					cmd := strings.Join(args, " ")
					switch {
					case kind == "drop-in" && strings.Contains(cmd, "DropInPaths"):
						return "/run/systemd/system/flea.service.d/custom.conf", nil
					case kind == "other unit path" && strings.Contains(cmd, "FragmentPath"):
						return "/usr/lib/systemd/system/flea.service", nil
					case kind == "other socket unit" && strings.Contains(cmd, "LoadState"):
						return "loaded", nil
					case kind == "pending changes" && strings.Contains(cmd, "NeedDaemonReload"):
						return "yes", nil
					case kind == "systemd unavailable":
						return "", errors.New("systemd unavailable")
					}
					return originalRun(ctx, args...)
				}
			}
			if err := u.Upgrade(context.Background(), candidate, binaryHash([]byte("new executable"))); err == nil {
				t.Fatal("accepted custom policy")
			}
			binary, _ := os.ReadFile(u.Binary)
			if string(binary) != "old executable" {
				t.Fatal("changed executable before policy review")
			}
			if _, err := os.Stat(upgradePending(u.Manager.ConfigPath)); !os.IsNotExist(err) {
				t.Fatal("created upgrade journal before review", err)
			}
			for _, command := range *commands {
				if !strings.HasPrefix(command, "show ") {
					t.Fatal("mutated systemd before review", command)
				}
			}
		})
	}
}

func TestInterruptedSocketConversionRecoveryPreservesIndependentChanges(t *testing.T) {
	u, candidate, commands := socketUpgrader(t)
	run := u.sockets.run
	u.sockets.run = func(ctx context.Context, args ...string) (string, error) {
		if args[0] == "enable" || args[0] == "disable" {
			return "", errors.New("systemd unavailable")
		}
		return run(ctx, args...)
	}
	if err := u.Upgrade(context.Background(), candidate, binaryHash([]byte("new executable"))); err == nil || !strings.Contains(err.Error(), "recovery failed") {
		t.Fatal("expected pending conversion", err)
	}
	if _, err := os.Stat(upgradePending(u.Manager.ConfigPath)); err != nil {
		t.Fatal("missing pending journal", err)
	}
	u.sockets.run = run
	path := filepath.Join(u.sockets.dir, "flea.socket")
	if err := os.WriteFile(path, []byte("administrator change"), 0644); err != nil {
		t.Fatal(err)
	}
	*commands = nil
	if err := u.Recover(context.Background()); err == nil || !strings.Contains(err.Error(), "independently") {
		t.Fatal("overwrote independent socket policy", err)
	}
	if len(*commands) != 0 {
		t.Fatal("changed systemd before validating recovery", *commands)
	}
	if err := os.WriteFile(path, []byte(serverunit.Socket), 0644); err != nil {
		t.Fatal(err)
	}
	if err := u.Recover(context.Background()); err != nil {
		t.Fatal(err)
	}
	service, err := u.sockets.read("flea.service")
	if err != nil || service != serverunit.DirectService {
		t.Fatal("interrupted conversion not undone", err)
	}
	binary, err := os.ReadFile(u.Binary)
	if err != nil || string(binary) != "old executable" {
		t.Fatal("interrupted executable swap not undone", err)
	}
}

func TestSocketRecoveryAfterPartialUnitPublication(t *testing.T) {
	for _, phase := range []string{"journal only", "service only", "socket only"} {
		t.Run(phase, func(t *testing.T) {
			u, _, _ := socketUpgrader(t)
			old := []byte("old executable")
			next := []byte("new executable")
			journal, _ := json.Marshal(upgradeJournal{OldSHA: binaryHash(old), NewSHA: binaryHash(next), SocketMigration: true})
			files := map[string][]byte{u.previous(): old, u.Binary: next, upgradePending(u.Manager.ConfigPath): journal}
			if phase == "service only" {
				files[filepath.Join(u.sockets.dir, "flea.service")] = []byte(serverunit.Service)
			}
			if phase == "socket only" {
				files[filepath.Join(u.sockets.dir, "flea.socket")] = []byte(serverunit.Socket)
			}
			for path, data := range files {
				if err := os.WriteFile(path, data, 0600); err != nil {
					t.Fatal(err)
				}
			}
			if err := u.Recover(context.Background()); err != nil {
				t.Fatal(err)
			}
			service, err := u.sockets.read("flea.service")
			if err != nil || service != serverunit.DirectService {
				t.Fatal("partial conversion not restored", err)
			}
			binary, err := os.ReadFile(u.Binary)
			if err != nil || !bytes.Equal(binary, old) {
				t.Fatal("executable not restored", err)
			}
			if _, err := os.Lstat(filepath.Join(u.sockets.dir, "flea.socket")); !os.IsNotExist(err) {
				t.Fatal("partial socket unit retained", err)
			}
		})
	}
}
