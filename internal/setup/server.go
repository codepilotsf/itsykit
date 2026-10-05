package setup

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"flea/internal/admin"
	"flea/internal/release"
)

// serverAdmin is an authenticated SSH session to an installed FleaCMS host,
// independent of any local site, Git repository or saved deployment settings.
type serverAdmin struct {
	target string
	ssh    func(context.Context, string, bool, io.Writer) error
	root   bool
}

// ServerSites lists the sites installed on an explicit SSH target. It never
// requires a local site and never mutates remote state.
func ServerSites(ctx context.Context, target string, in io.Reader, out, errOut io.Writer) error {
	return serverSites(ctx, target, in, out, errOut, execute)
}

func serverSites(ctx context.Context, target string, in io.Reader, out, errOut io.Writer, run runner) error {
	return withServerAdmin(ctx, target, in, out, errOut, run, func(r serverAdmin) error {
		raw, err := captureAdmin(ctx, "/usr/local/bin/flea-server -list-sites", r.ssh, r.root)
		if err != nil {
			return err
		}
		var sites []admin.SiteSummary
		if err = json.Unmarshal(raw, &sites); err != nil {
			return fmt.Errorf("could not read the server's site list")
		}
		printServerSites(out, r.target, sites)
		return nil
	})
}

// ServerRemove unregisters one site's routing after an exact typed confirmation.
// It never deletes releases, content, media, sessions, certificates or downloads.
func ServerRemove(ctx context.Context, target, name string, in io.Reader, out, errOut io.Writer) error {
	return serverRemove(ctx, target, name, in, out, errOut, execute)
}

func serverRemove(ctx context.Context, target, name string, in io.Reader, out, errOut io.Writer, run runner) error {
	if !admin.ValidSiteName(name) {
		return fmt.Errorf("invalid site name %q; use 1-64 letters, digits, hyphens or underscores", name)
	}
	return withServerAdmin(ctx, target, in, out, errOut, run, func(r serverAdmin) error {
		raw, err := captureAdmin(ctx, "/usr/local/bin/flea-server -site-info "+shellQuote(name), r.ssh, r.root)
		if err != nil {
			return err
		}
		var info admin.SiteSummary
		if err = json.Unmarshal(raw, &info); err != nil {
			return fmt.Errorf("could not read the site's current state")
		}
		if !info.Managed {
			return fmt.Errorf("site %q is not managed by this server; refusing to change it", name)
		}
		printSiteReview(out, r.target, info)
		answer, err := ask(ctx, in, out, "Type the site name to confirm", "")
		if err != nil {
			return err
		}
		if answer != name {
			return fmt.Errorf("cancelled; nothing changed")
		}
		command := "/usr/local/bin/flea-server -remove-site " + shellQuote(name) + " -expected-state " + shellQuote(info.State)
		if !r.root {
			command = "sudo -- " + command
		}
		return r.ssh(ctx, command, !r.root, out)
	})
}

// withServerAdmin verifies SSH, the installed binary and the deployment
// protocol before running fn. An older server is reported, never bypassed.
func withServerAdmin(ctx context.Context, target string, in io.Reader, out, errOut io.Writer, run runner, fn func(serverAdmin) error) error {
	target, err := targetName(target)
	if err != nil {
		return err
	}
	dir, err := os.MkdirTemp("", "flea-server-")
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
	var probeOutput bytes.Buffer
	script := `set -eu; printf 'FLEA_UID=%s\n' "$(id -u)"; if test -x /usr/local/bin/flea-server; then /usr/local/bin/flea-server -protocol || printf 'FLEA_OUTDATED=1\n'; elif test -e /usr/local/bin/flea-server; then printf 'FLEA_OUTDATED=1\n'; else printf 'FLEA_MISSING=1\n'; fi`
	if err = ssh(ctx, script, false, &probeOutput); err != nil {
		if errors.Is(err, errHostKeyChanged) {
			return err
		}
		return fmt.Errorf("Could not connect to or check %s.\n\nCheck your SSH username and key, then try again.", target)
	}
	raw := probeOutput.String()
	if strings.Contains(raw, "FLEA_MISSING=1") {
		return fmt.Errorf("FleaCMS isn't installed on %s.", target)
	}
	if strings.Contains(raw, "FLEA_OUTDATED=1") || !strings.Contains(raw, release.Protocol+"\n") {
		return fmt.Errorf("This server has an older or incomplete FleaCMS installation.\nUpdate the server before managing its sites; this command does not rewrite remote configuration by hand.")
	}
	root := strings.Contains(raw, "FLEA_UID=0\n")
	return fn(serverAdmin{target: target, ssh: ssh, root: root})
}

func printServerSites(out io.Writer, target string, sites []admin.SiteSummary) {
	if len(sites) == 0 {
		fmt.Fprintf(out, "No sites are registered on %s.\n", target)
		return
	}
	fmt.Fprintf(out, "Sites on %s:\n", target)
	for _, site := range sites {
		fmt.Fprintln(out)
		if site.Branch != "" {
			fmt.Fprintf(out, "%s  (branch %s)\n", site.Name, site.Branch)
		} else {
			fmt.Fprintf(out, "%s\n", site.Name)
		}
		for _, domain := range site.Domains {
			fmt.Fprintf(out, "  %s\n", domain)
		}
		if site.Release != "" {
			fmt.Fprintf(out, "  release %s\n", site.Release)
		}
		for _, redirect := range site.Redirects {
			fmt.Fprintf(out, "  redirect %s\n", redirect)
		}
		if len(site.HTTPS) > 0 {
			fmt.Fprintf(out, "  https %s\n", strings.Join(site.HTTPS, ", "))
		}
		if !site.Managed {
			fmt.Fprintln(out, "  not managed by this server")
		}
	}
	fmt.Fprintln(out)
}

func printSiteReview(out io.Writer, target string, info admin.SiteSummary) {
	fmt.Fprintf(out, "\nRemove site %q from %s?\n\n", info.Name, target)
	fmt.Fprintln(out, "Domains to unregister:")
	for _, domain := range info.Domains {
		fmt.Fprintf(out, "  %s\n", domain)
	}
	if len(info.Redirects) > 0 {
		fmt.Fprintln(out, "Redirects:")
		for _, redirect := range info.Redirects {
			fmt.Fprintf(out, "  %s\n", redirect)
		}
	}
	if len(info.HTTPS) > 0 {
		fmt.Fprintf(out, "HTTPS certificates: %s\n", strings.Join(info.HTTPS, ", "))
	}
	if info.Release != "" {
		fmt.Fprintf(out, "Current release: %s\n", info.Release)
	}
	if info.Retained != "" {
		fmt.Fprintf(out, "Retained files: %s\n", info.Retained)
	}
	fmt.Fprintln(out, "\nThis only unregisters routing; it deletes no files. Releases, content, media,\nsessions and certificates stay on disk. Removal rewrites the shared server\nconfig and restarts the service, so other sites are briefly unavailable. When no\nHTTPS hostnames remain afterwards, automatic HTTPS is disabled; cached\ncertificates stay on disk.")
	fmt.Fprintln(out, "\nRetained files are not restored automatically. Re-deploying the same site name\nstarts from the current project and cannot recover content left behind. Back it\nup or pull it before removing the site if a later deployment must keep it.")
}
