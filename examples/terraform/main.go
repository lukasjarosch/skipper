package main

import (
	"flag"
	"log"
	"path"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"

	"github.com/lukasjarosch/skipper"
)

var (
	inventoryPath string
	templatePath  string
	outputPath    string
	target        string

	targetPath string
	classPath  string

	fileSystem = afero.NewOsFs()
)

func init() {
	flag.StringVar(&inventoryPath, "data", "inventory", "path to the inventory folder")
	flag.StringVar(&templatePath, "templates", "templates", "path to the templates folder")
	flag.StringVar(&outputPath, "output", "output", "template output path")
	flag.StringVar(&target, "target", "", "name of the target to use")
	flag.Parse()

	targetPath = path.Join(inventoryPath, "targets")
	classPath = path.Join(inventoryPath, "classes")
}

func main() {
	if target == "" {
		log.Fatalln("target cannot be empty")
	}
	log.Printf("inventory path set to '%s'", inventoryPath)
	log.Printf("template path set to '%s'", templatePath)
	log.Printf("compiled path set to '%s'", outputPath)
	log.Printf("desired target is '%s'", target)

	// initialize and load inventory ----------------------------------------------------------------------------------
	inventory, err := skipper.NewInventory(afero.NewOsFs())
	if err != nil {
		panic(err)
	}

	err = inventory.Load(classPath, targetPath)
	if err != nil {
		panic(err)
	}

	// discover, load and parse the templates ----------------------------------------------------------------------------------
	myTemplateFuncs := map[string]any{
		"foo": func() string {
			return "foo-bar-baz"
		},
	}

	templateOutputPath := path.Join(outputPath, target)
	templater, err := skipper.NewTemplater(fileSystem, templatePath, templateOutputPath, myTemplateFuncs)
	if err != nil {
		panic(err)
	}

	for _, template := range templater.Files {
		log.Println("discovered template", template.Path)
	}

	// render inventory data based on target ----------------------------------------------------------------------------------
	predefinedVariables := map[string]interface{}{
		"target_name": target,
		"output_path": outputPath,
	}

	data, err := inventory.Data(target, predefinedVariables)
	if err != nil {
		panic(err)
	}

	out, _ := yaml.Marshal(data)
	log.Printf("\n%s", string(out))

	// pretend that we've got some other data source
	additional := map[string]any{
		"something": "else",
	}

	templateData := struct {
		Inventory  any
		Additional any
	}{
		Inventory:  data,
		Additional: additional,
	}

	// execute templates  ----------------------------------------------------------------------------------
	for _, template := range templater.Files {
		err := templater.Execute(template, templateData, false)
		if err != nil {
			panic(err)
		}
		log.Printf("executed template '%s' into: %s'", template.Path, path.Join(templateOutputPath, template.Path))
	}
	// Alternatively `templater.ExecuteAll(template, templateData)` can be used
}
