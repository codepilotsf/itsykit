package setup

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"itsy/internal/release"
	"itsy/internal/site"
	"itsy/internal/snips"
	"itsy/internal/syncstate"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPlanRequiresScopedBaseline(t *testing.T) {
	p := &site.Project{Content: snips.NewStore()}
	r := syncstate.Snapshot{Exists: true, Snips: snips.NewStore()}
	if _, e := makePlan(syncstate.Baseline{}, p, r, true, nil); e == nil {
		t.Fatal("missing baseline accepted")
	}
	b := syncstate.Baseline{Snips: snips.NewStore()}
	if _, e := makePlan(b, p, r, true, nil); e != nil {
		t.Fatal(e)
	}
	if _, e := makePlan(syncstate.Baseline{}, p, syncstate.Snapshot{}, true, nil); e != nil {
		t.Fatal("first deployment blocked", e)
	}
}

func TestPullPreservesNewLocalSnipsAndScopesBaseline(t *testing.T) {
	root := repoSite(t)
	local := snips.NewStore()
	local.Page("/").Snips["a7Km2pQ9xB"] = snips.Entry{Content: "Seed"}
	local.Page("/").Snips["b8Ln3rR0yC"] = snips.Entry{Content: "Local new"}
	if e := site.SaveContent(root, local); e != nil {
		t.Fatal(e)
	}
	os.WriteFile(filepath.Join(root, "site", "index.html"), []byte(`<p>{snip a7Km2pQ9xB}</p><p>{snip b8Ln3rR0yC}</p>`), 0644)
	s := DeploySettings{Target: "root@server", Name: "demo", Domain: "example.test", Branch: "main"}
	if e := saveDeploymentConfig(root, DeploymentConfig{Branches: map[string]DeploySettings{"main": s}}); e != nil {
		t.Fatal(e)
	}
	cssPath := filepath.Join(root, "site", "css", "theme.css")
	if e := os.MkdirAll(filepath.Dir(cssPath), 0755); e != nil {
		t.Fatal(e)
	}
	if e := os.WriteFile(cssPath, []byte(":root {--brand-1: #123456;}"), 0644); e != nil {
		t.Fatal(e)
	}
	testGit(t, root, "add", ".")
	testGit(t, root, "commit", "-m", "content")
	localCSS := ":root {--brand-1: #654321;}"
	if e := os.WriteFile(cssPath, []byte(localCSS), 0644); e != nil {
		t.Fatal(e)
	}
	remote := snips.NewStore()
	remote.Page("/").Snips["a7Km2pQ9xB"] = snips.Entry{Content: "Owner change"}
	run := func(ctx context.Context, c command) error {
		for _, arg := range c.args {
			if arg == "-O" {
				return nil
			}
		}
		last := c.args[len(c.args)-1]
		switch {
		case strings.Contains(last, "-protocol"):
			fmt.Fprintln(c.out, release.Protocol)
		case strings.Contains(last, "-export-state"):
			return json.NewEncoder(c.out).Encode(syncstate.Snapshot{Exists: true, Snips: remote})
		default:
			t.Fatal(last)
		}
		return nil
	}
	if e := pull(context.Background(), root, strings.NewReader("l\n"), io.Discard, io.Discard, run); e != nil {
		t.Fatal(e)
	}
	p, e := site.Load(root)
	if e != nil {
		t.Fatal(e)
	}
	if p.Content.Page("/").Snips["a7Km2pQ9xB"].Content != "Owner change" || p.Content.Page("/").Snips["b8Ln3rR0yC"].Content != "Local new" {
		t.Fatal(p.Content)
	}
	css, e := os.ReadFile(cssPath)
	if e != nil || string(css) != localCSS {
		t.Fatal("pull changed local CSS", e)
	}
	b, e := readBaseline(context.Background(), root, s)
	if e != nil || b.Snips == nil {
		t.Fatal(b, e)
	}
	if e := pull(context.Background(), root, strings.NewReader(""), io.Discard, io.Discard, run); e == nil {
		t.Fatal("dirty local content overwritten")
	}
	if !strings.Contains(testGit(t, root, "status", "--short"), "itsy/snips.json") {
		t.Fatal("pull committed changes")
	}
}
func TestConflictCancelAndNoDefault(t *testing.T) {
	r := resolveSync(context.Background(), strings.NewReader("\nc\n"), io.Discard)
	if _, e := r(syncstate.Conflict{Key: "test", Local: "a", Remote: "b"}); e == nil {
		t.Fatal("cancel accepted")
	}
}
