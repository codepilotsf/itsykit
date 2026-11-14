package buildinfo

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestBuildVersions(t *testing.T) {
	for _, tc := range []struct {
		version, tag, mode, want string
		valid                    bool
	}{
		{"", "", "", Version, true},
		{"0.2.0", "", "", "0.2.0", true},
		{"0.2.0-rc.1+build.007", "", "", "0.2.0-rc.1+build.007", true},
		{"1.0.0-01a", "", "", "1.0.0-01a", true},
		{"dev", "", "", "", false},
		{"v0.1.0", "", "", "", false},
		{"01.0.0", "", "", "", false},
		{"1.0.0-01", "", "", "", false},
		{"1.0.0-rc..1", "", "", "", false},
		{"1.0.0+", "", "", "", false},
		{"", "v" + Version, "release", Version, true},
		{"", "v99.0.0", "release", "", false},
		{"", "", "release", "", false},
	} {
		cmd := exec.Command("bash", "-c", `source ../../scripts/version.sh; itsy_version "$MODE"`)
		cmd.Env = append(os.Environ(), "VERSION="+tc.version, "CI_COMMIT_TAG="+tc.tag, "MODE="+tc.mode)
		output, err := cmd.CombinedOutput()
		if (err == nil) != tc.valid || (tc.valid && strings.TrimSpace(string(output)) != tc.want) {
			t.Errorf("version=%q tag=%q: output=%q err=%v", tc.version, tc.tag, output, err)
		}
	}
}
