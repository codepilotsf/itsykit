package setup

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func findGo() (string, error) {
	if binary, err := exec.LookPath("go"); err == nil {
		return binary, nil
	}
	for _, path := range []string{"/usr/local/go/bin/go", filepath.Join(runtime.GOROOT(), "bin", "go")} {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path, nil
		}
	}
	return "", errors.New("Go 1.26 or newer is required on the Mac to build the server")
}

func buildEnv(arch string) []string {
	var env []string
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if key != "GOOS" && key != "GOARCH" && key != "CGO_ENABLED" && key != "GOFLAGS" {
			env = append(env, entry)
		}
	}
	return append(env, "GOOS=linux", "GOARCH="+arch, "CGO_ENABLED=0", "GOFLAGS=")
}
