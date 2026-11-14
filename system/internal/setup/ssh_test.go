package setup

import (
	"strings"
	"testing"
)

func TestConnectionPolicyAndSCPHosts(t *testing.T) {
	options := connectionOptions("/tmp/test")
	got := strings.Join(options, " ")
	for _, required := range []string{"StrictHostKeyChecking=accept-new", "ControlPath=/tmp/test/ssh", "ConnectTimeout=10", "ServerAliveCountMax=3"} {
		if !strings.Contains(got, required) {
			t.Fatal(got)
		}
	}
	options[1] = "changed"
	if connectionOptions("/tmp/test")[1] != "StrictHostKeyChecking=accept-new" {
		t.Fatal("options share mutable state")
	}
	for target, want := range map[string]string{"root@example.test": "root@example.test", "my-alias": "my-alias", "root@2001:db8::1": "root@[2001:db8::1]", "2001:db8::1": "[2001:db8::1]"} {
		if got := scpHost(target); got != want {
			t.Errorf("%s: got %s, want %s", target, got, want)
		}
	}
}
