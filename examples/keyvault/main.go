package main

import (
	"flag"
	"log"
	"path"

	"github.com/lukasjarosch/skipper"
	"github.com/spf13/afero"
)

var (
	fileSystem = afero.NewOsFs()

	inventoryPath string
	templatePath  string
	outputPath    string

	targetPath string
	classPath  string
	secretPath string
	target     string
)

func init() {
	flag.StringVar(&inventoryPath, "data", "inventory", "path to the inventory folder")
	flag.StringVar(&templatePath, "templates", "templates", "path to the templates folder")
	flag.StringVar(&outputPath, "output", "compiled", "template output path")
	flag.StringVar(&target, "target", "dev", "name of the target to use")
	flag.Parse()

	targetPath = path.Join(inventoryPath, "targets")
	classPath = path.Join(inventoryPath, "classes")
	secretPath = path.Join(inventoryPath, "secrets")

	log.Printf("inventory path set to '%s'", inventoryPath)
	log.Printf("template path set to '%s'", templatePath)
	log.Printf("compiled path set to '%s'", outputPath)
	log.Printf("desired target is '%s'", target)
}

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
	data, err := inventory.Data("develop", predefinedVariables, false, true)
	if err != nil {
		panic(err)
	}

	templateOutputPath := path.Join(outputPath, target)
	templater, err := skipper.NewTemplater(fileSystem, templatePath, templateOutputPath, nil, nil)
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

	// {
	// 	components, err := inventory.GetComponents(target)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	//
	// 	err = templater.ExecuteComponents(templateData, components, false)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// }
	for _, template := range templater.Files {
		err := templater.Execute(template, templateData, false, nil)
		if err != nil {
			panic(err)
		}
		log.Printf("executed template '%s' into: %s'", template.Path, path.Join(templateOutputPath, template.Path()))
	}
}
