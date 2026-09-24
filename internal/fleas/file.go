package fleas

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ReadFile pins the administrator-selected content directory and rejects linked files.
func ReadFile(name string) (*Store, error) {
	absolute, err := filepath.Abs(name)
	if err != nil {
		return nil, err
	}
	root, err := os.OpenRoot(filepath.Dir(absolute))
	if err != nil {
		return nil, err
	}
	defer root.Close()
	current := root
	base := filepath.Base(absolute)
	info, err := current.Lstat(base)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("fleas must be a regular file")
	}
	f, err := current.Open(base)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	actual, err := f.Stat()
	if err != nil || !os.SameFile(info, actual) {
		return nil, fmt.Errorf("flea file changed")
	}
	data, err := io.ReadAll(io.LimitReader(f, 5<<20+1))
	if err != nil {
		return nil, err
	}
	return Decode(data)
}
