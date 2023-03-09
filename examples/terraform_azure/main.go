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
	templatePath  = "templates"
	secretPath    = path.Join(inventoryPath, "secrets")
	outputPath    = "compiled"

	target = "develop"
)

func main() {
	inventory, err := skipper.NewInventory(fileSystem, classPath, targetPath, secretPath)
	if err != nil {
		panic(err)
	}

	predefinedVariables := map[string]interface{}{
		"target_name": target,
		"output_path": outputPath,
	}

	// Process the inventory, given the target name
	data, err := inventory.Data("develop", predefinedVariables, true, false)
	if err != nil {
		panic(err)
	}

	{
	}

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

	err = templater.ExecuteAll(templateData, false, nil)
	if err != nil {
		panic(err)
	}

}
