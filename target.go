package skipper

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	targetKey string = "target"
	useKey    string = "use"
)

var wildcardUseRegex regexp.Regexp = *regexp.MustCompile(`^\w+\.\*$`)

// Target defines which classes to use for the compilation.
type Target struct {
	File *YamlFile
	// Name is the relative path of the file inside the inventory
	// where '/' is replaced with '.' and without file extension.
	Name string
	// UsedWildcardClasses holds all resolved wildcard class imports as specified in the `targets.skipper.use` key.
	UsedWildcardClasses []string
	// Configuration is the skipper-internal configuration which needs to be present on every target.
	Configuration TargetConfig
	// SkipperConfig is the generic Skipper configuration which can be use throughout targets and classes
	SkipperConfig *SkipperConfig
}

type TargetConfig struct {
	Secrets    TargetSecretConfig `mapstructure:"secrets,omitempty"`
	Components []ComponentConfig  `mapstructure:"components,omitempty"`
}

type TargetSecretConfig struct {
	Drivers map[string]interface{} `mapstructure:"drivers"`
}

func NewTarget(file *YamlFile, inventoryPath string) (*Target, error) {
	if file == nil {
		return nil, fmt.Errorf("file cannot be nil")
	}
	if inventoryPath == "" {
		return nil, fmt.Errorf("inventoryPath cannot be empty")
	}

	// create target name from inventory-relative path
	fileName := strings.TrimSuffix(inventoryPath, filepath.Ext(inventoryPath))
	name := strings.ReplaceAll(fileName, "/", ".")

	// every target must have the same root key
	if !file.Data.HasKey(targetKey) {
		return nil, fmt.Errorf("target must have valid top-level key")
	}

	// every target must have the 'skipper' key, which is used to load the Skipper-internal target configuration
	var config TargetConfig
	err := file.UnmarshalPath(&config, targetKey, skipperKey)
	if err != nil {
		return nil, fmt.Errorf("missing skipper key in target: %w", err)
	}

	// attempt to load the generic skipper config
	skipperConfig, err := LoadSkipperConfig(file, targetKey)
	if err != nil {
		return nil, err
	}

	target := &Target{
		File:          file,
		Name:          name,
		Configuration: config,
		SkipperConfig: skipperConfig,
	}

	err = target.loadUsedClasses()
	if err != nil {
		return nil, err
	}

	// TODO: do we allow to set the 'target.name' key or is it automatically populated with the target name?
	// Or do we handle it the same as kapitan where the value must match the filename?

	return target, nil
}

func (t *Target) ReloadConfiguration() {
	// every target must have the 'skipper' key, which is used to load the Skipper-internal target configuration
	var config TargetConfig
	t.File.UnmarshalPath(&config, targetKey, skipperKey)
	t.Configuration = config
}

func (t *Target) Data() Data {
	return t.File.Data.Get(targetKey)
}

// loadUsedClasses will check that the target has the 'use' key,
// with a value of kind []string which is not empty. At least one class must be used by every target.
// If these preconditions are met, the values are loaded into 'UsedClasses'.
func (t *Target) loadUsedClasses() error {

	// convert []interface to []string
	for _, class := range t.SkipperConfig.Classes {
		// load wildcard imports separately as they need to be resolved
		if match := wildcardUseRegex.FindAllString(class, 1); len(match) == 1 {
			wildcardUse := match[0]
			t.UsedWildcardClasses = append(t.UsedWildcardClasses, wildcardUse)
			continue
		}
		t.SkipperConfig.Classes = append(t.SkipperConfig.Classes, class)
	}

	return nil
}
