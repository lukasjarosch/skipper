package skipper_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lukasjarosch/skipper/v1"
	"github.com/lukasjarosch/skipper/v1/codec"
	"github.com/lukasjarosch/skipper/v1/data"
)

func TestNewClass(t *testing.T) {
	// Test case 1: Valid file path and codec
	filePath := "testdata/classes/person.yaml"
	class, err := skipper.NewClass(filePath, codec.NewYamlCodec(), data.NewPath("person"))
	assert.NoError(t, err, "No error expected when creating class")
	assert.NotNil(t, class, "Class should not be nil")
	assert.Equal(t, filePath, class.FilePath, "File path should match input")

	// Test case 2: Empty file path
	emptyFilePath := ""
	_, err = skipper.NewClass(emptyFilePath, codec.NewYamlCodec(), data.NewPath("something"))
	assert.Error(t, err)
	assert.EqualError(t, err, skipper.ErrEmptyFilePath.Error())

	// Test case 3: Non-existent file path
	nonExistentFilePath := "nonexistent.txt"
	_, err = skipper.NewClass(nonExistentFilePath, codec.NewYamlCodec(), data.NewPath("something"))
	assert.Error(t, err)
	assert.ErrorContains(t, err, "no such file or directory")

	// Test case 4: Multiple root keys in file
	_, err = skipper.NewClass("testdata/multiple_root_keys.yaml", codec.NewYamlCodec(), data.NewPath("something"))
	assert.ErrorContains(t, err, data.ErrMultipleRootKeys.Error())

	// Test case: empty class identifier
	_, err = skipper.NewClass(filePath, codec.NewYamlCodec(), data.Path{})
	assert.ErrorIs(t, err, skipper.ErrEmptyClassIdentifier)

	// Test case: class identifier's last segment must be class name
	_, err = skipper.NewClass(filePath, codec.NewYamlCodec(), data.NewPath("not.valid"))
	assert.ErrorIs(t, err, skipper.ErrInvalidClassIdentifier)
}

func TestClassLoader(t *testing.T) {
	rootPath := "testdata/classes"

	files, err := skipper.DiscoverFiles(rootPath, codec.YamlPathSelector)
	assert.NoError(t, err)
	assert.NotNil(t, files)

	classes, err := skipper.ClassLoader(rootPath, files, codec.NewYamlCodec())
	assert.NoError(t, err)
	assert.NotNil(t, classes)
}
