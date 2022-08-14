package skipper

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
)

// Class defines a single entity (yaml file) inside the inventory.
type Class struct {
	File *YamlFile
	// Name is the relative path of the file inside the inventory
	// where '/' is replaced with '.' and without file extension.
	Name string
}

func NewClass(file *YamlFile, inventoryPath string) (*Class, error) {
	if file == nil {
		return nil, fmt.Errorf("file cannot be nil")
	}
	if inventoryPath == "" {
		return nil, fmt.Errorf("inventoryPath cannot be empty")
	}

	// create class name from inventory-relative path
	fileName := strings.TrimSuffix(inventoryPath, filepath.Ext(inventoryPath))
	name := strings.ReplaceAll(fileName, "/", ".")

	// class file cannot be empty, there must be exactly one yaml root-key which must define a map
	val := reflect.ValueOf(file.Data)
	if val.Kind() != reflect.Map {
		return nil, fmt.Errorf("class '%s' root key does not define a map", name)
	}
	if len(val.MapKeys()) == 0 {
		return nil, fmt.Errorf("class '%s' does not have a root-key", name)
	}
	if len(val.MapKeys()) > 1 {
		return nil, fmt.Errorf("class '%s' has more than one root-key which is currently not supported", name)
	}

	return &Class{
		File: file,
		Name: name,
	}, nil
}

func (c *Class) Data() *Data {
	return &c.File.Data
}

// RootKey returns the root key name of the class.
func (c *Class) RootKey() string {
	val := reflect.ValueOf(c.Data()).Elem()
	return val.MapKeys()[0].String()
}
