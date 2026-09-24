package sitefiles

import "testing"

func TestSessionsAreNeverRuntimeContent(t *testing.T) {
	for _, name := range []string{"_/sessions.json", "_/.sessions-lock", "_/.sessions-temp"} {
		if Included(name) {
			t.Fatalf("private session state included: %s", name)
		}
	}
}
