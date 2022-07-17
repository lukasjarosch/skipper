package skipper

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
	Name        string
	UsedClasses []string
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

	target := &Target{
		File: file,
		Name: name,
	}

	err := target.loadUsedClasses()
	if err != nil {
		return nil, err
	}

	// TODO: do we allow to set the 'target.name' key or is it automatically populated with the target name?
	// Or do we handle it the same as kapitan where the value must match the filename?

	return target, nil
}

func (t *Target) Data() Data {
	return t.File.Data.Get(targetKey)
}

// loadUsedClasses will check that the target has the 'use' key,
// with a value of kind []string which is not empty. At least one class must be used by every target.
// If these preconditions are met, the values are loaded into 'UsedClasses'.
func (t *Target) loadUsedClasses() error {
	if !t.File.Data.Get(targetKey).HasKey(useKey) {
		return fmt.Errorf("target does not have a '%s.%s' key", targetKey, useKey)
	}

	useValue := t.File.Data.Get(targetKey)[useKey]
	if useValue == nil {
		return fmt.Errorf("target must use at least one class")
	}

	if reflect.TypeOf(useValue).Kind() != reflect.Slice {
		return fmt.Errorf("%s.%s must be a string array", targetKey, useKey)
	}

	// convert []interface to []string
	for _, class := range useValue.([]interface{}) {
		t.UsedClasses = append(t.UsedClasses, class.(string))
	}

	return nil
}
