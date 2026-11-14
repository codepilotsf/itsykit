package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"flea/internal/site"
)

// resetSite only operates on a recognized local site. Moving old entries into
// a private sibling backup preserves ignored media and secrets as well as code.
func resetSite(root string, in io.Reader, out io.Writer) error {
	root, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	if root == filepath.Dir(root) {
		return fmt.Errorf("run itsy reset from a site folder")
	}
	if home, _ := os.UserHomeDir(); root == home {
		return fmt.Errorf("refusing to reset your home folder")
	}
	info, err := os.Lstat(root)
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("site folder must be a real directory")
	}
	if info, err := os.Lstat(filepath.Join(root, "site")); err != nil || !info.IsDir() {
		return fmt.Errorf("site must be a real directory, not a link")
	}
	for _, name := range []string{".flea.config.json", "site/main.layout.html", "site/index.html"} {
		info, err := os.Lstat(filepath.Join(root, name))
		if err != nil || !info.Mode().IsRegular() {
			return fmt.Errorf("run itsy reset from a site with config, layout and index files")
		}
	}
	if info, err := os.Lstat(filepath.Join(root, ".git")); err == nil && !info.IsDir() {
		return fmt.Errorf("reset does not support Git worktrees or linked Git directories")
	}
	unlock, err := site.LockDevelopment(root)
	if err != nil {
		return err
	}
	defer unlock()
	parent := filepath.Dir(root)
	stage, err := os.MkdirTemp(parent, ".flea-reset-stage-")
	if err != nil {
		return err
	}
	// The generated child is known and owned by this invocation.
	defer os.RemoveAll(stage)
	fresh := filepath.Join(stage, "fresh")
	if err = createSite(fresh, false); err != nil {
		return err
	}
	fmt.Fprintf(out, "\nStart fresh in %s?\nThis replaces all local files and settings except Git history.\nYour old files will be kept in a private folder beside this site.\nThe remote site is not changed.\n\nReset this site? [y/N]: ", root)
	answer, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && err != io.EOF {
		return err
	}
	if a := strings.ToLower(strings.TrimSpace(answer)); a != "y" && a != "yes" {
		fmt.Fprintln(out, "Kept your site unchanged.")
		return nil
	}
	backup, err := os.MkdirTemp(parent, ".flea-reset-backup-")
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "\nRecovery copy: %s\n", backup)
	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}
	var moved, published []string
	rollback := func(cause error) error {
		var recoveryErr error
		for i := len(published) - 1; i >= 0; i-- {
			if e := os.Rename(filepath.Join(root, published[i]), filepath.Join(fresh, published[i])); e != nil {
				recoveryErr = e
			}
		}
		for i := len(moved) - 1; i >= 0; i-- {
			if e := os.Rename(filepath.Join(backup, moved[i]), filepath.Join(root, moved[i])); e != nil {
				recoveryErr = e
			}
		}
		if recoveryErr != nil {
			return fmt.Errorf("reset stopped: %w; recovery needs attention: %v; old files remain at %s", cause, recoveryErr, backup)
		}
		return fmt.Errorf("reset stopped and old files restored: %w", cause)
	}
	for _, entry := range entries {
		if entry.Name() == ".git" {
			continue
		}
		if err = os.Rename(filepath.Join(root, entry.Name()), filepath.Join(backup, entry.Name())); err != nil {
			return rollback(err)
		}
		moved = append(moved, entry.Name())
	}
	clean, err := os.ReadDir(fresh)
	if err != nil {
		return rollback(err)
	}
	for _, entry := range clean {
		if err = os.Rename(filepath.Join(fresh, entry.Name()), filepath.Join(root, entry.Name())); err != nil {
			return rollback(err)
		}
		published = append(published, entry.Name())
	}
	fmt.Fprintln(out, "Your site is reset. Run itsy dev to get started.\nGit history was kept. No commit or deployment was made.")
	return nil
}
