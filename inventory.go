package skipper

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

// Inventory is the collection of classes and targets.
// The inventory wraps everything together and is capable of producing a single, coherent [Data]
// which can then be used inside the templates.
type Inventory struct {
	fs             afero.Fs
	fileExtensions []string
	classPath      string
	targetPath     string
	secretPath     string
	secretFiles    []*SecretFile
	classFiles     []*Class
	targetFiles    []*Target
}

// NewInventory creates a new Inventory with the given afero.Fs.
// At least one extension must be provided, otherwise an error is returned.
func NewInventory(fs afero.Fs, classPath, targetPath, secretPath string) (*Inventory, error) {
	if fs == nil {
		return nil, fmt.Errorf("fs cannot be nil")
	}
	if classPath == "" {
		return nil, fmt.Errorf("classPath cannot be empty")
	}
	if targetPath == "" {
		return nil, fmt.Errorf("targetPath cannot be empty")
	}
	if secretPath == "" {
		return nil, fmt.Errorf("secretPath cannot be empty")
	}

	if strings.EqualFold(classPath, targetPath) {
		return nil, fmt.Errorf("classPath cannot be the same as targetPath")
	}
	if strings.EqualFold(classPath, secretPath) {
		return nil, fmt.Errorf("classPath cannot be the same as secretPath")
	}
	if strings.EqualFold(targetPath, secretPath) {
		return nil, fmt.Errorf("targetPath cannot be the same as secretPath")
	}

	inv := &Inventory{
		fs:             fs,
		classPath:      classPath,
		targetPath:     targetPath,
		secretPath:     secretPath,
		fileExtensions: []string{".yml", ".yaml", ""},
	}

	return inv, nil
}

// Load will discover and load all classes and targets given the paths.
// It will also ensure that all targets only use classes which are actually defined.
func (inv *Inventory) Load() error {
	err := inv.loadClassFiles(inv.classPath)
	if err != nil {
		return fmt.Errorf("unable to load class files: %w", err)
	}

	err = inv.loadTargetFiles(inv.targetPath)
	if err != nil {
		return fmt.Errorf("unable to load target files: %w", err)
	}

	// check for all targets whether they use classes which actually exist
	for _, target := range inv.targetFiles {
		for _, class := range target.UsedClasses {
			if inv.GetClass(class) == nil {
				return fmt.Errorf("target '%s' uses class '%s' which does not exist", target.Name, class)
			}
		}
	}

	// load all secret files which exist in the inventory
	err = inv.loadSecretFiles(inv.secretPath)
	if err != nil {
		return fmt.Errorf("unable to load secret files: %w", err)
	}

	return nil
}

// GetUsedClasses returns the loaded classes which are used by the given target.
func (inv *Inventory) GetUsedClasses(targetName string) ([]*Class, error) {
	target := inv.GetTarget(targetName)
	if target == nil {
		return nil, fmt.Errorf("target could not be loaded: %s", targetName)
	}

	var classes []*Class
	for _, className := range target.UsedClasses {
		class := inv.GetClass(className)
		if class == nil {
			return nil, fmt.Errorf("class could not be loaded: %s", className)
		}
		classes = append(classes, class)
	}

	return classes, nil
}

// GetComponents returns the ComponentConfig of the given target and its used classes.
func (inv *Inventory) GetComponents(targetName string) ([]ComponentConfig, error) {
	target := inv.GetTarget(targetName)
	if target == nil {
		return nil, fmt.Errorf("target could not be loaded: %s", targetName)
	}

	var components []ComponentConfig

	if target.SkipperConfig.IsSet() {
		components = append(components, target.SkipperConfig.Components...)
	}

	usedClasses, err := inv.GetUsedClasses(targetName)
	if err != nil {
		return nil, err
	}

	for _, class := range usedClasses {
		if class.Configuration.IsSet() {
			components = append(components, class.Configuration.Components...)
		}
	}

	return components, nil
}

