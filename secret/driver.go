package secret

import (
	"fmt"
	"strings"

	"github.com/lukasjarosch/skipper/secret/driver"
)

type Driver interface {
	Type() string
	Encrypt(data string) (string, error)
	Decrypt(encrypted string) (string, error)
}

type ConfigurableDriver interface {
	Driver
	Configure(config map[string]interface{}) error
}

var driverCache = map[string]Driver{}

func SecretDriverFactory(name string) (secretDriver Driver, err error) {
	name = strings.ToLower(name)

	// return a cached version of the driver if there is one
	if secretDriver, cached := driverCache[name]; cached {
		return secretDriver, nil
	}

	// create new driver and cache it
	switch name {
	case "plain":
		secretDriver, err = driver.NewPlain()
	case "base64":
		secretDriver, err = driver.NewBase64()
	case "aes":
		secretDriver, err = driver.NewAes()
	case "azurekv":
		secretDriver, err = driver.NewAzure()
	default:
		return nil, fmt.Errorf("driver '%s' cannot be loaded: not implemented", name)
	}

	driverCache[name] = secretDriver

	return secretDriver, nil
}
