package site

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// DefaultSiteName is a display name, never a hostname or deployment identifier.
func DefaultSiteName(root string) string {
	words := strings.FieldsFunc(filepath.Base(filepath.Clean(root)), func(r rune) bool { return r == '-' || r == '_' || r == '.' || unicode.IsSpace(r) })
	for i, word := range words {
		letters := []rune(strings.ToLower(word))
		if len(letters) > 0 {
			letters[0] = unicode.ToTitle(letters[0])
		}
		words[i] = string(letters)
	}
	if len(words) == 0 {
		return "Website"
	}
	return strings.Join(words, " ")
}

func appSettings(config map[string]json.RawMessage) (map[string]json.RawMessage, error) {
	var app map[string]json.RawMessage
	if err := json.Unmarshal(config["app"], &app); err != nil || app == nil {
		return nil, fmt.Errorf(".itsy.config.json: app must be a JSON object")
	}
	if raw, exists := app["site_name"]; exists {
		var name string
		if err := json.Unmarshal(raw, &name); err != nil || strings.TrimSpace(name) == "" {
			return nil, fmt.Errorf(".itsy.config.json: app.site_name must be a nonempty string")
		}
	}
	if _, exists := app["current_admin"]; exists {
		return nil, fmt.Errorf("app.current_admin is reserved for the authenticated request")
	}
	return app, nil
}

// SetSiteName preserves other settings and uses confined, durable writes.
func SetSiteName(directory, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("Please enter a site name.")
	}
	root, err := os.OpenRoot(directory)
	if err != nil {
		return err
	}
	defer root.Close()
	const path = ".itsy.config.json"
	original, err := readFile(root, path)
	if err != nil {
		return err
	}
	var config map[string]json.RawMessage
	if err := json.Unmarshal(original, &config); err != nil || config == nil {
		return fmt.Errorf("Please fix the JSON in .itsy.config.json before setting the site name.")
	}
	app, err := appSettings(config)
	if err != nil {
		return err
	}
	app["site_name"], _ = json.Marshal(name)
	config["app"], err = json.Marshal(app)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	current, err := readFile(root, path)
	if err != nil || !bytes.Equal(current, original) {
		return fmt.Errorf("Configuration changed while setting the site name. Please try again.")
	}
	return writeSiteFile(root, path, append(data, '\n'))
}
