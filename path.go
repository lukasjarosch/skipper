package skipper

import (
	"errors"
	"strings"
)

var (
	ErrEmptyPath = errors.New("empty path")
)

// PathSeparator is the separator used for string representations of [Path].
const PathSeparator = "."

// Path is the central way of indexing data in Skipper.
// Paths are used to traverse [Data] and to create namespaces in [Class]es
type Path []string

// P is a helper function to quickly create a [Path] from a string
//
// It takes a string representing a Path and returns a normalized Path.
// A normalized path is a Path without leading or trailing PathSeparator characters,
// and with empty segments removed.
// You can still use `skipper.Path{"foo", "bar"}`, but using `skipper.P("foo.bar")`
// is usually way more convenient.
func P(path string) Path {
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

// String returns a string representation of a Path.
//
// In case the Path was constructed using `Path{"foo"}` we need to re-normalize the string
// as it is not (yet) guaranteed that the Path looks like what we want it to.
// by calling `P()` first and then joining the result.
func (p Path) String() string {
	pathString := strings.Join(p, PathSeparator)
	path := P(pathString)

	return strings.Join(path, PathSeparator)
}
