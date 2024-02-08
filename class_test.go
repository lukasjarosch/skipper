package skipper_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lukasjarosch/skipper"
	"github.com/lukasjarosch/skipper/codec"
	"github.com/lukasjarosch/skipper/data"
)

func TestNewClass(t *testing.T) {
	// Test case 1: Valid file path and codec
	filePath := "testdata/simple_class.yaml"
	class, err := skipper.NewClass(filePath, codec.NewYamlCodec())
	assert.NoError(t, err, "No error expected when creating class")
	assert.NotNil(t, class, "Class should not be nil")
	assert.Equal(t, filePath, class.FilePath, "File path should match input")

	// Test case 2: Empty file path
	emptyFilePath := ""
	_, err = skipper.NewClass(emptyFilePath, codec.NewYamlCodec())
	assert.Error(t, err)
	assert.EqualError(t, err, skipper.ErrEmptyFilePath.Error())

	// Test case 3: Non-existent file path
	nonExistentFilePath := "nonexistent.txt"
	_, err = skipper.NewClass(nonExistentFilePath, codec.NewYamlCodec())
	assert.Error(t, err)
	assert.ErrorContains(t, err, "no such file or directory")

	// Test case 4: Multiple root keys in file
	_, err = skipper.NewClass("testdata/multiple_root_keys.yaml", codec.NewYamlCodec())
	assert.ErrorContains(t, err, data.ErrMultipleRootKeys.Error())
}
