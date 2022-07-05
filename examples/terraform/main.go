package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"

	"github.com/lukasjarosch/templater"
)

var (
	dataPath     string
	target       string
	targetPath   string
	templatePath string
	outputPath   string

	fileSystem = afero.NewOsFs()
)

func init() {
	flag.StringVar(&dataPath, "data", "inventory", "path to the data folder")
	flag.StringVar(&target, "target", "", "name of the target to use")
	flag.StringVar(&targetPath, "targetPath", "targets", "path to the targets directory")
	flag.StringVar(&templatePath, "templates", "templates", "path to the templates folder")
	flag.StringVar(&outputPath, "output", "output", "template output path")
	flag.Parse()
}

func main() {
	if dataPath == "" {
		log.Fatalln("data path cannot be empty")
	}
	if templatePath == "" {
		log.Fatalln("templates path cannot be empty")
	}
	if target == "" {
		log.Fatalln("target cannot be empty")
	}
	if targetPath == "" {
		log.Fatalln("targetPath cannot be empty")
	}
	if outputPath == "" {
		log.Fatalln("outputPath cannot be empty")
	}

	log.Printf("dataPath set to '%s'", dataPath)
	log.Printf("templatePath set to '%s'", templatePath)
	log.Printf("target set to '%s'", target)
	log.Printf("targetPath set to '%s'", targetPath)
	log.Printf("outputPath set to '%s'", outputPath)

	// load data inventory ----------------------------------------------------------------------------------

	afero.Walk(fileSystem, dataPath, func(path string, info fs.FileInfo, err error) error {
		log.Println(path)
		return nil
	})

	inventory, err := templater.NewInventory(afero.NewOsFs())
	if err != nil {
		// TODO: handle error instead of panicking
		panic(err)
	}

	err = inventory.Load(dataPath)
	if err != nil {
		// TODO: handle error instead of panicking
		panic(err)
	}

	// discover templates ----------------------------------------------------------------------------------

	// iterate over templatePath, where we set the filesystem root to templatePath
	// this allows us to only get relative paths to the template file
	var templateFiles []string
	err = afero.Walk(afero.NewBasePathFs(fileSystem, templatePath), "", func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		templateFiles = append(templateFiles, path)
		return nil
	})
	if err != nil {
		// TODO: handle error instead of panicking
		panic(err)
	}

	// load target file ----------------------------------------------------------------------------------

	exists, err := afero.Exists(fileSystem, targetPath)
	if err != nil {
		panic(fmt.Errorf("failed to check target path: %w", err))
	}
	if !exists {
		panic(fmt.Errorf("data path does not exist: %s", targetPath))
	}

	var targetFilePath string
	err = afero.Walk(fileSystem, targetPath, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		baseName := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		if baseName == target {
			targetFilePath = path
			return nil
		}

		return nil
	})
	if err != nil {
		// TODO: handle error instead of panicking
		panic(err)
	}

	if targetFilePath == "" {
		panic(fmt.Errorf("target '%s' not found in '%s'", target, targetPath))
	}

	targetFile, err := templater.NewFile(targetFilePath)
	if err != nil {
		// TODO: handle error instead of panicking
		panic(err)
	}
	err = targetFile.Load(fileSystem)
	if err != nil {
		// TODO: handle error instead of panicking
		panic(err)
	}

	log.Printf("%#v", targetFile.Data)

	// TODO: load target data and map to data into _target

	// render output  ----------------------------------------------------------------------------------

	// render templates
	for _, templateFile := range templateFiles {
		templateFilePath := path.Join(templatePath, templateFile)

		tpl := template.New(filepath.Base(templateFile)).Funcs(sprig.TxtFuncMap()).Funcs(map[string]any{
			"tfStringArray": func(input []interface{}) string {
				var s []string
				for _, v := range input {
					s = append(s, "\""+fmt.Sprintf(v.(string))+"\"")
				}
				return strings.Join(s, ", ")
			},
		})

		tplContent, err := afero.ReadFile(fileSystem, templateFilePath)
		if err != nil {
			// TODO: handle error instead of panicking
			panic(err)
		}

		tpl, err = tpl.Parse(string(tplContent))
		if err != nil {
			// TODO: handle error instead of panicking
			panic(err)
		}

		// create the file output path: <output>/<target>/<template>
		fileOutputPath := path.Join(outputPath, target, templateFile)

		err = fileSystem.MkdirAll(filepath.Dir(fileOutputPath), 0755)
		if err != nil {
			// TODO: handle error instead of panicking
			panic(err)
		}

		outFile, err := fileSystem.Create(fileOutputPath)
		if err != nil {
			// TODO: handle error instead of panicking
			panic(err)
		}

		// ensure '.Target.name' is always set to the file-basename of the target
		targetFile.Data["target"].(templater.Data)["name"] = target

		templateData := struct {
			Inventory templater.Data
			Target    templater.Data
		}{
			Inventory: inventory.Data,
			Target:    targetFile.Data["target"].(templater.Data), // TODO: ensure targets always use top-level-key 'target'
		}

		err = tpl.Execute(outFile, templateData)
		if err != nil {
			// TODO: handle error instead of panicking
			panic(err)
		}
	}

	s, _ := yaml.Marshal(inventory.Data)
	fmt.Println(string(s))

	log.Printf("%+v", inventory.Data)
}
