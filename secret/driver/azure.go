package driver

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
	ErrInvalidKeyUri = fmt.Errorf("invalid vault key id")
)

type Azure struct {
	client       *azkeys.Client
	loadedKey    string
	vaultBaseUrl string
	config       *azureConfig
}

type azureConfig struct {
	// IgnoreVersion will ignore any key version, even if given, and always use the latest version.
	IgnoreVersion bool `mapstructure:"ignore_version"`
	// The Azure Vault KeyId to use for encryption and decryption
	KeyId string `mapstructure:"key_id"`

	VaultName  string
	KeyName    string
	KeyVersion string
}

func NewAzure() (*Azure, error) {
	return &Azure{config: &azureConfig{}}, nil
}

func (driver *Azure) Configure(config map[string]interface{}) error {
	err := mapstructure.Decode(config, driver.config)
	if err != nil {
		return fmt.Errorf("failed to decode azure driver config: %w", err)
	}

	if len(driver.config.KeyId) == 0 {
		return fmt.Errorf("azure key id cannot be empty")
	}

	if !isAzureKeyId(driver.config.KeyId) {
		return fmt.Errorf("malformed azure key id")
	}

	driver.config.VaultName, driver.config.KeyName, driver.config.KeyVersion, err = parseAzureKeyVaultKeyId(driver.config.KeyId)
	if err != nil {
		return fmt.Errorf("failed to parse azure key id: %w", err)
	}

	driver.vaultBaseUrl = fmt.Sprintf("https://%s.vault.azure.net", driver.config.VaultName)

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return err
	}
	driver.client, err = azkeys.NewClient(driver.vaultBaseUrl, cred, nil)
	if err != nil {
		return err
	}

	return nil
}

func (driver *Azure) Encrypt(input string) (string, error) {
	encryptParams := azkeys.KeyOperationsParameters{
		Algorithm: to.Ptr(azkeys.JSONWebKeyEncryptionAlgorithmRSAOAEP256),
		Value:     []byte(input),
	}

	version := driver.config.KeyVersion
	if driver.config.IgnoreVersion {
		version = ""
	}

	// execute a get first because 'encrypt' produces weird errors if one does not
	// have access to the vault/key. At least `GetKey` produces a somewhat more usable error like a 401.
	_, err := driver.client.GetKey(context.TODO(), driver.config.KeyName, version, nil)
	if err != nil {
		return "", err
	}

	res, err := driver.client.Encrypt(context.TODO(), driver.config.KeyName, version, encryptParams, nil)
	if err != nil {
		return "", err
	}

	return base64.RawStdEncoding.EncodeToString(res.Result), nil
}

func (driver *Azure) Decrypt(input string) (string, error) {

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

	// execute a get first because 'decrypt' produces weird errors if one does not
	// have access to the vault/key. At least `GetKey` produces a somewhat more usable error like a 401.
	_, err = driver.client.GetKey(context.TODO(), driver.config.KeyName, version, nil)
	if err != nil {
		return "", err
	}

	res, err := driver.client.Decrypt(context.TODO(), driver.config.KeyName, version, encryptParams, nil)
	if err != nil {
		return "", err
	}

	return string(res.Result), nil
}

func isAzureKeyId(key string) bool {
	_, err := url.ParseRequestURI(key)
	if err != nil {
		return false
	}
	return true
}

func parseAzureKeyVaultKeyId(key string) (vaultName string, keyName string, keyVersion string, err error) {
	keyId, err := url.ParseRequestURI(key)
	if err != nil {
		return "", "", "", err
	}

	hostParts := strings.Split(keyId.Hostname(), ".")
	// the hostname part must have 4 segments (vaultname.vault.azure.net)
	if len(hostParts) < 4 {
		return "", "", "", ErrInvalidKeyUri
	}
	// the second segment must be 'vault'
	if hostParts[1] != "vault" {
		return "", "", "", ErrInvalidKeyUri
	}
	vaultName = hostParts[0]

	pathParts := strings.Split(strings.Trim(keyId.Path, "/"), "/")
	// The path must have at least two parts (keys/keyname/version)
	if len(pathParts) < 2 {
		return "", "", "", ErrInvalidKeyUri
	}
	if pathParts[0] != "keys" {
		return "", "", "", ErrInvalidKeyUri
	}
	keyName = pathParts[1]
	if pathParts[2] != "" {
		keyVersion = pathParts[2]
	}

	return vaultName, keyName, keyVersion, nil
}

func (driver *Azure) Type() string {
	return "azurekv"
}
