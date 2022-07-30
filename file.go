package skipper

import (
	"fmt"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

type file struct {
	Path  string
	Bytes []byte
}

func newFile(path string) (*file, error) {
	if path == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}

	return &file{Path: path}, nil
}

func (f *file) Load(fs afero.Fs) (err error) {
	f.Bytes, err = afero.ReadFile(fs, f.Path)
	if err != nil {
		return fmt.Errorf("failed to Load %s: %w", f.Path, err)
	}
	return nil
}

type YamlFile struct {
	file
	Data Data
}

// TODO: change name to NewYamlFile
func NewFile(path string) (*YamlFile, error) {
	f, err := newFile(path)
	if err != nil {
		return nil, err
	}

	return &YamlFile{
		file: *f,
	}, nil
}

func (f *YamlFile) Load(fs afero.Fs) error {
	err := f.file.Load(fs)
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

type TemplateFile struct {
	file
}

func NewTemplateFile(path string) (*TemplateFile, error) {
	f, err := newFile(path)
	if err != nil {
		return nil, err
	}

	return &TemplateFile{
		file: *f,
	}, nil
}
