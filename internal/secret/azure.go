package secret

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azkeys"
)

type Azure struct {
	*driver
	client       *azkeys.Client
	loadedKey    string
	vaultBaseUrl string
	keyName      string
}

// TODO: Secret version support (azure keys are versioned)

func NewAzure() (*Azure, error) {
	return &Azure{driver: &driver{initialized: false}}, nil
}

func (driver *Azure) Initialize(config map[string]interface{}) error {

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return err
	}

	vaultName, exists := config["vault_name"]
	if !exists {
		return fmt.Errorf("missing 'vault_name' key for %s configuration", driver.Type())
	}
	if vaultName == "" || vaultName == "UNDEFINED" {
		return fmt.Errorf("%s: configuration 'vault_name' cannot be empty", driver.Type())
	}

	driver.vaultBaseUrl = fmt.Sprintf("https://%s.vault.azure.net", vaultName)
	driver.client = azkeys.NewClient(driver.vaultBaseUrl, cred, nil)
	driver.initialized = true

	return nil
}

func (driver *Azure) SetKey(key string) error {
	if !driver.initialized {
		return fmt.Errorf("%s: %w", driver.Type(), ErrDriverNotInitialized)
	}

	if key == "" {
		return ErrDriverKeyEmpty
	}

	driver.key = key

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

	res, err := driver.client.Encrypt(context.TODO(), driver.key, "", encryptParams, nil)
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
	res, err := driver.client.Decrypt(context.TODO(), driver.key, "", encryptParams, nil)

	return string(res.Result), nil
}

func (driver *Azure) Type() string {
	return "azurekv"
}
