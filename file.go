package skipper

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

var (
	yamlFileExtensions = []string{".yml", ".yaml", ""}
	ErrFilePathEmpty   = fmt.Errorf("file path is empty")
)

// File is just an arbitrary description of a path and the data of the File to which Path points to.
// Note that the used filesystem is not relevant, only at the time of loading a File.
type File struct {
	path  string
	mode  fs.FileMode
	bytes []byte
}

func NewFile(path string) (*File, error) {
	if path == "" {
		return nil, ErrFilePathEmpty
	}

	f := &File{path: path}
	if err := f.load(); err != nil {
		return nil, err
	}

	return f, nil
}

// Load will attempt to read the File from the given filesystem implementation.
// The loaded data is stored in `File.Bytes`
func (f *File) load() (err error) {
	f.bytes, err = ioutil.ReadFile(f.path)
	if err != nil {
		return fmt.Errorf("failed to Load %s: %w", f.path, err)
	}
	info, err := os.Stat(f.path)
	if err != nil {
		return fmt.Errorf("unable to stat file %s: %w", f.path, err)
	}
	f.mode = info.Mode()

	return nil
}

func (f *File) Bytes() []byte {
	return f.bytes
}

func (f *File) Path() string {
	return f.path
}

func (f *File) Mode() fs.FileMode {
	return f.mode
}
func (f *File) BaseName() string {
	name := filepath.Base(f.path)
	name = strings.TrimSuffix(name, filepath.Ext(name))
	return name
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

		relativePath := strings.ReplaceAll(yamlFile.Path(), basePath, "")
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

// NewYamlFile returns a newly initialized `YamlFile`.
func NewYamlFile(path string) (*YamlFile, error) {
	f, err := NewFile(path)
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
	return NewYamlFile(path)
}

// Load will first load the underlying raw file-data and then attempt to `yaml.Unmarshal` it into `Data`
// The resulting Data is stored in `YamlFile.Data`.
func (f *YamlFile) Load(fs afero.Fs) error {
	var d Data
	if err := yaml.Unmarshal(f.Bytes(), &d); err != nil {
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
	var d SecretFileData
	if err := yaml.Unmarshal(sf.Bytes(), &d); err != nil {
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
