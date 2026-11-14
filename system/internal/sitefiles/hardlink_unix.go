//go:build unix

package sitefiles

import (
	"io/fs"
	"syscall"
)

// Hardlinked reports whether a regular file has more than one directory entry.
// A hardlinked package file could be replaced through its other name after
// validation, so it is rejected at both load and packaging time.
func Hardlinked(info fs.FileInfo) bool {
	if st, ok := info.Sys().(*syscall.Stat_t); ok {
		return st.Nlink > 1
	}
	return false
}
