// Package templates contains the starter site bundled with the executable.
package templates

import "embed"

// Files mirrors a generated site. The all: prefix includes private underscore
// files, which Go's default directory embedding would otherwise exclude.
//
//go:embed *.html all:css all:images all:_ README.md .gitignore
var Files embed.FS
