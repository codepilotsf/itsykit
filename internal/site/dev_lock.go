package site

import (
	"crypto/sha256"
	"encoding/hex"
	"flea/internal/state"
	"fmt"
	"os"
	"path/filepath"
)

// LockDevelopment prevents a pull from racing a running dev watcher or another
// pull. The OS releases it on crash; the empty lock file carries no site data.
func LockDevelopment(root string) (func(), error) {
	absolute, e := filepath.Abs(root)
	if e != nil {
		return nil, e
	}
	absolute, e = filepath.EvalSymlinks(absolute)
	if e != nil {
		return nil, e
	}
	h := sha256.Sum256([]byte(absolute))
	unlock, e := state.Lock(filepath.Join(os.TempDir(), "flea-dev-"+hex.EncodeToString(h[:])+".lock"))
	if e != nil {
		return nil, fmt.Errorf("stop flea dev or the other pull for this site, then retry: %w", e)
	}
	return unlock, nil
}
