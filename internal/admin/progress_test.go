package admin

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"flea/internal/host"
)

func TestActivationProgress(t *testing.T) {
	configured := host.Config{TLS: &host.TLSConfig{Domains: []string{"example.test"}}}
	for _, tc := range []struct {
		name    string
		config  host.Config
		options Options
		want    string
	}{
		{"first HTTPS", host.Config{}, Options{HTTPS: true, Domains: []string{"example.test"}}, "Setting up HTTPS certificates."},
		{"existing HTTPS", configured, Options{HTTPS: true, Domains: []string{"example.test"}}, "Checking site health and HTTPS..."},
		{"new alias", configured, Options{HTTPS: true, Domains: []string{"example.test", "www.example.test"}}, "Setting up HTTPS certificates."},
		{"disable HTTPS", configured, Options{Domains: []string{"example.test"}}, "Checking site health..."},
		{"HTTP only", host.Config{}, Options{Domains: []string{"example.test"}}, "Checking site health..."},
		{"rollback retains HTTPS", configured, Options{keepRouting: true, Domains: []string{"example.test"}}, "Checking site health and HTTPS..."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := activationProgress(tc.config, tc.options); !strings.HasPrefix(got, tc.want) {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}

func TestDeploymentProgressUsesInstalledHTTPSState(t *testing.T) {
	m, o := fixture(t)
	if err := os.WriteFile(m.ConfigPath, []byte(`{"listen":":80","sites":{}}`), 0644); err != nil {
		t.Fatal(err)
	}
	o.HTTPS = true
	o.Email = "dev@example.test"
	o.AgreeTOS = true
	var out bytes.Buffer
	m.Progress = &out
	checks := 0
	m.Verify = func(context.Context, host.Config, []string) error { checks++; return nil }
	if _, err := m.Apply(context.Background(), o); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Setting up HTTPS certificates.") {
		t.Fatal(out.String())
	}
	out.Reset()
	if _, err := m.Apply(context.Background(), o); err != nil {
		t.Fatal(err)
	}
	if out.String() != "Checking site health and HTTPS...\n" || checks != 2 {
		t.Fatal(out.String(), checks)
	}
}
