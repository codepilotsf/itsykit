package setup

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"flea/internal/release"
)

func TestDeploySSHWorkflow(t *testing.T) {
	root := t.TempDir()
	os.Mkdir(filepath.Join(root, "."), 0755)
	os.Mkdir(filepath.Join(root, "site"), 0755)
	os.WriteFile(filepath.Join(root, "site", "index.html"), []byte("Hello"), 0644)
	os.WriteFile(filepath.Join(root, ".flea.config.json"), []byte(`{"app":{"site_name":"test"}}`), 0644)
	initRepo(t, root)
	for _, fail := range []string{"", "connect", "dns", "upload", "activate", "cancel-upload"} {
		t.Run(fail, func(t *testing.T) {
			tempRoot := t.TempDir()
			t.Setenv("TMPDIR", tempRoot)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			var out bytes.Buffer
			var stages []string
			var scratch string
			var cleanedRemote, closedSSH bool
			runner := func(ctx context.Context, c command) error {
				for _, arg := range c.args {
					if strings.HasPrefix(arg, "ControlPath=") {
						scratch = filepath.Dir(strings.TrimPrefix(arg, "ControlPath="))
					}
				}
				if filepath.Dir(scratch) != tempRoot {
					t.Fatalf("scratch %q is outside TMPDIR %q", scratch, tempRoot)
				}
				args := strings.Join(c.args, " ")
				if strings.Contains(args, "-O exit") || strings.Contains(args, "rm -f --") {
					if ctx.Err() != nil {
						t.Fatal("cleanup used cancelled context", ctx.Err())
					}
					closedSSH = closedSSH || strings.Contains(args, "-O exit")
					cleanedRemote = cleanedRemote || strings.Contains(args, "rm -f --")
					return nil
				}
				if strings.Contains(args, "StrictHostKeyChecking=no") || strings.Contains(args, "BatchMode=yes") {
					t.Fatal(args)
				}
				stage := ""
				switch {
				case strings.Contains(args, "FLEA_SYSTEM"):
					stage = "connect"
					fmt.Fprintln(c.out, "FLEA_SYSTEM=ubuntu:24.04:x86_64:1000\nFLEA_DEPLOY=28")
				case strings.Contains(args, "-check-dns"):
					stage = "dns"
				case strings.Contains(args, "mktemp"):
					fmt.Fprintln(c.out, "/tmp/flea-deploy.ABCDE12345")
					return nil
				case c.name == "scp":
					stage = "upload"
					archive, err := os.Open(c.args[len(c.args)-2])
					if err != nil {
						t.Fatal(err)
					}
					defer archive.Close()
					extracted := t.TempDir()
					if err := release.Extract(archive, extracted); err != nil {
						t.Fatal(err)
					}
					if body, err := os.ReadFile(filepath.Join(extracted, "site", "index.html")); err != nil || string(body) != "Hello" {
						t.Fatalf("packaged page = %q, %v", body, err)
					}
				case strings.Contains(args, " -deploy "):
					stage = "activate"
					if !strings.Contains(args, "-tt -- ubuntu@server sudo --") || !strings.Contains(args, "-domain 'example.test'") {
						t.Fatal(args)
					}
				default:
					t.Fatal("unexpected command", args)
				}
				stages = append(stages, stage)
				if stage == "upload" && fail == "cancel-upload" {
					cancel()
					return ctx.Err()
				}
				if stage == fail {
					return errors.New("simulated " + stage)
				}
				return nil
			}
			err := deploy(ctx, root, DeploySettings{Target: "ubuntu@server", Domain: "example.test", Name: "demo"}, false, strings.NewReader(""), &out, io.Discard, runner)
			if _, statErr := os.Stat(scratch); !errors.Is(statErr, os.ErrNotExist) {
				t.Fatalf("scratch directory remains: %s (%v)", scratch, statErr)
			}
			if fail == "" {
				if err != nil || strings.Join(stages, ",") != "connect,dns,upload,activate" || !strings.Contains(out.String(), "http://example.test") {
					t.Fatal(err, stages, out.String())
				}
			} else if fail == "cancel-upload" {
				if !errors.Is(err, context.Canceled) || !cleanedRemote || strings.Contains(out.String(), "Deployed main branch") {
					t.Fatal(err, cleanedRemote, out.String())
				}
			} else if err == nil || len(stages) == 0 || stages[len(stages)-1] != fail || strings.Contains(out.String(), "Deployed main branch") {
				t.Fatal(err, out.String())
			}
			if !closedSSH {
				t.Fatal("SSH control connection was not closed")
			}
		})
	}
}

func TestHTTPSCommandAndInvalidInput(t *testing.T) {
	root := repoSite(t)
	var commandText string
	var progress bytes.Buffer
	runner := func(_ context.Context, c command) error {
		text := strings.Join(c.args, " ")
		if strings.Contains(text, "FLEA_SYSTEM") {
			fmt.Fprintln(c.out, "FLEA_SYSTEM=ubuntu:24.04:x86_64:0\nFLEA_DEPLOY=28")
			return nil
		}
		if strings.Contains(text, "-O exit") {
			return nil
		}
		if c.name != "ssh" {
			t.Fatal("HTTPS should not upload a site")
		}
		commandText = text
		return nil
	}
	s := DeploySettings{Target: "root@server", Domain: "example.test", Email: "dev@example.test", AgreeTOS: true}
	if err := deploy(context.Background(), root, s, true, strings.NewReader(""), &progress, io.Discard, runner); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(progress.String(), "HTTPS certificates") || strings.Contains(progress.String(), "two minutes") {
		t.Fatal("client inferred certificate state without server evidence", progress.String())
	}
	if !strings.Contains(commandText, "-https -email 'dev@example.test' -agree-tos") || strings.Contains(commandText, "sudo") {
		t.Fatal(commandText)
	}
	s.AgreeTOS = false
	if err := deploy(context.Background(), root, s, true, strings.NewReader(""), io.Discard, io.Discard, func(context.Context, command) error { t.Fatal("connected before validating"); return nil }); err == nil {
		t.Fatal("missing consent accepted")
	}
}

func TestDirtyDuringDNSStopsUpload(t *testing.T) {
	root := repoSite(t)
	runner := func(_ context.Context, c command) error {
		text := strings.Join(c.args, " ")
		switch {
		case strings.Contains(text, "-O exit"):
			return nil
		case strings.Contains(text, "FLEA_SYSTEM"):
			fmt.Fprintln(c.out, "FLEA_SYSTEM=ubuntu:24.04:x86_64:0\nFLEA_DEPLOY=28")
			return nil
		case strings.Contains(text, "-check-dns"):
			return os.WriteFile(filepath.Join(root, "new-file"), []byte("changed during preflight"), 0644)
		default:
			t.Fatal("continued after repository became dirty", text)
			return nil
		}
	}
	err := deploy(context.Background(), root, DeploySettings{Target: "server", Domain: "example.test", Name: "demo"}, false, strings.NewReader(""), io.Discard, io.Discard, runner)
	if err == nil || err.Error() != "Please commit or discard your changes before deploying." {
		t.Fatal(err)
	}
}
