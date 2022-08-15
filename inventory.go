package skipper

import (
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/spf13/afero"
)

// Inventory is the collection of data files.
type Inventory struct {
	fs             afero.Fs
	fileExtensions []string
	classFiles     []*Class
	targetFiles    []*Target
}

// NewInventory creates a new Inventory with the given afero.Fs.
// At least one extension must be provided, otherwise an error is returned.
func NewInventory(fs afero.Fs) (*Inventory, error) {
	if fs == nil {
		return nil, fmt.Errorf("fs cannot be nil")
	}

	inv := &Inventory{
		fs:             fs,
		fileExtensions: []string{".yml", ".yaml"},
	}

	return inv, nil
}

// Load will discover and load all classes and targets given the paths.
// It will also ensure that all targets only use classes which are actually defined.
func (inv *Inventory) Load(classPath, targetPath string) error {
	err := inv.loadClassFiles(classPath)
	if err != nil {
		return fmt.Errorf("unable to load class files: %w", err)
	}

	err = inv.loadTargetFiles(targetPath)
	if err != nil {
		return fmt.Errorf("unable to load target files: %w", err)
	}

	// check for all targets whether they use classes which actually exist
	for _, target := range inv.targetFiles {
		for _, class := range target.UsedClasses {
			if !inv.ClassExists(class) {
				return fmt.Errorf("target '%s' uses class '%s' which does not exist", target.Name, class)
			}
		}
	}

	return nil
}

