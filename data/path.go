package data

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

var (
	PathSeparator string = "."
	ErrEmptyPath  error  = fmt.Errorf("empty path")
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

func (p Path) AppendPath(appendPath Path) Path {
	return append(p, appendPath...)
}

// Prepend prepends the provided segment to the beginning of the Path and returns the modified Path.
func (p Path) Prepend(a string) Path {
	return append(Path{a}, p...)
}

// First returns the first segment of the Path.
func (p Path) First() string {
	return p[0]
}

// FirstSegment returns the first segment of the path as Path
func (p Path) FirstSegment() Path {
	return NewPath(p.First())
}

// Last returns the last segment of the Path.
func (p Path) Last() string {
	return p[len(p)-1]
}

// LastSegment returns the last segment of the path as Path
func (p Path) LastSegment() Path {
	return NewPath(p.Last())
}

// Equals returns true if the current path is equal to the provided path.
func (p Path) Equals(other Path) bool {
	if len(p) != len(other) {
		return false
	}

	for i := range p {
		if p[i] != other[i] {
			return false
		}
	}

	return true
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

// StripSuffix removes the specified suffix Path from the current Path.
func (p Path) StripSuffix(suffix Path) Path {
	if len(suffix) > len(p) {
		return p
	}

	for i := 0; i < len(suffix); i++ {
		if p[len(p)-1-i] != suffix[len(suffix)-1-i] {
			return p
		}
	}

	return p[:len(p)-len(suffix)]
}

// HasPrefix checks if the current Path has the specified prefix.
// It returns true if the Path has the prefix; otherwise, it returns false.
func (p Path) HasPrefix(prefix Path) bool {
	if len(prefix) > len(p) {
		return false
	}

	// an empty prefix is always there :shrug:
	if len(prefix) == 0 {
		return true
	}

	for i := range prefix {
		if p[i] != prefix[i] {
			return false
		}
	}

	return true
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

func (p Path) SegmentIndex(segment string) int {
	for i, seg := range p {
		if seg == segment {
			return i
		}
	}
	return -1
}

// SortPaths sorts a slice of Paths alphabetically.
func SortPaths(input []Path) {
	less := func(p1, p2 Path) bool {
		minLen := len(p1)
		if len(p2) < minLen {
			minLen = len(p2)
		}

		for i := 0; i < minLen; i++ {
			if p1[i] < p2[i] {
				return true
			} else if p1[i] > p2[i] {
				return false
			}
		}

		return len(p1) < len(p2)
	}
	sort.Slice(input, func(i, j int) bool {
		return less(input[i], input[j])
	})
}

// FindLongestMatchingPath finds the longest matching path within a list of paths for a given search path.
// Say that the list is: {"foo.bar.qux", "hello.world", "foo"}
// and search is {"foo.bar.baz"}
// Then FindLongestMatchingPath will return {"foo", "bar"} as this is the longest matching path.
func FindLongestMatchingPath(list []Path, search Path) Path {
	var longestMatch Path

	for _, path := range list {
		matchLength := 0
		for i := 0; i < len(path) && i < len(search); i++ {
			if path[i] != search[i] {
				break
			}
			matchLength++
		}

		if matchLength > len(longestMatch) {
			longestMatch = path[:matchLength]
		}
	}

	return longestMatch
}

// FindMostSimilarPath finds the most similar path within a list of paths for a given search path.
// Say that the list is: {"foo.bar.qux", "hello.world", "foo"}
// and search is {"foo.bar.baz"}
// Then FindLongestMatchingPath will return {"foo", "bar", "qux"} as this is the most similar path.
func FindMostSimilarPath(list []Path, search Path) Path {
	var mostSimilarPath Path
	longestMatchLength := 0

	for _, path := range list {
		matchLength := 0
		for i := 0; i < len(path) && i < len(search); i++ {
			if path[i] != search[i] {
				break
			}
			matchLength++
		}

		if matchLength > longestMatchLength {
			longestMatchLength = matchLength
			mostSimilarPath = path
		}
	}

	return mostSimilarPath
}

// FindLongestPrefixMatch finds the path within a list of paths that has the longest matching prefix with the search path.
func FindLongestPrefixMatch(list []Path, search Path) Path {
	var longestMatch Path
	longestMatchLength := 0

	for _, path := range list {
		matchLength := 0
		minLength := len(path)

		if len(search) < minLength {
			minLength = len(search)
		}

		for i := 0; i < minLength; i++ {
			if path[i] != search[i] {
				break
			}
			matchLength++
		}

		if matchLength > longestMatchLength {
			longestMatchLength = matchLength
			longestMatch = path
		}
	}

	return longestMatch
}
