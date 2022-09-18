package main

import (
	"log"
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

	target = "develop"
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

	predefinedVariables := map[string]interface{}{
		"target_name": target,
		"output_path": outputPath,
	}

	// Process the inventory, given the target name
	data, err := inventory.Data("develop", predefinedVariables)
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
	for _, template := range templater.Files {
		err := templater.Execute(template, templateData, false)
		if err != nil {
			panic(err)
		}
		log.Printf("executed template '%s' into: %s'", template.Path, path.Join(templateOutputPath, template.Path))
	}
}
