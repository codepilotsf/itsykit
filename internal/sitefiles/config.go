package sitefiles

import (
	"encoding/json"
	"fmt"
)

// RuntimeConfig removes local deployment destinations from a release or backup.
// Everything else remains part of the site's configuration.
func RuntimeConfig(data []byte) ([]byte, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil || fields == nil {
		return nil, fmt.Errorf(".flea.config.json must be a valid object")
	}
	delete(fields, "deploy")
	data, err := json.MarshalIndent(fields, "", "  ")
	return append(data, '\n'), err
}
