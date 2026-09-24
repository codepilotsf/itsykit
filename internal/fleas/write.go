package fleas

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

// WriteFile replaces a managed content snapshot, never templates/configuration.
// The caller must hold the appropriate deployment or local extraction lock.
func WriteFile(name string, store *Store) error {
	data, err := store.Bytes()
	if err != nil {
		return err
	}
	if _, err = Decode(data); err != nil {
		return err
	}
	root, err := os.OpenRoot(filepath.Dir(name))
	if err != nil {
		return err
	}
	defer root.Close()
	base := filepath.Base(name)
	if info, e := root.Lstat(base); e == nil && !info.Mode().IsRegular() {
		return fmt.Errorf("content must be a regular file")
	} else if e != nil && !os.IsNotExist(e) {
		return e
	}
	var random [16]byte
	if _, err = rand.Read(random[:]); err != nil {
		return err
	}
	temp := ".content-" + hex.EncodeToString(random[:])
	f, err := root.OpenFile(temp, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer root.Remove(temp)
	_, err = f.Write(data)
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	if err = root.Rename(temp, base); err != nil {
		return err
	}
	dir, err := root.Open(".")
	if err != nil {
		return err
	}
	defer dir.Close()
	return dir.Sync()
}
