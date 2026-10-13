package setup

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

func TestQAUpgradeRecoveryDoesNotRequireWorkingInstalledBinary(t *testing.T) {
	root := qaRecoverySite(t)
	recovered := false
	run := func(_ context.Context, c command) error {
		if c.name == "extract-server" {
			return os.WriteFile(c.args[1], []byte("replacement executable"), 0700)
		}
		for _, a := range c.args {
			if a == "-O" {
				return nil
			}
		}
		cmd := c.args[len(c.args)-1]
		switch {
		case strings.Contains(cmd, "/usr/local/bin/flea-server"):
			return errors.New("installed executable is unavailable")
		case strings.Contains(cmd, "FLEA_SYSTEM"):
			fmt.Fprintln(c.out, "FLEA_SYSTEM=ubuntu:24.04:x86_64:0")
		case strings.Contains(cmd, "mktemp"):
			fmt.Fprintln(c.out, "/tmp/flea-upgrade.abcdefghij")
		case strings.HasPrefix(cmd, "cat >"):
			_, e := io.Copy(io.Discard, c.in)
			return e
		case strings.Contains(cmd, "-recover-upgrade"):
			recovered = true
		case strings.HasPrefix(cmd, "chmod "), strings.HasPrefix(cmd, "rm -f"):
			return nil
		default:
			t.Fatal("unexpected command", cmd)
		}
		return nil
	}
	if e := upgradeServer(context.Background(), root, true, strings.NewReader("y\n"), io.Discard, io.Discard, false, run); e != nil {
		t.Fatal(e)
	}
	if !recovered {
		t.Fatal("did not execute uploaded recovery helper")
	}
}
