package setup

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"itsy/internal/admin"
	"itsy/internal/release"
	"itsy/internal/site"
)

// DeploySettings contains no secrets. Native SSH configuration owns credentials.
type DeploySettings struct {
	WithContent     bool              `json:"-"`
	Target          string            `json:"target"`
	Domain          string            `json:"domain"`
	Name            string            `json:"name"`
	HTTPS           bool              `json:"https"`
	Email           string            `json:"email"`
	AgreeTOS        bool              `json:"agree_tos"`
	WWWRedirect     bool              `json:"www_redirect"`
	DomainRedirects map[string]string `json:"domain_redirects,omitempty"`
	ExpectedIPs     []string          `json:"expected_ips,omitempty"`
	Branch          string            `json:"-"`
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

func deployWithSetup(ctx context.Context, root string, s DeploySettings, httpsOnly bool, in io.Reader, out, errOut io.Writer, runCommand runner, install func(context.Context, string) error) (result error) {
	state, err := currentGit(ctx, root, true)
	if err != nil {
		return err
	}
	if s.Branch != "" && s.Branch != state.Branch {
		return fmt.Errorf("branch changed since selecting the deployment destination")
	}
	fmt.Fprintf(out, "\nDeploying %s branch to %s.\nPress Ctrl-C to abort.\n\n", state.Branch, s.Domain)
	if !httpsOnly && !s.WithContent {
		fmt.Fprintln(out, "Existing live Blog content will be kept. Styles come from this commit. New posts use their committed content.")
	}
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
	opts := admin.Options{Name: s.Name, Domains: s.domains(), HTTPS: s.HTTPS || httpsOnly, Email: s.Email, AgreeTOS: s.AgreeTOS, Branch: state.Branch, Commit: state.Commit, DomainRedirects: s.DomainRedirects}
	if s.WWWRedirect {
		opts.Canonical = s.Domain
	}
	if !httpsOnly {
		opts.Archive = "site.tar.gz"
	}
	if err = opts.Validate(); err != nil {
		return err
	}
	temp, err := os.MkdirTemp("", "flea-deploy-")
	if err != nil {
		return temporaryStorageError(os.TempDir(), err)
	}
	defer os.RemoveAll(temp)
	defer func() { result = explainTemporaryStorageError(temp, result) }()
	archive := filepath.Join(temp, "site.tar.gz")
	snapshot := ""
	needsMail := false
	var localProject *site.Project
	planFile := ""
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
		if !p.SiteNameConfigured {
			return fmt.Errorf("Set app.site_name in itsy.config.json, or run itsy dev to choose a site name, then commit it before deploying.")
		}
		localProject = p
		// Admin login, Contact forms and every session-owning custom app require
		// HTTPS. The built-in Blog without admin access is public read-only, so it
		// keeps the previous HTTP allowance.
		needsForms := len(p.Admin.Emails) > 0
		needsMail = len(p.Admin.Emails) > 0
		for name := range p.Apps {
			if name == "blog" {
				continue
			}
			needsForms = true
			if name == "contact" {
				needsMail = true
			}
		}
		if needsForms && !opts.HTTPS {
			return fmt.Errorf("This site requires HTTPS for admin login and app forms. Enable it with itsy deploy -configure before deploying this site.")
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
		snapshot = filepath.Join(temp, "snapshot")
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
	serverProbe := probe + `; if test -x /usr/local/bin/flea-server; then /usr/local/bin/flea-server -protocol || printf 'FLEA_OUTDATED=1\n'; elif test -e /usr/local/bin/flea-server; then printf 'FLEA_OUTDATED=1\n'; else printf 'FLEA_MISSING=1\n'; fi`
	checkServer := func() error { output.Reset(); return ssh(ctx, "sh -c "+shellQuote(serverProbe), false, &output) }
	if err = checkServer(); err != nil {
		if errors.Is(err, errHostKeyChanged) {
			return err
		}
		return fmt.Errorf("Could not connect to or check %s.\n\nCheck your SSH username and key, then try again.\nTo change the saved username, run itsy deploy -configure.", target)
	}
	arch, isRoot, err := parseProbe(output.String())
	if err != nil {
		return err
	}
	// Portable package preflight: the target server's architecture is known now,
	// so a package that declares only the development platform is rejected
	// before anything is uploaded. Check the immutable staged snapshot that is
	// about to be uploaded, not the later mutable working tree.
	if !httpsOnly {
		if err = release.CheckAppPlatforms(snapshot, "linux", arch); err != nil {
			return err
		}
	}
	if strings.Contains(output.String(), "FLEA_MISSING=1\n") {
		fmt.Fprint(out, "\nItsyWeb isn't installed on this server yet.\nSetup will install the server and start its service.\n\n")
		confirmed, err := confirmSetup(ctx, in, out)
		if err != nil {
			return err
		}
		if !confirmed {
			return fmt.Errorf("Deployment stopped. You can install the server later with itsy setup.")
		}
		if err = install(ctx, target); err != nil {
			return fmt.Errorf("Server setup did not finish.\n\n%w", err)
		}
		if err = checkServer(); err != nil {
			if errors.Is(err, errHostKeyChanged) {
				return err
			}
			return fmt.Errorf("ItsyWeb was installed, but the server check failed.\nPlease check the service before deploying again.")
		}
		_, isRoot, err = parseProbe(output.String())
		if err != nil {
			return err
		}
	}
	if !strings.Contains(output.String(), release.Protocol+"\n") {
		return fmt.Errorf("This server has an older or incomplete ItsyWeb installation.\nPlease update the server before deploying.")
	}
	if s.WithContent {
		if httpsOnly {
			return fmt.Errorf("content options require itsy deploy")
		}
		s.Branch = state.Branch
		base, e := readBaseline(ctx, root, s)
		if e != nil {
			return e
		}
		remote, e := fetchState(ctx, s, ssh, isRoot)
		if e != nil {
			return e
		}
		plan, e := makePlan(base, localProject, remote, s.WithContent, resolveSync(ctx, in, out))
		if e != nil {
			return e
		}
		b, e := json.Marshal(plan)
		if e != nil {
			return e
		}
		planFile = filepath.Join(temp, "sync.json")
		if e = os.WriteFile(planFile, b, 0600); e != nil {
			return e
		}
	}
	fmt.Fprintln(out, "Checking DNS from the target server...")
	dnsCommand := "/usr/local/bin/flea-server -check-dns -domain " + shellQuote(strings.Join(opts.Domains, ","))
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
		if err = ensureMail(ctx, root, target, sshOptions, isRoot, in, out, errOut, runCommand); err != nil {
			return err
		}
	}
	text := "/usr/local/bin/flea-server"
	if !httpsOnly {
		output.Reset()
		if err = ssh(ctx, "umask 077; mktemp -d /tmp/flea-deploy.XXXXXXXXXX", false, &output); err != nil {
			return err
		}
		stage := strings.TrimSpace(output.String())
		if !regexp.MustCompile(`^/tmp/flea-deploy\.[A-Za-z0-9]{10}$`).MatchString(stage) {
			return fmt.Errorf("unexpected remote staging path %q", stage)
		}
		defer func() {
			c, stop := context.WithTimeout(context.Background(), 5*time.Second)
			defer stop()
			_ = ssh(c, "rm -f -- "+shellQuote(stage+"/site.tar.gz")+" "+shellQuote(stage+"/sync.json")+"; rmdir -- "+shellQuote(stage), false, io.Discard)
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
		text += " -deploy " + shellQuote(stage+"/site.tar.gz")
		if planFile != "" {
			args := append(append([]string{}, sshOptions...), "--", planFile, scpTarget+":"+stage+"/sync.json")
			if e := runCommand(ctx, command{name: "scp", args: args, in: in, out: out, errOut: errOut}); e != nil {
				return e
			}
			text += " -sync-plan " + shellQuote(stage+"/sync.json")
		}
	}
	text += " -name " + shellQuote(s.Name) + " -domain " + shellQuote(strings.Join(opts.Domains, ","))
	if len(opts.DomainRedirects) > 0 {
		rules, err := json.Marshal(opts.DomainRedirects)
		if err != nil {
			return err
		}
		text += " -domain-redirects " + shellQuote(string(rules))
	}
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
	if err = ssh(ctx, text, true, out); err != nil {
		return fmt.Errorf("remote activation failed: %w (if SSH disconnected, inspect the server; use sudo flea-server -recover if a recovery journal remains)", err)
	}
	scheme := "http"
	if opts.HTTPS {
		scheme = "https"
	}
	fmt.Fprintf(out, "\nDeployed %s branch to %s://%s\n", state.Branch, scheme, s.Domain)
	fmt.Fprintln(out, "Remote readiness verified. Public reachability depends on DNS and firewall settings.")
	return nil
}
