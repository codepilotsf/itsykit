package setup

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
)

const testSiteJSON = `{"name":"demo","branch":"main","domains":["example.test","www.example.test"],"release":"20260101-120000-abcdef","redirects":["www.example.test -> example.test"],"https":["example.test","www.example.test"],"state":"digest","managed":true,"retained":"/var/lib/flea/sites/demo"}`

// inventoryRunner answers the SSH probe and the site list, then delegates any
// remaining command to extra. It never performs real work.
func inventoryRunner(t *testing.T, listJSON string, extra func(c command) (bool, error)) runner {
	t.Helper()
	return func(_ context.Context, c command) error {
		text := strings.Join(c.args, " ")
		switch {
		case strings.Contains(text, "-O exit"):
			return nil
		case strings.Contains(text, "FLEA_UID"):
			fmt.Fprint(c.out, "FLEA_UID=0\nFLEA_DEPLOY=30\n")
			return nil
		case strings.Contains(text, "-list-sites"):
			fmt.Fprint(c.out, listJSON)
			return nil
		}
		if extra != nil {
			if handled, err := extra(c); handled {
				return err
			}
		}
		t.Fatal("unexpected command", text)
		return nil
	}
}

func TestServerListWorkflow(t *testing.T) {
	var out bytes.Buffer
	if err := serverList(context.Background(), "server", strings.NewReader(""), &out, io.Discard, inventoryRunner(t, "["+testSiteJSON+"]", nil)); err != nil {
		t.Fatal(err)
	}
	want := "Sites on server:\ndemo  (branch main)\n  https example.test, www.example.test\n  indexing example.test: allow (no restriction added)\n  indexing www.example.test: allow (no restriction added)\n  redirect www.example.test -> example.test\n\n"
	if out.String() != want {
		t.Fatalf("listing:\n got %q\nwant %q", out.String(), want)
	}
	if strings.Contains(out.String(), "release") {
		t.Fatalf("listing still shows the release:\n%s", out.String())
	}
	if strings.Contains(out.String(), "\n  example.test") {
		t.Fatalf("listing repeats the domains as a plain line:\n%s", out.String())
	}
}

func TestServerListHTTPAndMixedDisplay(t *testing.T) {
	payload := `[{"name":"alpha","domains":["a.test"],"https":["a.test"],"managed":true,"state":"s1"},` +
		`{"name":"beta","domains":["b.test","c.test"],"managed":true,"state":"s2"},` +
		`{"name":"gamma","domains":["d.test","e.test"],"https":["d.test"],"redirects":["e.test -> d.test"],"managed":false,"state":"s3"}]`
	var out bytes.Buffer
	if err := serverList(context.Background(), "server", strings.NewReader(""), &out, io.Discard, inventoryRunner(t, payload, nil)); err != nil {
		t.Fatal(err)
	}
	want := "Sites on server:\n" +
		"alpha\n  https a.test\n  indexing a.test: allow (no restriction added)\n\n" +
		"beta\n  http b.test, c.test\n  indexing b.test: allow (no restriction added)\n  indexing c.test: allow (no restriction added)\n\n" +
		"gamma\n  https d.test\n  http e.test\n  not managed by this server\n  indexing d.test: allow (no restriction added)\n  indexing e.test: allow (no restriction added)\n  redirect e.test -> d.test\n\n"
	if out.String() != want {
		t.Fatalf("listing:\n got %q\nwant %q", out.String(), want)
	}
}

func TestServerListEmptyAndOlderProtocol(t *testing.T) {
	var out bytes.Buffer
	if err := serverList(context.Background(), "server", strings.NewReader(""), &out, io.Discard, inventoryRunner(t, "[]", nil)); err != nil {
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
			fmt.Fprint(c.out, "FLEA_UID=0\nFLEA_DEPLOY=28\n")
			return nil
		}
		ran = true
		return nil
	}
	err := serverList(context.Background(), "server", strings.NewReader(""), io.Discard, io.Discard, old)
	if err == nil || !strings.Contains(err.Error(), "older or incomplete") {
		t.Fatalf("older protocol not reported: %v", err)
	}
	if ran {
		t.Fatal("ran a server command despite the older protocol")
	}
}

