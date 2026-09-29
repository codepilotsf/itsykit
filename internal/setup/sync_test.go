package setup

import (
	"context"
	"encoding/json"
	"flea/internal/release"
	"flea/internal/site"
	"flea/internal/snips"
	"flea/internal/syncstate"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPlanRequiresScopedBaseline(t *testing.T) {
	css := "a"
	p := &site.Project{Content: snips.NewStore(), AssetData: map[string][]byte{"/css/theme.css": []byte(css)}}
	r := syncstate.Snapshot{Exists: true, Snips: snips.NewStore(), Theme: &css}
	for _, scope := range []struct{ s, t bool }{{true, false}, {false, true}} {
		if _, e := makePlan(syncstate.Baseline{}, p, r, scope.s, scope.t, nil); e == nil {
			t.Fatal("missing baseline accepted")
		}
	}
	b := syncstate.Baseline{Snips: snips.NewStore(), HasTheme: true, Theme: &css}
	if _, e := makePlan(b, p, r, true, true, nil); e != nil {
		t.Fatal(e)
	}
	if _, e := makePlan(syncstate.Baseline{}, p, syncstate.Snapshot{}, true, true, nil); e != nil {
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
	testGit(t, root, "add", ".")
	testGit(t, root, "commit", "-m", "content")
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
	if e := pull(context.Background(), root, true, false, strings.NewReader("l\n"), io.Discard, io.Discard, run); e != nil {
		t.Fatal(e)
	}
	p, e := site.Load(root)
	if e != nil {
		t.Fatal(e)
	}
	if p.Content.Page("/").Snips["a7Km2pQ9xB"].Content != "Owner change" || p.Content.Page("/").Snips["b8Ln3rR0yC"].Content != "Local new" {
		t.Fatal(p.Content)
	}
	b, e := readBaseline(context.Background(), root, s)
	if e != nil || b.Snips == nil || b.HasTheme {
		t.Fatal(b, e)
	}
	if e := pull(context.Background(), root, true, false, strings.NewReader(""), io.Discard, io.Discard, run); e == nil {
		t.Fatal("dirty local content overwritten")
	}
	if !strings.Contains(testGit(t, root, "status", "--short"), "flea/snips.json") {
		t.Fatal("pull committed changes")
	}
}
func TestConflictCancelAndNoDefault(t *testing.T) {
	r := resolveSync(context.Background(), strings.NewReader("\nc\n"), io.Discard)
	if _, e := r(syncstate.Conflict{Key: "test", Local: "a", Remote: "b"}); e == nil {
		t.Fatal("cancel accepted")
	}
}