// GetCopyConfigs returns the CopyConfig of the given target and its used classes.
func (inv *Inventory) GetCopyConfigs(targetName string) ([]CopyConfig, error) {
	target := inv.GetTarget(targetName)
	if target == nil {
		return nil, fmt.Errorf("target could not be loaded: %s", targetName)
	}

	var copyConfigs []CopyConfig

	if target.SkipperConfig.IsSet() {
		copyConfigs = append(copyConfigs, target.SkipperConfig.Copies...)
	}

	usedClasses, err := inv.GetUsedClasses(targetName)
	if err != nil {
		return nil, err
	}

	for _, class := range usedClasses {
		if class.Configuration.IsSet() {
			copyConfigs = append(copyConfigs, class.Configuration.Copies...)
		}
	}

	return copyConfigs, nil
}

// Data loads the required inventory data map given the target.
// This is where variables and secrets are handled and eventually replaced.
// The resulting Data is what can be passed to the templates.
func (inv *Inventory) Data(targetName string, predefinedVariables map[string]interface{}, revealSecrets bool) (data Data, err error) {
	data = make(Data)

	target := inv.GetTarget(targetName)
	if target == nil {
		return nil, fmt.Errorf("target could not be loaded: %s", targetName)
	}

	// load all classes as defined by the target
	classes, err := inv.GetUsedClasses(targetName)
	if err != nil {
		return nil, err
	}

	// merge data from all classes into Data, preserving the class path.
	// A class with path "foo.bar.class" will be added like: Data["foo"]["bar"]["baz"] = classData
	{
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
	}

	// Merge target into Data, overwriting any existing values which were defined in classes because target data has precedence over class data.
	// Any key which is not added to the main Data (because the keys did not already exist), will be added.
	// Only exception is the 'skipper' key from the target, it will not be added to the data.
	targetData := target.Data()
	delete(targetData, skipperKey)
	data = data.MergeReplace(targetData)

	// add Skipper pre-defined variables
	if predefinedVariables == nil {
		predefinedVariables = make(map[string]interface{})
	}
	predefinedVariables["target_name"] = targetName

	// replace all ordinary variables (`${...}`) inside the data
	err = inv.replaceVariables(data, predefinedVariables)
	if err != nil {
		return nil, err
	}

	// call managment
	{
		calls, err := FindCalls(data)
		if err != nil {
			return nil, err
		}

		for _, call := range calls {
			log.Println("found call", call.FullName(), "at", call.Path(), "value", call.Execute())

			// replace call with function result
			// sourceValue is the value where the variable is. It needs to be replaced with an actual value
			sourceValue, err := data.GetPath(call.Identifier...)
			if err != nil {
				return nil, err
			}

			// Replace the full variable name (${variable}) with the targetValue
			sourceValue = strings.ReplaceAll(fmt.Sprint(sourceValue), call.FullName(), call.Execute())
			data.SetPath(sourceValue, call.Identifier...)
		}
	}

	// we need to reload the target configuration as it will derive it's configuration from the Data
	// of a previous state. Since the calls can modify the target configuration as well, we have to reload it.
	target.ReloadConfiguration()

	// secret management
	// initialize drivers, load or create secrets and eventually replace them if `revealSecrets` is true.
	{
		// fetch and configure secret drivers configured by the target
		for driverName, driverConfig := range target.Configuration.Secrets.Drivers {
			driver, err := SecretDriverFactory(driverName)
			if err != nil {
				return nil, fmt.Errorf("target contains invalid secret driver configuration: %w", err)
			}

			if drv, ok := driver.(ConfigurableSecretDriver); ok {
				if config, ok := driverConfig.(map[string]interface{}); ok {
					err = drv.Configure(config)
					if err != nil {
						return nil, fmt.Errorf("failed to configure driver '%s': %w", driverName, err)
					}
				} else {
					return nil, fmt.Errorf("driver configuration for '%s' is not map[string]interface{}", driverName)
				}
			}
		}

		// find all secrets or attempt to create them if an alternative action is set
		secrets, err := FindOrCreateSecrets(data, inv.secretFiles, inv.secretPath, inv.fs)
		if err != nil {
			return nil, err
		}

		// attempt load all secret files and replace the variables with the actual values if revealSecrets is true
		for _, secret := range secrets {
			log.Println("found secret", secret.FullName(), "at", secret.Path())

			if !secret.Exists(inv.fs) {
				return nil, fmt.Errorf("undefined secret '%s': file does not exist: %s", secret.FullName(), secret.SecretFile.Path)
			}

			err = secret.Load(inv.fs)
			if err != nil {
				return nil, fmt.Errorf("failed to load secret: %w", err)
			}

			// if the flag is true, all secret variables will be replaced by their actual value.
			// CAUTION: This is not something you want to do during local development, only inside your CI pipeline when the compiled output is ephemeral.
			if revealSecrets {
				err = inv.replaceSecret(data, secret)
				if err != nil {
					return nil, fmt.Errorf("failed to replace secret value: %w", err)
				}
			}
		}
	}

	return data, nil
}

