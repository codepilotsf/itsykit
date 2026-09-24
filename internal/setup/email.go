package setup

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"flea/internal/localenv"
	"flea/internal/mailer"
)

func Email(ctx context.Context, directory string, in io.Reader, out io.Writer) error {
	return configureEmail(ctx, directory, in, out, func(s mailer.Settings) mailer.Checker { return mailer.NewResend(s) })
}

func configureEmail(ctx context.Context, directory string, in io.Reader, out io.Writer, checker func(mailer.Settings) mailer.Checker) error {
	if _, err := os.Stat(filepath.Join(directory, "_", "flea.config")); err != nil {
		return fmt.Errorf("Run flea email from your site's folder.")
	}
	// If Git is present, fail closed on inspection errors and tracked secrets.
	if _, err := exec.LookPath("git"); err == nil {
		git := func(args ...string) *exec.Cmd {
			cmd := exec.CommandContext(ctx, "git", append([]string{"-C", directory}, args...)...)
			for _, entry := range os.Environ() {
				key, _, _ := strings.Cut(entry, "=")
				if !strings.HasPrefix(key, "GIT_") && key != "LC_ALL" {
					cmd.Env = append(cmd.Env, entry)
				}
			}
			cmd.Env = append(cmd.Env, "LC_ALL=C", "GIT_TERMINAL_PROMPT=0")
			return cmd
		}
		cmd := git("rev-parse", "--show-toplevel")
		if output, err := cmd.CombinedOutput(); err == nil {
			cmd = git("ls-files", "--", ".env")
			tracked, err := cmd.Output()
			if err != nil {
				return fmt.Errorf("Could not check whether .env is tracked by Git. Nothing was saved.")
			}
			if len(tracked) != 0 {
				return fmt.Errorf(".env is tracked by Git. Remove it from Git tracking before running flea email. Nothing was saved.")
			}
		} else if !strings.Contains(string(output), "not a git repository") {
			return fmt.Errorf("Could not check this site's Git repository. Nothing was saved.")
		}
	}
	if _, err := localenv.Read(directory); err != nil {
		return err
	}
	fmt.Fprint(out, "\nEmail setup · Resend\n\nUse an API key and an address on a domain verified in Resend.\nThese settings will be saved in this site's .env file.\n\n")
	key, err := askSecret(ctx, in, out)
	if err != nil {
		return err
	}
	name, err := ask(ctx, in, out, "Sender name (for example, MySite Support)", "")
	if err != nil {
		return err
	}
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("Please enter a sender name. Nothing was saved.")
	}
	from, err := ask(ctx, in, out, "From address (for example, hello@your-verified-domain.com)", "")
	if err != nil {
		return err
	}
	s := mailer.Settings{APIKey: key, Name: name, From: from}
	if err := s.Validate(); err != nil {
		return err
	}
	fmt.Fprintln(out, "\nChecking the Resend connection...")
	_, err = checker(s).Check(ctx)
	if err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := localenv.Save(directory, map[string]string{"RESEND_API_KEY": key, "FLEA_EMAIL_NAME": name, "FLEA_EMAIL_FROM": from}, []string{"RESEND_API_KEY", "FLEA_EMAIL_NAME", "FLEA_EMAIL_FROM"}); err != nil {
		return err
	}
	fmt.Fprintln(out, "\nSaved email settings to .env.")
	return nil
}
