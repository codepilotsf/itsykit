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

	"itsy/internal/admin"
	"itsy/internal/release"
)

// serverAdmin is an authenticated SSH session to an installed ItsyWeb host,
// independent of any local site, Git repository or saved deployment settings.
type serverAdmin struct {
	target string
	ssh    func(context.Context, string, bool, io.Writer) error
	root   bool
}

// ServerList lists the sites installed on an SSH target. An empty target is
// prompted for. It never requires a local site and never mutates remote state.
func ServerList(ctx context.Context, target string, in io.Reader, out, errOut io.Writer) error {
	return serverList(ctx, target, in, out, errOut, execute)
}

func serverList(ctx context.Context, target string, in io.Reader, out, errOut io.Writer, run runner) error {
	target, err := promptServerTarget(ctx, target, in, out)
	if err != nil {
		return err
	}
	return withServerAdmin(ctx, target, in, out, errOut, run, func(r serverAdmin) error {
		sites, err := fetchSites(ctx, r)
		if err != nil {
			return err
		}
		printServerSites(out, r.target, sites)
		return nil
	})
}

// fetchSites reads the server's read-only inventory. No local project or saved
// deployment settings are involved.
func fetchSites(ctx context.Context, r serverAdmin) ([]admin.SiteSummary, error) {
	raw, err := captureAdmin(ctx, "/usr/local/bin/flea-server -list-sites", r.ssh, r.root)
	if err != nil {
		return nil, err
	}
	var sites []admin.SiteSummary
	if err = json.Unmarshal(raw, &sites); err != nil {
		return nil, fmt.Errorf("could not read the server's site list")
	}
	return sites, nil
}

// ServerRemove unregisters one site's routing after an exact typed confirmation.
// An empty target or site name is prompted for; a prompted name is chosen from
// the managed sites installed on the server. It never deletes releases, content,
// media, sessions, certificates or downloads.
func ServerRemove(ctx context.Context, target, name string, in io.Reader, out, errOut io.Writer) error {
	return serverRemove(ctx, target, name, in, out, errOut, execute)
}

