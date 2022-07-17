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

	"github.com/lukasjarosch/skipper"
)

var (
	dataPath     string
	target       string
	templatePath string
	outputPath   string

	fileSystem = afero.NewOsFs()
)

func init() {
	flag.StringVar(&dataPath, "data", "inventory", "path to the data folder")
	flag.StringVar(&target, "target", "", "name of the target to use")
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
	if outputPath == "" {
		log.Fatalln("outputPath cannot be empty")
	}

	log.Printf("inventory path set to '%s'", dataPath)
	log.Printf("template path set to '%s'", templatePath)
	log.Printf("compiled path set to '%s'", outputPath)
	log.Printf("desired target is '%s'", target)

	// load inventory ----------------------------------------------------------------------------------

	inventory, err := skipper.NewInventory(afero.NewOsFs())
	if err != nil {
		// TODO: handle error instead of panicking
		panic(err)
	}

	// TODO: make configurable
	targetPath := path.Join(dataPath, "targets")
	classPath := path.Join(dataPath, "classes")
	log.Printf("target path: '%s'", targetPath)
	log.Printf("classes path: '%s'", classPath)

	err = inventory.Load(classPath, targetPath)
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

	// fetch target file ----------------------------------------------------------------------------------
	targetFile, err := inventory.Target(target)
	if err != nil {
		log.Fatalln(err)
	}

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

		templateData := struct {
			Inventory skipper.Data
			Target    skipper.Data
		}{
			Inventory: inventory.Data,
			Target:    targetFile.Data(), // TODO: ensure targets always use top-level-key 'target'
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
