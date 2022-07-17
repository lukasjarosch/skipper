package templater

import (
	"fmt"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

type YamlFile struct {
	Path string
	Data Data
}

func NewFile(path string) (*YamlFile, error) {
	if path == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}

	f := &YamlFile{
		Path: path,
	}

	return f, nil
}

// TODO: separate from yaml, depend only on an unmarshaller interface
func (f *YamlFile) Load(fs afero.Fs) error {
	fileBytes, err := afero.ReadFile(fs, f.Path)
	if err != nil {
		return fmt.Errorf("failed to Load: %w", err)
	}

	var d Data
	if err := yaml.Unmarshal(fileBytes, &d); err != nil {
		return err
	}
	f.Data = d
	return nil
}

func (f *YamlFile) LoadAs(fs afero.Fs, target interface{}) error {
	fileBytes, err := afero.ReadFile(fs, f.Path)
	if err != nil {
		return fmt.Errorf("failed to Load: %w", err)
	}

	if err := yaml.Unmarshal(fileBytes, target); err != nil {
		return err
	}
	return nil
}
