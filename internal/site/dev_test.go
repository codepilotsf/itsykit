package site

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDevPreviewConfiguration(t *testing.T) {
	for _, tc := range []struct {
		raw              string
		enabled, invalid bool
	}{
		{`{"app":{}}`, false, false},
		{`{"app":{},"dev":{"email_preview":true}}`, true, false},
		{`{"app":{},"dev":{"email_preview":false}}`, false, false},
		{`{"app":{},"dev":{"email_preview":"yes"}}`, false, true},
	} {
		t.Run(tc.raw, func(t *testing.T) {
			root := t.TempDir()
			if err := os.Mkdir(filepath.Join(root, "site"), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, ".itsy.config.json"), []byte(tc.raw), 0644); err != nil {
				t.Fatal(err)
			}
			dev, err := ReadDevSettings(root)
			if (err != nil) != tc.invalid || dev.EmailPreview != tc.enabled {
				t.Fatal(dev, err)
			}
			p, err := Load(root)
			if (err != nil) != tc.invalid {
				t.Fatal(err)
			}
			if err == nil && p.Dev.EmailPreview != tc.enabled {
				t.Fatal(p.Dev)
			}
		})
	}
}
