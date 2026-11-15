package admin

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"itsy/internal/host"
	"itsy/internal/serverunit"
)

// socketUpgrade is deliberately limited to converting the exact original unit.
// It is not a general service-policy updater. The executable-upgrade journal
// records whether recovery must also undo this one-time conversion.
type socketUpgrade struct {
	dir string
	run func(context.Context, ...string) (string, error)
}

func systemSocketUpgrade() *socketUpgrade {
	return &socketUpgrade{dir: "/etc/systemd/system", run: func(ctx context.Context, args ...string) (string, error) {
		output, err := exec.CommandContext(ctx, "systemctl", args...).CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("systemd socket activation: %w: %s", err, output)
		}
		return strings.TrimSpace(string(output)), nil
	}}
}

func (s *socketUpgrade) read(name string) (string, error) {
	path := filepath.Join(s.dir, name)
	info, err := os.Lstat(path)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() || info.Size() > 64<<10 {
		return "", fmt.Errorf("refusing to change nonstandard systemd unit %s", path)
	}
	data, err := os.ReadFile(path)
	return string(data), err
}

func (s *socketUpgrade) prepare(ctx context.Context, config host.Config) (bool, error) {
	service, err := s.read("flea.service")
	if err != nil {
		return false, err
	}
	socket, socketErr := s.read("flea.socket")
	if socketErr != nil && !os.IsNotExist(socketErr) {
		return false, socketErr
	}
	if service == serverunit.Service && socketErr == nil && socket == serverunit.Socket {
		// Already converted. Ordinary executable upgrades never rewrite units.
		return false, nil
	}
	if service != serverunit.DirectService || !os.IsNotExist(socketErr) {
		return false, fmt.Errorf("systemd units are customized or incomplete; review socket activation manually before updating the server")
	}
	if config.Listen != ":80" {
		return false, fmt.Errorf("socket activation requires listen :80; review the custom listener before updating the server")
	}
	for _, property := range []string{"FragmentPath", "DropInPaths", "NeedDaemonReload"} {
		value, err := s.run(ctx, "show", "flea.service", "--property="+property, "--value")
		if err != nil {
			return false, err
		}
		if (property == "FragmentPath" && value != filepath.Join(s.dir, "flea.service")) || (property == "DropInPaths" && value != "") || (property == "NeedDaemonReload" && value != "no") {
			return false, fmt.Errorf("systemd service has overrides or pending changes; review socket activation manually before updating the server")
		}
	}
	// Refuse a socket unit from another systemd search path or any drop-in,
	// including ones that would apply only after installing the new unit.
	value, err := s.run(ctx, "show", "flea.socket", "--property=LoadState", "--value")
	if err != nil {
		return false, err
	}
	if value != "not-found" {
		return false, fmt.Errorf("a systemd flea.socket unit already exists; review it before updating the server")
	}
	for _, directory := range []string{s.dir, "/run/systemd/system", "/usr/local/lib/systemd/system", "/usr/lib/systemd/system", "/lib/systemd/system"} {
		for _, unit := range []string{"flea.service", "flea.socket"} {
			if _, err := os.Lstat(filepath.Join(directory, unit+".d")); !os.IsNotExist(err) {
				return false, fmt.Errorf("systemd unit drop-ins require manual review before socket activation")
			}
		}
	}
	return true, nil
}

func (s *socketUpgrade) apply(ctx context.Context) error {
	service, err := s.read("flea.service")
	if err != nil {
		return err
	}
	if service != serverunit.DirectService {
		return fmt.Errorf("service unit changed before socket activation; conversion stopped")
	}
	if _, err := os.Lstat(filepath.Join(s.dir, "flea.socket")); !os.IsNotExist(err) {
		return fmt.Errorf("socket unit appeared before socket activation; conversion stopped")
	}
	if _, err := s.run(ctx, "stop", "flea.service"); err != nil {
		return err
	}
	for name, data := range map[string]string{"flea.service": serverunit.Service, "flea.socket": serverunit.Socket} {
		if err := atomicWrite(filepath.Join(s.dir, name), []byte(data), 0644); err != nil {
			return err
		}
	}
	if _, err := s.run(ctx, "daemon-reload"); err != nil {
		return err
	}
	_, err = s.run(ctx, "enable", "--now", "flea.socket")
	return err
}

func (s *socketUpgrade) restore(ctx context.Context) error {
	// Validate both files before stopping anything. A pending journal never
	// authorizes overwriting independent administrator changes.
	service, err := s.read("flea.service")
	if err != nil {
		return err
	}
	if service != serverunit.Service && service != serverunit.DirectService {
		return fmt.Errorf("service unit changed independently; socket activation recovery stopped")
	}
	socket, socketErr := s.read("flea.socket")
	if socketErr != nil && !os.IsNotExist(socketErr) {
		return socketErr
	}
	if socketErr == nil && socket != serverunit.Socket {
		return fmt.Errorf("socket unit changed independently; socket activation recovery stopped")
	}
	if _, err := s.run(ctx, "stop", "flea.service"); err != nil {
		return err
	}
	if socketErr == nil {
		if _, err := s.run(ctx, "disable", "--now", "flea.socket"); err != nil {
			return err
		}
		if err := os.Remove(filepath.Join(s.dir, "flea.socket")); err != nil {
			return err
		}
	}
	if err := atomicWrite(filepath.Join(s.dir, "flea.service"), []byte(serverunit.DirectService), 0644); err != nil {
		return err
	}
	_, err = s.run(ctx, "daemon-reload")
	return err
}
