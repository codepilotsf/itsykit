// Package templates contains the starter site bundled with the executable.
package templates

import "embed"

// Files mirrors a generated site. The all: prefix includes private underscore
// files and project-local agent skills, which Go's default directory embedding
// would otherwise exclude. Agent skills remain outside runtime/deploy content.
//
//go:embed all:site all:mailers all:flea all:.agents .flea.config.json README.md AGENTS.md .gitignore
var Files embed.FS

// MinimalFiles supplies the two content files used by the clean default.
//
//go:embed minimal/*
var MinimalFiles embed.FS
