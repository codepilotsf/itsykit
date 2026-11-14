package site

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"slices"
)

// AddAdmin changes only developer-owned configuration, never content or secrets.
// RawMessage preserves unknown settings and exact numeric values during encoding.
func AddAdmin(directory, email string) (string, bool, error) {
	address := AdminAccess{Emails: []string{email}}
	if err := address.validate(); err != nil {
		return "", false, fmt.Errorf("Please enter a valid email address, without a display name.")
	}
	email = address.Emails[0]
	root, err := os.OpenRoot(directory)
	if err != nil {
		return "", false, err
	}
	defer root.Close()
	const name = "itsy.config.json"
	original, err := readFile(root, name)
	if err != nil {
		return "", false, fmt.Errorf("Could not read itsy.config.json. Run this command from your site's folder.")
	}
	var config map[string]json.RawMessage
	if err := json.Unmarshal(original, &config); err != nil || config == nil {
		return "", false, fmt.Errorf("Please fix the JSON in itsy.config.json before adding an admin.")
	}
	if _, err := appSettings(config); err != nil {
		return "", false, err
	}
	admin := map[string]json.RawMessage{}
	if raw, ok := config["admin"]; ok {
		if err := json.Unmarshal(raw, &admin); err != nil {
			return "", false, fmt.Errorf("The admin setting must be a JSON object. Nothing was changed.")
		}
		if admin == nil {
			admin = map[string]json.RawMessage{}
		}
	}
	var access AdminAccess
	if raw, ok := admin["emails"]; ok {
		if err := json.Unmarshal(raw, &access.Emails); err != nil {
			return "", false, fmt.Errorf("admin.emails must be an array of email addresses. Nothing was changed.")
		}
	}
	if err := access.validate(); err != nil {
		return "", false, fmt.Errorf("Please fix admin.emails before adding an admin: %w", err)
	}
	if slices.Contains(access.Emails, email) {
		return email, false, nil
	}
	access.Emails = append(access.Emails, email)
	if err := access.validate(); err != nil {
		return "", false, fmt.Errorf("A site can have up to 20 admins. Nothing was changed.")
	}
	admin["emails"], err = json.Marshal(access.Emails)
	if err != nil {
		return "", false, err
	}
	config["admin"], err = json.Marshal(admin)
	if err != nil {
		return "", false, err
	}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", false, err
	}
	// Refuse an observed concurrent edit instead of replacing it with stale data.
	current, err := readFile(root, name)
	if err != nil || !bytes.Equal(current, original) {
		return "", false, fmt.Errorf("Configuration changed while adding the admin. Please try again.")
	}
	if err := writeSiteFile(root, name, append(data, '\n')); err != nil {
		return "", false, err
	}
	return email, true, nil
}
