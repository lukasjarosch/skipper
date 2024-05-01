package data_test

import (
	"os"
	"strings"
	"testing"

	. "github.com/lukasjarosch/skipper/v1/data"
)

func TestNewPath(t *testing.T) {
	tests := []struct {
		input    string
		expected Path
	}{
		{"", Path{}},
		{".", Path{}},
		{"..", Path{}},
		{"foo", Path{"foo"}},
		{"foo.bar.baz", Path{"foo", "bar", "baz"}},
		{".foo.bar.", Path{"foo", "bar"}},
		{"..foo.bar..", Path{"foo", "bar"}},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			actual := NewPath(test.input)
			if len(actual) != len(test.expected) {
				t.Errorf("Expected %v, but got %v", test.expected, actual)
			}
			for i := range actual {
				if actual[i] != test.expected[i] {
					t.Errorf("Expected %v, but got %v", test.expected, actual)
				}
			}
		})
	}
}

func TestNewPathFromFilePath(t *testing.T) {
	tests := []struct {
		input    string
		expected Path
	}{
		{"", Path{}},
		{string(os.PathSeparator), Path{}},
		{"path/to/somewhere", Path{"path", "to", "somewhere"}},
		{"path/to/file.txt", Path{"path", "to", "file"}},
		{"/absolute/path/file.txt", Path{"absolute", "path", "file"}},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			actual := NewPathFromOsPath(test.input)
			if len(actual) != len(test.expected) {
				t.Errorf("Expected %v, but got %v", test.expected, actual)
			}

			if !actual.Equals(test.expected) {
				t.Errorf("Paths are not equal. Expected '%+v' but got '%+v'", test.expected, actual)
			}
		})
	}
}

func TestNewPathVar(t *testing.T) {
	tests := []struct {
		segments []string
		expected Path
	}{
		{[]string{}, Path{}},
		{[]string{"."}, Path{}},
		{[]string{"foo"}, Path{"foo"}},
		{[]string{"foo", "bar", "baz"}, Path{"foo", "bar", "baz"}},
	}

	for _, test := range tests {
		t.Run(strings.Join(test.segments, "."), func(t *testing.T) {
			actual := NewPathVar(test.segments...)
			if len(actual) != len(test.expected) {
				t.Errorf("Expected %v, but got %v", test.expected, actual)
			}
			for i := range actual {
				if actual[i] != test.expected[i] {
					t.Errorf("Expected %v, but got %v", test.expected, actual)
				}
			}
		})
	}
}

func TestPathAppend(t *testing.T) {
	p := Path{"foo", "bar"}

	appended := p.Append("baz")
	expected := Path{"foo", "bar", "baz"}
	if !appended.Equals(expected) {
		t.Errorf("Expected %v, but got %v", expected, appended)
	}
}

func TestPathPrepend(t *testing.T) {
	p := Path{"bar", "baz"}

	prepended := p.Prepend("foo")
	expected := Path{"foo", "bar", "baz"}
	if !prepended.Equals(expected) {
		t.Errorf("Expected %v, but got %v", expected, prepended)
	}
}

func TestPathFirst(t *testing.T) {
	p := Path{"foo", "bar", "baz"}

	first := p.First()
	expected := "foo"
	if first != expected {
		t.Errorf("Expected %v, but got %v", expected, first)
	}
}

func TestPathLast(t *testing.T) {
	p := Path{"foo", "bar", "baz"}

	last := p.Last()
	expected := "baz"
	if last != expected {
		t.Errorf("Expected %v, but got %v", expected, last)
	}
}

func TestPathStripPrefix(t *testing.T) {
	p := Path{"foo", "bar", "baz", "qux"}

	stripped := p.StripPrefix(Path{"foo", "bar"})
	expected := Path{"baz", "qux"}
	if !stripped.Equals(expected) {
		t.Errorf("Expected %v, but got %v", expected, stripped)
	}
}

func TestPathString(t *testing.T) {
	p := Path{"foo", "bar", "baz"}

	str := p.String()
	expected := "foo.bar.baz"
	if str != expected {
		t.Errorf("Expected %v, but got %v", expected, str)
	}
}

func TestEquals(t *testing.T) {
	tests := []struct {
		path1   Path
		path2   Path
		isEqual bool
	}{
		{NewPath("foo.bar"), NewPath("foo.bar"), true},
		{NewPath("foo.bar"), NewPath("foo.baz"), false},
		{NewPath(""), NewPath(""), true},
	}

	for _, test := range tests {
		isEqual := test.path1.Equals(test.path2)
		if isEqual != test.isEqual {
			t.Errorf("Expected %v to be equal to %v: %v", test.path1, test.path2, test.isEqual)
		}
	}
}

func TestStripPrefix(t *testing.T) {
	tests := []struct {
		path     Path
		prefix   Path
		expected Path
	}{
		{NewPath("foo.bar.qux"), NewPath("foo.bar"), NewPath("qux")},
		{NewPath("foo.bar.qux"), NewPath("baz"), NewPath("foo.bar.qux")},
		{NewPath("foo.bar.qux"), NewPath(""), NewPath("foo.bar.qux")},
	}

	for _, test := range tests {
		stripped := test.path.StripPrefix(test.prefix)
		if stripped.String() != test.expected.String() {
			t.Errorf("Expected StripPrefix(%v, %v) to be %v, but got %v", test.path, test.prefix, test.expected, stripped)
		}
	}
}