// AddExternalClass can be used to dynamically create class files.
// The given data will be written into `classFilePath`, overwriting any existing file.
//
// The class path is first normalized to match the existing `Inventory.classPath`.
//
// After that, the root-key of the data is adjusted to match the fileName which is extracted from `classFilePath`.
// This has to be done in order to comply Skipper rules where the class-filename must also be the root-key of any given class.
//
// A new file inside the Skipper class path is created which makes it available for loading.
// In order to prevent confusion, a file header is added to indicate that the class was generated.
func (inv *Inventory) AddExternalClass(data map[string]any, classFilePath string) error {
	if data == nil {
		return fmt.Errorf("cannot add external class without data")
	}
	if classFilePath == "" {
		return fmt.Errorf("classFilePath cannot be empty")
	}

	// normalize classFilePath
	classFilePath = strings.TrimLeft(classFilePath, "./")
	if !strings.HasPrefix(classFilePath, inv.classPath) {
		classFilePath = filepath.Join(inv.classPath, classFilePath)
	}

	// adjust the root key to match the filename because this is what Skipper expects
	fileName := filepath.Base(classFilePath)
	rootKey := strings.TrimSuffix(fileName, filepath.Ext(fileName))

	// create new data and set the root key
	classData := make(Data)
	classData[rootKey] = data

	// warn the user that this class is generated and should not be edited manually
	classBytes := []byte("---\n# This is a dynamically generated class file. DO NOT EDIT!\n")
	classBytes = append(classBytes, classData.Bytes()...)

	// write the class into the inventory filesystem
	classFile, err := CreateNewYamlFile(inv.fs, classFilePath, classBytes)
	if err != nil {
		return err
	}

	newClass, err := NewClass(classFile, classFilePath)
	if err != nil {
		return err
	}

	inv.classFiles = append(inv.classFiles, newClass)

	return nil
}

// replaceSecret will replace the given secret inside Data.
func (inv *Inventory) replaceSecret(data Data, secret *Secret) error {
	// sourceValue is the value where the variable is. It needs to be replaced with an actual value
	sourceValue, err := data.GetPath(secret.Identifier...)
	if err != nil {
		return err
	}

	// Replace the full variable name (${variable}) with the actual secret value which will be fetched by the underlying driver.
	secretValue, err := secret.Value()
	if err != nil {
		return err
	}

	sourceValue = strings.ReplaceAll(fmt.Sprint(sourceValue), secret.FullName(), secretValue)
	data.SetPath(sourceValue, secret.Identifier...)

	return nil
}

