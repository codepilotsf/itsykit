package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"flea/internal/setup"
	"flea/internal/site"
)

func devTerminal(ctx context.Context, in *os.File) bool {
	cmd := exec.CommandContext(ctx, "stty", "-g")
	cmd.Stdin = in
	return cmd.Run() == nil
}

// Only interactive dev startup offers onboarding. Explicit commands remain
// available for later setup; missing credentials never create a login bypass.
func prepareDevOwner(ctx context.Context, p *site.Project, in io.Reader, out io.Writer, interactive bool) (bool, error) {
	if !interactive {
		return false, nil
	}
	changed := false
	if !p.SiteNameConfigured {
		name, err := setup.Prompt(ctx, in, out, "Site name ["+fmt.Sprint(p.App["site_name"])+"]")
		if err != nil {
			return false, err
		}
		if name == "" {
			name = fmt.Sprint(p.App["site_name"])
		}
		if err := ctx.Err(); err != nil {
			return false, err
		}
		if err := site.SetSiteName(p.Root, name); err != nil {
			return false, err
		}
		changed = true
		fmt.Fprintln(out, "Config file updated.")
	}
	if len(p.Admin.Emails) == 0 {
		for {
			address, err := setup.Prompt(ctx, in, out, "Admin email (optional)")
			if err != nil {
				return false, err
			}
			if address == "" {
				break
			}
			if err := ctx.Err(); err != nil {
				return false, err
			}
			_, _, err = site.AddAdmin(p.Root, address)
			if err != nil {
				return false, err
			}
			changed = true
			fmt.Fprintln(out, "Config file updated.")
			break
		}
	}
	return changed, nil
}
