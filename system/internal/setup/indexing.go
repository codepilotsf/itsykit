package setup

import (
	"context"
	"fmt"
	"io"

	"itsy/internal/admin"
	"itsy/internal/host"
)

// ServerIndexing configures one exact hostname, independent of website files or
// branch destinations. Review and confirmation precede the locked remote change.
func ServerIndexing(ctx context.Context, target, domain, mode string, in io.Reader, out, errOut io.Writer) error {
	return serverIndexing(ctx, target, domain, mode, in, out, errOut, execute)
}

func serverIndexing(ctx context.Context, target, domain, mode string, in io.Reader, out, errOut io.Writer, run runner) error {
	if !host.ValidHost(domain) || !host.ValidIndexingMode(mode) {
		return fmt.Errorf("use an exact lowercase hostname without a port, and mode noindex or allow")
	}
	target, err := promptServerTarget(ctx, target, in, out)
	if err != nil {
		return err
	}
	return withServerAdmin(ctx, target, in, out, errOut, run, func(r serverAdmin) error {
		sites, err := fetchSites(ctx, r)
		if err != nil {
			return err
		}
		var info *admin.SiteSummary
		for i := range sites {
			for _, d := range sites[i].Domains {
				if d == domain {
					info = &sites[i]
				}
			}
		}
		if info == nil {
			return fmt.Errorf("hostname %s is not registered on this server", domain)
		}
		if !info.Managed {
			return fmt.Errorf("hostname %s is not managed by this server", domain)
		}
		fmt.Fprintf(out, "\nIndexing on %s\n", r.target)
		printSiteSummary(out, *info)
		fmt.Fprintf(out, "\nSet %s to %s?\n", domain, mode)
		if mode == "noindex" {
			fmt.Fprintln(out, "Responses will send X-Robots-Tag: noindex, nofollow. Keep robots.txt crawlable so crawlers can discover it.")
		} else {
			fmt.Fprintln(out, "ItsyKit will add no indexing restriction. Owner-authored headers and meta tags still apply.")
		}
		fmt.Fprintln(out, "This restarts the shared service; other sites are briefly unavailable.")
		answer, err := ask(ctx, in, out, "Type the hostname to confirm", "")
		if err != nil {
			return err
		}
		if answer != domain {
			return fmt.Errorf("cancelled; nothing changed")
		}
		command := "/usr/local/bin/flea-server -set-indexing " + shellQuote(domain) + " -indexing-mode " + shellQuote(mode) + " -expected-state " + shellQuote(info.State)
		if !r.root {
			command = "sudo -- " + command
		}
		if err = r.ssh(ctx, command, !r.root, out); err != nil {
			return err
		}
		// Fetch the new effective policy and any crawl-block diagnostic.
		updated, err := fetchSites(ctx, r)
		if err != nil {
			return fmt.Errorf("policy updated, but inspection failed: %w; run itsy server list", err)
		}
		for _, item := range updated {
			if item.Name == info.Name {
				printSiteSummary(out, item)
			}
		}
		return nil
	})
}

func printIndexing(out io.Writer, info admin.SiteSummary) {
	for _, domain := range info.Domains {
		mode := info.Indexing[domain]
		if mode == "" {
			mode = "allow"
		}
		label := "allow (no restriction added)"
		if mode == "noindex" {
			label = "noindex, nofollow"
		}
		fmt.Fprintf(out, "  indexing %s: %s\n", domain, label)
	}
	for _, warning := range info.Warnings {
		fmt.Fprintf(out, "  Warning: %s\n", warning)
	}
}
