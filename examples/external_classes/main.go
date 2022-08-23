package main

import (
	"fmt"
	"io/fs"
	"log"
	"path"
	"path/filepath"
	"strings"

	"github.com/lukasjarosch/skipper"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

type NetworkConfiguration struct {
	Name    string `mapstructure:"name"`
	IpRange string `mapstructure:"ip_range"`
}

var fileSystem = afero.NewOsFs()

func main() {
	networkConfigs, err := loadNetworkConfigurations("networks", "network")
	if err != nil {
		panic(err)
	}

	// ...
	// Perform validation, business logic, ... on the network configurations
	// ...

	// ======================================================================================================

	// Now we'll initialize skipper
	inventory, err := skipper.NewInventory(fileSystem, "skipper/inventory/classes", "skipper/inventory/targets")
	if err != nil {
		panic(err)
	}

	// BEFORE we load the inventory of skipper, we can inject our own class data
	// Skipper will then dynamically create these files in the inventory and make them available.
	for fileName, networkConfig := range networkConfigs {
		data, err := skipper.NewData(networkConfig)
		if err != nil {
			panic(err)
		}
		inventory.AddExternalClass(data, path.Join("network", fileName))
	}

	// Load the inventory
	err = inventory.Load()
	if err != nil {
		panic(err)
	}

	// Process the inventory, given the target name
	data, err := inventory.Data("develop", nil)
	if err != nil {
		panic(err)
	}

	log.Printf("\n%s", data.String())
}

func loadNetworkConfigurations(path, rootKey string) (map[string]NetworkConfiguration, error) {
	networkConfigurations := map[string]NetworkConfiguration{}

	externalDataFs := afero.NewBasePathFs(fileSystem, path)
	err := afero.Walk(externalDataFs, "/", func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		if filepath.Ext(path) == ".yml" || filepath.Ext(path) == ".yaml" {
			var tmp NetworkConfiguration
			loadConfigFile(externalDataFs, path, rootKey, &tmp)

			raw := make(map[string]interface{})
			loadConfigFile(externalDataFs, path, rootKey, &raw)

			path = strings.TrimLeft(path, "/")
			networkConfigurations[path] = tmp
			log.Println("loaded network configuration:", path)
		}

		return err
	})

	if err != nil {
		return nil, err
	}

	return networkConfigurations, nil
}

func loadConfigFile(fs afero.Fs, filePath, key string, target interface{}) error {
	v := viper.New()
	v.SetFs(fs)
	v.SetConfigFile(filePath)
	if err := v.ReadInConfig(); err != nil {
		return err
	}

	if v.Get(key) == nil {
		return fmt.Errorf("key does not exist: %s", key)
	}

	if err := v.UnmarshalKey(key, target); err != nil {
		return err
	}

	return nil
}
