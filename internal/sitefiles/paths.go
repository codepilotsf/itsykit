// Package sitefiles defines the shared root-level site boundary.
package sitefiles

import (
	"io/fs"
	"strings"
)

// Included excludes developer metadata while retaining private runtime files.
func Included(name string) bool {
	if !fs.ValidPath(name) || strings.ContainsAny(name, "\\\x00") {
		return false
	}
	if name == "." {
		return true
	}
	if name == ".flea.config.json" {
		return true
	}
	for _, part := range strings.Split(name, "/") {
		if strings.HasPrefix(part, ".") {
			return false
		}
		switch strings.ToLower(part) {
		case "readme.md", "agents.md", "architecture.md", "spec.md", "next-steps.md", "checklist.md", "license", "license.md", "go.mod", "go.sum", "node_modules":
			return false
		}
	}
	if name == "flea" || name == "flea/snips.json" {
		return true
	}
	for _, directory := range []string{"site", "mailers", "flea/data", "flea/snips", "flea/icons", "flea/media"} {
		if name == directory || strings.HasPrefix(name, directory+"/") {
			return true
		}
	}
	return false
}
