package setup

import (
	"errors"
	"fmt"
	"syscall"
)

func temporaryStorageError(directory string, err error) error {
	return fmt.Errorf("temporary storage at %q: %w\nSet TMPDIR to a writable directory with enough free space and quota, then retry.", directory, err)
}

// Only local storage errors are annotated here. Remote failures must not suggest
// that changing the caller's TMPDIR can repair the server's storage.
func explainTemporaryStorageError(directory string, err error) error {
	if errors.Is(err, syscall.ENOSPC) || errors.Is(err, syscall.EDQUOT) {
		return temporaryStorageError(directory, err)
	}
	return err
}
