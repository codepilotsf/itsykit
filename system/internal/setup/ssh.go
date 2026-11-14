package setup

import (
	"path/filepath"
	"strings"
)

// Keep setup, deployment and SCP on the same host-key and liveness policy.
// Each invocation owns its control socket; credentials remain native SSH's job.
func connectionOptions(directory string) []string {
	return []string{
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", "ConnectTimeout=10",
		"-o", "ConnectionAttempts=1",
		"-o", "ServerAliveInterval=15",
		"-o", "ServerAliveCountMax=3",
		"-o", "ControlMaster=auto",
		"-o", "ControlPersist=60",
		"-o", "ControlPath=" + filepath.Join(directory, "ssh"),
	}
}

func scpHost(target string) string {
	at := strings.LastIndex(target, "@")
	if strings.Contains(target[at+1:], ":") {
		return target[:at+1] + "[" + target[at+1:] + "]"
	}
	return target
}
