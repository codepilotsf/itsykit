package setup

import (
	"bytes"
	"context"
	"debug/elf"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestTargetNames(t *testing.T) {
	for _, name := range []string{"1.2.3.4", "server.example.com", "ubuntu@my-vps", "root@2001:db8::1", "[::1]"} {
		if _, err := targetName(name); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
	for _, name := range []string{"", "-oProxyCommand=bad", "host;id", "user@host:22", "a@b@c", "https://host", "host\ncommand", "$(whoami)", "root@host/path"} {
		if _, err := targetName(name); err == nil {
			t.Errorf("accepted %q", name)
		}
	}
}

func TestProbe(t *testing.T) {
	for _, test := range []struct {
		data, arch string
		root       bool
	}{
		{"FLEA_SYSTEM=ubuntu:24.04:x86_64:0\n", "amd64", true},
		{"banner\nFLEA_SYSTEM=ubuntu:26.04:aarch64:1000\n", "arm64", false},
	} {
		arch, root, err := parseProbe(test.data)
		if err != nil || arch != test.arch || root != test.root {
			t.Fatal(arch, root, err)
		}
	}
	for _, data := range []string{"", "FLEA_SYSTEM=debian:12:x86_64:0", "FLEA_SYSTEM=ubuntu:25.10:x86_64:0", "FLEA_SYSTEM=ubuntu:24.04:riscv64:0", "FLEA_SYSTEM=ubuntu:24.04:x86_64:bad"} {
		if _, _, err := parseProbe(data); err == nil {
			t.Fatal(data)
		}
	}
}

func TestReinstallRequiresTypedConfirmation(t *testing.T) {
	for _, answer := range []string{"\n", "yes\n", "WIPE\n", "destroy\n", ""} {
		var output bytes.Buffer
		err := run(context.Background(), strings.NewReader("server\n"+answer), &output, io.Discard, func(_ context.Context, c command) error {
			if c.name == "extract-server" {
				return nil
			}
			args := strings.Join(c.args, " ")
			if strings.Contains(args, "-O exit") {
				return nil
			}
			if !strings.Contains(args, "FLEA_SYSTEM") || !strings.Contains(args, "root@server") {
				t.Fatalf("unexpected command: %s", args)
			}
			fmt.Fprintln(c.out, "FLEA_SYSTEM=ubuntu:24.04:x86_64:0\nFLEA_EXISTING=1")
			return nil
		})
		if err == nil || !strings.Contains(output.String(), "Type DESTROY") {
			t.Fatal(err, output.String())
		}
	}
}

func TestUnavailableReplacementDoesNotAskToDestroy(t *testing.T) {
	var output bytes.Buffer
	want := errors.New("replacement unavailable")
	err := run(context.Background(), strings.NewReader("server\n"), &output, io.Discard, func(_ context.Context, c command) error {
		args := strings.Join(c.args, " ")
		switch {
		case c.name == "extract-server":
			return want
		case strings.Contains(args, "-O exit"):
			return nil
		case strings.Contains(args, "FLEA_SYSTEM"):
			fmt.Fprintln(c.out, "FLEA_SYSTEM=ubuntu:24.04:x86_64:0\nFLEA_EXISTING=1")
			return nil
		default:
			t.Fatalf("unexpected remote change: %s", args)
			return nil
		}
	})
	if !errors.Is(err, want) || strings.Contains(output.String(), "DESTROY") || strings.Contains(output.String(), "WARNING") {
		t.Fatal(err, output.String())
	}
}

func TestSetupExtractsRealBundleBeforeResetPrompt(t *testing.T) {
	if os.Getenv("FLEA_REQUIRE_BUNDLE") != "1" {
		t.Skip("requires generated server bundles")
	}
	for _, arch := range []string{"x86_64", "aarch64"} {
		t.Run(arch, func(t *testing.T) {
			var output bytes.Buffer
			prepared := false
			err := run(context.Background(), strings.NewReader("server\n\n"), &output, io.Discard, func(ctx context.Context, c command) error {
				args := strings.Join(c.args, " ")
				switch {
				case c.name == "extract-server":
					if strings.Contains(output.String(), "DESTROY") {
						t.Fatal("prompted before extraction")
					}
					if err := execute(ctx, c); err != nil {
						return err
					}
					prepared = true
					return nil
				case strings.Contains(args, "FLEA_SYSTEM"):
					fmt.Fprintf(c.out, "FLEA_SYSTEM=ubuntu:24.04:%s:0\nFLEA_EXISTING=1\n", arch)
					return nil
				case strings.Contains(args, "-O exit"):
					return nil
				default:
					t.Fatalf("changed server after cancellation: %s", args)
					return nil
				}
			})
			if !prepared || err == nil || !strings.Contains(err.Error(), "Setup cancelled") || !strings.Contains(output.String(), "Type DESTROY") {
				t.Fatal(prepared, err, output.String())
			}
		})
	}
}

func TestSetupStages(t *testing.T) {
	for _, failure := range []string{"", "reinstall", "connect", "extract", "upload", "install", "check"} {
		t.Run(failure, func(t *testing.T) {
			tempRoot := t.TempDir()
			t.Setenv("TMPDIR", tempRoot)
			var scratch string
			var output bytes.Buffer
			var stages []string
			fake := func(_ context.Context, c command) error {
				for _, arg := range c.args {
					if strings.HasPrefix(arg, "ControlPath=") {
						scratch = filepath.Dir(strings.TrimPrefix(arg, "ControlPath="))
					}
				}
				if filepath.Dir(scratch) != tempRoot {
					t.Fatalf("setup scratch %q is outside TMPDIR %q", scratch, tempRoot)
				}
				joined := strings.Join(c.args, " ")
				if strings.Contains(joined, "-O exit") || strings.Contains(joined, "rm -f --") {
					return nil
				}
				if strings.Contains(joined, "StrictHostKeyChecking=no") || strings.Contains(joined, "BatchMode=yes") {
					t.Fatal("disabled host verification or interactive authentication")
				}
				stage := ""
				switch {
				case c.name == "ssh" && strings.Contains(joined, "FLEA_SYSTEM"):
					stage = "connect"
					fmt.Fprintln(c.out, "FLEA_SYSTEM=ubuntu:24.04:aarch64:1000")
					if failure == "reinstall" {
						fmt.Fprintln(c.out, "FLEA_EXISTING=1")
					}
				case c.name == "ssh" && strings.Contains(joined, "mktemp"):
					fmt.Fprintln(c.out, "/tmp/flea-setup.ABCDE12345")
					return nil
				case c.name == "scp":
					stage = "upload"
					for _, name := range c.args {
						if strings.HasSuffix(name, "/install.sh") {
							data, err := os.ReadFile(name)
							if err != nil || string(data) != installer {
								t.Fatal(err)
							}
						}
					}
				case c.name == "ssh" && strings.Contains(joined, "install.sh"):
					stage = "install"
					if failure == "reinstall" && !strings.Contains(joined, "--reinstall") {
						t.Fatal(joined)
					}
					if !strings.Contains(joined, "-tt -- root@server") || !strings.Contains(joined, "sudo -- sh") {
						t.Fatal(joined)
					}
				case c.name == "ssh" && strings.Contains(joined, "is-active"):
					stage = "check"
				default:
					stage = "extract"
					if c.name != "extract-server" || c.args[0] != "arm64" {
						t.Fatal(c)
					}
				}
				stages = append(stages, stage)
				if stage == failure {
					return errors.New("simulated " + stage + " failure")
				}
				return nil
			}
			input := "server\n"
			if failure == "reinstall" {
				input += "DESTROY\n"
			}
			err := run(context.Background(), strings.NewReader(input), &output, io.Discard, fake)
			if _, statErr := os.Stat(scratch); !errors.Is(statErr, os.ErrNotExist) {
				t.Fatalf("setup scratch remains: %s (%v)", scratch, statErr)
			}
			if failure == "" || failure == "reinstall" {
				if err != nil || strings.Join(stages, ",") != "connect,extract,upload,install,check" || !strings.Contains(output.String(), "ItsyWeb server is running") {
					t.Fatal(stages, err, output.String())
				}
			} else {
				if err == nil || stages[len(stages)-1] != failure || strings.Contains(output.String(), "server is running") {
					t.Fatal(stages, err, output.String())
				}
				if failure == "connect" && strings.Contains(output.String(), "Preparing bundled") {
					t.Fatal("extracted despite failed SSH")
				}
			}
		})
	}
}

func TestCancelledPromptAndInvalidTarget(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	reader, writer := io.Pipe()
	defer reader.Close()
	defer writer.Close()
	if _, err := ask(ctx, reader, io.Discard, "Server hostname or IP", ""); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	err := run(context.Background(), strings.NewReader("-bad\n"), io.Discard, io.Discard, func(context.Context, command) error { t.Fatal("ran SSH on invalid target"); return nil })
	if err == nil {
		t.Fatal("invalid target accepted")
	}
}

func TestLinuxServerBuilds(t *testing.T) {
	if testing.Short() {
		t.Skip("cross-compilation")
	}
	source, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	goBinary, err := findGo()
	if err != nil {
		t.Fatal(err)
	}
	for _, arch := range []string{"amd64", "arm64"} {
		t.Run(arch, func(t *testing.T) {
			binary := filepath.Join(t.TempDir(), "flea-server")
			cmd := exec.Command(goBinary, "build", "-o", binary, "./cmd/itsy-server")
			cmd.Dir = source
			cmd.Env = buildEnv(arch)
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("%v: %s", err, output)
			}
			file, err := elf.Open(binary)
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()
			want := elf.EM_X86_64
			if arch == "arm64" {
				want = elf.EM_AARCH64
			}
			if file.Machine != want || file.Section(".interp") != nil {
				t.Fatalf("wrong target or dynamically linked build: %v", file.Machine)
			}
		})
	}
}

func TestInstallerSyntaxAndService(t *testing.T) {
	var config struct {
		Listen string            `json:"listen"`
		Sites  map[string]string `json:"sites"`
	}
	if err := json.Unmarshal([]byte(defaultConfig), &config); err != nil || config.Listen != ":80" || len(config.Sites) != 0 {
		t.Fatalf("invalid default config: %s (%v)", defaultConfig, err)
	}
	cmd := exec.Command("sh", "-n", "install.sh")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatal(err, string(output))
	}
	for _, setting := range []string{"User=flea", "NoNewPrivileges=true", "AmbientCapabilities=CAP_NET_BIND_SERVICE", "CapabilityBoundingSet=CAP_NET_BIND_SERVICE", "ProtectSystem=strict", "ExecStart=/usr/local/bin/flea-server"} {
		if !strings.Contains(service, setting) {
			t.Fatal(setting)
		}
	}
	for _, guard := range []string{"Refusing to replace existing", "Another service is using port 80 or 443", "sport = :80", "sport = :443", "systemctl is-active --quiet", "-check", "trap cleanup EXIT"} {
		if !strings.Contains(installer, guard) {
			t.Fatal(guard)
		}
	}
}
