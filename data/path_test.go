package data_test

import (
	"os"
	"strings"
	"testing"

	. "github.com/lukasjarosch/skipper/data"
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
		{"path/to/file.txt", Path{"path", "to", "file", "txt"}},
		{"/absolute/path/file.txt", Path{"absolute", "path", "file", "txt"}},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			actual := NewPathFromOsPath(test.input)
			if len(actual) != len(test.expected) {
				t.Errorf("Expected %v, but got %v", test.expected, actual)
			}

			if !equalPaths(actual, test.expected) {
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
	if !equalPaths(appended, expected) {
		t.Errorf("Expected %v, but got %v", expected, appended)
	}
}

func TestPathPrepend(t *testing.T) {
	p := Path{"bar", "baz"}

	prepended := p.Prepend("foo")
	expected := Path{"foo", "bar", "baz"}
	if !equalPaths(prepended, expected) {
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
	if !equalPaths(stripped, expected) {
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

// equalPaths checks if two Paths are equal by comparing their elements.
func equalPaths(p1, p2 Path) bool {
	if len(p1) != len(p2) {
		return false
	}
	for i := range p1 {
		if p1[i] != p2[i] {
			return false
		}
	}
	return true
}
