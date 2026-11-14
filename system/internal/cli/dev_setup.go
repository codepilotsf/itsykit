package cli

import (
	"context"
	"fmt"
	"io"
	"itsy/internal/sitefiles"
	"os"
	"os/exec"

	"itsy/internal/setup"
	"itsy/internal/site"
)

func devTerminal(ctx context.Context, in *os.File) bool {
	cmd := exec.CommandContext(ctx, "stty", "-g")
	cmd.Stdin = in
	return cmd.Run() == nil
}

// Only interactive dev startup offers onboarding, and only on the first run
// (before the site name is saved). Explicit commands remain available for later
// setup; missing credentials never create a login bypass. The second return
// value reports whether the admin prompt was shown this run, so the caller can
// avoid repeating its reminder immediately afterwards.
func prepareDevOwner(ctx context.Context, p *site.Project, in io.Reader, out io.Writer, interactive bool) (bool, bool, error) {
	if !interactive {
		return false, false, nil
	}
	firstRun := !p.SiteNameConfigured
	changed := false
	if firstRun {
		fmt.Fprintln(out)
		name, err := setup.Prompt(ctx, in, out, "Site name ["+fmt.Sprint(p.App["site_name"])+"]")
		if err != nil {
			return false, false, err
		}
		if name == "" {
			name = fmt.Sprint(p.App["site_name"])
		}
		if err := ctx.Err(); err != nil {
			return false, false, err
		}
		if err := site.SetSiteName(p.Root, name); err != nil {
			return false, false, err
		}
		changed = true
		fmt.Fprintln(out, "Config file updated.")
	}
	offered := false
	if firstRun && len(p.Admin.Emails) == 0 {
		offered = true
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Add an admin who can sign in and edit your site. You can skip this and add one later.")
		address, err := setup.Prompt(ctx, in, out, "Admin email (optional)")
		if err != nil {
			return false, offered, err
		}
		if address == "" {
			fmt.Fprintln(out, "No problem. Add an admin whenever you're ready with: itsy admin")
			return changed, offered, nil
		}
		if err := ctx.Err(); err != nil {
			return false, offered, err
		}
		email, added, err := site.AddAdmin(p.Root, address)
		if err != nil {
			return false, offered, err
		}
		if added {
			changed = true
			fmt.Fprintf(out, "Added %s to %s.\n", email, sitefiles.ResolveAt(p.Root, sitefiles.ConfigName))
		} else {
			fmt.Fprintf(out, "%s already has admin access in this site's configuration. Nothing changed.\n", email)
		}
	}
	return changed, offered, nil
}
