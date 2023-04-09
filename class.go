package skipper

import (
	"errors"
	"fmt"
	"log"
)

const SkipperKey = "skipper"

type ClassConfig struct {
	Includes []Namespace `yaml:"use"`
}

//go:generate mockery --name DataSource
type DataSource interface {
	HasPath(path DataPath) bool
	UnmarshalPath(path DataPath, target interface{}) error
}

var (
	ErrNilDataSource = errors.New("DataSource is nil")
)

type Class struct {
	Data          DataSource
	Namespace     Namespace
	Configuration *ClassConfig
	RootKey       string
}

func NewClass(namespace Namespace, data DataSource) (*Class, error) {
	if namespace == "" {
		return nil, ErrEmptyNamespace
	}

	if data == nil {
		return nil, ErrNilDataSource
	}

	class := Class{
		Namespace: namespace,
		Data:      data,
		RootKey:   namespace.Segments()[len(namespace.Segments())-1],
	}

	configPath := DataPath{class.RootKey, SkipperKey}
	if data.HasPath(configPath) {
		log.Println("CONFIG EXISTS in:", configPath.String())

		var config ClassConfig
		err := class.Data.UnmarshalPath(configPath, &config)
		if err != nil {
			return nil, fmt.Errorf("unable to load config with path '%s': %w", configPath.String(), err)
		}
		class.Configuration = &config
	}

	return &class, nil
}
