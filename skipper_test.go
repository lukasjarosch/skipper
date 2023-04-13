package skipper_test

import (
	"testing"

	"github.com/lukasjarosch/skipper"
	"github.com/stretchr/testify/assert"
)

func TestFilePathToPath(t *testing.T) {
	testCases := []struct {
		name         string
		filePath     string
		commonPrefix string
		expected     skipper.Path
	}{
		{
			name:         "file path with no common prefix",
			filePath:     "/path/to/file.yaml",
			commonPrefix: "",
			expected:     skipper.P("path.to.file"),
		},
		{
			name:         "file path with single common prefix",
			filePath:     "/path/to/file.yaml",
			commonPrefix: "/path",
			expected:     skipper.P("to.file"),
		},
		{
			name:         "file path with multi common prefix",
			filePath:     "/path/to/the/file.yaml",
			commonPrefix: "/path/to/the",
			expected:     skipper.P("file"),
		},
		{
			name:         "too long prefix returns empty path",
			filePath:     "/path/to/the/file.yaml",
			commonPrefix: "/path/to/the/file",
			expected:     skipper.P(""),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			actual := skipper.FilePathToPath(tt.filePath, tt.commonPrefix)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