func TestServerListMissingInstall(t *testing.T) {
	runner := func(_ context.Context, c command) error {
		text := strings.Join(c.args, " ")
		if strings.Contains(text, "-O exit") {
			return nil
		}
		fmt.Fprint(c.out, "FLEA_UID=0\nFLEA_MISSING=1\n")
		return nil
	}
	err := serverList(context.Background(), "server", strings.NewReader(""), io.Discard, io.Discard, runner)
	if err == nil || !strings.Contains(err.Error(), "isn't installed") {
		t.Fatalf("missing install not reported: %v", err)
	}
}

func TestServerListPromptsForTarget(t *testing.T) {
	var out bytes.Buffer
	runner := func(_ context.Context, c command) error {
		text := strings.Join(c.args, " ")
		switch {
		case strings.Contains(text, "-O exit"):
			return nil
		case strings.Contains(text, "FLEA_UID"):
			if !strings.Contains(text, "-- root@fleacms.com ") {
				t.Fatalf("prompted target not used: %q", text)
			}
			fmt.Fprint(c.out, "FLEA_UID=0\nFLEA_DEPLOY=30\n")
			return nil
		case strings.Contains(text, "-list-sites"):
			fmt.Fprint(c.out, "[]")
			return nil
		}
		t.Fatal("unexpected command", text)
		return nil
	}
	if err := serverList(context.Background(), "", strings.NewReader("root@fleacms.com\n"), &out, io.Discard, runner); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Server login (for example root@fleacms.com)") {
		t.Fatalf("missing login prompt:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "No sites are registered on root@fleacms.com") {
		t.Fatalf("prompted target not reported:\n%s", out.String())
	}
}

