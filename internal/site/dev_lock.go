package site

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"itsy/internal/state"
	"os"
	"path/filepath"
)

var ErrDevelopmentBusy = errors.New("another running itsy dev, pull or reset is using this site; stop that command, then try again")

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
	if errors.Is(e, state.ErrBusy) {
		return nil, ErrDevelopmentBusy
	}
	if e != nil {
		return nil, e
	}
	return unlock, nil
}
