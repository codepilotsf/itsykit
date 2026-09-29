package setup

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
)

func TestSavedTargetAsksForUsernameOnly(t *testing.T) {
	root := repoSite(t)
	c := DeploymentConfig{Branches: map[string]DeploySettings{"main": {Target: "server", Domain: "example.test", Name: "demo", ExpectedIPs: []string{"192.0.2.1"}}}}
	if err := saveDeploymentConfig(root, c); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	s, err := SelectDeployment(context.Background(), root, DeploySettings{}, nil, false, false, strings.NewReader("\n"), &out)
	if err != nil || s.Target != "root@server" {
		t.Fatal(s, err)
	}
	if strings.Contains(out.String(), "proxy") || strings.Contains(out.String(), "Site domain") {
		t.Fatal("repeated unrelated questions", out.String())
	}
	saved, err := readDeploymentConfig(root)
	if err != nil || saved.Branches["main"].Target != "root@server" || len(saved.Branches["main"].ExpectedIPs) != 1 {
		t.Fatal(saved, err)
	}
	out.Reset()
	if _, err = SelectDeployment(context.Background(), root, DeploySettings{}, nil, false, false, strings.NewReader(""), &out); err != nil || out.Len() != 0 {
		t.Fatal(err, out.String())
	}
}

func TestUsernamePrompt(t *testing.T) {
	for _, tc := range []struct{ target, input, want string }{{"server", "\n", "root@server"}, {"server", "ubuntu\n", "ubuntu@server"}, {"sean@server", "", "sean@server"}} {
		got, err := sshUsername(context.Background(), tc.target, strings.NewReader(tc.input), io.Discard)
		if err != nil || got != tc.want {
			t.Fatal(got, err)
		}
	}
	input := strings.NewReader("ubuntu\nremaining\n")
	if _, err := sshUsername(context.Background(), "server", input, io.Discard); err != nil {
		t.Fatal(err)
	}
	rest, _ := io.ReadAll(input)
	if string(rest) != "remaining\n" {
		t.Fatal("consumed later input", string(rest))
	}
}

func TestDeployOffersSetupOnlyForMissingServer(t *testing.T) {
	for _, mode := range []string{"install", "decline", "auth-failed", "outdated", "setup-failed"} {
		t.Run(mode, func(t *testing.T) {
			root := repoSite(t)
			installed := false
			installs := 0
			activated := false
			dns := false
			runner := func(_ context.Context, c command) error {
				args := strings.Join(c.args, " ")
				switch {
				case strings.Contains(args, "-O exit") || strings.Contains(args, "rm -f --"):
					return nil
				case strings.Contains(args, "FLEA_SYSTEM"):
					if mode == "auth-failed" {
						return errors.New("exit status 255")
					}
					fmt.Fprintln(c.out, "FLEA_SYSTEM=ubuntu:24.04:x86_64:0")
					if installed {
						fmt.Fprintln(c.out, "FLEA_DEPLOY=22")
					} else if mode == "outdated" {
						fmt.Fprintln(c.out, "FLEA_DEPLOY=1")
					} else {
						fmt.Fprintln(c.out, "FLEA_MISSING=1")
					}
					return nil
				case strings.Contains(args, "-check-dns"):
					dns = true
					return nil
				case strings.Contains(args, "mktemp"):
					fmt.Fprintln(c.out, "/tmp/flea-deploy.ABCDE12345")
					return nil
				case c.name == "scp":
					return nil
				case strings.Contains(args, " -deploy "):
					activated = true
					return nil
				default:
					t.Fatal(args)
					return nil
				}
			}
			install := func(_ context.Context, target string) error {
				installs++
				if target != "root@server" {
					t.Fatal(target)
				}
				if mode == "setup-failed" {
					return errors.New("failed")
				}
				installed = true
				return nil
			}
			answer := "y\n"
			if mode == "decline" {
				answer = "n\n"
			}
			var out bytes.Buffer
			err := deployWithSetup(context.Background(), root, DeploySettings{Target: "root@server", Domain: "example.test", Name: "demo"}, false, strings.NewReader(answer), &out, io.Discard, runner, install)
			if mode == "install" {
				if err != nil || installs != 1 || !activated || !dns {
					t.Fatal(err, installs, activated, dns)
				}
			} else {
				if err == nil || activated || dns {
					t.Fatal(err, activated, dns)
				}
				if mode != "setup-failed" && installs != 0 {
					t.Fatal("installed without consent or verified absence")
				}
				if mode == "auth-failed" && (strings.Contains(err.Error(), "255") || strings.Contains(out.String(), "Install FleaCMS")) {
					t.Fatal(err, out.String())
				}
			}
		})
	}
}
