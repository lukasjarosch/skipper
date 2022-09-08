package skipper

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

// Class represents a single file containing a YAML struct which makes up the inventory.
type Class struct {
	// File is the underlying file in the filesystem.
	File *YamlFile
	// Name is the relative path of the file inside the inventory which uniquely identifies this class.
	// Because the name is path based, no two classes with the same name can exist.
	// For the name, the path-separator is replaced with '.' and the file extension is stripped.
	// Example: 'something/foo/bar.yaml' will have the name 'something.foo.bar'
	//
	// The Name is also what is used to reference classes throughout Skpper.
	Name string
	// Configuration holds Skipper-relevant configuration inside the class
	Configuration *SkipperConfig
}

// NewClass will create a new class, given a raw YamlFile and the relative filePath from inside the inventory.
// If your class file is at `foo/bar/inventory/classes/myClass.yaml`, the relativeClassPath will be `myClass.yaml`
func NewClass(file *YamlFile, relativeClassPath string) (*Class, error) {
	if file == nil {
		return nil, fmt.Errorf("file cannot be nil")
	}
	if relativeClassPath == "" {
		return nil, fmt.Errorf("relativeClassPath cannot be empty")
	}

	name := classNameFromPath(relativeClassPath)

	// class file cannot be empty, there must be exactly one yaml root-key which must define a map
	val := reflect.ValueOf(file.Data)
	if val.Kind() != reflect.Map {
		return nil, fmt.Errorf("class '%s' root key does not define a map", name)
	}
	if len(val.MapKeys()) == 0 {
		return nil, fmt.Errorf("class '%s' does not have a root-key", name)
	}
	if len(val.MapKeys()) > 1 {
		return nil, fmt.Errorf("class '%s' has more than one root-key which is currently not supported", name)
	}

	class := &Class{
		File: file,
		Name: name,
	}

	// load skipper config
	config, err := LoadSkipperConfig(file, class.RootKey())
	if err != nil {
		return nil, err
	}
	class.Configuration = config

	return class, nil
}

// Data returns the underlying class file-data map as Data
func (c *Class) Data() *Data {
	return &c.File.Data
}

// RootKey returns the root key name of the class.
func (c *Class) RootKey() string {
	val := reflect.ValueOf(c.Data()).Elem()
	return val.MapKeys()[0].String()
}

func (c *Class) NameAsIdentifier() (id []interface{}) {
	tmp := strings.Split(c.Name, ".")
	id = make([]interface{}, len(tmp))

	for i := 0; i < len(tmp); i++ {
		id[i] = tmp[i]
	}
	return id
}

// classNameFromPath returns the class name given a path.
// The path must be relative to the root class path.
// If the class-root is at: /foo/bar/classes/
// And the target class is: /foo/bar/classes/something/class.yaml
// Then the path must be: something/class.yaml
//
// The resulting class name would be: something.class
func classNameFromPath(path string) string {
	path = strings.Trim(path, string(os.PathSeparator))
	pathNoExt := strings.TrimSuffix(path, filepath.Ext(path))
	return strings.ReplaceAll(pathNoExt, "/", ".")
}
