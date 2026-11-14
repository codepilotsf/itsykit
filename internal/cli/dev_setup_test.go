package cli

import (
	"bytes"
	"context"
	"itsy/internal/site"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDevOnboardingFirstRun(t *testing.T) {
	t.Setenv("RESEND_API_KEY", "")
	t.Setenv("FLEA_EMAIL_FROM", "")
	for _, tc := range []struct {
		name                 string
		input                string
		interactive          bool
		wantChanged, wantOff bool
		wantErr              bool
	}{
		{name: "add admin", input: "My Site\nowner@example.com\n", interactive: true, wantChanged: true, wantOff: true},
		{name: "skip admin", input: "\n\n", interactive: true, wantChanged: true, wantOff: true},
		{name: "invalid admin", input: "My Site\nbad\n", interactive: true, wantOff: true, wantErr: true},
		{name: "noninteractive"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := filepath.Join(t.TempDir(), "sample")
			if err := Create(root); err != nil {
				t.Fatal(err)
			}
			p, err := site.Load(root)
			if err != nil {
				t.Fatal(err)
			}
			before, _ := os.ReadFile(filepath.Join(root, ".itsy.config.json"))
			var out bytes.Buffer
			changed, offered, err := prepareDevOwner(context.Background(), p, strings.NewReader(tc.input), &out, tc.interactive)
			if (err != nil) != tc.wantErr || changed != tc.wantChanged || offered != tc.wantOff {
				t.Fatalf("changed=%v offered=%v err=%v out=%q", changed, offered, err, out.String())
			}
			if strings.Contains(out.String(), "Resend") || strings.Contains(out.String(), "Sender") {
				t.Fatalf("unexpected email prompt: %q", out.String())
			}
			if !tc.interactive {
				if out.Len() != 0 {
					t.Fatalf("noninteractive prompted: %q", out.String())
				}
				return
			}
			if tc.wantErr {
				return
			}
			after, _ := os.ReadFile(filepath.Join(root, ".itsy.config.json"))
			if !bytes.Equal(before, after) && !tc.wantChanged {
				t.Fatal("unexpected config write")
			}
			loaded, err := site.Load(root)
			if err != nil {
				t.Fatal(err)
			}
			if !loaded.SiteNameConfigured {
				t.Fatal("site name not saved on the first run")
			}
			if tc.name == "add admin" && len(loaded.Admin.Emails) != 1 {
				t.Fatalf("admin not saved: %v", loaded.Admin.Emails)
			}
		})
	}
}

// The skip guidance appears once and later runs never prompt again.
func TestDevOnboardingDoesNotRepeat(t *testing.T) {
	root := filepath.Join(t.TempDir(), "sample")
	if err := Create(root); err != nil {
		t.Fatal(err)
	}
	p, err := site.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	changed, offered, err := prepareDevOwner(context.Background(), p, strings.NewReader("\n\n"), &out, true)
	if err != nil || !changed || !offered {
		t.Fatalf("first run: changed=%v offered=%v err=%v out=%q", changed, offered, err, out.String())
	}
	if !strings.Contains(out.String(), "Add an admin") || !strings.Contains(out.String(), "itsy admin") {
		t.Fatalf("skip guidance missing: %q", out.String())
	}
	p, err = site.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	out.Reset()
	changed, offered, err = prepareDevOwner(context.Background(), p, strings.NewReader(""), &out, true)
	if err != nil || changed || offered || out.Len() != 0 {
		t.Fatalf("second run: changed=%v offered=%v err=%v out=%q", changed, offered, err, out.String())
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
			changed, offered, err := prepareDevOwner(context.Background(), p, strings.NewReader(""), &out, false)
			if err != nil || changed || offered || out.Len() != 0 {
				t.Fatal("noninteractive prompted", err)
			}
			changed, offered, err = prepareDevOwner(context.Background(), p, strings.NewReader(answer+"\n"), &out, true)
			if err != nil || !changed || offered || !strings.Contains(out.String(), "Site name [My Project]:") {
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
			changed, offered, err = prepareDevOwner(context.Background(), p, strings.NewReader(""), &out, true)
			if err != nil || changed || offered || out.Len() != 0 {
				t.Fatal("prompt repeated", err, out.String())
			}
		})
	}
}
