package site

import (
	"bytes"
	"context"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCurrentAndLegacySiteLayouts(t *testing.T) {
	for _, config := range []string{"itsy.config.json", ".flea.config.json"} {
		for _, data := range []string{"itsy", "flea"} {
			t.Run(config+"/"+data, func(t *testing.T) {
				root := fixture(t)
				put(t, root, "index.html", `<p snip>Editable copy</p><img snip src="/images/seed.png" alt="Seed">{team.text}`)
				var seed bytes.Buffer
				if err := png.Encode(&seed, image.NewRGBA(image.Rect(0, 0, 8, 4))); err != nil {
					t.Fatal(err)
				}
				if err := os.MkdirAll(filepath.Join(root, "site", "images"), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(root, "site", "images", "seed.png"), seed.Bytes(), 0644); err != nil {
					t.Fatal(err)
				}
				if config != "itsy.config.json" {
					if err := os.Rename(filepath.Join(root, "itsy.config.json"), filepath.Join(root, config)); err != nil {
						t.Fatal(err)
					}
				}
				if data != "itsy" {
					if err := os.Rename(filepath.Join(root, "itsy"), filepath.Join(root, data)); err != nil {
						t.Fatal(err)
					}
				}
				p, err := LoadEditing(context.Background(), root)
				if err != nil {
					t.Fatal(err)
				}
				if !p.SnipsUpdated || p.RuntimeRoot != filepath.Join(root, data, "runtime") || p.MediaRoot != filepath.Join(root, data, "media") {
					t.Fatal("wrong site paths", p.RuntimeRoot, p.MediaRoot)
				}
				photos := p.Content.ImageAssets()
				if len(photos) != 1 {
					t.Fatal("photo not extracted", photos)
				}
				for _, photo := range photos {
					if _, err := ReadPhoto(p.MediaRoot, photo.File); err != nil {
						t.Fatal("legacy media not installed", err)
					}
					if got := request(NewProductionServer(p, nil), "GET", "/media/"+photo.File).Code; got != 200 {
						t.Fatal("legacy media not served", got)
					}
				}
				if err := SetSiteName(root, "Compatible"); err != nil {
					t.Fatal(err)
				}
				if _, _, err := AddAdmin(root, "admin@example.test"); err != nil {
					t.Fatal(err)
				}
				if err := IgnoreSessionFiles(root); err != nil {
					t.Fatal(err)
				}
				if _, err := ReadDevSettings(root); err != nil {
					t.Fatal(err)
				}
				if err := SaveContent(root, p.Content); err != nil {
					t.Fatal(err)
				}
				p, err = Load(root)
				if err != nil || p.App["site_name"] != "Compatible" || len(p.Admin.Emails) != 1 {
					t.Fatal("config writes did not round-trip", err)
				}
				if _, err := os.Stat(filepath.Join(root, data, "snips.json")); err != nil {
					t.Fatal(err)
				}
				if _, err := os.Stat(filepath.Join(root, data, "runtime", ".snips-pending.json")); !os.IsNotExist(err) {
					t.Fatal("extraction journal retained", err)
				}
				ignore, err := os.ReadFile(filepath.Join(root, ".gitignore"))
				if err != nil || !strings.Contains(string(ignore), "/"+data+"/runtime/") {
					t.Fatal("private runtime not ignored", string(ignore), err)
				}
				if config == ".flea.config.json" {
					if _, err := os.Stat(filepath.Join(root, "itsy.config.json")); !os.IsNotExist(err) {
						t.Fatal("created a second config")
					}
				}
				if data == "flea" {
					if _, err := os.Stat(filepath.Join(root, "itsy")); !os.IsNotExist(err) {
						t.Fatal("created a second content directory")
					}
				}
			})
		}
	}
}

func TestCurrentAndLegacyLayoutComments(t *testing.T) {
	for _, name := range []string{"itsy", "flea"} {
		body, layout, err := pageSettings("index.html", "<!-- "+name+"\nlayout: /admin.layout.html\n-->\n<p>Body</p>")
		if err != nil || layout != "admin.layout.html" || !strings.Contains(body, "<p>Body</p>") {
			t.Fatal(name, body, layout, err)
		}
	}
}

func TestLegacyNamingAliasesRejectSymlinks(t *testing.T) {
	for _, config := range []bool{true, false} {
		root := fixture(t)
		current, legacy := "itsy.config.json", ".flea.config.json"
		if !config {
			current, legacy = "itsy", "flea"
		}
		outside := filepath.Join(t.TempDir(), "private")
		if err := os.Rename(filepath.Join(root, current), outside); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(outside, filepath.Join(root, legacy)); err != nil {
			t.Fatal(err)
		}
		if _, err := Load(root); err == nil {
			t.Fatal("legacy alias followed a symlink", legacy)
		}
	}
}

func TestConfinedAssetReadsKeepExactNames(t *testing.T) {
	public := t.TempDir()
	if err := os.Mkdir(filepath.Join(public, "flea"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(public, "flea", "asset.txt"), []byte("different asset"), 0644); err != nil {
		t.Fatal(err)
	}
	root, err := os.OpenRoot(public)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	file, err := openSiteFile(root, "itsy/asset.txt", 1024)
	if file != nil {
		file.Close()
	}
	if !os.IsNotExist(err) {
		t.Fatal("a missing asset was redirected to a different public directory", err)
	}
}
