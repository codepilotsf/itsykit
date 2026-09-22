package setup

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"nanoweb/internal/admin"
	"nanoweb/internal/release"
	"nanoweb/internal/site"
)

// DeploySettings contains no secrets. Native SSH configuration owns credentials.
type DeploySettings struct {
	Target      string   `json:"target"`
	Domain      string   `json:"domain"`
	Name        string   `json:"name"`
	HTTPS       bool     `json:"https"`
	Email       string   `json:"email"`
	AgreeTOS    bool     `json:"agree_tos"`
	WWWRedirect bool     `json:"www_redirect"`
	ExpectedIPs []string `json:"expected_ips,omitempty"`
	Branch      string   `json:"-"`
}

func Deploy(ctx context.Context, root string, s DeploySettings, httpsOnly bool, in io.Reader, out, errOut io.Writer) error {
	if err := prepareCommit(ctx, root, s.Branch, in, out, errOut); err != nil {
		return err
	}
	return deploy(ctx, root, s, httpsOnly, in, out, errOut, execute)
}

func deploy(ctx context.Context, root string, s DeploySettings, httpsOnly bool, in io.Reader, out, errOut io.Writer, runCommand runner) error {
	return deployWithSetup(ctx, root, s, httpsOnly, in, out, errOut, runCommand, nil)
}