func serverRemove(ctx context.Context, target, name string, in io.Reader, out, errOut io.Writer, run runner) error {
	if name != "" && !admin.ValidSiteName(name) {
		return fmt.Errorf("invalid site name %q; use 1-64 letters, digits, hyphens or underscores", name)
	}
	target, err := promptServerTarget(ctx, target, in, out)
	if err != nil {
		return err
	}
	return withServerAdmin(ctx, target, in, out, errOut, run, func(r serverAdmin) error {
		if name == "" {
			name, err = selectManagedSite(ctx, r, in, out)
			if err != nil {
				return err
			}
		}
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

// promptServerTarget asks for the SSH login only when none was given. The exact
// native target is kept, including SSH aliases and bare hosts.
func promptServerTarget(ctx context.Context, target string, in io.Reader, out io.Writer) (string, error) {
	if strings.TrimSpace(target) != "" {
		return target, nil
	}
	answer, err := ask(ctx, in, out, "Server login (for example root@fleacms.com)", "")
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(answer) == "" {
		return "", fmt.Errorf("enter a server login, for example root@fleacms.com")
	}
	return answer, nil
}

// selectManagedSite lists the managed sites and asks for one by its exact name.
// Unmanaged sites are excluded and there are no numeric choices, so every valid
// site name (including one made only of digits) is unambiguous. Empty or unknown
// input re-prompts; EOF and cancellation return before any mutation.
func selectManagedSite(ctx context.Context, r serverAdmin, in io.Reader, out io.Writer) (string, error) {
	sites, err := fetchSites(ctx, r)
	if err != nil {
		return "", err
	}
	managed := make([]admin.SiteSummary, 0, len(sites))
	for _, site := range sites {
		if site.Managed {
			managed = append(managed, site)
		}
	}
	if len(managed) == 0 {
		return "", fmt.Errorf("no managed sites are installed on %s; nothing to remove", r.target)
	}
	fmt.Fprintf(out, "\nManaged sites on %s:\n", r.target)
	for _, site := range managed {
		fmt.Fprintf(out, "  %s  %s\n", site.Name, siteDomainLabel(site))
	}
	for {
		answer, err := ask(ctx, in, out, "\nSelect a site by exact name", "")
		if err != nil {
			return "", err
		}
		answer = strings.TrimSpace(answer)
		if answer == "" {
			fmt.Fprintln(out, "Enter the exact name of a listed site.")
			continue
		}
		for _, site := range managed {
			if site.Name == answer {
				return site.Name, nil
			}
		}
		fmt.Fprintf(out, "No managed site named %q. Enter the exact name of a listed site.\n", answer)
	}
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
		return fmt.Errorf("ItsyWeb isn't installed on %s.", target)
	}
	if strings.Contains(raw, "FLEA_OUTDATED=1") || !strings.Contains(raw, release.Protocol+"\n") {
		return fmt.Errorf("This server has an older or incomplete ItsyWeb installation.\nUpdate the server before managing its sites; this command does not rewrite remote configuration by hand.")
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
	for i, site := range sites {
		if i > 0 {
			fmt.Fprintln(out)
		}
		printSiteSummary(out, site)
	}
	fmt.Fprintln(out)
}

// printSiteSummary shows the site's name and branch heading first, then its
// HTTPS domains, then HTTP-only domains, then any unmanaged warning, and
// redirects last. Domains are labelled by scheme under the heading and never
// repeated as plain lines; the release is omitted.
func printSiteSummary(out io.Writer, site admin.SiteSummary) {
	if site.Branch != "" {
		fmt.Fprintf(out, "%s  (branch %s)\n", site.Name, site.Branch)
	} else {
		fmt.Fprintf(out, "%s\n", site.Name)
	}
	https, http := splitDomains(site)
	if len(https) > 0 {
		fmt.Fprintf(out, "  https %s\n", strings.Join(https, ", "))
	}
	if len(http) > 0 {
		fmt.Fprintf(out, "  http %s\n", strings.Join(http, ", "))
	}
	if !site.Managed {
		fmt.Fprintln(out, "  not managed by this server")
	}
	printIndexing(out, site)
	for _, redirect := range site.Redirects {
		fmt.Fprintf(out, "  redirect %s\n", redirect)
	}
}

// splitDomains separates a site's domains into HTTPS and HTTP-only groups,
// preserving the sorted order returned by the server.
func splitDomains(site admin.SiteSummary) (https, http []string) {
	secure := make(map[string]bool, len(site.HTTPS))
	for _, domain := range site.HTTPS {
		secure[domain] = true
	}
	for _, domain := range site.Domains {
		if secure[domain] {
			https = append(https, domain)
		} else {
			http = append(http, domain)
		}
	}
	return https, http
}

// siteDomainLabel is one compact scheme-labelled summary for selection lists.
func siteDomainLabel(site admin.SiteSummary) string {
	https, http := splitDomains(site)
	var parts []string
	if len(https) > 0 {
		parts = append(parts, "https "+strings.Join(https, ", "))
	}
	if len(http) > 0 {
		parts = append(parts, "http "+strings.Join(http, ", "))
	}
	if len(parts) == 0 {
		return "(no domains)"
	}
	return strings.Join(parts, "; ")
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
	fmt.Fprintln(out, "\nThis only unregisters routing; it deletes no files. Releases, content, media,\nsessions and certificates stay on disk. Removal rewrites the shared server\nconfig and restarts the service. With socket activation, new connections queue\nthrough the restart; other installations are briefly unavailable. When no\nHTTPS hostnames remain afterwards, automatic HTTPS is disabled; cached\ncertificates stay on disk.")
	fmt.Fprintln(out, "\nRetained files are not restored automatically. Re-deploying the same site name\nstarts from the current project and cannot recover content left behind. Back it\nup or pull it before removing the site if a later deployment must keep it.")
}
