// Package templates contains the starter site bundled with the executable.
package templates

import "embed"

// Files mirrors a generated site. The all: prefix includes private underscore
// files, which Go's default directory embedding would otherwise exclude.
//
//go:embed all:html all:_ README.md
var Files embed.FS
