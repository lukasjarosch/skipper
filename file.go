package skipper

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

var (
	yamlFileExtensions = []string{".yml", ".yaml", ""}
)

// File is just an arbitrary description of a path and the data of the File to which Path points to.
// Note that the used filesystem is not relevant, only at the time of loading a File.
type File struct {
	Path  string
	Mode  fs.FileMode
	Bytes []byte
}

func newFile(path string) (*File, error) {
	if path == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}

	return &File{Path: path}, nil
}

// Exists returns true if the File exists in the given filesystem, false otherwise.
func (f *File) Exists(fs afero.Fs) bool {
	exists, err := afero.Exists(fs, f.Path)
	if err != nil {
		return false
	}
	return exists
}

// Load will attempt to read the File from the given filesystem implementation.
// The loaded data is stored in `File.Bytes`
func (f *File) Load(fs afero.Fs) (err error) {
	f.Bytes, err = afero.ReadFile(fs, f.Path)
	if err != nil {
		return fmt.Errorf("failed to Load %s: %w", f.Path, err)
	}
	info, err := fs.Stat(f.Path)
	if err != nil {
		return fmt.Errorf("unable to stat file %s: %w", f.Path, err)
	}
	f.Mode = info.Mode()
	return nil
}

// YamlFileLoaderFunc is a function used to create specific types from a YamlFile and a relative path to that file.
type YamlFileLoaderFunc func(file *YamlFile, relativePath string) error

// YamlFileLoader is used to load Skipper specific yaml files from the inventory.
// It searches the basePath inside the given fileSystem for yaml files and loads them.
// Empty files are skipped.
// A path relative to the given pasePath is constructed.
// The loaded yaml file and the relative path are then passed to the given YamlFileLoaderFunc
// which is responsible for creating specific types from the YamlFile.
func YamlFileLoader(fileSystem afero.Fs, basePath string, loader YamlFileLoaderFunc) error {
	yamlFiles, err := DiscoverYamlFiles(fileSystem, basePath)
	if err != nil {
		return err
	}

	for _, yamlFile := range yamlFiles {
		err = yamlFile.Load(fileSystem)
		if err != nil {
			return err
		}

		// skip empty files
		if len(yamlFile.Data) == 0 {
			continue
		}

		relativePath := strings.ReplaceAll(yamlFile.Path, basePath, "")
		relativePath = strings.TrimLeft(relativePath, "/")

		err = loader(yamlFile, relativePath)
		if err != nil {
			return err
		}
	}

	return nil
}

// YamlFile is what is used for all inventory-relevant files (classes, secrets and targets).
type YamlFile struct {
	File
	Data Data
}

// NewFile returns a newly initialized `YamlFile`.
func NewFile(path string) (*YamlFile, error) {
	f, err := newFile(path)
	if err != nil {
		return nil, err
	}

	return &YamlFile{
		File: *f,
	}, nil
}

// CreateNewFile can be used to manually create a File inside the given filesystem.
// This is useful for dynamically creating classes or targets.
//
// The given path is attempted to be created and a file written.
func CreateNewYamlFile(fs afero.Fs, path string, data []byte) (*YamlFile, error) {
	err := fs.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return nil, err
	}

	err = afero.WriteFile(fs, path, data, 0644)
	if err != nil {
		return nil, err
	}
	return NewFile(path)
}

// Load will first load the underlying raw file-data and then attempt to `yaml.Unmarshal` it into `Data`
// The resulting Data is stored in `YamlFile.Data`.
func (f *YamlFile) Load(fs afero.Fs) error {
	err := f.File.Load(fs)
	if err != nil {
		return err
	}

	var d Data
	if err := yaml.Unmarshal(f.Bytes, &d); err != nil {
		return err
	}
	f.Data = d
	return nil
}

// UnmarshalPath can be used to unmarshall only a sub-map of the Data inside [YamlFile].
// The function errors if the file has not been loaded.
func (f *YamlFile) UnmarshalPath(target interface{}, path ...interface{}) error {
	if f.Data == nil {
		return fmt.Errorf("yaml file not loaded, no data exists")
	}
	data, err := f.Data.GetPath(path...)
	if err != nil {
		return err
	}

	bytes, err := yaml.Marshal(data)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(bytes, target)
}

type SecretFile struct {
	*YamlFile
	Data         SecretFileData
	RelativePath string
}

func (sf *SecretFile) LoadSecretFileData(fs afero.Fs) error {
	err := sf.File.Load(fs)
	if err != nil {
		return err
	}

	var d SecretFileData
	if err := yaml.Unmarshal(sf.Bytes, &d); err != nil {
		return err
	}
	sf.Data = d
	return nil
}

type SecretFileList []*SecretFile

func NewSecretFile(file *YamlFile, relativeSecretPath string) (*SecretFile, error) {
	return &SecretFile{
		YamlFile:     file,
		RelativePath: relativeSecretPath,
	}, nil
}

func (sfl SecretFileList) GetSecretFile(path string) *SecretFile {
	for _, secretFile := range sfl {
		if strings.EqualFold(secretFile.RelativePath, path) {
			return secretFile
		}
	}
	return nil
}
