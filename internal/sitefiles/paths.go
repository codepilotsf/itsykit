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
	for _, part := range strings.Split(name, "/") {
		if strings.HasPrefix(part, ".") {
			return false
		}
		switch strings.ToLower(part) {
		case "readme.md", "agents.md", "architecture.md", "spec.md", "next-steps.md", "license", "license.md", "go.mod", "go.sum", "node_modules":
			return false
		}
	}
	if name == "_" || strings.HasPrefix(name, "_/") {
		if name == "_/data" || strings.HasPrefix(name, "_/data/") {
			return true
		}
		return name == "_" || name == "_/flea.config" || name == "_/fleas.json" || name == "_/fleas" || strings.HasPrefix(name, "_/fleas/") || name == "_/icons" || strings.HasPrefix(name, "_/icons/")
	}
	return true
}
