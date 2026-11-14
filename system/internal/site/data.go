package site

import (
	"fmt"
	"io/fs"
	"itsy/internal/sitefiles"
	"os"
	"strings"
)

func loadDataDirectory(root *os.Root, directory, kind string, values map[string]map[string]any) error {
	directory = sitefiles.Resolve(root, directory)
	entries, err := fs.ReadDir(root.FS(), directory)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".json")
		if !namePattern.MatchString(name) {
			return fmt.Errorf("invalid %s name %q", kind, name)
		}
		if _, exists := values[name]; exists {
			return fmt.Errorf("duplicate legacy snip/data name %q", name)
		}
		var data map[string]any
		if err := readJSON(root, directory+"/"+entry.Name(), &data); err != nil {
			return err
		}
		if data == nil {
			return fmt.Errorf("%s %s must contain a JSON object", kind, name)
		}
		values[name] = data
	}
	return nil
}
