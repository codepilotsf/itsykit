//go:build !unix

package appmanager

import "io/fs"

// ownedByService has no portable notion of ownership outside POSIX systems.
func ownedByService(fs.FileInfo) bool { return true }
