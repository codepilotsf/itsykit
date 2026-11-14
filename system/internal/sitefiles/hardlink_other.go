//go:build !unix

package sitefiles

import "io/fs"

// Hardlinked is a no-op on platforms without POSIX link counts.
func Hardlinked(fs.FileInfo) bool { return false }
