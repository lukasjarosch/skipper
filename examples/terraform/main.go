package main

import (
	"flag"
	"fmt"
	"log"
	"path"
	"time"

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
	secretPath string

	fileSystem = afero.NewOsFs()
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
}

func main() {
	log.Printf("inventory path set to '%s'", inventoryPath)
	log.Printf("template path set to '%s'", templatePath)
	log.Printf("compiled path set to '%s'", outputPath)
	log.Printf("desired target is '%s'", target)

	// initialize and load inventory ----------------------------------------------------------------------------------
	inventory, err := skipper.NewInventory(afero.NewOsFs(), classPath, targetPath, secretPath)
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
	templater, err := skipper.NewTemplater(fileSystem, templatePath, templateOutputPath, myTemplateFuncs, []string{})
	if err != nil {
		panic(err)
	}

	for _, template := range templater.Files {
		log.Println("discovered template", template.Path())
	}

	// render inventory data based on target ----------------------------------------------------------------------------------
	predefinedVariables := map[string]interface{}{
		"output_path":  outputPath,
		"company_name": "AcmeCorp International",
		"year":         time.Now().Year(),
	}

	data, err := inventory.Data(target, predefinedVariables, false, false)
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

	cfg, err := inventory.GetSkipperConfig(target)
	if err != nil {
		panic(err)
	}
	fmt.Println(cfg)
	fmt.Println(cfg.Renames)
	// execute templates  ----------------------------------------------------------------------------------
	for _, template := range templater.Files {
		err := templater.Execute(template, templateData, false, cfg.Renames)
		if err != nil {
			panic(err)
		}
		log.Printf("executed template '%s'", template.Path)
	}
	// Alternatively `templater.ExecuteAll(template, templateData)` can be used
}