// replaceVariables iterates over the given Data map and replaces all variables with the required value.
// TODO enable custom variable definition inside classes
func (inv *Inventory) replaceVariables(data Data, predefinedVariables map[string]interface{}) (err error) {

	// Determine which variables exist in the Data map
	variables, err := FindVariables(data)
	if err != nil {
		return err
	}

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
		var targetValue interface{}
		if isPredefinedVariable(variable) {
			targetValue = predefinedVariables[variable.Name]
		} else {
			// targetValue is the value on which the variable points to.
			// This is the value we need to replace the variable with
			targetValue, err = data.GetPath(variable.NameAsIdentifier()...)
			if err != nil {

				// for any other error than a 'key not found' there is nothing we can do
				if !strings.Contains(err.Error(), "key not found") {
					return fmt.Errorf("reference to invalid variable '%s': %w", variable.FullName(), err)
				}

				// Local variable handling
				//
				// at this point we have failed to resolve the variable using 'absolute' paths
				// but the variable may be only locally defined which means we need to change the lookup path.
				// We iterate over all classes and attempt to resolve the variable within that limited scope.
				for i, class := range inv.classFiles {

					// if the value to which the variable points is valid inside the class scope, we just need to add the class identifier
					// if the combination works this means we have found ourselves a local variable and we can set the targetValue
					fullPath := []interface{}{}
					fullPath = append(fullPath, class.NameAsIdentifier()...)
					fullPath = append(fullPath, variable.NameAsIdentifier()...)

					targetValue, err = data.GetPath(fullPath...)

					// as long as not all classes have been checked, we cannot be sure that the variable is undefined (aka. key not found error)
					if targetValue == nil &&
						i < len(inv.classFiles) &&
						strings.Contains(err.Error(), "key not found") {
						continue
					}

					// the local variable is really not defined at this point
					if err != nil {
						return fmt.Errorf("reference to invalid variable '%s': %w", variable.FullName(), err)
					}

					break
				}
			}
		}

		// sourceValue is the value where the variable is located. It needs to be replaced with the 'targetValue'
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

// GetTarget attempts to return a target struct given a target name.
// If the target could not be found, nil is returned.
func (inv *Inventory) GetTarget(name string) *Target {
	for _, target := range inv.targetFiles {
		if strings.ToLower(name) == strings.ToLower(target.Name) {
			return target
		}
	}
	return nil
}

// GetClass attempts to return a Class, given a name.
// If the class does not exist, nil is returned.
func (inv *Inventory) GetClass(name string) *Class {
	for _, class := range inv.classFiles {
		if class.Name == name {
			return class
		}
	}
	return nil
}

// loadClassFiles
func (inv *Inventory) loadClassFiles(classPath string) error {
	classFiles, err := DiscoverYamlFiles(inv.fs, classPath)
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

func (inv *Inventory) loadSecretFiles(secretPath string) error {
	secretFiles, err := DiscoverYamlFiles(inv.fs, secretPath)
	if err != nil {
		return err
	}

	// load all secret files
	for _, secret := range secretFiles {
		err = secret.Load(inv.fs)
		if err != nil {
			return err
		}

		// skip empty files
		if len(secret.Data) == 0 {
			continue
		}

		relativePath := strings.ReplaceAll(secret.Path, secretPath, "")
		relativePath = strings.TrimLeft(relativePath, "/")

		c, err := NewSecretFile(secret, relativePath)
		if err != nil {
			return fmt.Errorf("%s: %w", secret.Path, err)
		}
		inv.secretFiles = append(inv.secretFiles, c)
	}
	return nil
}

// loadTargetFiles
// MUST be called after loadClassFiles as it depends on existing classes to handle wildcard imports
func (inv *Inventory) loadTargetFiles(targetPath string) error {
	targetFiles, err := DiscoverYamlFiles(inv.fs, targetPath)
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

		for _, use := range t.UsedWildcardClasses {
			for _, class := range inv.classFiles {

				usePrefix := strings.TrimRight(use, "*")

				if strings.HasPrefix(class.Name, usePrefix) {
					t.UsedClasses = append(t.UsedClasses, class.Name)
				}

			}
		}

		inv.targetFiles = append(inv.targetFiles, t)
	}

	return nil
}
