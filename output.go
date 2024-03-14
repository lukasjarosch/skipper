package skipper

import (
	"fmt"
	"reflect"

	"github.com/lukasjarosch/skipper/codec"
	"github.com/lukasjarosch/skipper/data"
)

type OutputType string

func (o OutputType) Validate() error {
	if len(o) == 0 {
		return ErrEmptyOutputType
	}
	return nil
}

var (
	ErrOutputAlreadyRegistered = fmt.Errorf("output already registered")
	ErrEmptyOutputType         = fmt.Errorf("empty output-type")
	ErrOutputDoesNotExist      = fmt.Errorf("output does not exist")
)

type (
	// OutputConstructorFunc returns an Output.
	// It is used to let the OutputManager know how to create instances of an output.
	OutputConstructorFunc func() Output
	// Configurable defines behaviour of anything which can be configured (duh).
	Configurable interface {
		// ConfigPointer must return a pointer (!) to the config struct of the output.
		// The configuration will be unmarshalled using YAML into the given struct.
		// Struct-tags can be used.
		ConfigPointer() interface{}
		// Configure is called after the configuration of the output is injected.
		// The output can then configure whatever it needs.
		Configure() error
	}
	Output interface {
		Run() error
	}
)

type OutputManager struct {
	registered map[OutputType]OutputConstructorFunc
	configured map[OutputType][]Output
}

func NewOutputManager() *OutputManager {
	pm := &OutputManager{
		registered: make(map[OutputType]OutputConstructorFunc),
		configured: make(map[OutputType][]Output),
	}

	return pm
}

func (om *OutputManager) RegisterOutput(outputType OutputType, outputConstructorFunc OutputConstructorFunc) error {
	if err := outputType.Validate(); err != nil {
		return err
	}
	if _, exists := om.registered[outputType]; exists {
		return fmt.Errorf("%w: %s", ErrOutputAlreadyRegistered, outputType)
	}
	om.registered[outputType] = outputConstructorFunc

	return nil
}

func (om *OutputManager) CreateOutputInstance(typ OutputType) (Output, error) {
	if err := typ.Validate(); err != nil {
		return nil, err
	}

	constructor, exists := om.registered[typ]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrOutputDoesNotExist, typ)
	}

	return constructor(), nil
}

func (om *OutputManager) ConfigureOutput(outputType OutputType, config data.Value) error {
	output, err := om.CreateOutputInstance(outputType)
	if err != nil {
		return err
	}

	configurableOutput, isConfigurable := output.(Configurable)
	if !isConfigurable {
		om.configured[outputType] = append(om.configured[outputType], output)
		return nil
	}

	// Handle single instance output config
	if config.IsMap() {
		err = om.unmarshalConfig(configurableOutput, config)
		if err != nil {
			return err
		}

		err = configurableOutput.Configure()
		if err != nil {
			return fmt.Errorf("failed to configure output: %s: %w", outputType, err)
		}

		om.configured[outputType] = append(om.configured[outputType], output)
	}

	// If the config is a slice, it means that we need to have multiple instances
	// of the output with different configs.
	if configs, err := config.Slice(); err == nil {
		for _, c := range configs {
			// create a new instance of the output and configure it using the data
			// The type assertion does not need to be checked, it's been validated above.
			output := om.registered[outputType]()
			err = om.unmarshalConfig(output.(Configurable), data.NewValue(c))
			if err != nil {
				return err
			}

			err = output.(Configurable).Configure()
			if err != nil {
				return fmt.Errorf("failed to configure output: %s: %w", outputType, err)
			}

			om.configured[outputType] = append(om.configured[outputType], output)
		}
	}

	return nil
}

func (om *OutputManager) RunAll() error {
	for _, configuredSlice := range om.configured {
		for _, output := range configuredSlice {
			err := output.Run()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (om *OutputManager) Run(typ OutputType) error {
	if err := typ.Validate(); err != nil {
		return err
	}

	for _, configured := range om.configured[typ] {
		err := configured.Run()
		if err != nil {
			return err
		}
	}

	return nil
}

func (om *OutputManager) unmarshalConfig(output Configurable, config data.Value) error {
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

	if reflect.ValueOf(output.ConfigPointer()).Kind() != reflect.Ptr {
		return fmt.Errorf("cannot unmarshal output config, ConfigPointer must return a pointer")
	}

	err = configCodec.UnmarshalTarget(b, output.ConfigPointer())
	if err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return nil
}
