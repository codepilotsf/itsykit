package setup

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// prepareCommit offers a local commit only before deployment starts. Later
// cleanliness checks must still fail rather than committing during a deployment.
func prepareCommit(ctx context.Context, root, branch string, in io.Reader, out, errOut io.Writer) error {
	state, err := currentGit(ctx, root, false)
	if err != nil {
		return err
	}
	if branch != "" && branch != state.Branch {
		return fmt.Errorf("Your branch changed. Please run itsy deploy again.")
	}
	statusArgs := []string{"status", "--porcelain=v1", "--untracked-files=all", "--ignore-submodules=none"}
	status, err := git(ctx, root, statusArgs...)
	if err != nil {
		return err
	}
	if status == "" {
		if state.Commit == "" {
			return fmt.Errorf("There are no files to commit. Add your site files before deploying.")
		}
		return nil
	}
	if state.Prefix != "" {
		return fmt.Errorf("This site is inside a larger Git repository. Please commit your changes yourself before deploying.")
	}
	fmt.Fprintln(out, "\nYour changes need to be committed before deploying.\n\n"+status)
	fmt.Fprintln(out, "This includes all changes listed above, not just public site files.")
	for {
		answer, err := ask(ctx, in, out, "Commit these changes and deploy? (y/n)", "y")
		if err != nil {
			return err
		}
		switch strings.ToLower(answer) {
		case "n", "no":
			return fmt.Errorf("Deployment stopped. Your changes have not been committed.")
		case "y", "yes":
		default:
			fmt.Fprintln(out, "Please enter y or n.")
			continue
		}
		break
	}
	defaultMessage := ""
	if state.Commit == "" {
		defaultMessage = "init"
	}
	var message string
	for {
		message, err = ask(ctx, in, out, "Commit message", defaultMessage)
		if err != nil {
			return err
		}
		if message != "" {
			break
		}
		fmt.Fprintln(out, "Please enter a commit message.")
	}
	latest, err := currentGit(ctx, root, false)
	if err != nil {
		return err
	}
	again, err := git(ctx, root, statusArgs...)
	if err != nil {
		return err
	}
	if latest != state || again != status {
		return fmt.Errorf("Your repository changed while you were answering. Please run itsy deploy again.")
	}
	run := func(args ...string) error {
		cmd := exec.CommandContext(ctx, "git", append([]string{"-C", root}, args...)...)
		cmd.Stdin, cmd.Stdout, cmd.Stderr = in, out, errOut
		return cmd.Run()
	}
	if err := run("add", "-A", "--", "."); err != nil {
		return fmt.Errorf("Could not stage your changes. Please check Git's message above and try again.")
	}
	if err := run("commit", "-m", message); err != nil {
		return fmt.Errorf("Could not commit your changes. Please check Git's message above and try again.\nYour changes may still be staged; nothing was deployed.")
	}
	_, err = currentGit(ctx, root, true)
	return err
}
