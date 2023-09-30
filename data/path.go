package data

import (
	"fmt"
	"os"
	"strings"
)

var (
	PathSeparator string = "."

	ErrEmptyPath error = fmt.Errorf("empty path")
)

// Path is used to uniquely identify a value within [Map]
type Path []string

type ErrPathNotFound struct {
	Err  error
	Path Path
}

func (e ErrPathNotFound) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("path not found: %s: %s", e.Path, e.Err)
	}
	return fmt.Sprintf("path not found: %s", e.Path)
}

// NewPath is a helper function to quickly create a [Path] from a string
//
// It takes a string representing a Path and returns a normalized Path.
// A normalized path is a Path without leading or trailing PathSeparator characters,
// and with empty segments removed.
// You can still use
//
//	Path{"foo", "bar"}
//
// but using
//
//	NewPath("foo.bar")
//
// is usually way more convenient.
func NewPath(path string) Path {
	path = strings.TrimSpace(path)
	path = strings.Trim(path, PathSeparator)
	pathSegments := strings.Split(path, PathSeparator)

	// remove empty segments
	var p Path
	for _, segment := range pathSegments {
		if len(segment) == 0 {
			continue
		}
		p = append(p, segment)
	}

	// if now no more segments are left, return a completely empty path
	if len(p) == 0 {
		return Path{}
	}

	return p
}

// NewPathFromOsPath creates a new [Path] from an OS path.
// This is usually meant to use paths to directories, not files.
//
// If a path to a file (indicated by a file extension) is passed,
// the file extension will become part of the path as last segment.
// So if you pass '/foo/bar.baz', the path will be [foo bar baz]
func NewPathFromOsPath(path string) Path {
	p := strings.Replace(path, string(os.PathSeparator), PathSeparator, -1)
	p = strings.Trim(p, PathSeparator)
	return NewPath(p)
}

// NewPathVar is the same as [NewPath], except it has a variadic parameter for each path segment
func NewPathVar(segments ...string) Path {
	return NewPath(strings.Join(segments, PathSeparator))
}

// Append appends the provided segment to the end of the Path and returns the modified Path.
func (p Path) Append(a string) Path {
	return append(p, a)
}

// Prepend prepends the provided segment to the beginning of the Path and returns the modified Path.
func (p Path) Prepend(a string) Path {
	return append(Path{a}, p...)
}

// First returns the first segment of the Path.
func (p Path) First() string {
	return p[0]
}

// Last returns the last segment of the Path.
func (p Path) Last() string {
	return p[len(p)-1]
}

// StripPrefix removes the specified prefix Path from the current Path.
func (p Path) StripPrefix(prefix Path) Path {
	if len(prefix) > len(p) {
		return p
	}

	for i := range prefix {
		if p[i] != prefix[i] {
			return p
		}
	}

	return p[len(prefix):]
}

// String returns a string representation of a Path.
// For the case that the Path was not constructed using
// the constructors, the string needs to be normalized by
// running it through a constructor.
func (p Path) String() string {
	pathString := strings.Join(p, PathSeparator)
	path := NewPath(pathString)

	return strings.Join(path, PathSeparator)
}
