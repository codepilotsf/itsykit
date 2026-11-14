//go:build unix

package appmanager

import (
	"io/fs"
	"os"
	"syscall"
)

// ownedByService reports whether an existing storage directory belongs to the
// account running the server, so another user cannot inject or read app state.
// A root process may service any directory; it is not the untrusted case.
func ownedByService(info fs.FileInfo) bool {
	if os.Geteuid() == 0 {
		return true
	}
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return true
	}
	return int(st.Uid) == os.Geteuid()
}
