package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"

	"github.com/lukasjarosch/skipper"
	"github.com/lukasjarosch/skipper/codec"
	"github.com/lukasjarosch/skipper/data"
	"github.com/mitchellh/mapstructure"
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

	// REFACTOR ZONE ===================================================
	{

		classFiles, err := skipper.DiscoverFiles(classPath, []string{".yml", ".yaml"})
		if err != nil {
			log.Fatal(fmt.Errorf("error discovering class files: %w", err))
		}

		var classContainer []*data.Container
		for _, file := range classFiles {
			class, err := data.NewContainer(file, codec.NewYamlCodec())
			if err != nil {
				log.Fatal(fmt.Errorf("cannot create data container from '%s': %w", file.Path(), err))
			}
			classContainer = append(classContainer, class)
			log.Info("created container from file", "name", class.Name)
		}

		inventory, err := data.NewInventory()
		if err != nil {
			log.Fatal(fmt.Errorf("failed to create inventory: %w", err))
		}

		// register class containers
		for _, container := range classContainer {

			// The namespace of a container is calculated by the path of the underlying file.
			// The classPath is removed as well as the filename.
			// This is then used to create a [data.Path] which is the actual namespace of the container.
			ns := container.File.Path()
			ns = strings.Replace(ns, classPath, "", 1)
			ns = strings.Trim(ns, string(os.PathSeparator))
			ns = filepath.Dir(ns)
			ns = strings.Trim(ns, ".")
			namespace := data.NewPathFromOsPath(ns)

			err := inventory.RegisterContainer(namespace, container)
			if err != nil {
				log.Fatalf("failed to register container: %s", err)
			}
			log.Info("registered container", "namespace", namespace, "name", container.Name)
		}

		_ = inventory

	}
	return
	// =================================================================

	inventory, err := skipper.NewInventory(fileSystem, classPath, targetPath, secretPath)
	if err != nil {
		panic(err)
	}

	predefinedVariables := map[string]interface{}{
		"target_name": target,
		"output_path": outputPath,
	}

	// Process the inventory, given the target name
	reveal := false
	data, err := inventory.Data(target, predefinedVariables, false, reveal)
	if err != nil {
		panic(err)
	}

	{

		type AutoGenerated struct {
			ProjectName string   `yaml:"project_name" mapstructure:"project_name"`
			Test        []string `yaml:"test" mapstructure:"test"`
			Var         string   `yaml:"var_iable" mapstructure:"var"`
		}
		out, err := data.GetPath("common")
		if err != nil {
			panic(err)
		}

		ag := AutoGenerated{}
		err = mapstructure.Decode(out, &ag)
		if err != nil {
			panic(fmt.Errorf("failed to decode struct: %w", err))
		}
		log.Printf("%#v", ag)

		data.SetPath(ag, "common", "transformed")
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

	err = templater.ExecuteAll(templateData, false, nil)
	if err != nil {
		panic(err)
	}

}
