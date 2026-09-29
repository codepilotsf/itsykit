package site

import "os"

// DevSettings is read by the local CLI only. The host must never select a
// delivery provider from this configuration. Restart dev after changes.
type DevSettings struct {
	EmailPreview bool `json:"email_preview"`
}

// ReadDevSettings supports local setup commands without compiling site templates.
func ReadDevSettings(directory string) (DevSettings, error) {
	root, err := os.OpenRoot(directory)
	if err != nil {
		return DevSettings{}, err
	}
	defer root.Close()
	var config struct {
		Dev DevSettings `json:"dev"`
	}
	err = readJSON(root, ".flea.config.json", &config)
	return config.Dev, err
}
