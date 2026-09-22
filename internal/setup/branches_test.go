package setup

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func testGit(t *testing.T, root string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", root, "-c", "core.hooksPath=/dev/null", "-c", "commit.gpgsign=false"}, args...)...)
	b, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v %s", args, err, b)
	}
	return strings.TrimSpace(string(b))
}
func initRepo(t *testing.T, root string) {
	t.Helper()
	testGit(t, root, "init", "-b", "main")
	testGit(t, root, "config", "user.name", "Nanoweb Test")
	testGit(t, root, "config", "user.email", "test@example.test")
	testGit(t, root, "add", ".")
	testGit(t, root, "commit", "--allow-empty", "-m", "fixture")
}
func repoSite(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	os.Mkdir(filepath.Join(root, "."), 0755)
	os.Mkdir(filepath.Join(root, "_"), 0755)
	os.WriteFile(filepath.Join(root, "index.html"), []byte("committed"), 0644)
	os.WriteFile(filepath.Join(root, "_", "nanoweb.config"), []byte(`{"site":{"name":"test"}}`), 0644)
	initRepo(t, root)
	return root
}

func TestFirstDeployAndBranchMappings(t *testing.T) {
	root := repoSite(t)
	var out bytes.Buffer
	answers := strings.Join([]string{"ubuntu@server", "example.test", "", "", "you@example.test", "y", "", ""}, "\n") + "\n"
	s, err := SelectDeployment(context.Background(), root, DeploySettings{}, nil, false, false, strings.NewReader(answers), &out)
	if err != nil {
		t.Fatal(err, out.String())
	}
	if !s.AgreeTOS || !strings.Contains(out.String(), "Enable HTTPS and accept Let's Encrypt terms?") || strings.Contains(out.String(), "Accept Let's Encrypt terms (") {
		t.Fatal("expected combined HTTPS and terms confirmation", s, out.String())
	}
	if s.Branch != "main" || !s.HTTPS || !s.WWWRedirect || strings.Contains(out.String(), "production") || strings.Contains(out.String(), "Remote site name") || !strings.HasSuffix(s.Name, "-main") {
		t.Fatal(s, out.String())
	}
	if _, err := currentGit(context.Background(), root, true); err == nil {
		t.Fatal("new config did not require commit")
	}
	testGit(t, root, "add", "_/deploy.json")
	testGit(t, root, "commit", "-m", "main mapping")
	out.Reset()
	got, err := SelectDeployment(context.Background(), root, DeploySettings{}, nil, false, false, strings.NewReader(""), &out)
	if err != nil || got.Target != s.Target || out.Len() != 0 {
		t.Fatal(got, err, out.String())
	}
	if _, err := SelectDeployment(context.Background(), root, DeploySettings{Domain: "evil.test"}, map[string]bool{"domain": true}, false, false, strings.NewReader(""), io.Discard); err == nil {
		t.Fatal("overrode saved mapping without confirmation")
	}
	testGit(t, root, "checkout", "-b", "staging")
	out.Reset()
	answers = strings.Join([]string{"", "", "", "", "", "y", "", ""}, "\n") + "\n"
	staged, err := SelectDeployment(context.Background(), root, DeploySettings{}, nil, false, false, strings.NewReader(answers), &out)
	if err != nil || staged.Domain != "staging.example.test" || staged.Name == s.Name {
		t.Fatal(staged, err, out.String())
	}
	c, err := readDeploymentConfig(root)
	if err != nil || len(c.Branches) != 2 {
		t.Fatal(c, err)
	}
	testGit(t, root, "checkout", "--detach")
	if _, err := SelectDeployment(context.Background(), root, DeploySettings{}, nil, false, false, strings.NewReader(""), io.Discard); err == nil {
		t.Fatal("detached HEAD accepted")
	}
}

func TestCertificateEmailDefaultsToGit(t *testing.T) {
	root := repoSite(t)
	var out bytes.Buffer
	s, err := SelectDeployment(context.Background(), root, DeploySettings{}, nil, false, false,
		strings.NewReader("root@server\nexample.test\n\n\n\n"), &out)
	if err != nil || s.Email != "test@example.test" || !strings.Contains(out.String(), "Certificate contact email [test@example.test]:") {
		t.Fatal(s, err, out.String())
	}
}

