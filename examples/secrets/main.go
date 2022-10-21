package main

import (
	"log"
	"path"
	"path/filepath"

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

	target = "azure_keyvault"
)

func main() {
	inventory, err := skipper.NewInventory(fileSystem, classPath, targetPath, secretPath)
	if err != nil {
		panic(err)
	}

	// Load the inventory
	err = inventory.Load()
	if err != nil {
		panic(err)
	}

	// Process the inventory, given the target name
	data, err := inventory.Data(target, nil, false)
	if err != nil {
		panic(err)
	}

	log.Printf("\n%s", data.String())

	templateOutputPath := path.Join(outputPath, target)
	templater, err := skipper.NewTemplater(fileSystem, templatePath, templateOutputPath, nil)
	if err != nil {
		panic(err)
	}

	templateData := struct {
		Inventory  any
		TargetName string
	}{
		Inventory:  data,
		TargetName: target,
	}

	// execute templates  ----------------------------------------------------------------------------------
	components, err := inventory.GetComponents(target)
	if err != nil {
		panic(err)
	}

	err = templater.ExecuteComponents(templateData, components, false)
	if err != nil {
		panic(err)
	}

	// TODO: refactor, this is not nice to use
	// copy files as specified in the target config (base path is template root)
	t, err := inventory.Target(target)
	if err != nil {
		panic(err)
	}
	for _, copyFile := range t.SkipperConfig.Copies {
		source := filepath.Join(templatePath, copyFile.SourcePath)
		target := filepath.Join(templateOutputPath, copyFile.TargetPath)
		err := skipper.CopyFile(fileSystem, source, target)
		if err != nil {
			panic(err)
		}
	}
}
