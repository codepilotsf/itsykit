// Package setup bootstraps a server through the user's native SSH client.
package setup

import (
	"bufio"
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
	"runtime"
	"strings"
	"time"
)

//go:embed install.sh
var installer string

//go:embed nanoweb.service
var service string

const defaultConfig = "{\n  \"listen\": \":80\",\n  \"sites\": {}\n}\n"
const probe = `set -eu; . /etc/os-release; command -v systemctl >/dev/null; test -d /run/systemd/system; printf 'NANOWEB_SYSTEM=%s:%s:%s:%s\n' "$ID" "$VERSION_ID" "$(uname -m)" "$(id -u)"`

type command struct {
	name        string
	args        []string
	dir         string
	env         []string
	in          io.Reader
	out, errOut io.Writer
}
type runner func(context.Context, command) error

func execute(ctx context.Context, c command) error {
	cmd := exec.CommandContext(ctx, c.name, c.args...)
	cmd.Dir = c.dir
	cmd.Env = c.env
	cmd.Stdin = c.in
	cmd.Stdout = c.out
	cmd.Stderr = c.errOut
	cmd.WaitDelay = 2 * time.Second
	return cmd.Run()
}

// Run is independent of the current site directory. Source may be empty when
// the executable is linked to a development checkout's bin/ directory.
func Run(ctx context.Context, source string, in io.Reader, out, errOut io.Writer) error {
	return run(ctx, source, in, out, errOut, execute)
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

func readTarget(ctx context.Context, in io.Reader) (string, error) {
	type result struct {
		text string
		err  error
	}
	done := make(chan result, 1)
	go func() { line, err := bufio.NewReader(in).ReadString('\n'); done <- result{line, err} }()
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case r := <-done:
		if r.err != nil && (r.err != io.EOF || strings.TrimSpace(r.text) == "") {
			return "", fmt.Errorf("read target: %w", r.err)
		}
		return targetName(r.text)
	}
}

func run(ctx context.Context, source string, in io.Reader, out, errOut io.Writer, runCommand runner) error {
	fmt.Fprint(out, "Target hostname or IP? ")
	target, err := readTarget(ctx, in)
	if err != nil {
		return err
	}
	temp, err := os.MkdirTemp("/tmp", "nanoweb-setup-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(temp)
	sshOptions := []string{"-o", "ConnectTimeout=10", "-o", "ConnectionAttempts=1", "-o", "ServerAliveInterval=15", "-o", "ServerAliveCountMax=3", "-o", "ControlMaster=auto", "-o", "ControlPersist=60", "-o", "ControlPath=" + filepath.Join(temp, "ssh")}
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
	if err := ssh("sh -c "+shellQuote(probe), false, &probeOutput); err != nil {
		return fmt.Errorf("SSH connection or server check failed for %s: %w", target, err)
	}
	arch, isRoot, err := parseProbe(probeOutput.String())
	if err != nil {
		return err
	}
	// Check local build prerequisites only after a successful connection.
	source, err = findSource(source)
	if err != nil {
		return err
	}
	goBinary, err := findGo()
	if err != nil {
		return err
	}
	fmt.Fprintln(out, "\nBuilding Nanoweb server...")
	buildArgs := []string{"build", "-trimpath", "-o", filepath.Join(temp, "nanoweb-server"), "./cmd/nanoweb-server"}
	if err := runCommand(ctx, command{name: goBinary, args: buildArgs, dir: source, env: buildEnv(arch), out: out, errOut: errOut}); err != nil {
		return fmt.Errorf("server build failed: %w", err)
	}
	for name, body := range map[string]string{"install.sh": installer, "nanoweb.service": service, "server.json": defaultConfig} {
		if err := os.WriteFile(filepath.Join(temp, name), []byte(body), 0600); err != nil {
			return err
		}
	}
	var stageOutput bytes.Buffer
	if err := ssh("umask 077; mktemp -d /tmp/nanoweb-setup.XXXXXXXXXX", false, &stageOutput); err != nil {
		return fmt.Errorf("create remote staging directory: %w", err)
	}
	stage := strings.TrimSpace(stageOutput.String())
	if !regexp.MustCompile(`^/tmp/nanoweb-setup\.[A-Za-z0-9]{10}$`).MatchString(stage) {
		return fmt.Errorf("unexpected remote staging path %q", stage)
	}
	defer func() {
		cleanup, stop := context.WithTimeout(context.Background(), 5*time.Second)
		defer stop()
		// Remove only the four files staged by this invocation, then the empty directory.
		text := "rm -f -- " + shellQuote(stage+"/nanoweb-server") + " " + shellQuote(stage+"/install.sh") + " " + shellQuote(stage+"/nanoweb.service") + " " + shellQuote(stage+"/server.json") + "; rmdir -- " + shellQuote(stage)
		args := append(append([]string{}, sshOptions...), "-T", "--", target, text)
		_ = runCommand(cleanup, command{name: "ssh", args: args, out: io.Discard, errOut: io.Discard})
	}()
	fmt.Fprintln(out, "Uploading Nanoweb server...")
	scpTarget := target
	if at := strings.LastIndex(target, "@"); strings.Contains(target[at+1:], ":") {
		scpTarget = target[:at+1] + "[" + target[at+1:] + "]"
	}
	args := append([]string{}, sshOptions...)
	args = append(args, "--", filepath.Join(temp, "nanoweb-server"), filepath.Join(temp, "install.sh"), filepath.Join(temp, "nanoweb.service"), filepath.Join(temp, "server.json"), scpTarget+":"+stage+"/")
	if err := runCommand(ctx, command{name: "scp", args: args, in: in, out: out, errOut: errOut}); err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}
	fmt.Fprintln(out, "Installing Nanoweb service...")
	installCommand := "sh " + shellQuote(stage+"/install.sh") + " " + shellQuote(stage)
	if !isRoot {
		installCommand = "sudo -- " + installCommand
	}
	if err := ssh(installCommand, true, out); err != nil {
		return fmt.Errorf("server installation failed: %w", err)
	}
	fmt.Fprintln(out, "Checking Nanoweb service...")
	verify := "systemctl is-active --quiet nanoweb && /usr/local/bin/nanoweb-server -config /etc/nanoweb/server.json -check"
	if err := ssh(verify, false, out); err != nil {
		return fmt.Errorf("server installed but readiness check failed: %w", err)
	}
	fmt.Fprintf(out, "Nanoweb server is running on %s (HTTP port 80).\nConfiguration: /etc/nanoweb/server.json\nLocal readiness verified. Allow inbound TCP port 80 in your server/cloud firewall for public access.\nNo sites configured yet. Deployment and HTTPS are not enabled yet.\n", target)
	return nil
}

