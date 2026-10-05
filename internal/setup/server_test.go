package setup

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
)

const testSiteJSON = `{"name":"demo","domains":["example.test","www.example.test"],"release":"20260101-120000-abcdef","state":"digest","managed":true,"retained":"/var/lib/flea/sites/demo"}`

func TestServerSitesWorkflow(t *testing.T) {
	runner := func(_ context.Context, c command) error {
		text := strings.Join(c.args, " ")
		switch {
		case strings.Contains(text, "-O exit"):
			return nil
		case strings.Contains(text, "FLEA_UID"):
			fmt.Fprint(c.out, "FLEA_UID=0\nFLEA_DEPLOY=23\n")
			return nil
		case strings.Contains(text, "-list-sites"):
			if !strings.Contains(text, "-T -- server /usr/local/bin/flea-server -list-sites") {
				t.Fatal(text)
			}
			fmt.Fprint(c.out, "["+testSiteJSON+"]")
			return nil
		}
		t.Fatal("unexpected command", text)
		return nil
	}
	var out bytes.Buffer
	if err := serverSites(context.Background(), "server", strings.NewReader(""), &out, io.Discard, runner); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Sites on server", "demo", "example.test", "www.example.test", "release 20260101-120000-abcdef"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("missing %q in output:\n%s", want, out.String())
		}
	}
}

func TestServerSitesEmptyAndOlderProtocol(t *testing.T) {
	empty := func(_ context.Context, c command) error {
		text := strings.Join(c.args, " ")
		switch {
		case strings.Contains(text, "-O exit"):
			return nil
		case strings.Contains(text, "FLEA_UID"):
			fmt.Fprint(c.out, "FLEA_UID=0\nFLEA_DEPLOY=23\n")
			return nil
		case strings.Contains(text, "-list-sites"):
			fmt.Fprint(c.out, "[]")
			return nil
		}
		t.Fatal("unexpected command", text)
		return nil
	}
	var out bytes.Buffer
	if err := serverSites(context.Background(), "server", strings.NewReader(""), &out, io.Discard, empty); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "No sites are registered") {
		t.Fatal(out.String())
	}

	ran := false
	old := func(_ context.Context, c command) error {
		text := strings.Join(c.args, " ")
		switch {
		case strings.Contains(text, "-O exit"):
			return nil
		case strings.Contains(text, "FLEA_UID"):
			fmt.Fprint(c.out, "FLEA_UID=0\nFLEA_DEPLOY=22\n")
			return nil
		}
		ran = true
		return nil
	}
	err := serverSites(context.Background(), "server", strings.NewReader(""), io.Discard, io.Discard, old)
	if err == nil || !strings.Contains(err.Error(), "older or incomplete") {
		t.Fatalf("older protocol not reported: %v", err)
	}
	if ran {
		t.Fatal("ran a server command despite the older protocol")
	}
}

func TestServerSitesMissingInstall(t *testing.T) {
	runner := func(_ context.Context, c command) error {
		text := strings.Join(c.args, " ")
		if strings.Contains(text, "-O exit") {
			return nil
		}
		fmt.Fprint(c.out, "FLEA_UID=0\nFLEA_MISSING=1\n")
		return nil
	}
	err := serverSites(context.Background(), "server", strings.NewReader(""), io.Discard, io.Discard, runner)
	if err == nil || !strings.Contains(err.Error(), "isn't installed") {
		t.Fatalf("missing install not reported: %v", err)
	}
}