// Data loads the required inventory data map given the target.
func (inv *Inventory) Data(targetName string, predefinedVariables map[string]interface{}) (data Data, err error) {
	data = make(Data)

	target, err := inv.Target(targetName)
	if err != nil {
		return nil, err
	}

	// load all classes as defined by the target
	var classes []*Class
	for _, className := range target.UsedClasses {
		class, err := inv.Class(className)
		if err != nil {
			return nil, err
		}
		classes = append(classes, class)
	}

	// ensure that the loaded class-data does not conflict
	// If two classes with the same root-key are selected, we cannot continue.
	// We could attempt to perform a 'smart' merge or apply some precendende rules, but
	// this will inevitably cause unexpected behaviour which is not what we want.
	for _, class := range classes {

		// If the class name has multiple segments (foo.bar.baz), we will need to
		// add the keys do Data, so that Data[foo][bar][baz] is where the data of the class will be added.
		classSegments := strings.Split(class.Name, ".")
		if len(classSegments) > 1 {
			tmp := data

			for _, segment := range classSegments {

				if !tmp.HasKey(segment) {
					tmp[segment] = make(Data)
				}

				// as long as the current segment is not the RootKey, shift tmp by the segment
				if segment != class.RootKey() {
					tmp = tmp[segment].(Data)
					continue
				}

				// add class data to RootKey. Since we're here, RootKey==segment, hence we can add it here.
				if class.Data().Get(class.RootKey()) == nil {
					continue
				}
				tmp[class.RootKey()] = class.Data().Get(class.RootKey())

			}
		} else {
			// class does not have a dot separator, hence we just check if the RootKey exists and add the data
			if _, exists := data[class.RootKey()]; exists {
				return nil, fmt.Errorf("duplicate key '%s' registered by class '%s'", class.RootKey(), class.Name)
			}
			data[class.RootKey()] = class.Data().Get(class.RootKey())
		}

	}

	// next we need to determine which keys are present in the target which are also defined by the classes
	// these keys need to be merged into the existing data, eventually overwriting values since the target always has precendende over classes.
	dataKeys := reflect.ValueOf(data).MapKeys()            // we know that it's a map so we skip some checks
	targetKeys := reflect.ValueOf(target.Data()).MapKeys() // we know that it's a map so we skip some checks

	targetData := target.Data()   // copy target data since we're going to delete keys and like to preserve the original
	targetMergeData := make(Data) // target data which needs to be merged into the main data

	// copy existing keys in target data into targetMergeData and remove the key from targetData.
	for _, dataKey := range dataKeys {
		for _, targetKey := range targetKeys {
			if dataKey.String() == targetKey.String() {
				targetMergeData[targetKey.String()] = targetData[targetKey.String()]
				delete(targetData, targetKey.String())
				break
			}
		}
	}
	data = data.MergeReplace(targetMergeData)

	// add 'leftover' keys from the target under the 'target' key
	// TODO: what if a class defines the 'target' key?
	data[targetKey] = targetData

	err = inv.replaceVariables(data, predefinedVariables)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// replaceVariables iterates over the given Data map and replaces all variables with the required value.
// TODO enable custom variable definition inside classes
func (inv *Inventory) replaceVariables(data Data, predefinedVariables map[string]interface{}) (err error) {

	// Determine which variables exist in the Data map
	variables := FindVariables(data)

	if len(variables) == 0 {
		return nil
	}

	isPredefinedVariable := func(variable Variable) bool {
		for name := range predefinedVariables {
			if strings.EqualFold(variable.Name, name) {
				return true
			}
		}
		return false
	}

	for _, variable := range variables {

		log.Println("replacing variable", variable.FullName(), "at", variable.Identifier)

		var targetValue interface{}

		if isPredefinedVariable(variable) {
			targetValue = predefinedVariables[variable.Name]
		} else {
			// targetValue is the value on which the variable points to.
			// This is the value we need to replace the variable with
			targetValue, err = data.GetPath(variable.NameAsIdentifier()...)
			if err != nil {
				return err
			}
		}

		// sourceValue is the value where the variable is. It needs to be replaced with an actual value
		sourceValue, err := data.GetPath(variable.Identifier...)
		if err != nil {
			return err
		}

		// Replace the full variable name (${variable}) with the targetValue
		sourceValue = strings.ReplaceAll(fmt.Sprint(sourceValue), variable.FullName(), fmt.Sprint(targetValue))
		data.SetPath(sourceValue, variable.Identifier...)
	}

	return nil
}

// Target returns a target given a name.
func (inv *Inventory) Target(name string) (*Target, error) {
	if !inv.TargetExists(name) {
		return nil, fmt.Errorf("target '%s' does not exist", name)
	}

	return inv.getTarget(name), nil
}

// TargetExists returns true if the given target name exists
func (inv *Inventory) TargetExists(name string) bool {
	if inv.getTarget(name) == nil {
		return false
	}
	return true
}

// getTarget attempts to return a target struct given a target name
func (inv *Inventory) getTarget(name string) *Target {
	for _, target := range inv.targetFiles {
		if strings.ToLower(name) == strings.ToLower(target.Name) {
			return target
		}
	}
	return nil
}

// Class attempts to return a Class, given a name.
// If the class does not exist, an error is returned.
func (inv *Inventory) Class(name string) (*Class, error) {
	if !inv.ClassExists(name) {
		return nil, fmt.Errorf("class '%s' does not exist", name)
	}
	return inv.getClass(name), nil
}

// ClassExists returns true if a class with the given name exists.
func (inv *Inventory) ClassExists(name string) bool {
	if inv.getClass(name) == nil {
		return false
	}
	return true
}

// getClass attempts to return a Class, given a name.
// If the class does not exist, nil is returned.
func (inv *Inventory) getClass(name string) *Class {
	for _, class := range inv.classFiles {
		if class.Name == name {
			return class
		}
	}
	return nil

}

// discoverFiles iterates over a given rootPath recursively, filters out all files with the appropriate file fileExtensions
// and finally creates a YamlFile slice which is then returned.
func (inv *Inventory) discoverFiles(rootPath string) ([]*YamlFile, error) {
	exists, err := afero.Exists(inv.fs, rootPath)
	if err != nil {
		return nil, fmt.Errorf("check if path exists: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("file path does not exist: %s", rootPath)
	}

	var files []*YamlFile
	err = afero.Walk(inv.fs, rootPath, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if inv.matchesExtension(path) {
			file, err := NewFile(path)
			if err != nil {
				return err
			}
			files = append(files, file)
		}
		return nil
	})
	return files, err
}

// loadClassFiles
func (inv *Inventory) loadClassFiles(classPath string) error {
	classFiles, err := inv.discoverFiles(classPath)
	if err != nil {
		return err
	}

	// load all class files, replacing the inventory-relative path with dot-separated style
	for _, class := range classFiles {
		err = class.Load(inv.fs)
		if err != nil {
			return err
		}

		// skip empty files
		if len(class.Data) == 0 {
			continue
		}

		relativePath := strings.ReplaceAll(class.Path, classPath, "")
		relativePath = strings.TrimLeft(relativePath, "/")

		c, err := NewClass(class, relativePath)
		if err != nil {
			return fmt.Errorf("%s: %w", class.Path, err)
		}
		inv.classFiles = append(inv.classFiles, c)
	}
	return nil
}

// loadTargetFiles
// MUST be called after loadClassFiles as it depends on existing classes to handle wildcard imports
func (inv *Inventory) loadTargetFiles(targetPath string) error {
	targetFiles, err := inv.discoverFiles(targetPath)
	if err != nil {
		return err
	}

	for _, target := range targetFiles {
		err = target.Load(inv.fs)
		if err != nil {
			return err
		}

		relativePath := strings.ReplaceAll(target.Path, targetPath, "")
		relativePath = strings.TrimLeft(relativePath, "/")

		t, err := NewTarget(target, relativePath)
		if err != nil {
			return fmt.Errorf("%s: %w", target.Path, err)
		}

		if len(t.UsedWildcardClasses) > 0 {
			for _, use := range t.UsedWildcardClasses {
				for _, class := range inv.classFiles {

					usePrefix := strings.TrimRight(use, "*")

					if strings.HasPrefix(class.Name, usePrefix) {
						log.Println("ADD", class.Name)
						t.UsedClasses = append(t.UsedClasses, class.Name)
					}

				}
			}
		}

		inv.targetFiles = append(inv.targetFiles, t)
	}

	return nil
}

// matchesExtension returns true if the given string has a valid extension as defined in `Inventory.fileExtensions`
func (inv *Inventory) matchesExtension(path string) bool {
	ext := filepath.Ext(path)
	for _, extension := range inv.fileExtensions {
		if extension == ext {
			return true
		}
	}
	return false
}
