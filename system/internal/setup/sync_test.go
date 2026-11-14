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

func TestPullPreservesNewLocalPostsAndScopesBaseline(t *testing.T) {
	root := repoSite(t)
	local := snips.NewStore()
	local.Blog = []snips.BlogPost{{ID: "a7Km2pQ9xB", Title: "Seed", Slug: "seed", Status: "draft"}, {ID: "b8Ln3rR0yC", Title: "Local new", Slug: "local", Status: "draft"}}
	if e := site.SaveContent(root, local); e != nil {
		t.Fatal(e)
	}
	os.WriteFile(filepath.Join(root, "site", "index.html"), []byte(`<p>Ordinary content</p>`), 0644)
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
	remote.Blog = []snips.BlogPost{{ID: "a7Km2pQ9xB", Title: "Owner change", Slug: "seed", Status: "draft"}}
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
	if p.Content.Blog[0].Title != "Owner change" || p.Content.Blog[1].Title != "Local new" {
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
