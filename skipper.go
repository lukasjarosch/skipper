package skipper

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/afero"
)

// skipperKey is the key used to load skipper-related configurations from YAML files
const skipperKey string = "skipper"

type SkipperConfig struct {
	Classes    []string          `yaml:"use,omitempty"`
	Components []ComponentConfig `mapstructure:"components,omitempty"`
	Copies     []CopyConfig      `yaml:"copy,omitempty"`
	Renames    []RenameConfig    `yaml:"rename,omitempty"`
}

type CopyConfig struct {
	// SourcePath is the source file to copy, relative to the template-root
	SourcePath string `yaml:"source"`
	// TargetPath is the target to copy the source file to, relative to the compile-root
	TargetPath string `yaml:"target"`
}

type ComponentConfig struct {
	OutputPath string         `yaml:"output_path"`
	InputPaths []string       `yaml:"input_paths"`
	Renames    []RenameConfig `yaml:"rename"`
}

type RenameConfig struct {
	InputPath string `yaml:"input_path"`
	Filename  string `yaml:"filename"`
}

// IsSet returns true if the config is not nil.
// The function is useful because LoadSkipperConfig can return nil.
func (config *SkipperConfig) IsSet() bool {
	return config != nil
}

// MergeSkipperConfig merges a list of configs into one
func MergeSkipperConfig(merge ...*SkipperConfig) (mergedConfig *SkipperConfig) {
	mergedConfig = new(SkipperConfig)
	for _, config := range merge {
		mergedConfig.Classes = append(mergedConfig.Classes, config.Classes...)
		mergedConfig.Components = append(mergedConfig.Components, config.Components...)
		mergedConfig.Copies = append(mergedConfig.Copies, config.Copies...)
	}
	return mergedConfig
}

// LoadSkipperConfig attempts to load a SkipperConfig from the given YamlFile with the passed rootKey
func LoadSkipperConfig(file *YamlFile, rootKey string) (*SkipperConfig, error) {
	if file == nil {
		return nil, fmt.Errorf("file cannot be nil")
	}
	if file.Data == nil {
		return nil, fmt.Errorf("file.Data cannot be nil")
	}

	// if not skipper config exists, return nothing
	if _, err := file.Data.GetPath(rootKey, skipperKey); err != nil {
		return nil, nil
	}

	var config SkipperConfig
	err := file.UnmarshalPath(&config, rootKey, skipperKey)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal SkipperConfig: %w", err)
	}

	return &config, nil
}

// CopyFilesByConfig uses a list of CopyConfigs and calls the CopyFile func on them.
func CopyFilesByConfig(fs afero.Fs, configs []CopyConfig, sourceBasePath, targetBasePath string) error {
	for _, copyFile := range configs {
		source := filepath.Join(sourceBasePath, copyFile.SourcePath)
		target := filepath.Join(targetBasePath, copyFile.TargetPath)
		err := CopyFile(fs, source, target)
		if err != nil {
			return err
		}
	}
	return nil
}
