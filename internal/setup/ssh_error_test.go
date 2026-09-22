package setup

import (
	"bytes"
	"testing"
)

func TestSSHChangedKeyDetection(t *testing.T) {
	for _, tc := range []struct {
		text    string
		changed bool
	}{
		{"WARNING: REMOTE HOST IDENTIFICATION HAS CHANGED!", true},
		{"WARNING: REVOKED HOST KEY DETECTED!", true},
		{"Permission denied (publickey).", false},
		{"Warning: Permanently added 'server' to the list of known hosts.", false},
	} {
		var output bytes.Buffer
		w := &sshErrorWriter{out: &output}
		for _, b := range []byte(tc.text) {
			w.Write([]byte{b})
		}
		if w.changed != tc.changed || output.String() != tc.text {
			t.Fatal(tc, w.changed, output.String())
		}
	}
}
