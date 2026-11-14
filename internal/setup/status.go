package setup

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"flea/internal/admin"
	"flea/internal/site"
	"flea/internal/snips"
	"flea/internal/syncstate"
)

// Status validates an already-loaded project and reads its branch's deployment.
// It never configures a destination, pulls content, or updates a baseline.
func Status(ctx context.Context, p *site.Project, in io.Reader, out, errOut io.Writer, terminal bool) error {
	return status(ctx, p, in, out, errOut, terminal, execute, checkWebsite)
}

func status(ctx context.Context, p *site.Project, in io.Reader, out, errOut io.Writer, terminal bool, run runner, check func(context.Context, string) error) error {
	index, err := p.Content.Index()
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "%s\n%s\n\nLocal\n  Files        Valid configuration, templates and content\n  Site         %s, %s, %s\n", display(fmt.Sprint(p.App["site_name"])), display(p.Root), quantity(len(p.Pages), "page"), quantity(len(p.Assets), "asset"), quantity(len(index), "editable area"))
	for _, warning := range p.Warnings {
		fmt.Fprintf(errOut, "Warning: %s\n", warning)
	}
	state, gitErr := currentGit(ctx, p.Root, false)
	if gitErr != nil {
		label := "No Git repository"
		if commit, e := git(ctx, p.Root, "rev-parse", "--verify", "HEAD^{commit}"); e == nil {
			label = "Detached HEAD, " + shortCommit(strings.TrimSpace(commit))
		}
		fmt.Fprintf(out, "  Git          %s\n\nDeployment\n  No named branch; remote check skipped.\n\nNext: run itsy dev to preview your site.\n", label)
		return ctx.Err()
	}
	dirty, err := git(ctx, p.Root, "--no-optional-locks", "status", "--porcelain=v1", "-z", "--untracked-files=all", "--ignore-submodules=none")
	if err != nil {
		return fmt.Errorf("could not inspect local Git changes: %w", err)
	}
	changes := countChanges(dirty)
	description := "no uncommitted changes"
	if changes > 0 {
		description = fmt.Sprintf("%d uncommitted changes", changes)
	}
	if state.Commit == "" {
		description += ", no commits yet"
	}
	fmt.Fprintf(out, "  Git          %s, %s\n", display(state.Branch), description)
	config, err := readDeploymentConfig(p.Root)
	if err != nil {
		return err
	}
	settings, exists := config.Branches[state.Branch]
	if !exists {
		fmt.Fprintf(out, "\nDeployment\n  No remote site configured for %s.\n\nNext: run itsy dev to preview your site.\n", display(state.Branch))
		return ctx.Err()
	}
	if err = validateMapping(state.Branch, settings, config); err != nil {
		return err
	}
	settings.Branch = state.Branch
	fmt.Fprintln(out)
	progress := newCommandProgress(out, terminal, "Checking remote...")
	defer progress.stop()
	// Stop animation before native SSH/sudo diagnostics or password prompts.
	diagnostic := progressOutput{Writer: errOut, stop: progress.stop}
	wrapped := func(ctx context.Context, c command) error {
		for _, arg := range c.args {
			if arg == "-tt" {
				progress.stop()
				break
			}
		}
		return run(ctx, c)
	}
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	var remote syncstate.Snapshot
	var record admin.ReleaseRecord
	var websiteErr error
	var liveInfo admin.SiteSummary
	baseline, err := readBaseline(ctx, p.Root, settings)
	if err == nil && baseline.Snips != nil {
		var raw []byte
		raw, err = baseline.Snips.Bytes()
		if err == nil {
			_, err = snips.Decode(raw)
		}
		if err != nil {
			err = fmt.Errorf("the content pull baseline is invalid; review it before comparing: %w", err)
		}
	}
	if err == nil {
		err = withServerAdmin(ctx, settings.Target, in, out, diagnostic, wrapped, func(r serverAdmin) error {
			var e error
			remote, e = fetchState(ctx, settings, r.ssh, r.root)
			if e != nil || !remote.Exists {
				return e
			}
			inventory, e := fetchSites(ctx, r)
			if e != nil {
				return e
			}
			found := map[string]bool{}
			for _, item := range inventory {
				if item.Name == settings.Name && item.Branch == state.Branch && item.Managed && item.Release == filepath.Base(remote.Revision) {
					liveInfo = item
					for _, domain := range item.Domains {
						found[domain] = true
					}
				}
			}
			for _, domain := range admin.Domains(settings.Domain) {
				if !found[domain] {
					return fmt.Errorf("the saved domain or active deployment changed; check the branch's deployment settings and retry")
				}
			}
			raw, e := captureAdmin(ctx, "/usr/local/bin/flea-server -history -name "+shellQuote(settings.Name)+" -branch "+shellQuote(settings.Branch), r.ssh, r.root)
			if e != nil {
				return e
			}
			var records []admin.ReleaseRecord
			if e = json.Unmarshal(raw, &records); e != nil {
				return fmt.Errorf("could not read deployment history")
			}
			if len(records) == 0 || records[0].ID != filepath.Base(remote.Revision) || records[0].Branch != state.Branch || !regexp.MustCompile(`^([0-9a-f]{40}|[0-9a-f]{64})$`).MatchString(records[0].Commit) {
				return fmt.Errorf("deployment history does not match the live snapshot; run itsy status again")
			}
			record = records[0]
			for _, url := range statusURLs(settings) {
				if e := check(ctx, url); e != nil {
					websiteErr = fmt.Errorf("%s: %w", url, e)
					break
				}
			}
			return nil
		})
	}
	progress.stop()
	fmt.Fprintf(out, "Deployment\n  Website      %s\n  Server       %s\n", statusURL(settings), display(settings.Target))
	if err != nil {
		fmt.Fprintln(out, "  Comparison   Unavailable\n\nLocal validation passed. Remote state could not be checked.")
		return err
	}
	if !remote.Exists {
		fmt.Fprintln(out, "  State        Not deployed\n\nNext: run itsy deploy to publish this branch.")
		return nil
	}
	printIndexing(out, liveInfo)
	if websiteErr == nil {
		fmt.Fprintln(out, "  Connection   Reachable")
	} else {
		fmt.Fprintf(out, "  Connection   Unavailable (%s)\n", display(websiteErr.Error()))
	}
	code := "Matches latest local commit"
	if state.Commit == "" {
		code = "No local commit to compare"
	} else if record.Commit != state.Commit {
		code = "Different commits (local " + shortCommit(state.Commit) + ", live " + shortCommit(record.Commit) + ")"
	}
	fmt.Fprintf(out, "  Code         %s\n", code)
	if changes > 0 {
		fmt.Fprintf(out, "  Local files  %d uncommitted changes, not deployed\n", changes)
	}
	content := syncstate.CompareContent(baseline.Snips, p.Content, remote.Snips)
	fmt.Fprintf(out, "  Content      %s\n", content.String())
	switch {
	case content.Different > 0:
		fmt.Fprintln(out, "\nNext: run itsy pull to review content differences.")
	case changes > 0 || state.Commit != record.Commit:
		fmt.Fprintln(out, "\nNext: review your local changes, then run itsy deploy to publish code.")
	default:
		fmt.Fprintln(out, "\nCode and content match. Styles are part of the code release.")
	}
	if websiteErr != nil {
		return fmt.Errorf("the public website could not be reached; check its DNS, TLS and server availability")
	}
	return nil
}

