package setup

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"itsy/internal/admin"
	"itsy/internal/site"
	"itsy/internal/snips"
	"itsy/internal/syncstate"
)

func statusFixture(t *testing.T, configured bool) (*site.Project, DeploySettings) {
	t.Helper()
	root := repoSite(t)
	s := DeploySettings{Target: "root@server", Name: "demo", Domain: "example.test", Branch: "main"}
	if configured {
		if err := saveDeploymentConfig(root, DeploymentConfig{Branches: map[string]DeploySettings{"main": s}}); err != nil {
			t.Fatal(err)
		}
		testGit(t, root, "add", ".")
		testGit(t, root, "commit", "-m", "destination")
	}
	p, err := site.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	return p, s
}
func TestStatusLocalOnly(t *testing.T) {
	for _, mode := range []string{"no mapping", "other branch", "detached", "no git", "unborn"} {
		t.Run(mode, func(t *testing.T) {
			p, _ := statusFixture(t, mode != "no mapping")
			switch mode {
			case "other branch":
				testGit(t, p.Root, "checkout", "-b", "feature")
			case "detached":
				testGit(t, p.Root, "checkout", "--detach")
			case "no git":
				if e := os.RemoveAll(filepath.Join(p.Root, ".git")); e != nil {
					t.Fatal(e)
				}
			case "unborn":
				testGit(t, p.Root, "checkout", "--orphan", "new")
			}
			var out bytes.Buffer
			err := status(context.Background(), p, strings.NewReader(""), &out, io.Discard, false, func(context.Context, command) error { t.Fatal("connected without a matching destination"); return nil }, func(context.Context, string) error { t.Fatal("HTTP check without destination"); return nil })
			if err != nil || !strings.Contains(out.String(), "Local\n") || strings.Contains(out.String(), "Checking remote") {
				t.Fatal(err, out.String())
			}
		})
	}
}
func TestStatusRemoteReadOnlyAndLocalFirst(t *testing.T) {
	p, s := statusFixture(t, true)
	commit := testGit(t, p.Root, "rev-parse", "HEAD")
	before := testGit(t, p.Root, "status", "--porcelain")
	revision := "/var/lib/flea/sites/demo/releases/20260101-120000-abcdef"
	var out bytes.Buffer
	payload := strings.TrimSuffix(testSiteJSON, "}") + `,"indexing":{"example.test":"noindex","www.example.test":"allow"},"warnings":["example.test: robots.txt contains a site-wide Disallow rule"]}`
	run := inventoryRunner(t, "["+payload+"]", func(c command) (bool, error) {
		if !strings.Contains(out.String(), "Local\n") || !strings.Contains(out.String(), "Checking remote...\n") {
			t.Fatal("network before local summary", out.String())
		}
		text := strings.Join(c.args, " ")
		if strings.Contains(text, "-export-state") {
			return true, json.NewEncoder(c.out).Encode(syncstate.Snapshot{Exists: true, Snips: p.Content, Revision: revision})
		}
		if strings.Contains(text, "-history") {
			return true, json.NewEncoder(c.out).Encode([]admin.ReleaseRecord{{ID: filepath.Base(revision), Branch: s.Branch, Commit: commit}})
		}
		return false, nil
	})
	checks := 0
	err := status(context.Background(), p, strings.NewReader(""), &out, io.Discard, false, run, func(_ context.Context, url string) error {
		checks++
		if url != "http://example.test" {
			t.Fatal(url)
		}
		return nil
	})
	if err != nil || checks != 1 || !strings.Contains(out.String(), "Code and content match. Styles are part of the code release.") || strings.Contains(out.String(), "\x1b") {
		t.Fatal(err, out.String())
	}
	if !strings.Contains(out.String(), "indexing example.test: noindex, nofollow") || !strings.Contains(out.String(), "indexing www.example.test: allow (no restriction added)") || !strings.Contains(out.String(), "site-wide Disallow") {
		t.Fatal("effective remote policy/diagnostic missing", out.String())
	}
	if after := testGit(t, p.Root, "status", "--porcelain"); after != before {
		t.Fatal("status changed files", before, after)
	}
	path, err := baselinePath(context.Background(), p.Root, s)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(filepath.Dir(path)); !os.IsNotExist(err) {
		t.Fatal("read created baseline directory", err)
	}
}
func TestStatusRemoteFailureAndCancellation(t *testing.T) {
	p, _ := statusFixture(t, true)
	for _, cancel := range []bool{false, true} {
		t.Run(fmt.Sprint(cancel), func(t *testing.T) {
			ctx, stop := context.WithCancel(context.Background())
			defer stop()
			var out bytes.Buffer
			run := func(ctx context.Context, c command) error {
				if strings.Contains(strings.Join(c.args, " "), "-O exit") {
					return nil
				}
				if cancel {
					stop()
					<-ctx.Done()
					return ctx.Err()
				}
				return errors.New("offline")
			}
			err := status(ctx, p, strings.NewReader(""), &out, io.Discard, true, run, func(context.Context, string) error { t.Fatal("HTTP after SSH failure"); return nil })
			if err == nil || !strings.Contains(out.String(), "Comparison   Unavailable") || strings.Contains(out.String(), "Matches") || !strings.Contains(out.String(), "\x1b[2KDeployment") {
				t.Fatal(err, out.String())
			}
		})
	}
}
func TestStatusRejectsWrongDeployment(t *testing.T) {
	for _, mode := range []string{"receipt", "branch", "domain", "snapshot race"} {
		t.Run(mode, func(t *testing.T) {
			p, s := statusFixture(t, true)
			commit := testGit(t, p.Root, "rev-parse", "HEAD")
			revision := "/var/lib/flea/sites/demo/releases/20260101-120000-abcdef"
			inventory := testSiteJSON
			if mode == "domain" {
				inventory = strings.ReplaceAll(inventory, "example.test", "other.test")
			}
			if mode == "snapshot race" {
				inventory = strings.ReplaceAll(inventory, "20260101-120000-abcdef", "newer-release")
			}
			run := inventoryRunner(t, "["+inventory+"]", func(c command) (bool, error) {
				text := strings.Join(c.args, " ")
				if strings.Contains(text, "-export-state") {
					return true, json.NewEncoder(c.out).Encode(syncstate.Snapshot{Exists: true, Snips: p.Content, Revision: revision})
				}
				if strings.Contains(text, "-history") {
					record := admin.ReleaseRecord{ID: filepath.Base(revision), Branch: s.Branch, Commit: commit}
					if mode == "receipt" {
						record.ID = "other"
					}
					if mode == "branch" {
						record.Branch = "other"
					}
					return true, json.NewEncoder(c.out).Encode([]admin.ReleaseRecord{record})
				}
				return false, nil
			})
			var out bytes.Buffer
			err := status(context.Background(), p, strings.NewReader(""), &out, io.Discard, false, run, func(context.Context, string) error { t.Fatal("checked wrong website"); return nil })
			if err == nil || strings.Contains(out.String(), "Matches") {
				t.Fatal(err, out.String())
			}
		})
	}
}
func TestStatusNotDeployed(t *testing.T) {
	p, _ := statusFixture(t, true)
	run := inventoryRunner(t, "[]", func(c command) (bool, error) {
		if strings.Contains(strings.Join(c.args, " "), "-export-state") {
			return true, json.NewEncoder(c.out).Encode(syncstate.Snapshot{Snips: snips.NewStore()})
		}
		return false, nil
	})
	var out bytes.Buffer
	if err := status(context.Background(), p, strings.NewReader(""), &out, io.Discard, false, run, func(context.Context, string) error { t.Fatal("checked undeployed site"); return nil }); err != nil || !strings.Contains(out.String(), "Not deployed") {
		t.Fatal(err, out.String())
	}
}
func TestWebsiteCheck(t *testing.T) {
	for _, code := range []int{200, 301, 404, 503} {
		t.Run(fmt.Sprint(code), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodHead {
					t.Error(r.Method)
				}
				w.Header().Set("Location", "/redirect")
				w.WriteHeader(code)
			}))
			defer server.Close()
			err := checkWebsite(context.Background(), server.URL)
			if (err == nil) != (code == 200) {
				t.Fatal(code, err)
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := checkWebsite(ctx, "https://example.test"); err == nil {
		t.Fatal("cancellation ignored")
	}
}
func TestStatusSpinner(t *testing.T) {
	var out bytes.Buffer
	p := newCommandProgress(&out, true, "Checking remote...")
	time.Sleep(150 * time.Millisecond)
	p.stop()
	p.stop()
	finished := out.String()
	time.Sleep(150 * time.Millisecond)
	if out.String() != finished || !strings.HasSuffix(finished, "\r\x1b[2K") || !strings.Contains(finished, "Checking remote... /") {
		t.Fatal(out.String())
	}
}
func TestCountGitChanges(t *testing.T) {
	if got := countChanges(" M file\x00R  new\x00old\x00?? newline\nname\x00"); got != 3 {
		t.Fatal(got)
	}
}

