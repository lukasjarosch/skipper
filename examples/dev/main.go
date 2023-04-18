package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/dominikbraun/graph"
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
	resolver := skipper.NewClassResolver()
	for _, class := range classes {
		err := resolver.RegisterClass(class)
		if err != nil {
			log.Fatal(err)
		}
	}

	// add dependencies to namespaces
	// needs to be done after all classes are registered otherwise the included classes might be unknown
	for _, class := range classes {
		for _, includedClass := range class.Includes() {
			err = resolver.DependsOn(class.Namespace, skipper.P(includedClass))
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	// Graph visualization
	{
		err = skipper.VisualizeGraph(resolver.Graph, "/tmp/test.gv", "before graph reduction")
		if err != nil {
			log.Fatalln(err)
		}

		reduced, _ := graph.TransitiveReduction(resolver.Graph)

		err = skipper.VisualizeGraph(reduced, "/tmp/reduced.gv", "after graph reduction")
		if err != nil {
			log.Fatalln(err)
		}
	}

	// sort the graph topologically to get the execution order
	order, err := resolver.ReduceAndSort()
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
		// 4. If two classes `A` and `B` both include `C` and also set `C.value`, then
		//    as soon as *both* classes are included together somewhere else, an error must occur.
		//    If class `D` includes `A` and `B`, the path `C.value` is not properly defined anymore.
		// 5. PATHS NEVER CHANGE.
		//    If A includes B, A can use `A.foo` and not `B.A.Foo`. Maybe there is reason to allow both, but
		//    the path to the original data (in this case B) does never change.
		//    Changing the paths introduces difficulties if you think about environments for example.
		//    Say you have target `dev` and `prod`, `dev` includes `dev.common` and `prod` includes `prod.common`
		//    Both `dev.common` and `prod.common` include `myresource`. If we allow chaning paths
		//    then you'd have too access `myresource` differently in dev (`dev.myresource.foo`) and in prod (`prod.myresource.foo`).
		//    This will inevitably cause mayhem in the templates or any other processing down the line.
		//
		// ===[ Notes ]====
		//
		//  - Figure out a way to determine the exact path on how a path has gotten it's value.
		//    This corresponds to rule 4. As long as the path of a value is straight (aka. each node only has one outgoing edge),
		//    then the path is also properly defined. Otherwise not.
		//
		// ===[ Workflow ]====
		//

		reduced, _ := graph.TransitiveReduction(resolver.Graph)

		adjm, err := reduced.AdjacencyMap()
		if err != nil {
			log.Fatalln(err)
		}

		// Find all leaf nodes starting from the current class.
		// the leaf nodes are the destination nodes to which we need to find all possible paths.
		var leafs []string
		graph.DFS(reduced, class.Namespace.String(), func(val string) bool {
			if len(adjm[val]) == 0 {
				log.Printf("-> '%s' is transitively including '%s'", class.Namespace, val)
				leafs = append(leafs, val)
			}
			return false
		})

		// For every leaf we need to find all possible paths from the current node.
		// If there are more than one paths leading from the current node to the
		// target leaf, there might be undefined values.
		// Given two paths from A to C:
		//    A -> B -> C
		//    A -> D -> C
		// Both B and D make C available in A because includes are transitive (see rule 2).
		// This is fine initially, but if B and D both modify the same path (say `C.foo`),
		// then the value for `C.foo` will be undefined in A.
		// NOTE: maybe 'write' access should only be granted to classes which actually 'use' the class they write to

		for _, leaf := range leafs {
			dependencyPaths := resolver.AllPaths(class.Namespace.String(), leaf)

			log.Printf("--> found %d paths from '%s' to '%s'", len(dependencyPaths), class.Namespace, leaf)
			if len(dependencyPaths) > 1 {

				// classWrites maps paths written to a class namespace to indicate which class writes which path.
				// Note that only paths of included classes are part of this map.
				// As soon as a class dependency attempts to register an existing key, there is a write conflict within the dependency graph.
				classWrites := make(map[string]skipper.Path)

				// function which returns whether testClass performs WRITE access on any path from sourceClass
				// all written paths are returned as absolute paths.
				// Absolute paths means that if the actual write happens in `my.class.anotherclass.value`,
				// the paths returned absolute for the included class `anotherclass.value`.
				writeAccess := func(testClass, sourceClass *skipper.Class) (writtenPaths []skipper.Path) {

					// if testClass writes to sourceClass, this means that testClass must have a path
					// which looks like: `[testClass.RootKey].[sourceClass.Namespace]`
					writePathPrefix := strings.Join([]string{testClass.RootKey, sourceClass.Namespace.String()}, skipper.PathSeparator)

					testClassPaths := skipper.ListAllPaths(testClass.Data, "")
					for _, p := range testClassPaths {
						if strings.HasPrefix(p.String(), writePathPrefix) {
							absPath := p[len(fmt.Sprintf("%s", testClass.RootKey)):] // Note: we can use len() instead of len()-1 to include the '.' after the rootkey
							writtenPaths = append(writtenPaths, absPath)
						}
					}

					return writtenPaths
				}

				getWritesOnDependencyPath := func(pathToInvestigate []string, leafClass *skipper.Class) (writes map[string]skipper.Path) {
					// FIXME: The path `targets.dev -> common -> test.d -> test.b -> test.c` is also found and evaluated
					// But it is not a problem if `common` overwrites any path in `test.c` as the 'path-split' occurs right after and hence
					// `common` would just overwrite any value anyway.

					writes = make(map[string]skipper.Path)

					for i := len(pathToInvestigate) - 1; i >= 0; i-- {
						currentClass := getClass(skipper.P(pathToInvestigate[i]))
						writtenPaths := writeAccess(currentClass, leafClass)
						if writtenPaths == nil {
							continue
						}

						for _, writtenPath := range writtenPaths {
							writes[writtenPath.String()] = currentClass.Namespace
						}

						// paths := skipper.ListAllPaths(currentClass.Data, "")
						// for _, path := range paths {
						//
						// 	path = path[1:] // remove the first key because that is the classes root key
						// 	// log.Printf("WRITE ACCESS in '%s' on path '%s': '%s'", currentClass.Namespace, path, leafClass.Namespace)
						//
						// 	// NOTE: can you add a new path in an included class?
						//
						// 	if strings.HasPrefix(path.String(), leafClass.Namespace.String()) {
						// 		log.Printf("WRITE ACCESS: class '%s' writes '%s'", currentClass.Namespace, path)
						// 	}
						// }
					}
					return writes
				}

				for _, depPath := range dependencyPaths {
					log.Printf("---> path: %s", strings.Join(depPath, " -> "))

					// Walk eath path from the back (from the leaf) and note each modification they perform (if they do).
					// The first and last path segments can be skipped because source and destination nodes don't matter in this case.
					// Source: Includes the path and may even overwrite stuff, but that does not affect how the value is defined downstream.
					// Destination (leaf): does not have any dependencies and defines the initial data.
					pathToInvestigate := depPath[1 : len(depPath)-1]

					leafClass := getClass(skipper.P(depPath[len(depPath)-1]))

					writtenPaths := getWritesOnDependencyPath(pathToInvestigate, leafClass)

					for wp, classNs := range writtenPaths {
						if c, exists := classWrites[wp]; exists {
							log.Fatalf("resolve error: multi-write: '%s' and '%s' write path '%s' on different dependency paths", c, classNs, wp)
						}
						classWrites[wp] = classNs
					}
				}
			}
		}

		// TODO: if any leaf node has more than one path leading from the current class to it, it is a candidate for further investigation
		// TODO: figure out if the paths (say: the classes along it), modify the leaf node
		// TODO: if they modify the leaf node (destination class), then figure out which paths they modify
		// TODO: if two classes modify the same value (same path), then we have a non-resolvable path.

		// classDeps := adjm[class.Namespace.String()]
		// for _, dep := range classDeps {
		// 	log.Println("-->", dep.Source, "-->", dep.Target)
		// }
		//
		// log.Printf("-> fetch %d included classes: %s", len(class.Includes()), strings.Join(class.Includes(), ","))
		// for _, include := range class.Includes() {
		// includedClass := getClass(skipper.P(include))
		// log.Printf("--> loaded: %s", includedClass.Namespace)

		// err = class.LoadIncludedClass(includedClass)
		// if err != nil {
		// 	log.Fatalln(err)
		// }

		// log.Println("class:", class.Namespace, class.Data.(skipper.Data).Pretty())
		// log.Println("include:", includedClass.Namespace, includedClass.Data.(skipper.Data).Pretty())
		//
		// targetPath := strings.Join([]string{class.RootKey, includedClass.Namespace.String()}, skipper.PathSeparator)
		// log.Println("target path in class", class.Namespace, ":", targetPath)
		//
		// tmp, _ := skipper.NewData(class.Data.(skipper.Data)[class.RootKey])
		// tree := tmp
		//
		// for _, segment := range includedClass.Namespace {
		// 	tree[segment] = skipper.Data{}
		// 	tree = tree[segment].(skipper.Data)
		// }
		// tmp[class.RootKey] = tree
		// log.Println("tmp2", class.Namespace, tmp.Pretty())
		//
		// newData, err := tmp.GetPath(skipper.P(class.RootKey))
		// if err != nil {
		// 	log.Fatalln(err)
		// }
		// log.Println("newdata", class.Namespace, newData.(skipper.Data).Pretty())
		// }
		log.Printf("-> finished")
		log.Println()
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