func parseProbe(output string) (string, bool, error) {
	for _, line := range strings.Split(output, "\n") {
		if !strings.HasPrefix(line, "NANOWEB_SYSTEM=") {
			continue
		}
		parts := strings.Split(strings.TrimSpace(strings.TrimPrefix(line, "NANOWEB_SYSTEM=")), ":")
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

func findGo() (string, error) {
	if binary, err := exec.LookPath("go"); err == nil {
		return binary, nil
	}
	for _, path := range []string{"/usr/local/go/bin/go", filepath.Join(runtime.GOROOT(), "bin", "go")} {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path, nil
		}
	}
	return "", errors.New("Go 1.25 or newer is required on the Mac to build the server")
}

func findSource(explicit string) (string, error) {
	starts := []string{}
	if explicit != "" {
		starts = append(starts, explicit)
	} else {
		if exe, err := os.Executable(); err == nil {
			if resolved, err := filepath.EvalSymlinks(exe); err == nil {
				starts = append(starts, filepath.Dir(resolved))
			}
		}
		if cwd, err := os.Getwd(); err == nil {
			starts = append(starts, cwd)
		}
	}
	for _, start := range starts {
		dir, err := filepath.Abs(start)
		if err != nil {
			continue
		}
		for {
			mod, err := os.ReadFile(filepath.Join(dir, "go.mod"))
			if err == nil && strings.HasPrefix(string(mod), "module nanoweb\n") {
				if _, err := os.Stat(filepath.Join(dir, "cmd/nanoweb-server/main.go")); err == nil {
					return dir, nil
				}
			}
			parent := filepath.Dir(dir)
			if parent == dir || explicit != "" {
				break
			}
			dir = parent
		}
	}
	return "", errors.New("Nanoweb source checkout not found; use nanoweb setup -source /path/to/nanoweb")
}

func buildEnv(arch string) []string {
	var env []string
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if key != "GOOS" && key != "GOARCH" && key != "CGO_ENABLED" && key != "GOFLAGS" {
			env = append(env, entry)
		}
	}
	return append(env, "GOOS=linux", "GOARCH="+arch, "CGO_ENABLED=0", "GOFLAGS=")
}