func statusURL(s DeploySettings) string { return strings.Join(statusURLs(s), ", ") }
func statusURLs(s DeploySettings) []string {
	scheme := "http"
	if s.HTTPS {
		scheme = "https"
	}
	var urls []string
	for _, domain := range admin.Domains(s.Domain) {
		urls = append(urls, scheme+"://"+domain)
	}
	return urls
}
func quantity(n int, noun string) string {
	if n != 1 {
		noun += "s"
	}
	return fmt.Sprintf("%d %s", n, noun)
}
func checkWebsite(ctx context.Context, url string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return err
	}
	client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("HTTP %d", res.StatusCode)
	}
	return nil
}
func shortCommit(s string) string {
	if len(s) > 8 {
		return s[:8]
	}
	return s
}

// Config and branch values may contain terminal controls; never render them raw.
func display(s string) string {
	return strings.Map(func(r rune) rune {
		if r < 32 || r == 127 || (r >= 128 && r <= 159) {
			return ' '
		}
		return r
	}, s)
}
func countChanges(raw string) int {
	entries := strings.Split(raw, "\x00")
	count := 0
	for i := 0; i < len(entries); i++ {
		if entries[i] == "" {
			continue
		}
		count++
		if len(entries[i]) >= 2 && (entries[i][0] == 'R' || entries[i][0] == 'C' || entries[i][1] == 'R' || entries[i][1] == 'C') {
			i++
		}
	}
	return count
}