func TestFirstDeployConfiguresBeforeInitialCommit(t *testing.T) {
	root := t.TempDir()
	os.Mkdir(filepath.Join(root, "_"), 0755)
	testGit(t, root, "init", "-b", "main")
	answers := strings.Join([]string{"server", "", "example.test", "", "", "dev@example.test", "y", ""}, "\n") + "\n"
	s, err := SelectDeployment(context.Background(), root, DeploySettings{}, nil, false, false, strings.NewReader(answers), io.Discard)
	if err != nil || s.Branch != "main" {
		t.Fatal(s, err)
	}
	if _, err := os.Stat(filepath.Join(root, "_/deploy.json")); err != nil {
		t.Fatal("settings were not saved", err)
	}
	err = deploy(context.Background(), root, s, false, strings.NewReader(""), io.Discard, io.Discard, func(context.Context, command) error { t.Fatal("connected without a commit"); return nil })
	if err == nil || err.Error() != "Please make an initial Git commit before deploying." {
		t.Fatal(err)
	}
}

func TestCommittedSnapshotAndDirtyRejection(t *testing.T) {
	for _, kind := range []string{"unstaged", "staged", "untracked"} {
		t.Run(kind, func(t *testing.T) {
			root := repoSite(t)
			file := filepath.Join(root, "index.html")
			if kind == "untracked" {
				file = filepath.Join(root, "new.txt")
			}
			os.WriteFile(file, []byte("dirty"), 0644)
			if kind == "staged" {
				testGit(t, root, "add", ".")
			}
			err := deploy(context.Background(), root, DeploySettings{Target: "server", Name: "demo", Domain: "example.test"}, false, strings.NewReader(""), io.Discard, io.Discard, func(context.Context, command) error { t.Fatal("connected despite dirty status"); return nil })
			if err == nil || err.Error() != "Please commit or discard your changes before deploying." {
				t.Fatal(err)
			}
		})
	}
	root := repoSite(t)
	s, err := currentGit(context.Background(), root, true)
	if err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(root, "index.html"), []byte("changed after commit"), 0644)
	dest := t.TempDir()
	if err := committedSnapshot(context.Background(), root, dest, s); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(filepath.Join(dest, "index.html"))
	if string(got) != "committed" {
		t.Fatal("snapshot used working files", string(got))
	}
}

func TestMappingCollisionAndCancellation(t *testing.T) {
	root := repoSite(t)
	s := DeploySettings{Target: "server", Domain: "example.test", Name: "demo", HTTPS: false}
	c := DeploymentConfig{Branches: map[string]DeploySettings{"main": s}}
	if err := validateMapping("staging", s, c); err == nil {
		t.Fatal("branch destination collision accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := SelectDeployment(ctx, root, DeploySettings{}, nil, false, false, strings.NewReader("root@server\n192.0.2.1\n"), io.Discard)
	if err == nil {
		t.Fatal("cancellation ignored")
	}
	if _, err := os.Stat(filepath.Join(root, "_/deploy.json")); !os.IsNotExist(err) {
		t.Fatal("saved cancelled mapping")
	}
}

func TestNestedCommittedSnapshotAndIgnoredFiles(t *testing.T) {
	root := t.TempDir()
	sub := filepath.Join(root, "sites/demo")
	os.MkdirAll(filepath.Join(sub, "."), 0755)
	os.MkdirAll(filepath.Join(sub, "_"), 0755)
	os.WriteFile(filepath.Join(sub, "index.html"), []byte("nested commit"), 0644)
	os.WriteFile(filepath.Join(sub, "_/nanoweb.config"), []byte(`{"site":{"name":"nested"}}`), 0644)
	os.WriteFile(filepath.Join(root, ".gitignore"), []byte("sites/demo/ignored.txt\n"), 0644)
	initRepo(t, root)
	os.WriteFile(filepath.Join(sub, "ignored.txt"), []byte("not committed"), 0644)
	s, err := currentGit(context.Background(), sub, true)
	if err != nil {
		t.Fatal(err)
	}
	dest := t.TempDir()
	if err := committedSnapshot(context.Background(), sub, dest, s); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dest, "index.html"))
	if err != nil || string(b) != "nested commit" {
		t.Fatal(string(b), err)
	}
	if _, err := os.Stat(filepath.Join(dest, "ignored.txt")); !os.IsNotExist(err) {
		t.Fatal("ignored working file deployed")
	}
}

func TestCommittedSymlinkRejected(t *testing.T) {
	root := repoSite(t)
	if err := os.Symlink("index.html", filepath.Join(root, "link.html")); err != nil {
		t.Fatal(err)
	}
	testGit(t, root, "add", ".")
	testGit(t, root, "commit", "-m", "link")
	s, err := currentGit(context.Background(), root, true)
	if err != nil {
		t.Fatal(err)
	}
	if err := committedSnapshot(context.Background(), root, t.TempDir(), s); err == nil {
		t.Fatal("committed link accepted")
	}
}
