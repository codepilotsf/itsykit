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
	data, err := readFile(root, ".gitignore")
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	const rules = "/_/sessions.json\n/_/.sessions-*"
	text := strings.TrimRight(string(data), "\n")
	if text == rules || strings.HasSuffix(text, "\n"+rules) {
		return nil
	}
	return writeSiteFile(root, ".gitignore", []byte(text+"\n"+rules+"\n"))
}
