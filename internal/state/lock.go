// Package state provides the lock shared by deployment and browser content saves.
package state

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// ErrBusy reports that another process already holds this lock. Callers that
// know what the lock protects can replace it with more specific guidance.
var ErrBusy = errors.New("Content or deployment is busy. Please try again shortly.")

func Lock(name string) (func(), error) {
	root, err := os.OpenRoot(filepath.Dir(name))
	if err != nil {
		return nil, err
	}
	defer root.Close()
	return LockRoot(root, filepath.Base(name))
}

func LockRoot(root *os.Root, name string) (func(), error) {
	if info, err := root.Lstat(name); err == nil && !info.Mode().IsRegular() {
		return nil, fmt.Errorf("invalid content lock")
	} else if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	f, err := root.OpenFile(name, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	info, err := root.Lstat(name)
	actual, statErr := f.Stat()
	if err != nil || statErr != nil || !info.Mode().IsRegular() || !os.SameFile(info, actual) {
		f.Close()
		return nil, fmt.Errorf("content lock changed")
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		return nil, ErrBusy
	}
	return func() { _ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN); _ = f.Close() }, nil
}
