package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"itsy/internal/buildinfo"
)

func TestRetiredCommandStopsBeforeDispatch(t *testing.T) {
	for _, args := range [][]string{nil, {"dev"}, {"reset"}, {"deploy"}, {"version"}, {"--help"}, {"update", "--help"}, {"update", "unexpected"}} {
		var out, errOut bytes.Buffer
		if code := RunExecutable(context.Background(), "/tmp/flea", args, &out, &errOut); code != 2 || out.Len() != 0 || !strings.Contains(errOut.String(), "From now on, please use itsy.") {
			t.Fatalf("%v: code=%d out=%q err=%q", args, code, out.String(), errOut.String())
		}
	}
}

func TestCanonicalAndReleasedUpdaterVersionProbes(t *testing.T) {
	for _, tc := range []struct{ executable, want string }{
		{"/tmp/itsy", "ItsyKit "},
		{"/tmp/.itsy-update-123", "ItsyWeb "},
		{"/tmp/.itsykit-update-123", "ItsyKit "},
		{"/tmp/.flea-update-123", "FleaCMS "},
	} {
		var out, errOut bytes.Buffer
		if code := RunExecutable(context.Background(), tc.executable, []string{"version"}, &out, &errOut); code != 0 || out.String() != tc.want+buildinfo.Version+"\n" || errOut.Len() != 0 {
			t.Fatalf("%s: code=%d out=%q err=%q", tc.executable, code, out.String(), errOut.String())
		}
	}
}

func TestRetiredUpdateDispatchesWithProminentNotice(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Exercise dispatch without making a release request.
	var out, errOut bytes.Buffer
	code := RunExecutable(ctx, "/tmp/flea", []string{"update"}, &out, &errOut)
	if code != 1 || !strings.Contains(errOut.String(), "FLEA HAS CHANGED TO ITSYKIT") || !strings.Contains(errOut.String(), "itsy update") || !strings.Contains(errOut.String(), "itsy:") {
		t.Fatalf("code=%d out=%q err=%q", code, out.String(), errOut.String())
	}
}
