package site

import (
	"itsy/internal/sitefiles"
	"os"
	"sync"

	"itsy/internal/snips"
)

var localContentMu sync.Mutex

// LockContent coordinates dev extraction and browser saves within this process.
// Running multiple dev processes against the same project is unsupported.
// Acquire this before Server.reloadMu to keep publication and disk writes ordered.
func LockContent() func() {
	localContentMu.Lock()
	return localContentMu.Unlock
}
func SaveContent(directory string, content *snips.Store) error {
	data, err := content.Bytes()
	if err != nil {
		return err
	}
	if _, err = snips.Decode(data); err != nil {
		return err
	}
	root, err := os.OpenRoot(directory)
	if err != nil {
		return err
	}
	defer root.Close()
	if err := root.Mkdir(sitefiles.Resolve(root, "itsy"), 0755); err != nil && !os.IsExist(err) {
		return err
	}
	return writeSiteFile(root, "itsy/snips.json", data)
}