func deployWithSetup(ctx context.Context, root string, s DeploySettings, httpsOnly bool, in io.Reader, out, errOut io.Writer, runCommand runner, install func(context.Context, string) error) error {
	state, err := currentGit(ctx, root, true)
	if err != nil {
		return err
	}
	if s.Branch != "" && s.Branch != state.Branch {
		return fmt.Errorf("branch changed since selecting the deployment destination")
	}
	fmt.Fprintf(out, "\nDeploying %s branch to %s.\nPress Ctrl-C to abort.\n\n", state.Branch, s.Domain)
	target, err := targetName(s.Target)
	if err != nil {
		return fmt.Errorf("deploy target: %w", err)
	}
	root, err = filepath.Abs(root)
	if err != nil {
		return err
	}
	if s.Name == "" {
		s.Name = filepath.Base(root)
	}
	opts := admin.Options{Name: s.Name, Domains: s.domains(), HTTPS: s.HTTPS || httpsOnly, Email: s.Email, AgreeTOS: s.AgreeTOS, Branch: state.Branch, Commit: state.Commit}
	if s.WWWRedirect {
		opts.Canonical = s.Domain
	}
	if !httpsOnly {
		opts.Archive = "site.tar.gz"
	}
	if err = opts.Validate(); err != nil {
		return err
	}
	temp, err := os.MkdirTemp("/tmp", "nanoweb-deploy-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(temp)
	archive := filepath.Join(temp, "site.tar.gz")
	needsMail := false
	if !httpsOnly {
		fmt.Fprintln(out, "Validating and packaging site...")
		committed := filepath.Join(temp, "commit")
		if err = os.Mkdir(committed, 0700); err != nil {
			return err
		}
		if err = committedSnapshot(ctx, root, committed, state); err != nil {
			return err
		}
		p, err := site.LoadDev(ctx, committed)
		if err != nil {
			return err
		}
		if err = p.CheckExtractedSnips(); err != nil {
			return err
		}
		needsMail = len(p.Admin.Emails) > 0
		if needsMail && !opts.HTTPS {
			return fmt.Errorf("Admin login requires HTTPS. Enable it with nanoweb deploy -configure before deploying this site.")
		}
		for _, warning := range p.Warnings {
			fmt.Fprintln(errOut, "Warning:", warning)
		}
		f, err := os.OpenFile(archive, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		err = release.Pack(committed, f)
		closeErr := f.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
		snapshot := filepath.Join(temp, "snapshot")
		if err = os.Mkdir(snapshot, 0700); err != nil {
			return err
		}
		input, err := os.Open(archive)
		if err != nil {
			return err
		}
		err = release.Extract(input, snapshot)
		input.Close()
		if err != nil {
			return err
		}
		if _, err = site.Load(snapshot); err != nil {
			return fmt.Errorf("packaged site: %w", err)
		}
	}
	sshOptions := connectionOptions(temp)
	if install == nil {
		install = func(ctx context.Context, target string) error {
			// Reuse the authenticated deployment connection during installation.
			// Only the outer deployment closes it when all work is finished.
			sharedRunner := func(ctx context.Context, c command) error {
				if c.name == "ssh" || c.name == "scp" {
					for i, arg := range c.args {
						if arg == "-O" && i+1 < len(c.args) && c.args[i+1] == "exit" {
							return nil
						}
						if strings.HasPrefix(arg, "ControlPath=") {
							c.args[i] = "ControlPath=" + filepath.Join(temp, "ssh")
						}
					}
				}
				return runCommand(ctx, c)
			}
			return runTarget(ctx, target, in, out, errOut, sharedRunner)
		}
	}
	ssh := func(ctx context.Context, text string, tty bool, output io.Writer) error {
		args := append([]string{}, sshOptions...)
		if tty {
			args = append(args, "-tt")
		} else {
			args = append(args, "-T")
		}
		args = append(args, "--", target, text)
		return runCommand(ctx, command{name: "ssh", args: args, in: in, out: output, errOut: errOut})
	}
	defer func() {
		c, stop := context.WithTimeout(context.Background(), 5*time.Second)
		defer stop()
		args := append(append([]string{}, sshOptions...), "-O", "exit", "--", target)
		_ = runCommand(c, command{name: "ssh", args: args, out: io.Discard, errOut: io.Discard})
	}()
	var output bytes.Buffer
	serverProbe := probe + `; if test -x /usr/local/bin/nanoweb-server; then /usr/local/bin/nanoweb-server -protocol || printf 'NANOWEB_OUTDATED=1\n'; elif test -e /usr/local/bin/nanoweb-server; then printf 'NANOWEB_OUTDATED=1\n'; else printf 'NANOWEB_MISSING=1\n'; fi`
	checkServer := func() error { output.Reset(); return ssh(ctx, "sh -c "+shellQuote(serverProbe), false, &output) }
	if err = checkServer(); err != nil {
		if errors.Is(err, errHostKeyChanged) {
			return err
		}
		return fmt.Errorf("Could not connect to or check %s.\n\nCheck your SSH username and key, then try again.\nTo change the saved username, run nanoweb deploy -configure.", target)
	}
	_, isRoot, err := parseProbe(output.String())
	if err != nil {
		return err
	}
	if strings.Contains(output.String(), "NANOWEB_MISSING=1\n") {
		fmt.Fprint(out, "\nNanoweb isn't installed on this server yet.\nSetup will install the server and start its service.\n\n")
		confirmed, err := confirmSetup(ctx, in, out)
		if err != nil {
			return err
		}
		if !confirmed {
			return fmt.Errorf("Deployment stopped. You can install the server later with nanoweb setup.")
		}
		if err = install(ctx, target); err != nil {
			return fmt.Errorf("Server setup did not finish.\n\n%w", err)
		}
		if err = checkServer(); err != nil {
			if errors.Is(err, errHostKeyChanged) {
				return err
			}
			return fmt.Errorf("Nanoweb was installed, but the server check failed.\nPlease check the service before deploying again.")
		}
		_, isRoot, err = parseProbe(output.String())
		if err != nil {
			return err
		}
	}
	if !strings.Contains(output.String(), release.Protocol+"\n") {
		return fmt.Errorf("This server has an older or incomplete Nanoweb installation.\nPlease update the server before deploying.")
	}
	fmt.Fprintln(out, "Checking DNS from the target server...")
	dnsCommand := "/usr/local/bin/nanoweb-server -check-dns -domain " + shellQuote(strings.Join(opts.Domains, ","))
	if len(s.ExpectedIPs) > 0 {
		dnsCommand += " -expected-ips " + shellQuote(strings.Join(s.ExpectedIPs, ","))
	}
	if err = ssh(ctx, dnsCommand, false, out); err != nil {
		return fmt.Errorf("remote DNS check failed; nothing uploaded or activated: %w", err)
	}
	latest, err := currentGit(ctx, root, true)
	if err != nil {
		return err
	}
	if latest.Branch != state.Branch || latest.Commit != state.Commit {
		return fmt.Errorf("branch or commit changed during deployment; retry")
	}
	if needsMail {
		if err = ensureMail(ctx, target, sshOptions, isRoot, in, out, errOut, runCommand); err != nil {
			return err
		}
	}
	text := "/usr/local/bin/nanoweb-server"
	if !httpsOnly {
		output.Reset()
		if err = ssh(ctx, "umask 077; mktemp -d /tmp/nanoweb-deploy.XXXXXXXXXX", false, &output); err != nil {
			return err
		}
		stage := strings.TrimSpace(output.String())
		if !regexp.MustCompile(`^/tmp/nanoweb-deploy\.[A-Za-z0-9]{10}$`).MatchString(stage) {
			return fmt.Errorf("unexpected remote staging path %q", stage)
		}
		defer func() {
			c, stop := context.WithTimeout(context.Background(), 5*time.Second)
			defer stop()
			_ = ssh(c, "rm -f -- "+shellQuote(stage+"/site.tar.gz")+"; rmdir -- "+shellQuote(stage), false, io.Discard)
		}()
		scpTarget := target
		if at := strings.LastIndex(target, "@"); strings.Contains(target[at+1:], ":") {
			scpTarget = target[:at+1] + "[" + target[at+1:] + "]"
		}
		fmt.Fprintln(out, "Uploading release...")
		args := append(append([]string{}, sshOptions...), "--", archive, scpTarget+":"+stage+"/site.tar.gz")
		if err = runCommand(ctx, command{name: "scp", args: args, in: in, out: out, errOut: errOut}); err != nil {
			return fmt.Errorf("upload failed: %w", err)
		}
		text += " -deploy " + shellQuote(stage+"/site.tar.gz") + " -name " + shellQuote(s.Name)
	}
	text += " -domain " + shellQuote(strings.Join(opts.Domains, ","))
	text += " -branch " + shellQuote(state.Branch) + " -commit " + shellQuote(state.Commit)
	if opts.Canonical != "" {
		text += " -canonical " + shellQuote(opts.Canonical)
	}
	if opts.HTTPS {
		text += " -https -email " + shellQuote(opts.Email) + " -agree-tos"
	}
	if !isRoot {
		text = "sudo -- " + text
	}
	latest, err = currentGit(ctx, root, true)
	if err != nil {
		return err
	}
	if latest.Branch != state.Branch || latest.Commit != state.Commit {
		return fmt.Errorf("branch or commit changed before activation; retry")
	}
	fmt.Fprintln(out, "Activating and checking release...")
	if opts.HTTPS {
		fmt.Fprintln(out, "Requesting/verifying HTTPS certificates. DNS must point here; allow inbound 80/443. This may take two minutes.")
	}
	if err = ssh(ctx, text, true, out); err != nil {
		return fmt.Errorf("remote activation failed: %w (if SSH disconnected, inspect the server; use sudo nanoweb-server -recover if a recovery journal remains)", err)
	}
	scheme := "http"
	if opts.HTTPS {
		scheme = "https"
	}
	fmt.Fprintf(out, "\nDeployed %s branch to %s://%s\n", state.Branch, scheme, s.Domain)
	fmt.Fprintln(out, "Remote readiness verified. Public reachability depends on DNS and firewall settings.")
	return nil
}
