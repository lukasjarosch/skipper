package templater

import (
	"fmt"
	"path/filepath"
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

	return &Class{
		File: file,
		Name: name,
	}, nil
}
