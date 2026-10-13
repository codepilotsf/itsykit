package setup

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

type upgradeOutput struct {
	mu     sync.Mutex
	buffer bytes.Buffer
	tick   chan struct{}
}

func (w *upgradeOutput) Write(b []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	n, err := w.buffer.Write(b)
	if strings.HasPrefix(string(b), "\rUpdating server...") {
		select {
		case w.tick <- struct{}{}:
		default:
		}
	}
	return n, err
}
func (w *upgradeOutput) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buffer.String()
}

func TestUpgradeProgress(t *testing.T) {
	for _, tt := range []struct {
		name                                           string
		terminal, recover, fail, decline, sudo, cancel bool
		diagnostic                                     string
	}{
		{name: "terminal", terminal: true},
		{name: "redirected"},
		{name: "recovery", terminal: true, recover: true},
		{name: "failure", terminal: true, fail: true},
		{name: "cancelled", terminal: true, cancel: true},
		{name: "declined", terminal: true, decline: true},
		{name: "stderr", terminal: true, diagnostic: "stderr"},
		{name: "stdout", terminal: true, diagnostic: "stdout"},
		{name: "sudo", terminal: true, sudo: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			root := qaRecoverySite(t)
			if tt.sudo {
				config, err := readDeploymentConfig(root)
				if err != nil {
					t.Fatal(err)
				}
				s := config.Branches["main"]
				s.Target = "admin@server"
				config.Branches["main"] = s
				if err := saveDeploymentConfig(root, config); err != nil {
					t.Fatal(err)
				}
			}
			out := &upgradeOutput{tick: make(chan struct{}, 1)}
			var diagnostic bytes.Buffer
			upgraded := false
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			run := func(ctx context.Context, c command) error {
				if c.name == "extract-server" {
					if strings.Contains(out.String(), "Updating server...") {
						t.Fatal("animation started before confirmation")
					}
					return os.WriteFile(c.args[1], []byte("replacement executable"), 0700)
				}
				for _, arg := range c.args {
					if arg == "-O" {
						return nil
					}
				}
				cmd := c.args[len(c.args)-1]
				switch {
				case strings.Contains(cmd, "FLEA_SYSTEM"):
					fmt.Fprintln(c.out, "FLEA_SYSTEM=ubuntu:24.04:x86_64:0")
				case strings.Contains(cmd, "mktemp"):
					if tt.terminal {
						select {
						case <-out.tick:
						case <-time.After(2 * time.Second):
							t.Fatal("spinner did not animate")
						}
					}
					fmt.Fprintln(c.out, "/tmp/flea-upgrade.abcdefghij")
				case strings.HasPrefix(cmd, "cat >"):
					_, err := io.Copy(io.Discard, c.in)
					return err
				case strings.Contains(cmd, "-upgrade ") || strings.Contains(cmd, "-recover-upgrade"):
					upgraded = true
					if tt.sudo && !strings.HasSuffix(out.String(), "\r\x1b[2K") {
						t.Fatal("spinner continued into sudo prompt")
					}
					if tt.diagnostic != "" {
						target := c.out
						if tt.diagnostic == "stderr" {
							target = c.errOut
						}
						fmt.Fprintln(target, "Remote diagnostic")
						if !strings.Contains(out.String(), "\r\x1b[2K") {
							t.Fatal("spinner not cleared before diagnostics")
						}
					}
					if tt.cancel {
						cancel()
						return ctx.Err()
					}
					if tt.fail {
						return errors.New("upgrade failed")
					}
				case strings.HasPrefix(cmd, "chmod "), strings.HasPrefix(cmd, "rm -f"):
				default:
					t.Fatal("unexpected command", cmd)
				}
				return ctx.Err()
			}
			answer := "y\n"
			if tt.decline {
				answer = "n\n"
			}
			err := upgradeServer(ctx, root, tt.recover, strings.NewReader(answer), out, &diagnostic, tt.terminal, run)
			if (err != nil) != (tt.fail || tt.cancel || tt.decline) {
				t.Fatal(err)
			}
			text := out.String()
			if tt.decline {
				if upgraded || strings.Contains(text, "Updating server...") {
					t.Fatal("declined update started", text)
				}
				return
			}
			if !upgraded || !strings.Contains(text, "Updating server...") {
				t.Fatal(text)
			}
			if !tt.terminal && strings.ContainsAny(text, "\r\x1b") {
				t.Fatal("redirected output contains terminal controls", text)
			}
			if tt.fail || tt.cancel {
				if !strings.HasSuffix(text, "\r\x1b[2K") || strings.Contains(text, "complete.") {
					t.Fatal("failure did not clear spinner", text)
				}
			} else {
				want := "Server update complete.\n"
				if tt.recover {
					want = "Server update recovery complete.\n"
				}
				if !strings.HasSuffix(text, want) {
					t.Fatal(text)
				}
			}
		})
	}
}
