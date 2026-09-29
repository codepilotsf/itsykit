package cli

import (
	"bytes"
	"context"
	"flea/internal/site"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDevOwnerSetup(t *testing.T) {
	for _, tc := range []struct {
		name, input                                string
		interactive, owner, mail, changed, wantErr bool
	}{
		{name: "new", input: "owner@example.com\ny\n", interactive: true, changed: true},
		{name: "skip owner", input: "\n", interactive: true},
		{name: "skip mail", input: "owner@example.com\n\n", interactive: true, changed: true},
		{name: "noninteractive"},
		{name: "configured", interactive: true, owner: true, mail: true},
		{name: "owner only", input: "\n", interactive: true, owner: true},
		{name: "mail only", input: "owner@example.com\n", interactive: true, mail: true, changed: true},
		{name: "invalid owner", input: "bad\n", interactive: true, wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("RESEND_API_KEY", "")
			t.Setenv("FLEA_EMAIL_FROM", "")
			if tc.mail {
				t.Setenv("RESEND_API_KEY", "re_test12345678")
				t.Setenv("FLEA_EMAIL_FROM", "hello@example.com")
			}
			root := filepath.Join(t.TempDir(), "sample")
			if err := Create(root); err != nil {
				t.Fatal(err)
			}
			if tc.owner {
				if _, _, err := site.AddAdmin(root, "owner@example.com"); err != nil {
					t.Fatal(err)
				}
			}
			if err := site.SetSiteName(root, "Sample"); err != nil {
				t.Fatal(err)
			}
			p, err := site.Load(root)
			if err != nil {
				t.Fatal(err)
			}
			before, _ := os.ReadFile(filepath.Join(root, ".flea.config.json"))
			var out bytes.Buffer
			changed, err := prepareDevOwner(context.Background(), p, strings.NewReader(tc.input), &out, tc.interactive)
			if (err != nil) != tc.wantErr || changed != tc.changed {
				t.Fatal(changed, err, out.String())
			}
			if strings.Contains(out.String(), "Resend") || strings.Contains(out.String(), "Sender") {
				t.Fatal("unexpected email prompt", out.String())
			}
			after, _ := os.ReadFile(filepath.Join(root, ".flea.config.json"))
			if !tc.changed && !bytes.Equal(before, after) {
				t.Fatal("unexpected config write")
			}
			if tc.changed {
				loaded, err := site.Load(root)
				if err != nil || len(loaded.Admin.Emails) != 1 {
					t.Fatal("owner missing", err)
				}
			}
			if !tc.interactive || tc.owner && tc.mail {
				if out.Len() != 0 {
					t.Fatal("unexpected prompts", out.String())
				}
			}
		})
	}
}

func TestDevNonTerminal(t *testing.T) {
	file, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if devTerminal(context.Background(), file) {
		t.Fatal("nonterminal prompted")
	}
}

func TestFirstDevSiteName(t *testing.T) {
	for _, answer := range []string{"", "My Custom Name"} {
		t.Run(answer, func(t *testing.T) {
			t.Setenv("RESEND_API_KEY", "re_test12345678")
			t.Setenv("FLEA_EMAIL_FROM", "hello@example.com")
			root := filepath.Join(t.TempDir(), "my-project")
			if err := Create(root); err != nil {
				t.Fatal(err)
			}
			if _, _, err := site.AddAdmin(root, "owner@example.com"); err != nil {
				t.Fatal(err)
			}
			p, err := site.Load(root)
			if err != nil {
				t.Fatal(err)
			}
			var out bytes.Buffer
			changed, err := prepareDevOwner(context.Background(), p, strings.NewReader(""), &out, false)
			if err != nil || changed || out.Len() != 0 {
				t.Fatal("noninteractive prompted", err)
			}
			changed, err = prepareDevOwner(context.Background(), p, strings.NewReader(answer+"\n"), &out, true)
			if err != nil || !changed || !strings.Contains(out.String(), "Site name [My Project]:") {
				t.Fatal(out.String(), err)
			}
			p, err = site.Load(root)
			if err != nil {
				t.Fatal(err)
			}
			want := answer
			if want == "" {
				want = "My Project"
			}
			if p.App["site_name"] != want {
				t.Fatal(p.App)
			}
			out.Reset()
			changed, err = prepareDevOwner(context.Background(), p, strings.NewReader(""), &out, true)
			if err != nil || changed || out.Len() != 0 {
				t.Fatal("prompt repeated", err, out.String())
			}
		})
	}
}
