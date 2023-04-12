package main

import (
	"log"
	"path"

	"github.com/lukasjarosch/skipper"
	"github.com/spf13/afero"
)

type Test struct {
	Foo   string `yaml:"foo"`
	Hello Hello  `yaml:"hello"`
}

type Hello struct {
	World string `yaml:"world"`
}

func main() {
	classPath := "classes"
	fs := afero.NewOsFs()

	file, err := skipper.LoadYamlFile(path.Join(classPath, "common.yaml"), fs)
	handleErr(err)

	class, err := skipper.NewClass(skipper.P("common"), skipper.Data(file.Data))
	handleErr(err)

	if foo, exists := class.Get(skipper.P("common.foo")); exists {
		log.Println("FOO:", foo)
	}

	var t Test
	err = class.Data.UnmarshalPath(skipper.P("common"), &t)
	handleErr(err)
	log.Println(t)
}

func handleErr(err error) {
	if err != nil {
		panic(err)
	}
}
