package skipper

import (
	"fmt"
	"plugin"
	"reflect"

	"github.com/lukasjarosch/skipper/codec"
	"github.com/lukasjarosch/skipper/data"
)

type PluginType string

const (
	// SymbolName defines the symbol which will be loaded from the shared library.
	SymbolName string = "Plugin"

	OutputPlugin PluginType = "output"
)

var (
	ErrEmptyPluginName      = fmt.Errorf("empty plugin name")
	ErrEmptyPluginType      = fmt.Errorf("empty plugin type")
	ErrUnknownPluginType    = fmt.Errorf("unknown plugin type")
	ErrPluginNotLoaded      = fmt.Errorf("plugin not loaded")
	ErrUnconfigurablePlugin = fmt.Errorf("plugin cannot be configured, does not implement the 'ConfigurablePlugin' interface")
)

type PluginMetadataProvider interface {
	Name() string
	Type() PluginType
}

type ConfigurablePlugin interface {
	// ConfigPointer must return a pointer (!) to the config struct of the plugin.
	// The configuration will be unmarshalled using YAML into the given struct.
	// Struct-tags can be used.
	ConfigPointer() interface{}
	// Configure is called after the configuration of the plugin is injected.
	// The plugin can then configure whatever it needs.
	Configure() error
}

type Plugin interface {
	PluginMetadataProvider
	Run() error
}

type PluginManager struct {
	plugins           map[PluginType]map[string]Plugin
	configuredPlugins map[PluginType]map[string][]Plugin
	pluginPaths       map[PluginType]map[string]string
}

func NewPluginManager() *PluginManager {
	pm := &PluginManager{
		plugins:           make(map[PluginType]map[string]Plugin),
		pluginPaths:       make(map[PluginType]map[string]string),
		configuredPlugins: make(map[PluginType]map[string][]Plugin),
	}

	for _, t := range []PluginType{OutputPlugin} {
		pm.plugins[t] = make(map[string]Plugin)
		pm.pluginPaths[t] = make(map[string]string)
		pm.configuredPlugins[t] = make(map[string][]Plugin)
	}

	return pm
}

func (pm *PluginManager) LoadPlugin(path string) error {
	plugin, err := pm.loadPlugin(path)
	if err != nil {
		return err
	}

	if _, exists := pm.plugins[plugin.Type()][plugin.Name()]; exists {
		return fmt.Errorf("plugin %s already loaded", PluginID(plugin.Type(), plugin.Name()))
	}

	pm.plugins[plugin.Type()][plugin.Name()] = plugin

	return nil
}

func (pm *PluginManager) loadPlugin(path string) (Plugin, error) {
	p, err := plugin.Open(path)
	if err != nil {
		return nil, fmt.Errorf("unable to open plugin: %w", err)
	}

	symbol, err := p.Lookup(SymbolName)
	if err != nil {
		return nil, fmt.Errorf("unable to find symbol '%s': %w", SymbolName, err)
	}

	loadedPlugin, validPlugin := symbol.(Plugin)
	if !validPlugin {
		return nil, fmt.Errorf("invalid plugin: does not implement Plugin interface")
	}

	if err := pm.validateMetadata(loadedPlugin); err != nil {
		return nil, err
	}

	return loadedPlugin, nil
}

// TODO: maybe add a 'MustInitializePlugin' to check for the 'ConfigurePlugin' interface?
func (pm *PluginManager) ConfigurePlugin(typ PluginType, name string, config data.Value) error {
	plugin, err := pm.GetPlugin(typ, name)
	if err != nil {
		return err
	}

	// If the plugin does not need to be initialized, pretend we did and return an error.
	// If the user doesn't mind about plugins not being initialized, just ignore the error.
	iplugin, initializable := plugin.(ConfigurablePlugin)
	if !initializable {
		pm.configuredPlugins[typ][name] = append(pm.configuredPlugins[typ][name], plugin)
		return ErrUnconfigurablePlugin
	}

	// If the config is a slice, it means that we need to have multiple instances
	// of the plugin with different configs.
	// WARN: how do we copy plugins? re-load? deep-copy?
	if config.IsSlice() {
		configs, _ := config.Slice()
		for _, c := range configs {
			err = pm.unmarshalConfig(iplugin, data.NewValue(c))
			if err != nil {
				return err
			}

			err = iplugin.Configure()
			if err != nil {
				return fmt.Errorf("failed to initialize %s: %w", PluginID(plugin.Type(), plugin.Name()), err)
			}
		}
	}

	// Handle single instance plugin config
	if config.IsMap() {
		err = pm.unmarshalConfig(iplugin, config)
		if err != nil {
			return err
		}

		err = iplugin.Configure()
		if err != nil {
			return fmt.Errorf("failed to initialize %s: %w", PluginID(plugin.Type(), plugin.Name()), err)
		}
		return nil
	}

	return nil
}

func (pm *PluginManager) unmarshalConfig(plugin ConfigurablePlugin, config data.Value) error {
	// TODO: make configurable to allow user to use different struct-tags?
	configCodec := codec.NewYamlCodec()

	configMap, err := config.Map()
	if err != nil {
		return fmt.Errorf("config must be a map: %w", err)
	}

	b, err := configCodec.Marshal(configMap)
	if err != nil {
		return fmt.Errorf("cannot marshal config: %w", err)
	}

	if reflect.ValueOf(plugin.ConfigPointer()).Kind() != reflect.Ptr {
		return fmt.Errorf("cannot unmarshal plugin config, ConfigType must return a pointer")
	}

	err = configCodec.UnmarshalTarget(b, plugin.ConfigPointer())
	if err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return nil
}

func (pm *PluginManager) GetPlugin(typ PluginType, name string) (Plugin, error) {
	if len(typ) == 0 {
		return nil, ErrEmptyPluginType
	}
	if !ValidPluginType(typ) {
		return nil, fmt.Errorf("%w: %s", ErrUnknownPluginType, typ)
	}
	if err := ValidatePluginName(name); err != nil {
		return nil, err
	}

	plugin, exists := pm.plugins[typ][name]
	if !exists {
		return nil, fmt.Errorf("%s: %w", PluginID(typ, name), ErrPluginNotLoaded)
	}

	return plugin, nil
}

func (pm *PluginManager) validateMetadata(meta PluginMetadataProvider) error {
	if err := ValidatePluginName(meta.Name()); err != nil {
		return err
	}

	if !ValidPluginType(meta.Type()) {
		return fmt.Errorf("%w: %s", ErrUnknownPluginType, meta.Type())
	}
	return nil
}

func PluginID(typ PluginType, name string) string {
	return fmt.Sprintf("%s/%s", typ, name)
}

func ValidPluginType(typ PluginType) bool {
	switch typ {
	case OutputPlugin:
		return true
	}
	return false
}

func ValidatePluginName(name string) error {
	if len(name) == 0 {
		return ErrEmptyPluginName
	}
	return nil
}
