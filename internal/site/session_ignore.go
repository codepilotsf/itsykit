package site

import (
	"os"
	"strings"
)

// IgnoreSessionFiles preserves developer rules and keeps private runtime state
// out of ordinary Git adds. The CLI first checks for already tracked files.
func IgnoreSessionFiles(directory string) error {
	root, err := os.OpenRoot(directory)
	if err != nil {
		return err
	}
	defer root.Close()
	if err := root.MkdirAll("flea/runtime", 0700); err != nil {
		return err
	}
	private, err := pinnedDir(root, "flea/runtime")
	if err != nil {
		return err
	}
	private.Close()
	data, err := readFile(root, ".gitignore")
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	const rules = "/flea/runtime/\n/flea/media/*.mp4"
	text := strings.TrimRight(string(data), "\n")
	if text == rules || strings.HasSuffix(text, "\n"+rules) {
		return nil
	}
	return writeSiteFile(root, ".gitignore", []byte(text+"\n"+rules+"\n"))
}
