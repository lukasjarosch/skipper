package main

import (
	"path"

	"github.com/lukasjarosch/skipper"
	"github.com/spf13/afero"
)

var (
	fileSystem    = afero.NewOsFs()
	inventoryPath = "inventory"
	classPath     = path.Join(inventoryPath, "classes")
	targetPath    = path.Join(inventoryPath, "targets")
	secretPath    = path.Join(inventoryPath, "secrets")
	templatePath  = path.Join(inventoryPath, "..", "templates")
	outputPath    = "compiled"

	target = "example"
)

func main() {
	inventory, err := skipper.NewInventory(fileSystem, classPath, targetPath, secretPath)
	if err != nil {
		panic(err)
	}

	// Process the inventory, given the target name
	data, err := inventory.Data(target, nil, false)
	if err != nil {
		panic(err)
	}

	templateOutputPath := path.Join(outputPath, target)
	templater, err := skipper.NewTemplater(fileSystem, templatePath, templateOutputPath, nil)
	if err != nil {
		panic(err)
	}

	skipperConfig, err := inventory.GetSkipperConfig(target)
	if err != nil {
		panic(err)
	}

	// execute templates  ----------------------------------------------------------------------------------
	err = templater.ExecuteAll(skipper.DefaultTemplateContext(data, target), false, nil)
	if err != nil {
		panic(err)
	}

	// copy files as specified in the target config (base path is template root)
	err = skipper.CopyFilesByConfig(fileSystem, skipperConfig.Copies, templatePath, templateOutputPath)
	if err != nil {
		panic(err)
	}
}
