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
		{"NANOWEB_SYSTEM=ubuntu:24.04:x86_64:0\n", "amd64", true},
		{"banner\nNANOWEB_SYSTEM=ubuntu:26.04:aarch64:1000\n", "arm64", false},
	} {
		arch, root, err := parseProbe(test.data)
		if err != nil || arch != test.arch || root != test.root {
			t.Fatal(arch, root, err)
		}
	}
	for _, data := range []string{"", "NANOWEB_SYSTEM=debian:12:x86_64:0", "NANOWEB_SYSTEM=ubuntu:25.10:x86_64:0", "NANOWEB_SYSTEM=ubuntu:24.04:riscv64:0", "NANOWEB_SYSTEM=ubuntu:24.04:x86_64:bad"} {
		if _, _, err := parseProbe(data); err == nil {
			t.Fatal(data)
		}
	}
}

func TestSetupStages(t *testing.T) {
	source, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, failure := range []string{"", "connect", "build", "upload", "install", "check"} {
		t.Run(failure, func(t *testing.T) {
			var output bytes.Buffer
			var stages []string
			fake := func(_ context.Context, c command) error {
				joined := strings.Join(c.args, " ")
				if strings.Contains(joined, "-O exit") || strings.Contains(joined, "rm -f --") {
					return nil
				}
				if strings.Contains(joined, "StrictHostKeyChecking=no") || strings.Contains(joined, "BatchMode=yes") {
					t.Fatal("disabled host verification or interactive authentication")
				}
				stage := ""
				switch {
				case c.name == "ssh" && strings.Contains(joined, "NANOWEB_SYSTEM"):
					stage = "connect"
					fmt.Fprintln(c.out, "NANOWEB_SYSTEM=ubuntu:24.04:aarch64:1000")
				case c.name == "ssh" && strings.Contains(joined, "mktemp"):
					fmt.Fprintln(c.out, "/tmp/nanoweb-setup.ABCDE12345")
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
					if !strings.Contains(joined, "-tt -- ubuntu@server") || !strings.Contains(joined, "sudo -- sh") {
						t.Fatal(joined)
					}
				case c.name == "ssh" && strings.Contains(joined, "is-active"):
					stage = "check"
				default:
					stage = "build"
					if c.dir != source || !strings.Contains(strings.Join(c.env, "\n"), "GOARCH=arm64") || !strings.Contains(strings.Join(c.env, "\n"), "CGO_ENABLED=0") {
						t.Fatal(c)
					}
				}
				stages = append(stages, stage)
				if stage == failure {
					return errors.New("simulated " + stage + " failure")
				}
				return nil
			}
			err := run(context.Background(), source, strings.NewReader("ubuntu@server\n"), &output, io.Discard, fake)
			if failure == "" {
				if err != nil || strings.Join(stages, ",") != "connect,build,upload,install,check" || !strings.Contains(output.String(), "Nanoweb server is running") {
					t.Fatal(stages, err, output.String())
				}
			} else {
				if err == nil || stages[len(stages)-1] != failure || strings.Contains(output.String(), "server is running") {
					t.Fatal(stages, err, output.String())
				}
				if failure == "connect" && strings.Contains(output.String(), "Building") {
					t.Fatal("built despite failed SSH")
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
	if _, err := readTarget(ctx, reader); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	err := run(context.Background(), "", strings.NewReader("-bad\n"), io.Discard, io.Discard, func(context.Context, command) error { t.Fatal("ran SSH on invalid target"); return nil })
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
			binary := filepath.Join(t.TempDir(), "nanoweb-server")
			cmd := exec.Command(goBinary, "build", "-o", binary, "./cmd/nanoweb-server")
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
	for _, setting := range []string{"User=nanoweb", "NoNewPrivileges=true", "AmbientCapabilities=CAP_NET_BIND_SERVICE", "CapabilityBoundingSet=CAP_NET_BIND_SERVICE", "ProtectSystem=strict", "ExecStart=/usr/local/bin/nanoweb-server"} {
		if !strings.Contains(service, setting) {
			t.Fatal(setting)
		}
	}
	for _, guard := range []string{"Refusing to replace existing", "Port 80 is already in use", "sport = :80", "systemctl is-active --quiet", "-check", "trap cleanup EXIT"} {
		if !strings.Contains(installer, guard) {
			t.Fatal(guard)
		}
	}
}
