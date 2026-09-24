// Package setup bootstraps a server through the user's native SSH client.
package setup

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"flea/internal/serverbundle"
)

//go:embed install.sh
var installer string

//go:embed flea.service
var service string

const defaultConfig = "{\n  \"listen\": \":80\",\n  \"sites\": {}\n}\n"
const probe = `set -eu; . /etc/os-release; command -v systemctl >/dev/null; test -d /run/systemd/system; printf 'FLEA_SYSTEM=%s:%s:%s:%s\n' "$ID" "$VERSION_ID" "$(uname -m)" "$(id -u)"`

type command struct {
	name        string
	args        []string
	dir         string
	env         []string
	in          io.Reader
	out, errOut io.Writer
}
type runner func(context.Context, command) error

var errHostKeyChanged = errors.New("The server's SSH identity has changed.\n\nConnection stopped to protect your site. This can happen after a server rebuild,\nbut it could also mean you are connecting to a different server.\nVerify the new fingerprint with your server provider before updating its saved\nentry in ~/.ssh/known_hosts. FleaCMS has not removed or replaced the saved key.")

type sshErrorWriter struct {
	out     io.Writer
	tail    string
	changed bool
}

func (w *sshErrorWriter) Write(p []byte) (int, error) {
	w.tail += string(p)
	if strings.Contains(w.tail, "REMOTE HOST IDENTIFICATION HAS CHANGED") || strings.Contains(w.tail, "REVOKED HOST KEY DETECTED") {
		w.changed = true
	}
	if len(w.tail) > 4096 {
		w.tail = w.tail[len(w.tail)-4096:]
	}
	if w.out == nil {
		return len(p), nil
	}
	return w.out.Write(p)
}

func execute(ctx context.Context, c command) error {
	if c.name == "extract-server" {
		return serverbundle.Extract(ctx, c.args[0], c.args[1])
	}
	cmd := exec.CommandContext(ctx, c.name, c.args...)
	cmd.Dir = c.dir
	cmd.Env = c.env
	cmd.Stdin = c.in
	cmd.Stdout = c.out
	cmd.Stderr = c.errOut
	var sshErrors *sshErrorWriter
	if c.name == "ssh" || c.name == "scp" {
		sshErrors = &sshErrorWriter{out: c.errOut}
		cmd.Stderr = sshErrors
	}
	cmd.WaitDelay = 2 * time.Second
	err := cmd.Run()
	if err != nil && sshErrors != nil && sshErrors.changed {
		return errHostKeyChanged
	}
	return err
}

// Run installs the bundled server independently of the site directory.
func Run(ctx context.Context, in io.Reader, out, errOut io.Writer) error {
	return run(ctx, in, out, errOut, execute)
}

func targetName(input string) (string, error) {
	input = strings.TrimSpace(input)
	parts := strings.Split(input, "@")
	if len(parts) > 2 {
		return "", fmt.Errorf("enter a hostname, IP, SSH alias, or user@host")
	}
	host := parts[len(parts)-1]
	if len(parts) == 2 && !regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.-]*$`).MatchString(parts[0]) {
		return "", fmt.Errorf("invalid SSH username")
	}
	host = strings.TrimPrefix(strings.TrimSuffix(host, "]"), "[")
	if host == "" || (net.ParseIP(host) == nil && !regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]*$`).MatchString(host)) {
		return "", fmt.Errorf("invalid hostname or IP; configure custom ports in ~/.ssh/config")
	}
	if len(parts) == 2 {
		return parts[0] + "@" + host, nil
	}
	return host, nil
}

