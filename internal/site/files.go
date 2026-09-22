package site

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"
)

// Open each path component through pinned directory handles. Comparing lstat
// with the opened handle rejects symlink swaps during a reload as well as
// symlinks already present at validation time.
func readFile(root *os.Root, name string) ([]byte, error) {
	if !fs.ValidPath(name) || strings.ContainsAny(name, "\\\x00") {
		return nil, fmt.Errorf("invalid site path: %q", name)
	}
	current, err := pinnedDir(root, path.Dir(name))
	if err != nil {
		return nil, err
	}
	defer current.Close()
	base := path.Base(name)
	info, err := current.Lstat(base)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("not a regular file (symlinks unsupported): %s", name)
	}
	f, err := current.Open(base)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	actual, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !os.SameFile(info, actual) {
		return nil, fmt.Errorf("file changed while loading: %s", name)
	}
	return io.ReadAll(f)
}

func readJSON(root *os.Root, path string, target any) error {
	data, err := readFile(root, path)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("%s: invalid JSON: %w", path, err)
	}
	return nil
}

func pinnedDir(root *os.Root, directory string) (*os.Root, error) {
	current, err := root.OpenRoot(".")
	if err != nil {
		return nil, err
	}
	if directory == "." {
		return current, nil
	}
	for _, part := range strings.Split(directory, "/") {
		info, err := current.Lstat(part)
		if err != nil {
			current.Close()
			return nil, err
		}
		if !info.IsDir() {
			current.Close()
			return nil, fmt.Errorf("not a real directory: %s", directory)
		}
		next, err := current.OpenRoot(part)
		if err != nil {
			current.Close()
			return nil, err
		}
		actual, err := next.Stat(".")
		current.Close()
		if err != nil || !os.SameFile(info, actual) {
			next.Close()
			return nil, fmt.Errorf("directory changed: %s", directory)
		}
		current = next
	}
	return current, nil
}

func writeSiteFile(root *os.Root, name string, data []byte) error {
	if !fs.ValidPath(name) || strings.ContainsAny(name, "\\\x00") {
		return fmt.Errorf("invalid site path: %q", name)
	}
	parent, err := pinnedDir(root, path.Dir(name))
	if err != nil {
		return err
	}
	defer parent.Close()
	if info, err := parent.Lstat(path.Base(name)); err == nil && !info.Mode().IsRegular() {
		return fmt.Errorf("not a regular file: %s", name)
	} else if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := replaceFile(parent, path.Base(name), data); err != nil {
		return err
	}
	directory, err := parent.Open(".")
	if err != nil {
		return err
	}
	defer directory.Close()
	return directory.Sync()
}

// replaceFile atomically replaces a basename within an already pinned directory.
// Callers that need crash-durable directory entries also sync the directory.
func replaceFile(root *os.Root, name string, data []byte) error {
	var random [16]byte
	if _, err := rand.Read(random[:]); err != nil {
		return err
	}
	temp := ".flea-" + hex.EncodeToString(random[:])
	f, err := root.OpenFile(temp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
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
	return root.Rename(temp, name)
}
