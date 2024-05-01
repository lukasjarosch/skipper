package skipper_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lukasjarosch/skipper/v1"
	"github.com/lukasjarosch/skipper/v1/codec"
)

func TestFileSelector(t *testing.T) {
	rootPath := "testdata/fs"
	selector := codec.YamlPathSelector

	// Test: empty root path
	_, err := skipper.DiscoverFiles("", selector)
	assert.ErrorContains(t, err, "rootPath is empty")

	// Test: nil selector
	_, err = skipper.DiscoverFiles(rootPath, nil)
	assert.ErrorContains(t, err, "selectorRegex is nil")

	// Test: rootPath doesn't exist
	_, err = skipper.DiscoverFiles("thereisnowaythispatexists", selector)
	assert.ErrorIs(t, err, os.ErrNotExist)

	// Test: rootPath is not a directory
	_, err = skipper.DiscoverFiles(filepath.Join(rootPath, "hans.yaml"), selector)
	assert.ErrorContains(t, err, "rootPath is not a directory")

	// Test: valid preconditions, properly select files
	expected := []string{
		filepath.Join(rootPath, "hans.yaml"),
		filepath.Join(rootPath, "nested", "jane.yml"),
		filepath.Join(rootPath, "john.yaml"),
	}
	files, err := skipper.DiscoverFiles(rootPath, selector)
	assert.NoError(t, err)
	assert.ElementsMatch(t, files, expected)
}

func TestStripCommonPathPrefix(t *testing.T) {
	rootPath := "testdata/fs"
	selector := codec.YamlPathSelector

	files, err := skipper.DiscoverFiles(rootPath, selector)
	assert.NoError(t, err)
	assert.NotNil(t, files)

	expected := []string{"hans.yaml", "nested/jane.yml", "john.yaml"}
	strippedFiles := skipper.StripCommonPathPrefix(files)
	assert.ElementsMatch(t, strippedFiles, expected)
}
