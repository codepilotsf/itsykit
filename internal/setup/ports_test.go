package setup

import (
	"os/exec"
	"strings"
	"testing"
)

func TestInstallerPortOwnership(t *testing.T) {
	start := strings.Index(installer, "check_web_ports() {")
	end := strings.Index(installer, "# This must precede")
	if start < 0 || end < start {
		t.Fatal("missing preflight")
	}
	function := installer[start:end]
	for _, tc := range []struct {
		name, tcp, udp, pid, exe, mode string
		fail                           bool
	}{
		{name: "free"},
		{name: "http", tcp: `LISTEN 0 128 0.0.0.0:80 0.0.0.0:* users:(("nginx",pid=42,fd=1))`, fail: true},
		{name: "https-ipv6", tcp: `LISTEN 0 128 [::]:443 [::]:* users:(("apache",pid=42,fd=1))`, fail: true},
		{name: "quic", udp: `UNCONN 0 0 *:443 *:* users:(("server",pid=42,fd=1))`, fail: true},
		{name: "unknown-owner", tcp: `LISTEN 0 128 *:80 *:*`, fail: true},
		{name: "flea-reinstall", tcp: `LISTEN 0 128 *:80 *:* users:(("flea-server",pid=42,fd=1))`, pid: "42", exe: "/usr/local/bin/flea-server", mode: "--reinstall"},
		{name: "flea-fresh", tcp: `LISTEN 0 128 *:80 *:* users:(("flea-server",pid=42,fd=1))`, pid: "42", exe: "/usr/local/bin/flea-server", fail: true},
		{name: "wrong-executable", tcp: `LISTEN 0 128 *:80 *:* users:(("server",pid=42,fd=1))`, pid: "42", exe: "/usr/sbin/nginx", mode: "--reinstall", fail: true},
		{name: "shared-socket", tcp: `LISTEN 0 128 *:443 *:* users:(("flea-server",pid=42,fd=1),("other",pid=43,fd=2))`, pid: "42", exe: "/usr/local/bin/flea-server", mode: "--reinstall", fail: true},
		{name: "ss-error", tcp: "ERROR", fail: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			script := "set -eu\n" + function + `
ss() {
 case "$*" in *'-lntp'*) value=$TEST_TCP;; *'-lnup'*) value=$TEST_UDP;; *) exit 99;; esac
 test "$value" != ERROR || return 1
 test -z "$value" || printf '%s\n' "$value"
 return 0
}
systemctl() { printf '%s\n' "$TEST_PID"; }
readlink() { printf '%s\n' "$TEST_EXE"; }
check_web_ports "$TEST_MODE"
`
			cmd := exec.Command("sh", "-c", script)
			cmd.Env = append(cmd.Environ(), "TEST_TCP="+tc.tcp, "TEST_UDP="+tc.udp, "TEST_PID="+tc.pid, "TEST_EXE="+tc.exe, "TEST_MODE="+tc.mode)
			out, err := cmd.CombinedOutput()
			if (err != nil) != tc.fail {
				t.Fatalf("%v: %s", err, out)
			}
		})
	}
	if strings.Index(installer, "check_web_ports \"${2:-}\"") > strings.Index(installer, "backup=$(mktemp") {
		t.Fatal("port check occurs after changes")
	}
}
