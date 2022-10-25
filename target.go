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
	// UsedClasses holds the resolved class names which are specified in the `target.skipper.use` key.
	UsedClasses []string
	// UsedWildcardClasses holds all resolved wildcard class imports as specified in the `targets.skipper.use` key.
	UsedWildcardClasses []string
	// Configuration is the skipper-internal configuration which needs to be present on every target.
	Configuration TargetConfig
	// SkipperConfig is the generic Skipper configuration which can be use throughout targets and classes
	SkipperConfig *SkipperConfig
}

type TargetConfig struct {
	Use        []string           `mapstructure:"use"`
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

	err = target.loadUsedWildcardClasses()
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

// loadUsedWildcardClasses will extract all wildcards-uses from the configuration,
// remove them from the loaded configuration and store them in UsedWildcardClasses for
// further processing.
func (t *Target) loadUsedWildcardClasses() error {

	// convert []interface to []string
	for _, class := range t.Configuration.Use {
		// load wildcard imports separately as they need to be resolved
		if match := wildcardUseRegex.FindAllString(class, 1); len(match) == 1 {
			wildcardUse := match[0]
			t.UsedWildcardClasses = append(t.UsedWildcardClasses, wildcardUse)
			continue
		}

		t.UsedClasses = append(t.UsedClasses, class)
	}

	return nil
}

// targetYamlFileLoader returns a YamlFileLoaderFunc which is capable of
// creating Targets from a given YamlFile.
// The created targets are then appended to the passed targetList.
func targetYamlFileLoader(targetList *[]*Target) YamlFileLoaderFunc {
	return func(file *YamlFile, relativePath string) error {
		target, err := NewTarget(file, relativePath)
		if err != nil {
			return fmt.Errorf("%s: %w", file.Path, err)
		}

		(*targetList) = append((*targetList), target)

		return nil
	}
}
