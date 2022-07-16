package templater

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
)

const (
	targetKey string = "target"
	useKey    string = "use"
)

// Target defines which classes to use for the compilation.
type Target struct {
	File *YamlFile
	// Name is the relative path of the file inside the inventory
	// where '/' is replaced with '.' and without file extension.
	Name string
}

type TargetConfig struct {
	Use []string `mapstructure:"use"`
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

	// Targets must at least use one class, otherwise there would be no data available from the inventory.
	if !file.Data.Get(targetKey).HasKey(useKey) {
		return nil, fmt.Errorf("target must use at least one class")
	}

	// ensure that the useKey is an array
	if reflect.TypeOf(file.Data.Get(targetKey)[useKey]).Kind() != reflect.Slice {
		return nil, fmt.Errorf("%s.%s must be a string array", targetKey, useKey)
	}

	return &Target{
		File: file,
		Name: name,
	}, nil
}
