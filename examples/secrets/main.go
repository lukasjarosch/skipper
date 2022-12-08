package main

import (
	"bytes"
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

	target = "azure_keyvault"
)

func main() {
	inventory, err := skipper.NewInventory(fileSystem, classPath, targetPath, secretPath)
	if err != nil {
		panic(err)
	}

	// Process the inventory, given the target name
	data, err := inventory.Data(target, nil, true)
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

	skipperConfig, err := inventory.GetSkipperConfig(target)
	if err != nil {
		panic(err)
	}

	driver, err := skipper.SecretDriverFactory("azurekv")
	if err != nil {
		log.Fatalf("cannot get secret driver %q: %w", "azurekv", err)
	}

	source := bytes.NewBuffer([]byte("Hallo Welt, das hab ich ganz alleine verschlüsselt"))
	sink := bytes.NewBuffer([]byte{})
	err = driver.(skipper.SecretFileEncrypter).EncryptFile(source, sink)
	if err != nil {
		panic(err)
	}

	// execute templates  ----------------------------------------------------------------------------------
	err = templater.ExecuteComponents(templateData, skipperConfig.Components, false)
	if err != nil {
		panic(err)
	}

	// copy files as specified in the target config (base path is template root)
	err = skipper.CopyFilesByConfig(fileSystem, skipperConfig.Copies, templatePath, templateOutputPath)
	if err != nil {
		panic(err)
	}
}
