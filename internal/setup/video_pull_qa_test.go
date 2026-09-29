package setup

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"flea/internal/backup"
	"flea/internal/release"
	"flea/internal/site"
	"flea/internal/snips"
	"flea/internal/syncstate"
	"flea/internal/video"
)

func TestQAPullHydratesIgnoredVideoInFreshClone(t *testing.T) {
	root := qaRecoverySite(t)
	b, e := os.ReadFile("../video/testdata/valid.mp4")
	if e != nil {
		t.Fatal(e)
	}
	meta, e := video.Inspect(bytes.NewReader(b), int64(len(b)))
	if e != nil {
		t.Fatal(e)
	}
	content := snips.NewStore()
	content.Videos = map[string]video.Video{"Abcdefghij": {Metadata: meta, Name: "Clip"}}
	if e = site.SaveContent(root, content); e != nil {
		t.Fatal(e)
	}
	if e = site.IgnoreSessionFiles(root); e != nil {
		t.Fatal(e)
	}
	testGit(t, root, "add", ".")
	testGit(t, root, "commit", "-m", "video catalogue without ignored bytes")
	if _, e = site.Load(root); e == nil {
		t.Fatal("fresh clone unexpectedly has playable video")
	}
	remoteRoot := t.TempDir()
	if e = os.CopyFS(remoteRoot, os.DirFS(root)); e != nil {
		t.Fatal(e)
	}
	if e = os.MkdirAll(filepath.Join(remoteRoot, "flea", "media"), 0755); e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(remoteRoot, "flea", "media", meta.File), b, 0644); e != nil {
		t.Fatal(e)
	}
	remote := syncstate.Snapshot{Exists: true, Snips: content, Revision: "remote-release"}
	var archive bytes.Buffer
	if e = backup.Write(remoteRoot, backup.Manifest{Name: "demo", Branch: "main", Commit: strings.Repeat("a", 40), StateVersion: remote.Version()}, &archive); e != nil {
		t.Fatal(e)
	}
	run := func(_ context.Context, c command) error {
		for _, a := range c.args {
			if a == "-O" {
				return nil
			}
		}
		cmd := c.args[len(c.args)-1]
		switch {
		case strings.Contains(cmd, "-protocol"):
			fmt.Fprintln(c.out, release.Protocol)
		case strings.Contains(cmd, "-export-state"):
			return json.NewEncoder(c.out).Encode(remote)
		case strings.Contains(cmd, "-backup"):
			_, e := c.out.Write(archive.Bytes())
			return e
		default:
			t.Fatal("unexpected SSH command", cmd)
		}
		return nil
	}
	if e = pull(context.Background(), root, true, false, strings.NewReader(""), io.Discard, io.Discard, run); e != nil {
		t.Fatal(e)
	}
	p, e := site.Load(root)
	if e != nil || p.Content.Videos["Abcdefghij"].File != meta.File {
		t.Fatal("hydration failed", e)
	}
	got, e := os.ReadFile(filepath.Join(root, "flea", "media", meta.File))
	if e != nil || !bytes.Equal(got, b) {
		t.Fatal("incorrect hydrated bytes", e)
	}
	if status := testGit(t, root, "status", "--porcelain"); status != "" {
		t.Fatal("hydration dirtied committed source", status)
	}
}