func shellQuote(s string) string { return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'" }

func run(ctx context.Context, in io.Reader, out, errOut io.Writer, runCommand runner) error {
	target, err := ask(ctx, in, out, "Server hostname or IP", "")
	if err != nil {
		return err
	}
	target, err = targetName(target)
	if err != nil {
		return err
	}
	if at := strings.LastIndex(target, "@"); at >= 0 {
		target = target[at+1:]
	}
	target = "root@" + target
	return runTarget(ctx, target, in, out, errOut, runCommand)
}

func runTarget(ctx context.Context, target string, in io.Reader, out, errOut io.Writer, runCommand runner) error {
	target, err := targetName(target)
	if err != nil {
		return err
	}
	temp, err := os.MkdirTemp("/tmp", "flea-setup-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(temp)
	sshOptions := connectionOptions(temp)
	ssh := func(commandText string, tty bool, output io.Writer) error {
		args := append([]string{}, sshOptions...)
		if tty {
			args = append(args, "-tt")
		} else {
			args = append(args, "-T")
		}
		args = append(args, "--", target, commandText)
		return runCommand(ctx, command{name: "ssh", args: args, in: in, out: output, errOut: errOut})
	}
	defer func() {
		cleanup, stop := context.WithTimeout(context.Background(), 5*time.Second)
		defer stop()
		args := append(append([]string{}, sshOptions...), "-O", "exit", "--", target)
		_ = runCommand(cleanup, command{name: "ssh", args: args, out: io.Discard, errOut: io.Discard})
	}()
	var probeOutput bytes.Buffer
	if err := ssh("sh -c "+shellQuote(probe+`; if test -f /usr/local/bin/flea-server && test -f /etc/flea/server.json; then printf 'FLEA_EXISTING=1\n'; fi`), false, &probeOutput); err != nil {
		if errors.Is(err, errHostKeyChanged) {
			return err
		}
		return fmt.Errorf("SSH connection or server check failed for %s: %w", target, err)
	}
	arch, isRoot, err := parseProbe(probeOutput.String())
	if err != nil {
		return err
	}
	reinstall := strings.Contains(probeOutput.String(), "FLEA_EXISTING=1\n")
	// Verify the replacement before asking permission to reset a working server.
	// Extraction failures must never lead to remote installation changes.
	fmt.Fprintln(out, "\nPreparing bundled FleaCMS server...")
	if err := runCommand(ctx, command{name: "extract-server", args: []string{arch, filepath.Join(temp, "flea-server")}}); err != nil {
		return err
	}
	if reinstall {
		fmt.Fprint(out, "\nWARNING: FleaCMS is already installed on this server.\nReinstalling will take ALL existing sites offline and replace the installation,\nincluding site configuration and certificates, with an empty installation.\nThe previous files will be kept in a root-only backup on the server.\nThere is no automatic undo.\n\n")
		answer, err := ask(ctx, in, out, "Type DESTROY to replace this installation (Enter cancels)", "")
		if err != nil {
			return err
		}
		if answer != "DESTROY" {
			return fmt.Errorf("Setup cancelled. The existing installation was not changed.")
		}
	}
	for name, body := range map[string]string{"install.sh": installer, "flea.service": service, "server.json": defaultConfig} {
		if err := os.WriteFile(filepath.Join(temp, name), []byte(body), 0600); err != nil {
			return err
		}
	}
	var stageOutput bytes.Buffer
	if err := ssh("umask 077; mktemp -d /tmp/flea-setup.XXXXXXXXXX", false, &stageOutput); err != nil {
		return fmt.Errorf("create remote staging directory: %w", err)
	}
	stage := strings.TrimSpace(stageOutput.String())
	if !regexp.MustCompile(`^/tmp/flea-setup\.[A-Za-z0-9]{10}$`).MatchString(stage) {
		return fmt.Errorf("unexpected remote staging path %q", stage)
	}
	defer func() {
		cleanup, stop := context.WithTimeout(context.Background(), 5*time.Second)
		defer stop()
		// Remove only the four files staged by this invocation, then the empty directory.
		text := "rm -f -- " + shellQuote(stage+"/flea-server") + " " + shellQuote(stage+"/install.sh") + " " + shellQuote(stage+"/flea.service") + " " + shellQuote(stage+"/server.json") + "; rmdir -- " + shellQuote(stage)
		args := append(append([]string{}, sshOptions...), "-T", "--", target, text)
		_ = runCommand(cleanup, command{name: "ssh", args: args, out: io.Discard, errOut: io.Discard})
	}()
	fmt.Fprintln(out, "Uploading FleaCMS server...")
	scpTarget := scpHost(target)
	args := append([]string{}, sshOptions...)
	args = append(args, "--", filepath.Join(temp, "flea-server"), filepath.Join(temp, "install.sh"), filepath.Join(temp, "flea.service"), filepath.Join(temp, "server.json"), scpTarget+":"+stage+"/")
	if err := runCommand(ctx, command{name: "scp", args: args, in: in, out: out, errOut: errOut}); err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}
	fmt.Fprintln(out, "Installing FleaCMS service...")
	installCommand := "sh " + shellQuote(stage+"/install.sh") + " " + shellQuote(stage)
	if reinstall {
		installCommand += " --reinstall"
	}
	if !isRoot {
		installCommand = "sudo -- " + installCommand
	}
	if err := ssh(installCommand, true, out); err != nil {
		return fmt.Errorf("server installation failed: %w", err)
	}
	fmt.Fprintln(out, "Checking FleaCMS service...")
	verify := "systemctl is-active --quiet flea && /usr/local/bin/flea-server -config /etc/flea/server.json -check"
	if err := ssh(verify, false, out); err != nil {
		return fmt.Errorf("server installed but readiness check failed: %w", err)
	}
	fmt.Fprintf(out, "FleaCMS server is running on %s (HTTP port 80).\nConfiguration: /etc/flea/server.json\nLocal readiness verified. Allow inbound TCP ports 80 and 443 for HTTP and HTTPS.\nNext: run flea deploy -target %s -domain YOUR_DOMAIN from your site.\n", target, target)
	return nil
}

func parseProbe(output string) (string, bool, error) {
	for _, line := range strings.Split(output, "\n") {
		if !strings.HasPrefix(line, "FLEA_SYSTEM=") {
			continue
		}
		parts := strings.Split(strings.TrimSpace(strings.TrimPrefix(line, "FLEA_SYSTEM=")), ":")
		if len(parts) != 4 {
			break
		}
		if parts[0] != "ubuntu" || (parts[1] != "22.04" && parts[1] != "24.04" && parts[1] != "26.04") {
			return "", false, fmt.Errorf("setup requires Ubuntu LTS 22.04, 24.04, or 26.04 with systemd; found %s %s", parts[0], parts[1])
		}
		arch := ""
		switch parts[2] {
		case "x86_64":
			arch = "amd64"
		case "aarch64", "arm64":
			arch = "arm64"
		default:
			return "", false, fmt.Errorf("unsupported server architecture %q", parts[2])
		}
		if !regexp.MustCompile(`^[0-9]+$`).MatchString(parts[3]) {
			break
		}
		return arch, parts[3] == "0", nil
	}
	return "", false, fmt.Errorf("could not identify remote Ubuntu version and architecture")
}
