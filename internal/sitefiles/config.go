package sitefiles

import (
	"encoding/json"
	"fmt"
	"os"
)

// ExecutableMode is the immutable mode for a confined custom-app executable:
// readable and executable by the service user (release trees are owned by the
// deployment root, not the service account) and writable by no one. It is the
// single policy shared by committed snapshots and release archives.
const ExecutableMode os.FileMode = 0o555

// RuntimeConfig removes local deployment destinations from a release or backup.
// Everything else remains part of the site's configuration.
func RuntimeConfig(data []byte) ([]byte, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil || fields == nil {
		return nil, fmt.Errorf("itsy.config.json must be a valid object")
	}
	delete(fields, "deploy")
	data, err := json.MarshalIndent(fields, "", "  ")
	return append(data, '\n'), err
}
