package site

import (
	"os"
	"strings"
)

// sessionIgnoreRules are the lines IgnoreSessionFiles ensures are present in the
// project's .gitignore.
var sessionIgnoreRules = []string{"/flea/runtime/", "/flea/media/*.mp4"}

// IgnoreSessionFiles preserves developer rules and keeps private runtime state
// out of ordinary Git adds. The CLI first checks for already tracked files.
//
// It appends only the rules that are not already present, so repeated runs (for
// example every `flea dev` start) never duplicate existing entries.
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

	text := string(data)
	present := make(map[string]bool)
	for _, line := range strings.Split(text, "\n") {
		present[strings.TrimSpace(line)] = true
	}

	var missing []string
	for _, rule := range sessionIgnoreRules {
		if !present[rule] {
			missing = append(missing, rule)
		}
	}
	if len(missing) == 0 {
		return nil
	}

	// Preserve existing developer rules verbatim; append only what is missing.
	updated := strings.TrimRight(text, "\n")
	if updated != "" {
		updated += "\n"
	}
	updated += strings.Join(missing, "\n") + "\n"
	return writeSiteFile(root, ".gitignore", []byte(updated))
}