func TestStatusDifferencesWithAndWithoutBaseline(t *testing.T) {
	for _, known := range []bool{false, true} {
		t.Run(fmt.Sprint(known), func(t *testing.T) {
			p, s := statusFixture(t, true)
			remote := snips.NewStore()
			remote.Page("/").Title = "Live title"
			if known {
				if err := saveBaseline(context.Background(), p.Root, s, syncstate.Baseline{Snips: p.Content}); err != nil {
					t.Fatal(err)
				}
			}
			revision := "/var/lib/flea/sites/demo/releases/20260101-120000-abcdef"
			commit := testGit(t, p.Root, "rev-parse", "HEAD")
			run := inventoryRunner(t, "["+testSiteJSON+"]", func(c command) (bool, error) {
				text := strings.Join(c.args, " ")
				if strings.Contains(text, "-export-state") {
					return true, json.NewEncoder(c.out).Encode(syncstate.Snapshot{Exists: true, Snips: remote, Revision: revision})
				}
				if strings.Contains(text, "-history") {
					return true, json.NewEncoder(c.out).Encode([]admin.ReleaseRecord{{ID: filepath.Base(revision), Branch: s.Branch, Commit: commit}})
				}
				return false, nil
			})
			var out bytes.Buffer
			err := status(context.Background(), p, strings.NewReader(""), &out, io.Discard, false, run, func(context.Context, string) error { return errors.New("website offline") })
			want := "direction unknown (no pull baseline)"
			if known {
				want = "0 local, 1 remote, 0 conflicting changes"
			}
			if err == nil || strings.Count(out.String(), want) != 1 || !strings.Contains(out.String(), "Connection   Unavailable") || !strings.Contains(out.String(), "itsy pull") {
				t.Fatal(err, out.String())
			}
		})
	}
}
