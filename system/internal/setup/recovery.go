package setup

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"itsy/internal/admin"
	"itsy/internal/release"
)

type remoteAdmin struct {
	settings DeploySettings
	ssh      func(context.Context, string, bool, io.Writer) error
	root     bool
	upload   func(context.Context, string, io.Reader) error
}

func withRemoteAdmin(ctx context.Context, root string, in io.Reader, out, errOut io.Writer, run runner, fn func(remoteAdmin) error) error {
	return withRemoteAdminProtocol(ctx, root, in, out, errOut, run, true, fn)
}
func withRemoteAdminProtocol(ctx context.Context, root string, in io.Reader, out, errOut io.Writer, run runner, checkProtocol bool, fn func(remoteAdmin) error) error {
	g, err := currentGit(ctx, root, false)
	if err != nil {
		return err
	}
	config, err := readDeploymentConfig(root)
	if err != nil {
		return err
	}
	s, ok := config.Branches[g.Branch]
	if !ok {
		return fmt.Errorf("configure this branch with itsy deploy -configure first")
	}
	s.Branch = g.Branch
	if err = validateMapping(g.Branch, s, config); err != nil {
		return err
	}
	target, err := targetName(s.Target)
	if err != nil {
		return err
	}
	dir, err := os.MkdirTemp("", "flea-recovery-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	options := connectionOptions(dir)
	defer func() {
		cleanup, stop := context.WithTimeout(context.Background(), 5*time.Second)
		defer stop()
		_ = run(cleanup, command{name: "ssh", args: append(append([]string{}, options...), "-O", "exit", "--", target), out: io.Discard, errOut: io.Discard})
	}()
	ssh := func(ctx context.Context, text string, tty bool, w io.Writer) error {
		args := append([]string{}, options...)
		if tty {
			args = append(args, "-tt")
		} else {
			args = append(args, "-T")
		}
		args = append(args, "--", target, text)
		return run(ctx, command{name: "ssh", args: args, in: in, out: w, errOut: errOut})
	}
	if checkProtocol {
		var probe bytes.Buffer
		if err = ssh(ctx, "/usr/local/bin/flea-server -protocol", false, &probe); err != nil {
			return err
		}
		if strings.TrimSpace(probe.String()) != release.Protocol {
			return fmt.Errorf("update the server before using recovery commands")
		}
	}
	upload := func(ctx context.Context, path string, input io.Reader) error {
		return run(ctx, command{name: "ssh", args: append(append([]string{}, options...), "-T", "--", target, "cat > "+shellQuote(path)), in: input, out: out, errOut: errOut})
	}
	return fn(remoteAdmin{s, ssh, strings.HasPrefix(target, "root@"), upload})
}

func Rollback(ctx context.Context, root, target string, list bool, in io.Reader, out, errOut io.Writer) error {
	return rollback(ctx, root, target, list, in, out, errOut, execute)
}
func rollback(ctx context.Context, root, target string, list bool, in io.Reader, out, errOut io.Writer, run runner) error {
	return withRemoteAdmin(ctx, root, in, out, errOut, run, func(r remoteAdmin) error {
		suffix := " -name " + shellQuote(r.settings.Name) + " -branch " + shellQuote(r.settings.Branch)
		raw, err := captureAdmin(ctx, "/usr/local/bin/flea-server -history"+suffix, r.ssh, r.root)
		if err != nil {
			return err
		}
		var records []admin.ReleaseRecord
		if err = json.Unmarshal(raw, &records); err != nil {
			return err
		}
		if len(records) == 0 {
			return fmt.Errorf("no release history")
		}
		if list {
			for i, record := range records {
				label := ""
				if i == 0 {
					label = " (current)"
				}
				fmt.Fprintf(out, "%s  %s%s\n", record.ID, record.Commit, label)
			}
			return nil
		}
		if target == "" {
			if len(records) < 2 {
				return fmt.Errorf("no previous release")
			}
			target = records[1].ID
		}
		found := false
		for _, record := range records[1:] {
			if record.ID == target {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("choose a release from itsy rollback --list")
		}
		fmt.Fprintf(out, "Rollback %s to %s. Live Blog content will be kept. Styles come from the selected release.\n", r.settings.Domain, target)
		answer, err := ask(ctx, in, out, "Continue? (y/N)", "n")
		if err != nil {
			return err
		}
		if strings.ToLower(answer) != "y" && strings.ToLower(answer) != "yes" {
			return fmt.Errorf("cancelled; nothing changed")
		}
		latest, err := currentGit(ctx, root, false)
		if err != nil {
			return err
		}
		if latest.Branch != r.settings.Branch {
			return fmt.Errorf("branch changed; retry")
		}
		cmd := "/usr/local/bin/flea-server -rollback " + shellQuote(target) + " -expected-release " + shellQuote(records[0].ID) + suffix
		if !r.root {
			cmd = "sudo -- " + cmd
		}
		return r.ssh(ctx, cmd, !r.root, out)
	})
}