func TestHasPrefix(t *testing.T) {
	tests := []struct {
		path      Path
		prefix    Path
		hasPrefix bool
	}{
		{NewPath("foo.bar.qux"), NewPath("foo.bar"), true},
		{NewPath("foo.bar.qux"), NewPath("foo.baz"), false},
		{NewPath("foo.bar.qux"), NewPath(""), true}, // Empty prefix is always present
		{NewPath(""), NewPath("foo.bar"), false},    // Empty path cannot have a prefix
	}

	for _, test := range tests {
		hasPrefix := test.path.HasPrefix(test.prefix)
		if hasPrefix != test.hasPrefix {
			t.Errorf("Expected HasPrefix(%v, %v) to be %v, but got %v", test.path, test.prefix, test.hasPrefix, hasPrefix)
		}
	}
}

func TestSortPaths(t *testing.T) {
	tests := []struct {
		unsortedPaths []Path
		sortedPaths   []Path
	}{
		{
			[]Path{
				NewPath("foo.bar.qux"),
				NewPath("foo.baz"),
				NewPath("baz.bar.foo"),
			},
			[]Path{
				NewPath("baz.bar.foo"),
				NewPath("foo.bar.qux"),
				NewPath("foo.baz"),
			},
		},
		{
			[]Path{},
			[]Path{},
		},
		{
			[]Path{
				NewPath("z"),
				NewPath("a"),
				NewPath("b"),
			},
			[]Path{
				NewPath("a"),
				NewPath("b"),
				NewPath("z"),
			},
		},
	}

	for _, test := range tests {
		SortPaths(test.unsortedPaths)
		if !arePathsEqual(test.unsortedPaths, test.sortedPaths) {
			t.Errorf("Expected SortPaths(%v) to be %v, but got %v", test.unsortedPaths, test.sortedPaths, test.unsortedPaths)
		}
	}
}

func TestFindLongestMatchingPath(t *testing.T) {
	tests := []struct {
		paths         []Path
		searchPath    Path
		expectedMatch Path
	}{
		{
			[]Path{
				NewPath("foo.bar.qux"),
				NewPath("foo.bar.baz"),
				NewPath("baz.bar.foo"),
				NewPath("hello.world"),
			},
			NewPath("foo.bar.baz.qux"),
			NewPath("foo.bar.baz"),
		},
		{
			[]Path{
				NewPath("abc.def"),
				NewPath("abc"),
				NewPath("xyz"),
				NewPath("xyz.abc.def"),
			},
			NewPath("xyz.abc.def.ghi"),
			NewPath("xyz.abc.def"),
		},
		{
			[]Path{
				NewPath("path.to.some.value"),
				NewPath("path.to"),
				NewPath("another.path"),
			},
			NewPath("yet.another.path"),
			NewPath(""),
		},
	}

	for _, test := range tests {
		match := FindLongestMatchingPath(test.paths, test.searchPath)
		if !match.Equals(test.expectedMatch) {
			t.Errorf("Expected FindLongestMatchingPath(%v, %v) to be %v, but got %v", test.paths, test.searchPath, test.expectedMatch, match)
		}
	}
}

func TestFindMostSimilarPath(t *testing.T) {
	tests := []struct {
		paths         []Path
		searchPath    Path
		expectedMatch Path
	}{
		{
			[]Path{
				NewPath("foo.bar.qux"),
				NewPath("foo.baz.bar"),
				NewPath("hello.world"),
			},
			NewPath("foo.bar.baz.qux"),
			NewPath("foo.bar.qux"),
		},
		{
			[]Path{
				NewPath("abc.def"),
				NewPath("abc"),
				NewPath("xyz"),
				NewPath("xyz.abc.def"),
			},
			NewPath("xyz.abc.def.ghi"),
			NewPath("xyz.abc.def"),
		},
		{
			[]Path{
				NewPath("path.to.some.value"),
				NewPath("path.to"),
				NewPath("another.path"),
			},
			NewPath("another.path.test"),
			NewPath("another.path"),
		},
	}

	for _, test := range tests {
		match := FindMostSimilarPath(test.paths, test.searchPath)
		if !match.Equals(test.expectedMatch) {
			t.Errorf("Expected FindMostSimilarPath(%v, %v) to be %v, but got %v", test.paths, test.searchPath, test.expectedMatch, match)
		}
	}
}

func TestFindLongestPrefixMatch(t *testing.T) {
	tests := []struct {
		paths         []Path
		searchPath    Path
		expectedMatch Path
	}{
		{
			[]Path{
				NewPath("foo.bar.qux"),
				NewPath("foo.bar.baz"),
				NewPath("baz.bar.foo"),
				NewPath("hello.world"),
			},
			NewPath("foo.bar.baz.qux"),
			NewPath("foo.bar.baz"),
		},
		{
			[]Path{
				NewPath("abc.def"),
				NewPath("abc"),
				NewPath("xyz"),
				NewPath("xyz.abc.def"),
			},
			NewPath("xyz.abc.def.ghi"),
			NewPath("xyz.abc.def"),
		},
		{
			[]Path{
				NewPath("path.to.some.value"),
				NewPath("path.to"),
				NewPath("another.path"),
			},
			NewPath("yet.another.path"),
			NewPath(""),
		},
	}

	for _, test := range tests {
		match := FindLongestPrefixMatch(test.paths, test.searchPath)
		if !match.Equals(test.expectedMatch) {
			t.Errorf("Expected FindLongestPrefixMatch(%v, %v) to be %v, but got %v", test.paths, test.searchPath, test.expectedMatch, match)
		}
	}
}

// Helper function to compare two slices of paths
func arePathsEqual(paths1, paths2 []Path) bool {
	if len(paths1) != len(paths2) {
		return false
	}

	for i := range paths1 {
		if paths1[i].String() != paths2[i].String() {
			return false
		}
	}

	return true
}
