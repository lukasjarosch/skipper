package skipper

import (
	"fmt"
)

// skipperKey is the key used to load skipper-related configurations from YAML files
const skipperKey string = "skipper"

type SkipperConfig struct {
	// TODO: for some reason, this is not yet used
	Classes    []string          `yaml:"use,omitempty"`
	Components []ComponentConfig `mapstructure:"components,omitempty"`
	Copies     []CopyConfig      `yaml:"copy,omitempty"`
}

type CopyConfig struct {
	// SourcePath is the source file to copy, relative to the template-root
	SourcePath string `yaml:"source"`
	// TargetPath is the target to copy the source file to, relative to the compile-root
	TargetPath string `yaml:"target"`
}

type ComponentConfig struct {
	OutputPath string                  `yaml:"output_path"`
	InputPaths []string                `yaml:"input_paths"`
	Renames    []RenameComponentConfig `yaml:"rename"`
}

type RenameComponentConfig struct {
	InputPath string `yaml:"input_path"`
	Filename  string `yaml:"filename"`
}

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

// IsSet returns true if the config is not nil.
// The function is useful because LoadSkipperConfig can return nil.
func (config *SkipperConfig) IsSet() bool {
	return config != nil
}
