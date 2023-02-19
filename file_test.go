package skipper_test

// These tests were written by ChatGPT. They are far from perfect, but a very nice starting point.

import (
	"errors"
	"testing"

	"github.com/lukasjarosch/skipper"
	"github.com/spf13/afero"
)

func TestNewFile(t *testing.T) {
	file, err := skipper.NewFile("example.txt")
	if err != nil {
		t.Errorf("NewFile() returned an error: %v", err)
	}

	if file.Path != "example.txt" {
		t.Errorf("NewFile() did not set Path correctly")
	}

	if file.Bytes != nil {
		t.Errorf("NewFile() did not create an empty slice for Bytes")
	}

	_, err = skipper.NewFile("")
	if !errors.Is(err, skipper.ErrFilePathEmpty) {
		t.Errorf("NewFile() did not return the correct error for an empty path")
	}
}

func TestFileExists(t *testing.T) {
	memFs := afero.NewMemMapFs()

	file := &skipper.File{Path: "/test/file.txt"}

	exists := file.Exists(memFs)
	if exists {
		t.Errorf("Exists() returned true for a file that does not exist")
	}

	afero.WriteFile(memFs, "/test/file.txt", []byte("test data"), 0644)

	exists = file.Exists(memFs)
	if !exists {
		t.Errorf("Exists() returned false for a file that exists")
	}
}

func TestFileLoad(t *testing.T) {
	memFs := afero.NewMemMapFs()

	afero.WriteFile(memFs, "/test/file.txt", []byte("test data"), 0644)

	file := &skipper.File{Path: "/test/file.txt"}

	err := file.Load(memFs)
	if err != nil {
		t.Errorf("Load() returned an error: %v", err)
	}

	if file.Mode != 0644 {
		t.Errorf("Load() did not set the correct file mode")
	}

	if string(file.Bytes) != "test data" {
		t.Errorf("Load() did not load the file data correctly")
	}

	file.Path = "nonexistent"
	err = file.Load(memFs)
	if err == nil {
		t.Errorf("Load() did not return an error when loading a nonexistent file")
	}
}

func TestNewYamlFile(t *testing.T) {
	file, err := skipper.NewYamlFile("./test.yaml")
	if err != nil {
		t.Errorf("Error creating YamlFile: %v", err)
	}

	if file.Path != "./test.yaml" {
		t.Errorf("Expected path %s, got %s", "./test.yaml", file.Path)
	}
}

func TestCreateNewYamlFile(t *testing.T) {
	fileSystem := afero.NewMemMapFs()
	path := "./new-test.yaml"
	data := []byte("test data")

	file, err := skipper.CreateNewYamlFile(fileSystem, path, data)
	if err != nil {
		t.Errorf("Error creating new YamlFile: %v", err)
	}

	if file.Path != path {
		t.Errorf("Expected path %s, got %s", path, file.Path)
	}

	_, err = fileSystem.Stat(path)
	if err != nil {
		t.Errorf("Error getting file info: %v", err)
	}
}

func TestYamlFile_Load(t *testing.T) {
	fileSystem := afero.NewMemMapFs()
	path := "./test.yaml"
	data := []byte("name: Test\n")

	err := afero.WriteFile(fileSystem, path, data, 0644)
	if err != nil {
		t.Errorf("Error writing test file: %v", err)
	}

	file, err := skipper.NewYamlFile(path)
	if err != nil {
		t.Errorf("Error creating YamlFile: %v", err)
	}

	err = file.Load(fileSystem)
	if err != nil {
		t.Errorf("Error loading file: %v", err)
	}

	if file.Data == nil {
		t.Errorf("Expected non-nil Data")
	}

	if file.Data["name"] != "Test" {
		t.Errorf("Expected name %s, got %s", "Test", file.Data["name"])
	}
}

func TestYamlFileLoader(t *testing.T) {
	fileSystem := afero.NewMemMapFs()
	rootPath := "/test/"
	basePath := "/test/skipper/"
	data := "foo: bar"

	// Create test YAML file
	file, err := skipper.NewYamlFile(basePath + "test.yaml")
	if err != nil {
		t.Fatalf("error creating test YAML file: %v", err)
	}
	file.Bytes = []byte(data)

	// Add file to filesystem
	err = afero.WriteFile(fileSystem, basePath+"test.yaml", file.Bytes, 0644)
	if err != nil {
		t.Fatalf("error writing test YAML file to filesystem: %v", err)
	}

	var loadedData string
	loaderFunc := func(file *skipper.YamlFile, relativePath string) error {
		loadedData = string(file.Bytes)
		return nil
	}

	// Test loading YAML file
	err = skipper.YamlFileLoader(fileSystem, rootPath, loaderFunc)
	if err != nil {
		t.Fatalf("error loading YAML file: %v", err)
	}

	// Check that data was loaded correctly
	if loadedData != data {
		t.Errorf("loaded data doesn't match expected data: expected=%q, got=%q", data, loadedData)
	}
}

func TestYamlFile_UnmarshalPath(t *testing.T) {
	fileSystem := afero.NewMemMapFs()
	path := "./test.yaml"
	data := []byte("person:\n  name: John\n  age: 30\n")

	err := afero.WriteFile(fileSystem, path, data, 0644)
	if err != nil {
		t.Errorf("Error writing test file: %v", err)
	}

	file, err := skipper.NewYamlFile(path)
	if err != nil {
		t.Errorf("Error creating YamlFile: %v", err)
	}

	err = file.Load(fileSystem)
	if err != nil {
		t.Errorf("Error loading file: %v", err)
	}

	var person struct {
		Name string `yaml:"name"`
		Age  int    `yaml:"age"`
	}

	err = file.UnmarshalPath(&person, "person")
	if err != nil {
		t.Errorf("Error unmarshaling path: %v", err)
	}

	if person.Name != "John" {
		t.Errorf("Expected name %s, got %s", "John", person.Name)
	}

	if person.Age != 30 {
		t.Errorf("Expected age %d, got %d", 30, person.Age)
	}
}

func TestSecretFile_LoadSecretFileData(t *testing.T) {
	fs := afero.NewMemMapFs()
	content := "data: secret data\n"
	err := afero.WriteFile(fs, "/path/to/file", []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %s", err)
	}
	sf := &skipper.SecretFile{
		YamlFile: &skipper.YamlFile{
			File: skipper.File{
				Path: "/path/to/file",
			},
		},
		RelativePath: "/path/to/file",
	}
	err = sf.LoadSecretFileData(fs)
	if err != nil {
		t.Fatalf("LoadSecretFileData failed: %s", err)
	}
	if sf.Data.Data != "secret data" {
		t.Errorf("LoadSecretFileData returned incorrect data: %s", sf.Data.Data)
	}
}

func TestSecretFileList_GetSecretFile(t *testing.T) {
	sfl := skipper.SecretFileList{
		&skipper.SecretFile{RelativePath: "file1"},
		&skipper.SecretFile{RelativePath: "file2"},
		&skipper.SecretFile{RelativePath: "file3"},
	}
	result := sfl.GetSecretFile("file2")
	if result == nil {
		t.Fatalf("GetSecretFile returned nil")
	}
	if result.RelativePath != "file2" {
		t.Errorf("GetSecretFile returned incorrect result: %s", result.RelativePath)
	}
	result = sfl.GetSecretFile("file4")
	if result != nil {
		t.Errorf("GetSecretFile returned incorrect result: %v", result)
	}
}
