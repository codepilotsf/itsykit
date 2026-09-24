package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"flea/internal/editor"
	"flea/internal/site"
)

func prepareLocalSessions(ctx context.Context, p *site.Project) error {
	if _, err := exec.LookPath("git"); err == nil {
		git := func(args ...string) *exec.Cmd {
			cmd := exec.CommandContext(ctx, "git", append([]string{"-C", p.Root}, args...)...)
			for _, entry := range os.Environ() {
				name, _, _ := strings.Cut(entry, "=")
				if !strings.HasPrefix(name, "GIT_") && name != "LC_ALL" {
					cmd.Env = append(cmd.Env, entry)
				}
			}
			cmd.Env = append(cmd.Env, "LC_ALL=C", "GIT_TERMINAL_PROMPT=0")
			return cmd
		}
		if out, err := git("rev-parse", "--show-toplevel").CombinedOutput(); err == nil {
			tracked, err := git("ls-files", "--", "_/sessions.json", "_/.sessions-*").Output()
			if err != nil {
				return fmt.Errorf("Could not check Git tracking for session files.")
			}
			if len(tracked) > 0 {
				return fmt.Errorf("Session files are tracked by Git. Remove _/sessions.json and _/.sessions-* from Git tracking before starting dev.")
			}
		} else if !strings.Contains(string(out), "not a git repository") {
			return fmt.Errorf("Could not check this site's Git repository before storing sessions.")
		}
	}
	if err := site.IgnoreSessionFiles(p.Root); err != nil {
		return err
	}
	return editor.PruneSessions(editor.Options{SessionPath: filepath.Join(p.Root, "_", "sessions.json"), Emails: p.Admin.Emails, SessionIdleDays: p.Admin.SessionIdleDays})
}
