package skipper

import (
	"fmt"
	"io"
	"strings"

	driver "github.com/lukasjarosch/skipper/internal/secret"
)

type SecretDriver interface {
	Type() string
	Encrypt(data string) (string, error)
	Decrypt(encrypted string) (string, error)
}

type SecretFileEncrypter interface {
	SecretDriver
	EncryptFile(source io.Reader, sink io.Writer) error
}

type ConfigurableSecretDriver interface {
	SecretDriver
	Configure(config map[string]interface{}) error
}

var driverCache = map[string]SecretDriver{}

func SecretDriverFactory(name string) (secretDriver SecretDriver, err error) {
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
