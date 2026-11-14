package sitefiles

import (
	"os"
	"path/filepath"
	"strings"
)

const ConfigName = ".itsy.config.json"
const DataDir = "itsy"
const LegacyConfigName = ".flea.config.json"
const LegacyDataDir = "flea"

// CanonicalName is for policy comparisons only. Files keep their actual names.
func CanonicalName(name string) string {
	if name == LegacyConfigName {
		return ConfigName
	}
	if name == LegacyDataDir || strings.HasPrefix(name, LegacyDataDir+"/") {
		return DataDir + strings.TrimPrefix(name, LegacyDataDir)
	}
	return name
}

// resolve prefers the current spelling and falls back only when it is absent.
// Invalid, linked or unreadable current paths must still fail normal validation.
func resolve(name string, stat func(string) (os.FileInfo, error)) string {
	current, legacy := "", ""
	if name == ConfigName {
		current, legacy = ConfigName, LegacyConfigName
	} else if name == DataDir || strings.HasPrefix(name, DataDir+"/") {
		current, legacy = DataDir, LegacyDataDir
	} else {
		return name
	}
	if _, err := stat(current); os.IsNotExist(err) {
		if _, err := stat(legacy); err == nil {
			return legacy + strings.TrimPrefix(name, current)
		}
	}
	return name
}

// Resolve is only for fixed, caller-selected paths under a confined site root.
func Resolve(root *os.Root, name string) string { return resolve(name, root.Lstat) }

func ResolveAt(directory, name string) string {
	return resolve(name, func(name string) (os.FileInfo, error) {
		return os.Lstat(filepath.Join(directory, filepath.FromSlash(name)))
	})
}

// Path selects a local site's actual path. It does not authorize a read or write;
// callers retain their normal confinement, type, ownership and locking checks.
func Path(directory, name string) string {
	return filepath.Join(directory, filepath.FromSlash(ResolveAt(directory, name)))
}
