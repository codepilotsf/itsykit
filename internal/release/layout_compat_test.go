package release

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestArchivePreservesLegacyLayoutAndExcludesPrivateState(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		".flea.config.json":          `{"app":{"site_name":"Legacy"},"deploy":{"branches":{"main":{"target":"private-host"}}}}`,
		"site/index.html":            "Legacy site",
		"flea/snips.json":            `{}`,
		"flea/runtime/sessions.json": "private session",
		"flea/skills/private.md":     "private guidance",
	}
	for name, data := range files {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(data), 0644); err != nil {
			t.Fatal(err)
		}
	}
	var archive bytes.Buffer
	if err := Pack(root, &archive); err != nil {
		t.Fatal(err)
	}
	dest := t.TempDir()
	if err := Extract(&archive, dest); err != nil {
		t.Fatal(err)
	}
	config, err := os.ReadFile(filepath.Join(dest, ".flea.config.json"))
	if err != nil || bytes.Contains(config, []byte("private-host")) || bytes.Contains(config, []byte(`"deploy"`)) {
		t.Fatal("legacy deployment config not stripped", string(config), err)
	}
	if _, err := os.Stat(filepath.Join(dest, "flea", "snips.json")); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{".itsy.config.json", "itsy", "flea/runtime", "flea/skills"} {
		if _, err := os.Stat(filepath.Join(dest, name)); !os.IsNotExist(err) {
			t.Fatal("unexpected archive entry", name, err)
		}
	}
}
