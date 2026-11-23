package sitefiles

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Custom-app package policy. A package lives at itsy/apps/<name>/ and contains
// app.json plus one executable per <goos>-<goarch>. The same rules gate the site
// loader, the release packer/extractor and the deployment preflight so they
// cannot diverge.
const (
	MaxManifestBytes = 64 << 10
	MaxAppExecutable = 100 << 20
	MaxAppFiles      = 256
)

var (
	// AppName is a confined package directory name.
	AppName = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	// AppPlatform is a declared <goos>-<goarch> key.
	AppPlatform = regexp.MustCompile(`^[a-z][a-z0-9]*-[a-z0-9]+$`)
	// AppFileName is a single confined executable basename.
	AppFileName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)
)

// AppManifest is the strict custom-app manifest. Exactly these four keys are
// accepted; nothing else is carried from app.json into a release.
type AppManifest struct {
	Name       string            `json:"name"`
	Version    string            `json:"version"`
	Protocol   int               `json:"protocol"`
	Executable map[string]string `json:"executable"`
}

// DecodeAppManifest accepts exactly one JSON object with exactly the four
// documented fields, rejecting duplicate keys, trailing bytes, unknown keys and
// any case-variant spelling. encoding/json alone would accept all of these.
func DecodeAppManifest(raw []byte) (AppManifest, error) {
	var manifest AppManifest
	trimmed := bytes.TrimSpace(raw)
	if !utf8.Valid(raw) || len(trimmed) == 0 || trimmed[0] != '{' {
		return manifest, fmt.Errorf("must be one JSON object")
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := checkAppJSON(decoder, 0); err != nil {
		return manifest, fmt.Errorf("must be one JSON object")
	}
	if _, err := decoder.Token(); err != io.EOF {
		return manifest, fmt.Errorf("must contain exactly one object")
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return manifest, fmt.Errorf("must be one JSON object")
	}
	if len(fields) != 4 {
		return manifest, fmt.Errorf("must contain exactly name, version, protocol and executable")
	}
	for _, key := range []string{"name", "version", "protocol", "executable"} {
		if _, ok := fields[key]; !ok {
			return manifest, fmt.Errorf("must contain exactly name, version, protocol and executable")
		}
	}
	if err := json.Unmarshal(fields["name"], &manifest.Name); err != nil {
		return manifest, fmt.Errorf("has an invalid name")
	}
	if err := json.Unmarshal(fields["version"], &manifest.Version); err != nil {
		return manifest, fmt.Errorf("has an invalid version")
	}
	if err := json.Unmarshal(fields["protocol"], &manifest.Protocol); err != nil {
		return manifest, fmt.Errorf("has an invalid protocol")
	}
	if err := json.Unmarshal(fields["executable"], &manifest.Executable); err != nil || len(manifest.Executable) == 0 {
		return manifest, fmt.Errorf("must declare platform executables")
	}
	return manifest, nil
}

// checkAppJSON walks tokens once to reject duplicate keys and deep nesting
// before any typed decode.
func checkAppJSON(d *json.Decoder, depth int) error {
	if depth > 16 {
		return fmt.Errorf("JSON nesting exceeds the limit")
	}
	t, err := d.Token()
	if err != nil {
		return err
	}
	delim, ok := t.(json.Delim)
	if !ok {
		return nil
	}
	if delim != '{' && delim != '[' {
		return fmt.Errorf("invalid JSON")
	}
	seen := map[string]bool{}
	for d.More() {
		if delim == '{' {
			key, err := d.Token()
			if err != nil {
				return err
			}
			name, ok := key.(string)
			if !ok || seen[name] {
				return fmt.Errorf("duplicate JSON field")
			}
			seen[name] = true
		}
		if err := checkAppJSON(d, depth+1); err != nil {
			return err
		}
	}
	_, err = d.Token()
	return err
}

// ValidateAppManifest applies the semantic rules shared by the loader and the
// release tooling: the configured name matches, the version is a bounded
// printable string, the protocol matches, and every declared platform maps to a
// confined executable filename. It returns the set of declared filenames.
func ValidateAppManifest(manifest AppManifest, name string, protocol int) (map[string]bool, error) {
	if manifest.Name != name {
		return nil, fmt.Errorf("app package name %q must match the configured name %q", manifest.Name, name)
	}
	if manifest.Protocol != protocol {
		return nil, fmt.Errorf("app %q declares protocol %d; this ItsyKit supports protocol %d", name, manifest.Protocol, protocol)
	}
	if manifest.Version == "" || len(manifest.Version) > 64 || strings.IndexFunc(manifest.Version, func(r rune) bool { return r < 0x20 || r > 0x7e }) >= 0 {
		return nil, fmt.Errorf("app %q needs a printable version of at most 64 bytes", name)
	}
	return AppExecutables(name, manifest)
}

// AppExecutables validates a manifest's platform map. It rejects unknown
// platform keys, the reserved app.json name, the runtime data name, dotfiles and
// any filename that is not a single confined basename. Every declared entry must
// be well formed regardless of the platform being deployed.
func AppExecutables(name string, manifest AppManifest) (map[string]bool, error) {
	if len(manifest.Executable) == 0 {
		return nil, fmt.Errorf("app %q does not declare platform executables", name)
	}
	files := map[string]bool{}
	for platform, executable := range manifest.Executable {
		if !AppPlatform.MatchString(platform) || executable == "app.json" || executable == "data" || strings.HasPrefix(executable, ".") || !AppFileName.MatchString(executable) {
			return nil, fmt.Errorf("app %q declares an invalid executable for %s", name, platform)
		}
		files[executable] = true
	}
	return files, nil
}

// AppPackageFile reports whether name is a static candidate package file: the
// manifest or a single executable filename directly inside itsy/apps/<name>.
// Runtime storage, nested files and dotfiles are never candidates.
func AppPackageFile(name string) bool {
	name = CanonicalName(name)
	rest, ok := strings.CutPrefix(name, "itsy/apps/")
	if !ok {
		return false
	}
	parts := strings.Split(rest, "/")
	if len(parts) != 2 || !AppName.MatchString(parts[0]) {
		return false
	}
	file := parts[1]
	if file == "data" || strings.HasPrefix(file, ".") {
		return false
	}
	return file == "app.json" || AppFileName.MatchString(file)
}

// AppPackageDir reports whether name is a package directory the packer must
// descend into to reach a package file. It never admits the runtime data
// directory.
func AppPackageDir(name string) bool {
	name = CanonicalName(name)
	if name == "itsy/apps" {
		return true
	}
	rest, ok := strings.CutPrefix(name, "itsy/apps/")
	return ok && !strings.Contains(rest, "/") && AppName.MatchString(rest)
}

// AppExecutablePath reports whether a packaged file must be extracted
// executable. The manifest is never executable.
func AppExecutablePath(name string) bool {
	return AppPackageFile(name) && path.Base(name) != "app.json"
}
