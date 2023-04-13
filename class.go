package skipper

import (
	"errors"
	"fmt"
)

// SkipperKey is the expected key within [Data] under which
// the [ClassConfig] is meant to be located.
const SkipperKey = "skipper"

var (
	ErrInvalidRootKey = errors.New("invalid root key")
)

type ClassConfig struct {
	Includes []Path `yaml:"use"`
}

//go:generate mockery --name DataProvider
type DataProvider interface {
	GetPath(path Path) (interface{}, error)
	HasPath(path Path) bool
	UnmarshalPath(path Path, target interface{}) error
	Keys() []string
}

var (
	ErrNilDataProvider  = errors.New("DataProvider is nil")
	ErrMultipleRootKeys = errors.New("multiple root keys")
)

type Class struct {
	Data          DataProvider
	Namespace     Path
	Configuration *ClassConfig
	RootKey       string
}

// NewClass creates a new instance of Class for the given namespace and data source.
// If the namespace is empty, returns (nil, ErrEmptyNamespace).
// If the data source is nil, returns (nil, ErrNilDataSource).
// If the data source doesn't contain a key matching the last segment of the namespace,
// returns (nil, ErrInvalidRootKey).
// Otherwise, returns a pointer to a new Class instance.
func NewClass(namespace Path, data DataProvider) (*Class, error) {
	if len(namespace) == 0 {
		return nil, ErrEmptyPath
	}

	if data == nil {
		return nil, ErrNilDataProvider
	}

	// The expected root key in the [DataProvider] is the last segment of the namespace.
	rootKey := namespace[len(namespace)-1]
	if !data.HasPath(P(rootKey)) {
		return nil, fmt.Errorf("%w: expected '%s'", ErrInvalidRootKey, rootKey)
	}

	// Because the root key is derived from the namespace a class can only have one root key.
	// This also forces users to not shove everything into one file and actually use namespaces.
	if len(data.Keys()) > 1 {
		return nil, ErrMultipleRootKeys
	}

	class := Class{
		Namespace: namespace,
		Data:      data,
		RootKey:   namespace[len(namespace)-1],
	}

	err := class.loadConfig()
	if err != nil {
		return nil, err
	}

	return &class, nil
}

// loadConfig attempts to load the configuration for this class using the [DataProvider].
// The configuration is expected to be accessible in the Data at the key `[RootKey].[SkipperKey]`.
// If the configuration exists, it is unmarshaled into a ClassConfig struct and stored in
// the class's Configuration field. If the configuration is not found the Configuration field remains nil and no error is returned.
// If unmarshalling fails, an error is returned. This indicates that the raw input (yaml data) is wrong.
func (class *Class) loadConfig() error {
	configPath := Path{class.RootKey, SkipperKey}
	if class.Data.HasPath(configPath) {
		var config ClassConfig
		err := class.Data.UnmarshalPath(configPath, &config)
		if err != nil {
			return fmt.Errorf("unable to load config with path '%s': %w", configPath.String(), err)
		}
		class.Configuration = &config
	}

	return nil
}

// Get returns the value at the given path in the class's data source.
// If the path is not found, the second return value is false.
// If the data is found it is returned and the second return value is true
// which allows idiomatic existence checks.
func (class *Class) Get(path Path) (interface{}, bool) {
	out, err := class.Data.GetPath(path)
	if err != nil {
		return nil, false
	}
	return out, true
}
