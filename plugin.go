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
	// PluginSymbolName defines the symbol which will be loaded from the shared library.
	PluginSymbolName string = "Plugin"

	OutputPlugin PluginType = "output"
)

var (
	ErrEmptyPluginName      = fmt.Errorf("empty plugin name")
	ErrEmptyPluginType      = fmt.Errorf("empty plugin type")
	ErrUnknownPluginType    = fmt.Errorf("unknown plugin type")
	ErrPluginNotLoaded      = fmt.Errorf("plugin not loaded")
	ErrUnconfigurablePlugin = fmt.Errorf("plugin cannot be configured, does not implement the 'ConfigurablePlugin' interface")
)

type (
	// PluginConstructor is a function which needs to be exported with the symbol SymbolName by the plugins.
	PluginConstructor  func() Plugin
	ConfigurablePlugin interface {
		// ConfigPointer must return a pointer (!) to the config struct of the plugin.
		// The configuration will be unmarshalled using YAML into the given struct.
		// Struct-tags can be used.
		ConfigPointer() interface{}
		// Configure is called after the configuration of the plugin is injected.
		// The plugin can then configure whatever it needs.
		Configure() error
	}
	PluginMetadataProvider interface {
		Name() string
		Type() PluginType
	}
	Plugin interface {
		PluginMetadataProvider
		Run() error
	}
)

type PluginManager struct {
	plugins            map[PluginType]map[string]Plugin
	configuredPlugins  map[PluginType]map[string][]Plugin
	pluginConstructors map[PluginType]map[string]PluginConstructor
}

func NewPluginManager() *PluginManager {
	pm := &PluginManager{
		plugins:            make(map[PluginType]map[string]Plugin),
		pluginConstructors: make(map[PluginType]map[string]PluginConstructor),
		configuredPlugins:  make(map[PluginType]map[string][]Plugin),
	}

	// for every plugin type, create the sub-maps
	for _, t := range []PluginType{OutputPlugin} {
		pm.plugins[t] = make(map[string]Plugin)
		pm.pluginConstructors[t] = make(map[string]PluginConstructor)
		pm.configuredPlugins[t] = make(map[string][]Plugin)
	}

	return pm
}

func (pm *PluginManager) LoadPlugin(path string) error {
	pluginConstructor, err := pm.loadPluginConstructor(path)
	if err != nil {
		return err
	}

	plugin := pluginConstructor()
	if err := pm.validateMetadata(plugin); err != nil {
		return err
	}

	if _, exists := pm.plugins[plugin.Type()][plugin.Name()]; exists {
		return fmt.Errorf("plugin %s already loaded", PluginID(plugin.Type(), plugin.Name()))
	}

	pm.plugins[plugin.Type()][plugin.Name()] = plugin
	pm.pluginConstructors[plugin.Type()][plugin.Name()] = pluginConstructor

	return nil
}

func (pm *PluginManager) loadPluginConstructor(path string) (PluginConstructor, error) {
	p, err := plugin.Open(path)
	if err != nil {
		return nil, fmt.Errorf("unable to open plugin: %w", err)
	}

	symbol, err := p.Lookup(PluginSymbolName)
	if err != nil {
		return nil, fmt.Errorf("unable to find symbol '%s': %w", PluginSymbolName, err)
	}

	pluginConstructor, validPlugin := symbol.(*PluginConstructor)
	if !validPlugin {
		return nil, fmt.Errorf("invalid plugin: invalid symbol '%s', want=PluginConstructor, have=%T", PluginSymbolName, pluginConstructor)
	}

	return *pluginConstructor, nil
}

func (pm *PluginManager) ConfigurePlugin(typ PluginType, name string, config data.Value) error {
	plugin, err := pm.GetPlugin(typ, name)
	if err != nil {
		return err
	}

	// If the plugin does not need to be configured, pretend we did and return an error.
	// If the user doesn't mind about plugins not being configured, just ignore the error.
	iplugin, configurable := plugin.(ConfigurablePlugin)
	if !configurable {
		pm.configuredPlugins[typ][name] = append(pm.configuredPlugins[typ][name], plugin)
		return ErrUnconfigurablePlugin
	}

	// If the config is a slice, it means that we need to have multiple instances
	// of the plugin with different configs.
	if configs, err := config.Slice(); err == nil {
		for _, c := range configs {
			// create a new instance of the plugin and configure it using the data
			plugin := pm.pluginConstructors[plugin.Type()][plugin.Name()]()
			err = pm.unmarshalConfig(plugin.(ConfigurablePlugin), data.NewValue(c))
			if err != nil {
				return err
			}

			err = plugin.(ConfigurablePlugin).Configure()
			if err != nil {
				return fmt.Errorf("failed to configure plugin %s: %w", PluginID(plugin.Type(), plugin.Name()), err)
			}

			pm.configuredPlugins[typ][name] = append(pm.configuredPlugins[typ][name], plugin)

			// TODO: remove
			err = plugin.Run()
			if err != nil {
				return err
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

		pm.configuredPlugins[typ][name] = append(pm.configuredPlugins[typ][name], plugin)
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
