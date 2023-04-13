package skipper

import (
	"fmt"
	"io/fs"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

var (
	ErrFilePathEmpty    = fmt.Errorf("file path is empty")
	ErrPathDoesNotExist = fmt.Errorf("path does not exist")
)

// File is just an arbitrary description of a path and the data of the File to which Path points to.
// Note that the used filesystem is not relevant, only at the time of loading a File.
type File struct {
	Path  string
	Mode  fs.FileMode
	Bytes []byte
}

// NewFile creates a new [File] instance.
// If the path is empty (nil, [ErrFilePathEmpty]) is returned.
func NewFile(path string) (*File, error) {
	if path == "" {
		return nil, ErrFilePathEmpty
	}

	return &File{Path: path}, nil
}

// LoadFile creates a new [File], checks if the path exists in the given [afero.Fs] and
// loads the file contents.
func LoadFile(path string, fs afero.Fs) (*File, error) {
	f, err := NewFile(path)
	if err != nil {
		return nil, err
	}
	if !f.Exists(fs) {
		return nil, fmt.Errorf("%w: %s", ErrPathDoesNotExist, path)
	}
	err = f.Load(fs)
	if err != nil {
		return nil, err
	}

	return f, nil
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

type YamlFile struct {
	File
	Data map[string]interface{}
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

func LoadYamlFile(path string, fs afero.Fs) (*YamlFile, error) {
	f, err := NewYamlFile(path)
	if err != nil {
		return nil, err
	}

	if !f.Exists(fs) {
		return nil, fmt.Errorf("%w: %s", ErrPathDoesNotExist, path)
	}

	err = f.Load(fs)
	if err != nil {
		return nil, err
	}

	return f, nil
}

// Load will first load the underlying raw file-data and then attempt to `yaml.Unmarshal` it into `Data`
// The resulting Data is stored in `YamlFile.Data`.
func (f *YamlFile) Load(fs afero.Fs) error {
	err := f.File.Load(fs)
	if err != nil {
		return err
	}

	data := make(map[string]interface{})
	if err := yaml.Unmarshal(f.Bytes, &data); err != nil {
		return err
	}
	f.Data = data
	return nil
}
