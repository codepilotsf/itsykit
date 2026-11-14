package site

import (
	"fmt"
	"strings"
)

// SourceError captures compiler-owned context at failure time. Error pages never
// derive filesystem reads from request paths or parse file names out of messages.
type SourceError struct {
	File  string
	Line  int
	Lines []sourceLine
	Err   error
}
type sourceLine struct {
	Number    int
	Text      string
	Highlight bool
}

func (e *SourceError) Error() string { return fmt.Sprintf("%s:%d: %v", e.File, e.Line, e.Err) }
func (e *SourceError) Unwrap() error { return e.Err }

func sourceError(file, source string, pos int, err error) error {
	line := strings.Count(source[:pos], "\n") + 1
	e := &SourceError{File: file, Line: line, Err: err}
	lines := strings.Split(source, "\n")
	for i := max(0, line-6); i < min(len(lines), line+5); i++ {
		text := lines[i]
		if len(text) > 300 {
			text = text[:300] + " …"
		}
		e.Lines = append(e.Lines, sourceLine{i + 1, text, i+1 == line})
	}
	return e
}
