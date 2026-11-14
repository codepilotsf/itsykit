package setup

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"itsy/internal/release"
	"itsy/internal/sitefiles"
)

type gitState struct{ Branch, Commit, Prefix string }

func git(ctx context.Context, root string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", root}, args...)...)
	data, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return string(data), nil
}

func currentGit(ctx context.Context, root string, clean bool) (gitState, error) {
	var s gitState
	branch, err := git(ctx, root, "symbolic-ref", "--quiet", "--short", "HEAD")
	if err != nil {
		return s, fmt.Errorf("deployment requires a Git repository on a named branch (detached HEAD is unsupported)")
	}
	s.Branch = strings.TrimSpace(branch)
	commit, err := git(ctx, root, "rev-parse", "--verify", "HEAD^{commit}")
	if err != nil && clean {
		return s, fmt.Errorf("Please make an initial Git commit before deploying.")
	}
	s.Commit = strings.TrimSpace(commit)
	prefix, err := git(ctx, root, "rev-parse", "--show-prefix")
	if err != nil {
		return s, err
	}
	s.Prefix = strings.TrimSuffix(prefix, "\n")
	if clean {
		status, err := git(ctx, root, "status", "--porcelain=v1", "--untracked-files=all", "--ignore-submodules=none")
		if err != nil {
			return s, err
		}
		if status != "" {
			return s, fmt.Errorf("Please commit or discard your changes before deploying.")
		}
	}
	return s, nil
}

func committedSnapshot(ctx context.Context, root, dest string, s gitState) error {
	tree := s.Commit
	if s.Prefix != "" {
		tree += ":" + strings.TrimSuffix(s.Prefix, "/")
	}
	listing, err := git(ctx, root, "ls-tree", "-r", "-z", "--full-tree", tree)
	if err != nil {
		return err
	}
	var total int64
	count := 0
	for _, entry := range strings.Split(listing, "\x00") {
		if entry == "" {
			continue
		}
		meta, name, ok := strings.Cut(entry, "\t")
		if !ok {
			return fmt.Errorf("invalid Git tree entry")
		}
		if !release.Included(name) {
			continue
		}
		fields := strings.Fields(meta)
		if len(fields) != 3 || fields[1] != "blob" || (fields[0] != "100644" && fields[0] != "100755") {
			return fmt.Errorf("deployment does not support links or submodules: %s", name)
		}
		sizeText, err := git(ctx, root, "cat-file", "-s", fields[2])
		if err != nil {
			return err
		}
		size, err := strconv.ParseInt(strings.TrimSpace(sizeText), 10, 64)
		count++
		if err != nil || size < 0 || size > release.MaxBytes-total || count > release.MaxFiles {
			return fmt.Errorf("committed site exceeds release limits")
		}
		total += size
		body, err := git(ctx, root, "cat-file", "blob", fields[2])
		if err != nil {
			return err
		}
		file := filepath.Join(dest, filepath.FromSlash(name))
		if err = os.MkdirAll(filepath.Dir(file), 0755); err != nil {
			return err
		}
		// Preserve committed executables (a custom app binary) so load-time
		// validation sees a runnable file; everything else is ordinary data.
		mode := os.FileMode(0644)
		if fields[0] == "100755" {
			mode = sitefiles.ExecutableMode
		}
		if err = os.WriteFile(file, []byte(body), mode); err != nil {
			return err
		}
	}
	return nil
}
