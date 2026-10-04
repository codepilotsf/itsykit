package appapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestJSONRedirectPolicy(t *testing.T) {
	for _, tc := range []struct {
		destination string
		valid       bool
	}{
		{"thanks", true}, {"nested/thanks", true}, {"", false}, {"/thanks", false}, {"//evil.test", false}, {"https://evil.test", false}, {"../thanks", false}, {"%2e%2e/thanks", false}, {"thanks?next=x", false}, {"thanks#x", false}, {"_flea/handshake", false}, {"thanks\\other", false},
	} {
		t.Run(tc.destination, func(t *testing.T) {
			c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]any{"redirect": tc.destination, "flash": map[string]string{"notice": "Sent"}})
			})
			c.ready.Store(true)
			e, status, err := c.Request(context.Background(), Mount{"contact", "/deep/contact"}, "POST", "/deep/contact", "", map[string]any{})
			if (err == nil) != tc.valid {
				t.Fatal(status, err)
			}
			if tc.valid && e.Redirect != tc.destination {
				t.Fatal(e)
			}
		})
	}
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Envelope{Flash: map[string]string{"notice": strings.Repeat("x", 4096)}})
	})
	c.ready.Store(true)
	if _, _, err := c.Request(context.Background(), Mount{"contact", "/contact"}, "GET", "/contact", "", nil); err == nil {
		t.Fatal("unbounded flash")
	}
}
