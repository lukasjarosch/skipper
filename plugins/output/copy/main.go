package main

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/lukasjarosch/skipper"
	"github.com/lukasjarosch/skipper/codec"
	"github.com/lukasjarosch/skipper/data"
)

// Plugin symbol which is loaded by skipper.
// If this does not exist, the plugin cannot be used.
var Plugin = Copy{}

type Config struct {
	SourcePaths []string `yaml:"sourcePaths"`
	TargetPath  string   `yaml:"targetPath"`
	// OverwriteChmod os.FileMode `yaml:"overwriteChmod,omitempty"`
}

func NewConfig(raw map[string]interface{}) (Config, error) {
	yaml := codec.NewYamlCodec()
	b, err := yaml.Marshal(raw)
	if err != nil {
		return Config{}, err
	}

	var config Config
	err = yaml.UnmarshalTarget(b, &config)
	if err != nil {
		return Config{}, err
	}
	return config, nil
}

var ConfigPath = data.NewPath("config.plugins.output.copy")

type Copy struct {
	config Config
}

// ConfigPointer returns a pointer to [Config] which is going to be used by skipper
// to automatically load the config into said pointer, nice.
func (copy *Copy) ConfigPointer() interface{} {
	return &copy.config
}

func (copy *Copy) Configure() error {
	spew.Dump(copy.config)
	// configValue, err := inventory.GetPath(ConfigPath)
	// if err != nil {
	// 	return fmt.Errorf("config path does not exist: %w", err)
	// }
	//
	// // TODO: this should be done by the manager, not the plugin
	// // The plugin supports multiple instances of its configuration
	// switch typ := configValue.Raw.(type) {
	// case map[string]interface{}:
	// 	config, err := NewConfig(typ)
	// 	if err != nil {
	// 		return fmt.Errorf("cannot create config: %w", err)
	// 	}
	// 	copy.configs = append(copy.configs, config)
	// 	break
	// case []interface{}:
	// 	for _, c := range typ {
	// 		if _, ok := c.(map[string]interface{}); !ok {
	// 			return fmt.Errorf("malformed config: expected slice of maps, got %T", c)
	// 		}
	// 		conf, err := NewConfig(c.(map[string]interface{}))
	// 		if err != nil {
	// 			return fmt.Errorf("cannot create config: %w", err)
	// 		}
	// 		copy.configs = append(copy.configs, conf)
	// 	}
	// 	break
	// default:
	// 	return fmt.Errorf("malformed config at '%s': either a slice of configs or a simple config is expected", ConfigPath)
	// }
	//
	// // At ConfigPath there could be a slice of maps, indicating
	// // multiple configurations of this plugin.
	// if configs, ok := configValue.Raw.([]interface{}); ok {
	// 	for _, c := range configs {
	// 		if _, ok := c.(map[string]interface{}); !ok {
	// 			return fmt.Errorf("malformed config: expected slice of maps, got %T", c)
	// 		}
	// 		conf, err := NewConfig(c.(map[string]interface{}))
	// 		if err != nil {
	// 			return fmt.Errorf("cannot create config: %w", err)
	// 		}
	// 		copy.configs = append(copy.configs, conf)
	// 	}
	// }
	//
	// // There could also be just one map with the plugin config
	// else if c, ok := configValue.Raw.(map[string]interface{}); ok {
	// 	config, err := NewConfig(c)
	// 	if err != nil {
	// 		return fmt.Errorf("cannot create config: %w", err)
	// 	}
	// 	copy.configs = append(copy.configs, config)
	// }

	return nil
}

func (copy *Copy) Run() error {
	return nil
}

func (copy *Copy) Name() string {
	return "copy"
}

func (copy *Copy) Type() skipper.PluginType {
	return skipper.OutputPlugin
}
