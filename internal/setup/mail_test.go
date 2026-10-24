package setup

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMailSetupKeepsSecretsOutOfArgumentsAndOutput(t *testing.T) {
	t.Setenv("RESEND_API_KEY", "")
	t.Setenv("FLEA_EMAIL_FROM", "")
	for _, ready := range []bool{false, true} {
		var output bytes.Buffer
		transferred, configured, cleaned := false, false, false
		run := func(_ context.Context, c command) error {
			args := strings.Join(c.args, " ")
			if strings.Contains(args, "re_test123456") {
				t.Fatal("key in command arguments")
			}
			switch {
			case strings.Contains(args, "-mail-status"):
				if ready {
					fmt.Fprintln(c.out, "FLEA_MAIL=ready")
				} else {
					fmt.Fprintln(c.out, "FLEA_MAIL=missing")
				}
			case strings.Contains(args, "mktemp"):
				fmt.Fprintln(c.out, "/tmp/flea-mail.ABCDE12345")
			case strings.Contains(args, "cat >"):
				if strings.Contains(args, "-tt") {
					t.Fatal("secret transfer has a PTY")
				}
				data, _ := io.ReadAll(c.in)
				if !bytes.Contains(data, []byte("re_test123456")) {
					t.Fatal("key not transferred")
				}
				transferred = true
			case strings.Contains(args, "-configure-mail"):
				if !strings.Contains(args, "sudo --") {
					t.Fatal("missing sudo")
				}
				configured = true
			case strings.Contains(args, "rm -f --"):
				cleaned = true
			default:
				t.Fatal(args)
			}
			return nil
		}
		err := ensureMail(context.Background(), t.TempDir(), "user@server", connectionOptions("/tmp/test"), false, strings.NewReader("re_test123456\nlogin@example.test\n"), &output, io.Discard, run)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(output.String(), "re_test123456") {
			t.Fatal("key in output")
		}
		if ready && (transferred || configured || cleaned) {
			t.Fatal("reconfigured saved settings")
		}
		if !ready && (!transferred || !configured || !cleaned) {
			t.Fatal("incomplete settings workflow")
		}
	}
}
func TestAdminDeployRequiresHTTPSBeforeSSH(t *testing.T) {
	root := repoSite(t)
	if err := os.WriteFile(filepath.Join(root, ".flea.config.json"), []byte(`{"app":{"site_name":"Test"},"admin":{"emails":["owner@example.test"]}}`), 0644); err != nil {
		t.Fatal(err)
	}
	testGit(t, root, "add", ".")
	testGit(t, root, "commit", "-m", "Enable admin")
	err := deploy(context.Background(), root, DeploySettings{Target: "server", Domain: "example.test", Name: "demo"}, false, strings.NewReader(""), io.Discard, io.Discard, func(context.Context, command) error { t.Fatal("connected before rejecting HTTP admin"); return nil })
	if err == nil || !strings.Contains(err.Error(), "requires HTTPS") {
		t.Fatal(err)
	}
}

func TestAdminDeployConfiguresMailBeforeActivation(t *testing.T) {
	t.Setenv("RESEND_API_KEY", "re_test123456")
	t.Setenv("FLEA_EMAIL_FROM", "login@example.test")
	root := repoSite(t)
	if err := os.WriteFile(filepath.Join(root, ".flea.config.json"), []byte(`{"app":{"site_name":"Test"},"admin":{"emails":["owner@example.test"]}}`), 0644); err != nil {
		t.Fatal(err)
	}
	testGit(t, root, "add", ".")
	testGit(t, root, "commit", "-m", "Enable admin")
	configured, activated := false, false
	run := func(_ context.Context, c command) error {
		args := strings.Join(c.args, " ")
		if strings.Contains(args, "re_test123456") {
			t.Fatal("secret in argv")
		}
		switch {
		case strings.Contains(args, "FLEA_SYSTEM"):
			fmt.Fprintln(c.out, "FLEA_SYSTEM=ubuntu:24.04:x86_64:0\nFLEA_DEPLOY=27")
		case strings.Contains(args, "-mail-status"):
			fmt.Fprintln(c.out, "FLEA_MAIL=missing")
		case strings.Contains(args, "mktemp -d /tmp/flea-mail"):
			fmt.Fprintln(c.out, "/tmp/flea-mail.ABCDE12345")
		case strings.Contains(args, "mktemp -d /tmp/flea-deploy"):
			fmt.Fprintln(c.out, "/tmp/flea-deploy.ABCDE12345")
		case strings.Contains(args, "cat >"):
			data, _ := io.ReadAll(c.in)
			if !bytes.Contains(data, []byte("re_test123456")) {
				t.Fatal("missing mail payload")
			}
		case strings.Contains(args, "-configure-mail"):
			configured = true
		case strings.Contains(args, " -deploy "):
			if !configured {
				t.Fatal("activated before mail configuration")
			}
			activated = true
		}
		return nil
	}
	err := deploy(context.Background(), root, DeploySettings{Target: "root@server", Domain: "example.test", Name: "demo", HTTPS: true, Email: "owner@example.test", AgreeTOS: true}, false, strings.NewReader(""), io.Discard, io.Discard, run)
	if err != nil || !activated {
		t.Fatal(err, activated)
	}
}
