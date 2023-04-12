package skipper_test

import (
	"testing"

	"github.com/lukasjarosch/skipper"
	"github.com/stretchr/testify/assert"
)

func TestP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		path     string
		expected skipper.Path
	}{
		{
			name:     "simple path",
			path:     "foo.bar.baz",
			expected: skipper.Path{"foo", "bar", "baz"},
		},
		{
			name:     "empty path",
			path:     "",
			expected: skipper.Path{},
		},
		{
			name:     "single segment path",
			path:     "foo",
			expected: skipper.Path{"foo"},
		},
		{
			name:     "leading separator",
			path:     ".foo.bar",
			expected: skipper.Path{"foo", "bar"},
		},
		{
			name:     "trailing separator",
			path:     "foo.bar.",
			expected: skipper.Path{"foo", "bar"},
		},
		{
			name:     "only separator",
			path:     ".",
			expected: skipper.Path{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := skipper.P(tt.path)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestPath_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		path     skipper.Path
		expected string
	}{
		{
			name:     "simple path",
			path:     skipper.Path{"foo", "bar", "baz"},
			expected: "foo.bar.baz",
		},
		{
			name:     "empty path",
			path:     skipper.Path{},
			expected: "",
		},
		{
			name:     "single segment path",
			path:     skipper.Path{"foo"},
			expected: "foo",
		},
		{
			name:     "leading separator",
			path:     skipper.Path{"", "foo", "bar"},
			expected: "foo.bar",
		},
		{
			name:     "trailing separator",
			path:     skipper.Path{"foo", "bar", ""},
			expected: "foo.bar",
		},
		{
			name:     "only separator",
			path:     skipper.Path{"", ""},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.path.String()
			assert.Equal(t, tt.expected, actual)
		})
	}
}
