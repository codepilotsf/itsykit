package setup

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestInvalidTemporaryStorage(t *testing.T) {
	root := repoSite(t)
	parent := t.TempDir()
	file := filepath.Join(parent, "file")
	if err := os.WriteFile(file, nil, 0600); err != nil {
		t.Fatal(err)
	}
	locked := filepath.Join(parent, "unwritable")
	if err := os.Mkdir(locked, 0500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(locked, 0700) })
	for _, location := range []string{filepath.Join(parent, "missing"), file, locked} {
		t.Run(filepath.Base(location), func(t *testing.T) {
			if location == locked && os.Geteuid() == 0 {
				t.Skip("root can write directories without write permission")
			}
			t.Setenv("TMPDIR", location)
			noRemote := func(context.Context, command) error { t.Fatal("connected with invalid TMPDIR"); return nil }
			for _, err := range []error{
				deploy(context.Background(), root, DeploySettings{Target: "server", Domain: "example.test", Name: "demo"}, false, strings.NewReader(""), io.Discard, io.Discard, noRemote),
				runTarget(context.Background(), "server", strings.NewReader(""), io.Discard, io.Discard, noRemote),
			} {
				var pathErr *os.PathError
				if err == nil || !strings.Contains(err.Error(), location) || !strings.Contains(err.Error(), "Set TMPDIR") || !errors.As(err, &pathErr) {
					t.Fatalf("temporary storage error lost location, guidance or cause: %v", err)
				}
			}
		})
	}
}

func TestTemporarySpaceDiagnostics(t *testing.T) {
	for _, cause := range []error{syscall.ENOSPC, syscall.EDQUOT} {
		original := &os.PathError{Op: "write", Path: "/scratch/flea-deploy-test/site.tar.gz", Err: cause}
		err := explainTemporaryStorageError("/scratch/flea-deploy-test", original)
		if !errors.Is(err, cause) || !errors.Is(err, original) || !strings.Contains(err.Error(), "Set TMPDIR") || !strings.Contains(err.Error(), "/scratch/flea-deploy-test") {
			t.Fatal(err)
		}
	}
	original := errors.New("remote upload failed")
	if explainTemporaryStorageError("/scratch", original) != original || explainTemporaryStorageError("/scratch", nil) != nil {
		t.Fatal("unrelated errors changed")
	}
}

func TestDeploymentPackagingFailureCleansScratch(t *testing.T) {
	root := repoSite(t)
	if err := os.WriteFile(filepath.Join(root, "site", "index.html"), []byte("{#if missing}"), 0644); err != nil {
		t.Fatal(err)
	}
	testGit(t, root, "add", "site/index.html")
	testGit(t, root, "commit", "-m", "invalid template")
	tempRoot := t.TempDir()
	t.Setenv("TMPDIR", tempRoot)
	err := deploy(context.Background(), root, DeploySettings{Target: "server", Domain: "example.test", Name: "demo"}, false, strings.NewReader(""), io.Discard, io.Discard, func(context.Context, command) error {
		t.Fatal("connected before validating the package")
		return nil
	})
	entries, readErr := os.ReadDir(tempRoot)
	if err == nil || readErr != nil || len(entries) != 0 {
		t.Fatal(err, entries, readErr)
	}
}

func TestConcurrentDeploymentsKeepPrivateScratch(t *testing.T) {
	root := repoSite(t)
	tempRoot := t.TempDir()
	t.Setenv("TMPDIR", tempRoot)
	keep := filepath.Join(tempRoot, "other-job")
	if err := os.WriteFile(keep, []byte("preserve"), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	arrived := make(chan string, 2)
	resume := make(chan struct{})
	done := make(chan error, 2)
	stop := errors.New("stop before remote work")
	for range 2 {
		go func() {
			done <- deploy(ctx, root, DeploySettings{Target: "server", Domain: "example.test", Name: "demo"}, false, strings.NewReader(""), io.Discard, io.Discard, func(ctx context.Context, c command) error {
				if strings.Contains(strings.Join(c.args, " "), "-O exit") {
					return nil
				}
				for _, arg := range c.args {
					if strings.HasPrefix(arg, "ControlPath=") {
						arrived <- filepath.Dir(strings.TrimPrefix(arg, "ControlPath="))
					}
				}
				select {
				case <-resume:
				case <-ctx.Done():
				}
				return stop
			})
		}()
	}
	seen := make(map[string]bool)
	for range 2 {
		select {
		case dir := <-arrived:
			info, err := os.Stat(dir)
			if err != nil || filepath.Dir(dir) != tempRoot || seen[dir] || info.Mode().Perm() != 0700 {
				t.Fatalf("scratch is not private and unique: %q, %v", dir, err)
			}
			seen[dir] = true
		case <-ctx.Done():
			t.Fatal("deployments did not both reach preflight")
		}
	}
	close(resume)
	for range 2 {
		if err := <-done; err == nil {
			t.Fatal("accepted failed connection")
		}
	}
	for dir := range seen {
		if _, err := os.Stat(dir); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("deployment scratch remains: %s, %v", dir, err)
		}
	}
	if body, err := os.ReadFile(keep); err != nil || string(body) != "preserve" {
		t.Fatalf("other job's data changed: %q, %v", body, err)
	}
}
