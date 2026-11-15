package setup

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestDomainRedirectsInDeploymentConfig(t *testing.T) {
	root := repoSite(t)
	s := DeploySettings{Target: "root@server", Name: "old", Domain: "old.test", WWWRedirect: true, HTTPS: true, Email: "dev@example.test", AgreeTOS: true, DomainRedirects: map[string]string{"old.test": "current.test", "www.old.test": "current.test"}}
	config := DeploymentConfig{Branches: map[string]DeploySettings{"main": s}}
	if err := saveDeploymentConfig(root, config); err != nil {
		t.Fatal(err)
	}
	got, err := readDeploymentConfig(root)
	if err != nil || !reflect.DeepEqual(got.Branches["main"].DomainRedirects, s.DomainRedirects) {
		t.Fatal(got, err)
	}
	if err := validateMapping("main", s, got); err != nil {
		t.Fatal(err)
	}
	for _, rules := range []map[string]string{
		{"foreign.test": "current.test"},
		{"old.test": "old.test"},
		{"old.test": "https://current.test/docs"},
	} {
		invalid := s
		invalid.DomainRedirects = rules
		if err := validateMapping("main", invalid, got); err == nil {
			t.Fatal("accepted invalid config redirect", rules)
		}
	}
}

func TestDeployTransmitsConfiguredDomainRedirects(t *testing.T) {
	root := repoSite(t)
	settings := DeploySettings{Target: "root@server", Name: "old", Domain: "old.test", HTTPS: true, Email: "dev@example.test", AgreeTOS: true, DomainRedirects: map[string]string{"old.test": "current.test"}}
	var activation string
	runner := func(_ context.Context, c command) error {
		text := strings.Join(c.args, " ")
		switch {
		case strings.Contains(text, "-O exit"):
		case strings.Contains(text, "FLEA_SYSTEM"):
			fmt.Fprintln(c.out, "FLEA_SYSTEM=ubuntu:24.04:x86_64:0\nFLEA_DEPLOY=30")
		case strings.Contains(text, "-check-dns"):
		default:
			activation = text
		}
		return nil
	}
	var out bytes.Buffer
	if err := deploy(context.Background(), root, settings, true, strings.NewReader(""), &out, io.Discard, runner); err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(settings.DomainRedirects)
	if !strings.Contains(activation, "-domain-redirects "+shellQuote(string(encoded))) || !strings.Contains(activation, "-name 'old'") {
		t.Fatal("configured redirect missing from activation", activation)
	}
}
