package main

import (
	"log"
	"strings"

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

	// discover and load all yaml files in [classPath]
	classYamlFiles, err := skipper.DiscoverYamlFiles(fs, classPath)
	handleErr(err)

	// load classes from the loaded yaml files
	classes := make([]*skipper.Class, 0)
	for _, yamlFile := range classYamlFiles {
		namespace := skipper.FilePathToPath(yamlFile.Path, classPath)
		class, err := skipper.NewClass(namespace, skipper.Data(yamlFile.Data))
		if err != nil {
			log.Printf("cannot load class '%s': %s", namespace, err)
			continue
		}

		log.Printf("loaded class: %s", class.Namespace.String())
		classes = append(classes, class)
	}

	getClass := func(namespace skipper.Path) *skipper.Class {
		for _, class := range classes {
			if class.Namespace.String() == namespace.String() {
				return class
			}
		}
		return nil
	}

	// make sure that the class includes reference existing classes
	for _, class := range classes {
		for _, include := range class.Includes() {
			if getClass(skipper.P(include)) == nil {
				log.Fatalf("class '%s' includes non existing namespace '%s'", class.Namespace, include)
			}
		}
	}

	// create resolver and register class namespaces
	resolver := skipper.NewResolver()
	for _, class := range classes {
		err := resolver.RegisterPath(class.Namespace)
		if err != nil {
			log.Fatal(err)
		}
	}

	// add dependencies to namespaces
	for _, class := range classes {
		for _, includedClass := range class.Includes() {
			err = resolver.DependsOn(class.Namespace, skipper.P(includedClass))
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	// sort the graph topologically to get the execution order
	order, err := resolver.TopologicalSort()
	if err != nil {
		log.Fatal(err)
	}
	for i, path := range order {
		log.Printf("Step %d: get class '%s'", i, path)
		class := getClass(path)

		if len(class.Includes()) == 0 {
			log.Printf("-> finished because class has no dependencies")
			continue
		}

		// ===[ Inclusion rules ]====
		//
		// 1. The data of included classes will never be modified directly to preserve the original.
		//    This ensures that another class can still include the original data set.
		// 	  Otherwise class `foo` could modify `bar` by including it. Once `baz` attempts to include `bar` as well,
		//    the data might've already changed
		// 2. Includes are transitive.
		//    If `A` includes `B`, and `B` includes `C`, then `A` includes `C` as well.
		// 3. Modifications are passed 'up' the dependency graph.
		//    If class `A` is included and modified by `B` which is then again included by `C`,
		// 	  `C` will see the modifications made by `B`.
		//    In order to see the original data, `B` has to include `A` directly.
		//
		// ===[ Workflow ]====
		//

		log.Printf("-> fetch %d included classes: %s", len(class.Includes()), strings.Join(class.Includes(), ","))
		for _, include := range class.Includes() {
			includedClass := getClass(skipper.P(include))
			log.Printf("--> loaded: %s", includedClass.Namespace)

			log.Println("class:", class.Namespace, class.Data.(skipper.Data).Pretty())
			log.Println("include:", includedClass.Namespace, includedClass.Data.(skipper.Data).Pretty())

			targetPath := strings.Join([]string{class.RootKey, includedClass.Namespace.String()}, skipper.PathSeparator)
			log.Println("target path in class", class.Namespace, ":", targetPath)

			tmp, _ := skipper.NewData(class.Data.(skipper.Data)[class.RootKey])
			tree := tmp

			for _, segment := range includedClass.Namespace {
				tree[segment] = skipper.Data{}
				tree = tree[segment].(skipper.Data)
			}
			tmp[class.RootKey] = tree
			log.Println("tmp2", class.Namespace, tmp.Pretty())

			newData, err := tmp.GetPath(skipper.P(class.RootKey))
			if err != nil {
				log.Fatalln(err)
			}
			log.Println("newdata", class.Namespace, newData.(skipper.Data).Pretty())
		}
		log.Printf("-> finished")
		break
	}

	// _, err = skipper.NewInventory(classes)
	// handleErr(err)

	// sandbox := getClass(skipper.P("environments.sandbox"))
	// common := getClass(skipper.P("azure.common"))
	// err = sandbox.ResolveInclude(common)
	// if err != nil {
	// 	handleErr(fmt.Errorf("class '%s': %w", sandbox.Namespace, err))
	// }

	//
	// file, err := skipper.LoadYamlFile(path.Join(classPath, "common.yaml"), fs)
	// handleErr(err)
	//
	// class, err := skipper.NewClass(skipper.P("common"), skipper.Data(file.Data))
	// handleErr(err)
	//
	// if foo, exists := class.Get(skipper.P("common.foo")); exists {
	// 	log.Println("FOO:", foo)
	// }
	//
	// var t Test
	// err = class.Data.UnmarshalPath(skipper.P("common"), &t)
	// handleErr(err)
	// log.Println(t)
}

func handleErr(err error) {
	if err != nil {
		panic(err)
	}
}
