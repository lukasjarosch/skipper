package skipper_test

import (
	"testing"

	"github.com/lukasjarosch/skipper"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFile_EmptyPath(t *testing.T) {
	_, err := skipper.NewFile("")
	assert.EqualError(t, err, skipper.ErrFilePathEmpty.Error())
}

func TestFile_LoadFile_EmptyPath(t *testing.T) {
	fs := afero.NewMemMapFs()
	_, err := skipper.LoadFile("", fs)
	assert.EqualError(t, err, skipper.ErrFilePathEmpty.Error())
}

func TestFile_LoadFile_FileNotExisting(t *testing.T) {
	fs := afero.NewMemMapFs()
	_, err := skipper.LoadFile("/something/random", fs)
	assert.ErrorContains(t, err, skipper.ErrPathDoesNotExist.Error())
}

func TestFile_LoadFile_Success(t *testing.T) {
	fs := afero.NewMemMapFs()

	testData := []byte("test data")
	testPath := "/path/to/test/file"
	err := afero.WriteFile(fs, testPath, testData, 0644)
	require.NoError(t, err)

	file, err := skipper.LoadFile(testPath, fs)
	require.NoError(t, err)

	assert.Equal(t, testData, file.Bytes)
}

func TestFile_Exists(t *testing.T) {
	fs := afero.NewMemMapFs()
	err := afero.WriteFile(fs, "test.txt", []byte("Hello, world!"), 0644)
	assert.NoError(t, err)

	file := &skipper.File{Path: "test.txt"}
	assert.True(t, file.Exists(fs))

	file.Path = "notexist.txt"
	assert.False(t, file.Exists(fs))
}

func TestNewYamlFile_EmptyPath(t *testing.T) {
	_, err := skipper.NewYamlFile("")
	assert.EqualError(t, err, skipper.ErrFilePathEmpty.Error())
}

func TestYamlFile_Load(t *testing.T) {
	fs := afero.NewMemMapFs()
	err := afero.WriteFile(fs, "test.yaml", []byte("foo: bar\n"), 0644)
	assert.NoError(t, err)

	yamlFile, err := skipper.LoadYamlFile("test.yaml", fs)
	assert.NoError(t, err)

	expected := map[string]interface{}{"foo": "bar"}
	assert.Equal(t, expected, yamlFile.Data)
}
