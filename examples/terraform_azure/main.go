package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/davecgh/go-spew/spew"

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

func Path(s string) data.Path {
	return data.NewPath(s)
}

func main() {

	// REFACTOR ZONE ===================================================
	{
		log.SetLevel(log.DebugLevel)

		classFiles, err := skipper.DiscoverFiles(classPath, []string{".yml", ".yaml"})
		if err != nil {
			log.Fatal(fmt.Errorf("error discovering class files: %w", err))
		}
		targetFiles, err := skipper.DiscoverFiles(targetPath, []string{".yml", ".yaml"})
		if err != nil {
			log.Fatal(fmt.Errorf("error discovering target files: %w", err))
		}

		allFiles := classFiles
		allFiles = append(allFiles, targetFiles...)

		var fileContainerList []*data.FileContainer
		for _, file := range allFiles {
			container, err := data.NewFileContainer(file, codec.NewYamlCodec())
			if err != nil {
				log.Fatal(fmt.Errorf("cannot create data container from '%s': %w", file.Path(), err))
			}
			fileContainerList = append(fileContainerList, container)
			log.Info("created container from file", "name", container.Name(), "file", file.Path())
		}

		findContainerByName := func(name string) data.Container {
			for _, container := range fileContainerList {
				if container.Name() == name {
					return container
				}
			}
			return nil
		}

		target := findContainerByName("develop")
		tfIdentifiers := findContainerByName("identifiers")

		spew.Dump(target.MustGet(Path("develop.terraform.identifiers.*")).Map())
		spew.Dump(tfIdentifiers.MustGet(Path("identifiers.*")).Map())

		mergeData, err := target.MustGet(Path("develop.terraform.identifiers")).Map()
		if err != nil {
			log.Fatal(err)
		}
		log.Warn(tfIdentifiers.Merge(Path("identifiers.*"), mergeData))
		spew.Dump(tfIdentifiers.MustGet(Path("identifiers.*")).Map())
		tfIdentifiers.Set(Path("identifiers.vnet"), data.NewValue("GEILES VNET"))
		spew.Dump(tfIdentifiers.MustGet(Path("identifiers.")).Map())

		return

		inventory, err := data.NewInventory()
		if err != nil {
			log.Fatal(fmt.Errorf("failed to create inventory: %w", err))
		}

		// register class containers
		for _, container := range fileContainerList {

			// The namespace of a container is calculated by the path of the underlying file.
			// The classPath is removed as well as the filename.
			// This is then used to create a [data.Path] which is the actual namespace of the container.
			ns := container.File.Path()
			ns = strings.Replace(ns, inventoryPath, "", 1)
			ns = strings.Trim(ns, string(os.PathSeparator))
			ns = filepath.Dir(ns)
			ns = strings.Trim(ns, ".")
			namespace := data.NewPathFromOsPath(ns)

			err := inventory.RegisterContainer(namespace, container)
			if err != nil {
				log.Fatalf("failed to register container: %s", err)
			}
			log.Info("registered container", "namespace", namespace, "name", container.Name())
		}

		// classInventory := inventory.Scoped(Path("classes"))

		// p := Path("components.documentation.skipper.components.0.output_path")
		// val, err := inventory.GetValue(p)
		// if err != nil {
		// 	log.Fatalf("cannot resolve path %s: %s", p, err)
		// }
		// log.Warnf("%s: %s", p, val)
		// spew.Dump(val.Scope.Container.Get(Path("*")))

		// overwrite container tests
		{
			// test := struct {
			// 	Foo data.Map `yaml:"common"`
			// }{
			// 	Foo: data.Map{
			// 		"bar": "baz",
			// 	},
			// }
			//
			// container, err := data.NewRawContainer("common", test, codec.NewYamlCodec())
			// if err != nil {
			// 	log.Fatal(err)
			// }
			//
			// namespace := Path("azure")
			// err = inventory.RegisterContainer(namespace, container, data.ReplaceContainer())
			// if err != nil {
			// 	log.Fatal(err)
			// }
			// log.Info("registered container", "namespace", namespace, "name", container.Name(), "replace", true)
			// spew.Dump(inventory.GetValue(Path("azure.common.bar")))
		}

		type Dummy struct {
			Data data.Map `yaml:"data"`
		}

		// container patching
		{
			originalData := Dummy{
				Data: data.Map{
					"foo": data.Map{
						"bar": data.Map{
							"baz": "qux",
							"old": "hello",
						},
					},
				},
			}
			patchedData := Dummy{
				Data: data.Map{
					"foo": data.Map{
						"bar": data.Map{
							"baz": "CHANGED",
							"new": "ADDED",
						},
					},
				},
			}

			originalContainer, _ := data.NewRawContainer("data", originalData, codec.NewYamlCodec())
			patchedContainer, _ := data.NewRawContainer("data", patchedData, codec.NewYamlCodec())

			// register original (to be patched) container
			err = inventory.RegisterContainer(Path("patch"), originalContainer)
			if err != nil {
				log.Error(err)
			}
			log.Warn(inventory.MustGetValue(Path("patch.data.foo.bar.baz")))

			// patch the original container with the patchedContainer
			err = inventory.RegisterContainer(Path("patch"), patchedContainer, data.Patch())
			if err != nil {
				log.Error(err)
			}
			log.Warn(inventory.MustGetValue(Path("patch.data.foo.bar.baz")))
			log.Warn(inventory.MustGetValue(Path("patch.data.foo.bar.new")))
			log.Warn(inventory.MustGetValue(Path("patch.data.foo.bar.old")))

			// TODO: now patch the container with a different (different name, different namespace) container

		}

		targetPath := data.NewPathVar("targets", "develop")
		classPath := Path("classes")
		targetPaths := inventory.RegisteredPrefixedPaths(targetPath)

		for _, p := range targetPaths {

			if p.First() == "skipper" {
				continue
			}

			sourceValue, _ := inventory.GetValue(classPath.AppendPath(p))
			log.Debug("attempt to set value from target",
				"path", classPath.AppendPath(p),
				"targetValue", inventory.MustGetValue(targetPath.AppendPath(p)),
				"sourceValue", sourceValue,
			)
			err = inventory.SetValue(classPath.AppendPath(p), inventory.MustGetValue(targetPath.AppendPath(p)).Raw)
			if err != nil {
				log.Error(fmt.Errorf("failed to set value: %w", err))
				continue
			}

			log.Info("set value from target",
				"path", classPath.AppendPath(p),
				"value", inventory.MustGetValue(targetPath.AppendPath(p)),
			)
		}

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
