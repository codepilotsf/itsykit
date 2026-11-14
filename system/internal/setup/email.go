package setup

import (
	"context"
	"fmt"
	"io"
	"itsy/internal/sitefiles"
	"os"
	"os/exec"
	"strings"

	"itsy/internal/localenv"
	"itsy/internal/mailer"
)

func Email(ctx context.Context, directory string, in io.Reader, out io.Writer) error {
	return configureEmail(ctx, directory, in, out, func(s mailer.Settings) mailer.Checker { return mailer.NewResend(s) })
}

func configureEmail(ctx context.Context, directory string, in io.Reader, out io.Writer, checker func(mailer.Settings) mailer.Checker) error {
	return configureEmailFlow(ctx, directory, in, out, checker, false)
}

func EmailOptional(ctx context.Context, directory string, in io.Reader, out io.Writer) error {
	return configureEmailFlow(ctx, directory, in, out, func(s mailer.Settings) mailer.Checker { return mailer.NewResend(s) }, true)
}

func configureEmailFlow(ctx context.Context, directory string, in io.Reader, out io.Writer, checker func(mailer.Settings) mailer.Checker, optional bool) error {
	if _, err := os.Stat(sitefiles.Path(directory, "itsy.config.json")); err != nil {
		return fmt.Errorf("Run itsy email from your site's folder.")
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
				return fmt.Errorf(".env is tracked by Git. Remove it from Git tracking before running itsy email. Nothing was saved.")
			}
		} else if !strings.Contains(string(output), "not a git repository") {
			return fmt.Errorf("Could not check this site's Git repository. Nothing was saved.")
		}
	}
	if _, err := localenv.Read(directory); err != nil {
		return err
	}
	s := mailer.Settings{}
	var err error
	if optional {
		s, err = mailer.LocalSettings(directory)
		if err != nil {
			return err
		}
	}
	if s.APIKey == "" {
		label := "Resend API key (masked)"
		if optional {
			label = "Resend API key (optional, masked)"
		}
		s.APIKey, err = askMasked(ctx, in, out, label)
		if err != nil {
			return err
		}
		if optional && s.APIKey == "" {
			return nil
		}
	}
	if s.Name == "" {
		s.Name, err = ask(ctx, in, out, "Sender name", "")
	}
	if err != nil {
		return err
	}
	if strings.TrimSpace(s.Name) == "" {
		return fmt.Errorf("Please enter a sender name. Nothing was saved.")
	}
	if s.From == "" {
		s.From, err = ask(ctx, in, out, "Sender email (on a domain verified in Resend)", "")
	}
	if err != nil {
		return err
	}
	if err := s.Validate(); err != nil {
		return err
	}
	_, err = checker(s).Check(ctx)
	if err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := localenv.Save(directory, map[string]string{"RESEND_API_KEY": s.APIKey, "FLEA_EMAIL_NAME": s.Name, "FLEA_EMAIL_FROM": s.From}, []string{"RESEND_API_KEY", "FLEA_EMAIL_NAME", "FLEA_EMAIL_FROM"}); err != nil {
		return err
	}
	fmt.Fprintln(out, ".env updated.")
	return nil
}