func TestServerListMissingTargetStopsBeforeConnect(t *testing.T) {
	runner := func(context.Context, command) error {
		t.Fatal("connected before a target was entered")
		return nil
	}
	if err := serverList(context.Background(), "", strings.NewReader(""), io.Discard, io.Discard, runner); err == nil {
		t.Fatal("EOF target accepted")
	}
	if err := serverList(context.Background(), "", strings.NewReader("\n"), io.Discard, io.Discard, runner); err == nil || !strings.Contains(err.Error(), "enter a server login") {
		t.Fatalf("blank target: %v", err)
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
			fmt.Fprint(c.out, "FLEA_UID=0\nFLEA_DEPLOY=30\n")
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

func TestServerRemovePromptsForTargetAndSelectsByName(t *testing.T) {
	var mutation string
	var out bytes.Buffer
	runner := inventoryRunner(t, "["+testSiteJSON+"]", func(c command) (bool, error) {
		text := strings.Join(c.args, " ")
		switch {
		case strings.Contains(text, "-site-info"):
			if !strings.Contains(text, "-site-info 'demo'") {
				t.Fatal(text)
			}
			fmt.Fprint(c.out, testSiteJSON)
			return true, nil
		case strings.Contains(text, "-remove-site"):
			mutation = text
			return true, nil
		}
		return false, nil
	})
	// target prompt, selection by exact name, then exact-name confirmation.
	if err := serverRemove(context.Background(), "", "", strings.NewReader("root@server\ndemo\ndemo\n"), &out, io.Discard, runner); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Server login (for example root@fleacms.com)") {
		t.Fatalf("missing login prompt:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "Managed sites on root@server:") || !strings.Contains(out.String(), "  demo  https example.test, www.example.test") {
		t.Fatalf("selection list:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "Select a site by exact name") {
		t.Fatalf("selection prompt:\n%s", out.String())
	}
	if !strings.Contains(mutation, "-remove-site 'demo'") {
		t.Fatalf("mutation: %q", mutation)
	}
}

// A site whose name is only digits must be selectable by that exact name, with
// no numeric-choice interpretation.
func TestServerRemoveSelectsNumericSiteName(t *testing.T) {
	numeric := `[{"name":"1","domains":["one.test"],"https":["one.test"],"managed":true,"state":"s1"},` +
		`{"name":"2","domains":["two.test"],"managed":true,"state":"s2"}]`
	var mutation string
	var out bytes.Buffer
	runner := inventoryRunner(t, numeric, func(c command) (bool, error) {
		text := strings.Join(c.args, " ")
		switch {
		case strings.Contains(text, "-site-info"):
			if !strings.Contains(text, "-site-info '1'") {
				t.Fatal(text)
			}
			fmt.Fprint(c.out, `{"name":"1","domains":["one.test"],"https":["one.test"],"managed":true,"state":"s1"}`)
			return true, nil
		case strings.Contains(text, "-remove-site"):
			mutation = text
			return true, nil
		}
		return false, nil
	})
	// Exact-name selection of the digits-only name, then confirmation.
	if err := serverRemove(context.Background(), "", "", strings.NewReader("root@server\n1\n1\n"), &out, io.Discard, runner); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "  1  https one.test") || !strings.Contains(mutation, "-remove-site '1'") {
		t.Fatalf("numeric-name selection:\nout=%s\nmutation=%q", out.String(), mutation)
	}
}

func TestServerRemoveInvalidSelectionThenDelete(t *testing.T) {
	var mutation string
	var out bytes.Buffer
	runner := inventoryRunner(t, "["+testSiteJSON+"]", func(c command) (bool, error) {
		text := strings.Join(c.args, " ")
		switch {
		case strings.Contains(text, "-site-info"):
			fmt.Fprint(c.out, testSiteJSON)
			return true, nil
		case strings.Contains(text, "-remove-site"):
			mutation = text
			return true, nil
		}
		return false, nil
	})
	// A bare number is not a selection and 99 does not name a site; both must
	// re-prompt rather than resolve to a listed site.
	in := strings.NewReader("root@server\n1\n99\nnope\n\ndemo\ndemo\n")
	if err := serverRemove(context.Background(), "", "", in, &out, io.Discard, runner); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`No managed site named "1".`, `No managed site named "99".`, `No managed site named "nope".`, "Enter the exact name of a listed site."} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("missing %q:\n%s", want, out.String())
		}
	}
	if !strings.Contains(mutation, "-remove-site 'demo'") {
		t.Fatalf("mutation: %q", mutation)
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
			fmt.Fprint(c.out, "FLEA_UID=0\nFLEA_DEPLOY=30\n")
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

func TestServerRemoveSelectionEOFStopsBeforeMutation(t *testing.T) {
	mutated := false
	runner := inventoryRunner(t, "["+testSiteJSON+"]", func(c command) (bool, error) {
		if strings.Contains(strings.Join(c.args, " "), "-remove-site") {
			mutated = true
			return true, nil
		}
		return false, nil
	})
	if err := serverRemove(context.Background(), "", "", strings.NewReader("root@server\n"), io.Discard, io.Discard, runner); err == nil {
		t.Fatal("EOF selection accepted")
	}
	if mutated {
		t.Fatal("selection EOF still mutated the server")
	}
}

func TestServerRemoveNoManagedSites(t *testing.T) {
	unmanaged := `[{"name":"solo","domains":["solo.test"],"managed":false,"state":"other"}]`
	runner := inventoryRunner(t, unmanaged, func(c command) (bool, error) {
		t.Fatal("unexpected command", strings.Join(c.args, " "))
		return true, nil
	})
	err := serverRemove(context.Background(), "", "", strings.NewReader("root@server\n"), io.Discard, io.Discard, runner)
	if err == nil || !strings.Contains(err.Error(), "no managed sites") {
		t.Fatalf("no managed sites: %v", err)
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
			fmt.Fprint(c.out, "FLEA_UID=1000\nFLEA_DEPLOY=30\n")
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
	if err := serverList(context.Background(), "bad target!", strings.NewReader(""), io.Discard, io.Discard, runner); err == nil {
		t.Fatal("invalid target accepted")
	}
}
