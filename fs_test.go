package skipper_test

// These tests were written by ChatGPT. They are far from perfect, but a very nice starting point.

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lukasjarosch/skipper"
	"github.com/spf13/afero"
)

func TestDiscoverYamlFiles(t *testing.T) {
	fileSystem := afero.NewMemMapFs()
	rootPath := "/test/"

	// Create test files
	for _, path := range []string{
		"skipper/test.yaml",
		"skipper/test.yml",
		"other/test.yaml",
	} {
		file, err := skipper.NewYamlFile(rootPath + path)
		if err != nil {
			t.Fatalf("error creating test YAML file: %v", err)
		}
		file.Bytes = []byte("foo: bar")

		err = afero.WriteFile(fileSystem, rootPath+path, file.Bytes, 0644)
		if err != nil {
			t.Fatalf("error writing test YAML file to filesystem: %v", err)
		}
	}

	// Test discovering YAML files
	files, err := skipper.DiscoverYamlFiles(fileSystem, rootPath+"skipper/")
	if err != nil {
		t.Fatalf("error discovering YAML files: %v", err)
	}

	// Check that only skipper/*.yaml and skipper/*.yml files were discovered
	if len(files) != 2 {
		t.Fatalf("discovered the wrong number of YAML files: expected=2, got=%d", len(files))
	}
	for _, file := range files {
		if !strings.HasPrefix(file.Path, rootPath+"skipper/") {
			t.Errorf("discovered a file from the wrong directory: %q", file.Path)
		}
		if ext := filepath.Ext(file.Path); ext != ".yaml" && ext != ".yml" {
			t.Errorf("discovered a file with the wrong extension: %q", file.Path)
		}
	}
}

func TestCopyFile(t *testing.T) {
	testFs := afero.NewMemMapFs()

	sourcePath := "/path/to/source.txt"
	targetPath := "/path/to/target.txt"
	sourceData := []byte("hello world")

	err := skipper.WriteFile(testFs, sourcePath, sourceData, 0755)
	if err != nil {
		t.Fatalf("WriteFile failed with error: %v", err)
	}

	err = skipper.CopyFile(testFs, sourcePath, targetPath)
	if err != nil {
		t.Fatalf("CopyFile failed with error: %v", err)
	}

	targetData, err := afero.ReadFile(testFs, targetPath)
	if err != nil {
		t.Fatalf("ReadFile failed with error: %v", err)
	}

	if !bytes.Equal(sourceData, targetData) {
		t.Errorf("The source and target data should be equal, but they are not")
	}
}

func TestWriteFile(t *testing.T) {
	testFs := afero.NewMemMapFs()

	targetPath := "/path/to/file.txt"
	data := []byte("hello world")

	err := skipper.WriteFile(testFs, targetPath, data, 0755)
	if err != nil {
		t.Fatalf("WriteFile failed with error: %v", err)
	}

	exists, err := afero.Exists(testFs, targetPath)
	if err != nil {
		t.Fatalf("Exists failed with error: %v", err)
	}
	if !exists {
		t.Errorf("The target file should exist, but it does not")
	}

	fileStat, err := testFs.Stat(targetPath)
	if err != nil {
		t.Fatalf("Could not stat file %s", targetPath)
	}
	fileMode := fileStat.Mode()
	if err != nil {
		t.Fatalf("GetFileMode failed with error: %v", err)
	}
	if fileMode != 0755 {
		t.Errorf("The target file mode should be %d, but it is %d", 0755, fileMode)
	}

	fileData, err := afero.ReadFile(testFs, targetPath)
	if err != nil {
		t.Fatalf("ReadFile failed with error: %v", err)
	}
	if !bytes.Equal(data, fileData) {
		t.Errorf("The file data should be equal to the data passed to WriteFile, but it is not")
	}
}
