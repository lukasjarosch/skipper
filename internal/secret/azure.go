package secret

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azkeys"
	"github.com/mitchellh/mapstructure"
)

var (
	ErrEmptyVaultName = fmt.Errorf("vault_name is required if only the key name is set")
	ErrInvalidKeyUri  = fmt.Errorf("invalid vault key id")
)

type Azure struct {
	*driver
	client       *azkeys.Client
	loadedKey    string
	vaultBaseUrl string
	config       *azureConfig
}

type azureConfig struct {
	// IgnoreVersion will ignore any key version, even if given, and always use the latest version.
	IgnoreVersion bool `mapstructure:"ignore_version"`
	// VaultName is required if only the key name is passed.
	// The vault name is then required to build the base url.
	VaultName string `mapstructure:"vault_name"`
	// Version of the key to use. Will be ignored if IgnoreVersion is true.
	KeyVersion string `mapstructure:"key_version"`
}

// TODO: allow using only the KeyID for configuration as it contains all required information
// TODO: allow using the VaultURI instead of only the keyvault name?
// TODO: Secret version support (azure keys are versioned) (or just allow usage of the KeyID as Key name)

func NewAzure() (*Azure, error) {
	return &Azure{driver: &driver{initialized: false}}, nil
}

func (driver *Azure) Initialize(config map[string]interface{}) error {

	var driverConfig azureConfig
	err := mapstructure.Decode(config, &driverConfig)
	if err != nil {
		return fmt.Errorf("failed to decode azure driver config: %w", err)
	}
	driver.config = &driverConfig
	driver.initialized = true

	return nil
}

func (driver *Azure) SetKey(key string) error {
	if !driver.initialized {
		// there is no need to initialize the driver with custom configuration
		driver.config = &azureConfig{}
		driver.initialized = true
	}

	// at this point, there are two options on how the driver can be configured
	// 1. The key is a KeyID (URI), then the vault name, key name and key version are extracted from it (version is optional)
	// 2. The key is just a string. In that case the key is just the name and the driver must be configured with the VaultName and optionally a version in order to build the base url

	isKeyId := func(key string) bool {
		_, err := url.ParseRequestURI(key)
		if err != nil {
			return false
		}
		return true
	}

	// TODO: move keyId parsing to own func
	if isKeyId(key) {
		keyId, err := url.ParseRequestURI(key)
		if err != nil {
			return err
		}

		hostParts := strings.Split(keyId.Hostname(), ".")
		// the hostname part must have 4 segments (vaultname.vault.azure.net)
		if len(hostParts) < 4 {
			return ErrInvalidKeyUri
		}
		// the second segment must be 'vault'
		if hostParts[1] != "vault" {
			return ErrInvalidKeyUri
		}
		driver.config.VaultName = hostParts[0]

		pathParts := strings.Split(strings.Trim(keyId.Path, "/"), "/")
		// The path must have at least two parts (keys/keyname/version)
		if len(pathParts) < 2 {
			return ErrInvalidKeyUri
		}
		if pathParts[0] != "keys" {
			return ErrInvalidKeyUri
		}
		if pathParts[2] != "" {
			driver.config.KeyVersion = pathParts[2]
		}
		driver.key = pathParts[1]
	} else {
		if driver.config.VaultName == "" {
			return ErrEmptyVaultName
		}
		if key == "" {
			return ErrDriverKeyEmpty
		}
		driver.key = key
	}

	driver.vaultBaseUrl = fmt.Sprintf("https://%s.vault.azure.net", driver.config.VaultName)

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return err
	}
	driver.client = azkeys.NewClient(driver.vaultBaseUrl, cred, nil)

	// Check if the key exists by attempting to fetch it
	version := driver.config.KeyVersion
	if driver.config.IgnoreVersion {
		version = ""
	}
	_, err = driver.client.GetKey(context.TODO(), driver.key, version, &azkeys.GetKeyOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (driver *Azure) Encrypt(input string) (string, error) {
	if !driver.initialized {
		return "", fmt.Errorf("%s: %w", driver.Type(), ErrDriverNotInitialized)
	}
	if !driver.isKeySet() {
		return "", fmt.Errorf("%s: %w", driver.Type(), ErrDriverKeyEmpty)
	}
	encryptParams := azkeys.KeyOperationsParameters{
		Algorithm: to.Ptr(azkeys.JSONWebKeyEncryptionAlgorithmRSAOAEP256),
		Value:     []byte(input),
	}

	version := driver.config.KeyVersion
	if driver.config.IgnoreVersion {
		version = ""
	}
	res, err := driver.client.Encrypt(context.TODO(), driver.key, version, encryptParams, nil)
	if err != nil {
		return "", err
	}

	return base64.RawStdEncoding.EncodeToString(res.Result), nil
}

func (driver *Azure) Decrypt(input string) (string, error) {
	if !driver.initialized {
		return "", fmt.Errorf("%s: %w", driver.Type(), ErrDriverNotInitialized)
	}
	if !driver.isKeySet() {
		return "", fmt.Errorf("%s: %w", driver.Type(), ErrDriverKeyEmpty)
	}

	decoded, err := base64.RawStdEncoding.DecodeString(input)
	if err != nil {
		return "", err
	}

	encryptParams := azkeys.KeyOperationsParameters{
		Algorithm: to.Ptr(azkeys.JSONWebKeyEncryptionAlgorithmRSAOAEP256),
		Value:     []byte(decoded),
	}

	version := driver.config.KeyVersion
	if driver.config.IgnoreVersion {
		version = ""
	}
	res, err := driver.client.Decrypt(context.TODO(), driver.key, version, encryptParams, nil)
	if err != nil {
		return "", err
	}

	return string(res.Result), nil
}

func (driver *Azure) Type() string {
	return "azurekv"
}
