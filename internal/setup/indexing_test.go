package setup

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
)

func TestServerIndexingWorkflow(t *testing.T) {
	for _, nonroot := range []bool{false, true} {
		for _, mode := range []string{"noindex", "allow"} {
			t.Run(fmt.Sprint(nonroot, "/", mode), func(t *testing.T) {
				var out bytes.Buffer
				mutation := ""
				reads := 0
				runner := func(_ context.Context, c command) error {
					text := strings.Join(c.args, " ")
					switch {
					case strings.Contains(text, "-O exit"):
						return nil
					case strings.Contains(text, "FLEA_UID"):
						uid := "0"
						if nonroot {
							uid = "1000"
						}
						fmt.Fprintf(c.out, "FLEA_UID=%s\nFLEA_DEPLOY=27\n", uid)
						return nil
					case strings.Contains(text, "mktemp"):
						fmt.Fprintln(c.out, "/tmp/flea-state.ABCDE12345")
						return nil
					case strings.Contains(text, "rm -f --"):
						return nil
					case nonroot && strings.Contains(text, "-list-sites"):
						return nil
					case strings.Contains(text, "-list-sites") || strings.Contains(text, "cat --"):
						reads++
						payload := strings.TrimSuffix(testSiteJSON, "}") + `,"indexing":{"example.test":"` + mode + `","www.example.test":"allow"},"warnings":["example.test: robots.txt contains a site-wide Disallow rule"]}`
						fmt.Fprint(c.out, "["+payload+"]")
						return nil
					case strings.Contains(text, "-set-indexing"):
						mutation = text
						return nil
					}
					t.Fatal("unexpected command", text)
					return nil
				}
				if err := serverIndexing(context.Background(), "server", "example.test", mode, strings.NewReader("example.test\n"), &out, io.Discard, runner); err != nil {
					t.Fatal(err)
				}
				for _, want := range []string{"-set-indexing 'example.test'", "-indexing-mode '" + mode + "'", "-expected-state 'digest'"} {
					if !strings.Contains(mutation, want) {
						t.Fatal(mutation)
					}
				}
				if nonroot && (!strings.Contains(mutation, "sudo -- ") || !strings.Contains(mutation, "-tt")) {
					t.Fatal("missing sudo/PTY", mutation)
				}
				if reads != 2 || !strings.Contains(out.String(), "site-wide Disallow") || !strings.Contains(out.String(), "Type the hostname to confirm") {
					t.Fatal(reads, out.String())
				}
			})
		}
	}
}

func TestServerIndexingRejectsBeforeMutation(t *testing.T) {
	for _, tc := range []struct{ domain, mode, input, list string }{
		{"example.test", "noindex", "wrong\n", "[" + testSiteJSON + "]"},
		{"example.test", "noindex", "", "[" + testSiteJSON + "]"},
		{"unknown.test", "noindex", "unknown.test\n", "[" + testSiteJSON + "]"},
		{"example.test", "noindex", "example.test\n", "[" + strings.Replace(testSiteJSON, `"managed":true`, `"managed":false`, 1) + "]"},
		{"example.test", "index", "", "[]"},
		{"EXAMPLE.TEST", "noindex", "", "[]"},
		{"example.test:80", "noindex", "", "[]"},
	} {
		runner := inventoryRunner(t, tc.list, func(c command) (bool, error) { t.Fatal("mutation attempted", c.args); return true, nil })
		if err := serverIndexing(context.Background(), "server", tc.domain, tc.mode, strings.NewReader(tc.input), io.Discard, io.Discard, runner); err == nil {
			t.Fatal("invalid indexing request accepted", tc)
		}
	}
}