func TestServerRemoveWorkflow(t *testing.T) {
	var mutation string
	runner := func(_ context.Context, c command) error {
		text := strings.Join(c.args, " ")
		switch {
		case strings.Contains(text, "-O exit"):
			return nil
		case strings.Contains(text, "FLEA_UID"):
			fmt.Fprint(c.out, "FLEA_UID=0\nFLEA_DEPLOY=23\n")
			return nil
		case strings.Contains(text, "-site-info"):
			if !strings.Contains(text, "-site-info 'demo'") {
				t.Fatal(text)
			}
			fmt.Fprint(c.out, testSiteJSON)
			return nil
		case strings.Contains(text, "-remove-site"):
			mutation = text
			return nil
		}
		t.Fatal("unexpected command", text)
		return nil
	}
	var out bytes.Buffer
	if err := serverRemove(context.Background(), "server", "demo", strings.NewReader("demo\n"), &out, io.Discard, runner); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(mutation, "-remove-site 'demo' -expected-state 'digest'") || strings.Contains(mutation, "sudo") {
		t.Fatalf("root mutation quoting: %q", mutation)
	}
	if !strings.Contains(out.String(), "Retained files: /var/lib/flea/sites/demo") || !strings.Contains(out.String(), "www.example.test") {
		t.Fatalf("review output:\n%s", out.String())
	}
	for _, want := range []string{
		"restarts the service",
		"briefly unavailable",
		"automatic HTTPS is disabled",
		"certificates stay on disk",
		"not restored automatically",
		"Back it\nup or pull it",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("review output missing %q:\n%s", want, out.String())
		}
	}
}

func TestServerRemoveCancellation(t *testing.T) {
	mutated := false
	runner := func(_ context.Context, c command) error {
		text := strings.Join(c.args, " ")
		switch {
		case strings.Contains(text, "-O exit"):
			return nil
		case strings.Contains(text, "FLEA_UID"):
			fmt.Fprint(c.out, "FLEA_UID=0\nFLEA_DEPLOY=23\n")
			return nil
		case strings.Contains(text, "-site-info"):
			fmt.Fprint(c.out, testSiteJSON)
			return nil
		case strings.Contains(text, "-remove-site"):
			mutated = true
			return nil
		}
		t.Fatal("unexpected command", text)
		return nil
	}
	err := serverRemove(context.Background(), "server", "demo", strings.NewReader("not-the-name\n"), io.Discard, io.Discard, runner)
	if err == nil || !strings.Contains(err.Error(), "cancelled; nothing changed") {
		t.Fatalf("cancellation: %v", err)
	}
	if mutated {
		t.Fatal("cancelled removal still mutated the server")
	}
}

func TestServerRemoveNonRootUsesSudo(t *testing.T) {
	const stage = "/tmp/flea-state.ABCDE12345"
	var mutation string
	runner := func(_ context.Context, c command) error {
		text := strings.Join(c.args, " ")
		switch {
		case strings.Contains(text, "-O exit"):
			return nil
		case strings.Contains(text, "FLEA_UID"):
			fmt.Fprint(c.out, "FLEA_UID=1000\nFLEA_DEPLOY=23\n")
			return nil
		case strings.Contains(text, "mktemp"):
			fmt.Fprintln(c.out, stage)
			return nil
		case strings.Contains(text, "rm -f --"):
			return nil
		case strings.Contains(text, "cat --"):
			fmt.Fprint(c.out, testSiteJSON)
			return nil
		case strings.Contains(text, "-remove-site"):
			mutation = text
			return nil
		case strings.Contains(text, "sudo --"):
			return nil
		}
		t.Fatal("unexpected command", text)
		return nil
	}
	if err := serverRemove(context.Background(), "server", "demo", strings.NewReader("demo\n"), io.Discard, io.Discard, runner); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(mutation, "sudo -- /usr/local/bin/flea-server -remove-site 'demo'") || !strings.Contains(mutation, "-tt") {
		t.Fatalf("non-root mutation: %q", mutation)
	}
}

func TestServerRemoveRejectsInvalidInput(t *testing.T) {
	runner := func(context.Context, command) error {
		t.Fatal("connected before validating input")
		return nil
	}
	if err := serverRemove(context.Background(), "server", "bad name", strings.NewReader(""), io.Discard, io.Discard, runner); err == nil || !strings.Contains(err.Error(), "invalid site name") {
		t.Fatalf("invalid site name: %v", err)
	}
	if err := serverSites(context.Background(), "bad target!", strings.NewReader(""), io.Discard, io.Discard, runner); err == nil {
		t.Fatal("invalid target accepted")
	}
}
