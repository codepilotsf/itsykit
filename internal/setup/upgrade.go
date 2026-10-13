package setup

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func Upgrade(ctx context.Context, root string, recover bool, in io.Reader, out, errOut io.Writer, terminal bool) error {
	return upgradeServer(ctx, root, recover, in, out, errOut, terminal, execute)
}
func upgradeServer(ctx context.Context, root string, recover bool, in io.Reader, out, errOut io.Writer, terminal bool, run runner) error {
	var progress *commandProgress
	stop := func() {
		if progress != nil {
			progress.stop()
		}
	}
	defer stop()
	wrapped := func(ctx context.Context, c command) error {
		for _, arg := range c.args {
			if arg == "-tt" {
				stop()
				break
			}
		}
		return run(ctx, c)
	}
	output := progressOutput{Writer: out, stop: stop}
	diagnostic := progressOutput{Writer: errOut, stop: stop}
	err := withRemoteAdminProtocol(ctx, root, in, output, diagnostic, wrapped, false, func(r remoteAdmin) error {
		var result bytes.Buffer
		if err := r.ssh(ctx, probe, false, &result); err != nil {
			return err
		}
		arch, _, err := parseProbe(result.String())
		if err != nil {
			return err
		}
		local, err := os.MkdirTemp("", "flea-upgrade-")
		if err != nil {
			return err
		}
		defer os.RemoveAll(local)
		binary := filepath.Join(local, "flea-server")
		if err = run(ctx, command{name: "extract-server", args: []string{arch, binary}}); err != nil {
			return err
		}
		file, err := os.Open(binary)
		if err != nil {
			return err
		}
		defer file.Close()
		h := sha256.New()
		if _, err = io.Copy(h, file); err != nil {
			return err
		}
		sum := hex.EncodeToString(h.Sum(nil))
		if _, err = file.Seek(0, io.SeekStart); err != nil {
			return err
		}
		fmt.Fprintf(out, "Update the server on %s. All sites share this executable; a brief restart is required.\n", r.settings.Target)
		answer, err := ask(ctx, in, out, "Continue? (y/N)", "n")
		if err != nil {
			return err
		}
		if strings.ToLower(answer) != "y" && strings.ToLower(answer) != "yes" {
			return fmt.Errorf("cancelled; nothing changed")
		}
		progress = newCommandProgress(out, terminal, "Updating server...")
		result.Reset()
		if err = r.ssh(ctx, "umask 077; mktemp /tmp/flea-upgrade.XXXXXXXXXX", false, &result); err != nil {
			return err
		}
		stage := strings.TrimSpace(result.String())
		if !regexp.MustCompile(`^/tmp/flea-upgrade\.[A-Za-z0-9]{10}$`).MatchString(stage) {
			return fmt.Errorf("unexpected upgrade staging path")
		}
		defer func() {
			cleanup, stop := context.WithTimeout(context.Background(), 5*time.Second)
			defer stop()
			_ = r.ssh(cleanup, "rm -f -- "+shellQuote(stage), false, io.Discard)
		}()
		if err = r.upload(ctx, stage, file); err != nil {
			return err
		}
		if err = r.ssh(ctx, "chmod 700 -- "+shellQuote(stage), false, io.Discard); err != nil {
			return err
		}
		command := shellQuote(stage) + " -upgrade " + shellQuote(stage) + " -sha256 " + shellQuote(sum)
		if recover {
			command = shellQuote(stage) + " -recover-upgrade"
		}
		if !r.root {
			command = "sudo -- " + command
		}
		return r.ssh(ctx, command, !r.root, progressOutput{Writer: out, stop: stop})
	})
	stop()
	if err == nil {
		if recover {
			fmt.Fprintln(out, "Server update recovery complete.")
		} else {
			fmt.Fprintln(out, "Server update complete.")
		}
	}
	return err
}
