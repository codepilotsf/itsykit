package site

import (
	"os"
	"sync"

	"flea/internal/fleas"
)

var localContentMu sync.Mutex

// LockContent coordinates dev extraction and browser saves within this process.
// Running multiple dev processes against the same project is unsupported.
// Acquire this before Server.reloadMu to keep publication and disk writes ordered.
func LockContent() func() {
	localContentMu.Lock()
	return localContentMu.Unlock
}
func SaveContent(directory string, content *fleas.Store) error {
	data, err := content.Bytes()
	if err != nil {
		return err
	}
	if _, err = fleas.Decode(data); err != nil {
		return err
	}
	root, err := os.OpenRoot(directory)
	if err != nil {
		return err
	}
	defer root.Close()
	return writeSiteFile(root, "_/fleas.json", data)
}
