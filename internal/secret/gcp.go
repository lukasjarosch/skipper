package secret

import (
	"context"
	"encoding/base64"
	"fmt"
	"hash/crc32"
	"strings"

	"github.com/mitchellh/mapstructure"

	kms "cloud.google.com/go/kms/apiv1"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type GCP struct {
	client       *kms.KeyManagementClient
	loadedKey    string
	vaultBaseUrl string
	config       *gcpConfig
}

type gcpConfig struct {
	// The GCP KMS Key name to use for encryption and decryption
	// The name is the full URI of the key, including project, location, key ring, name and optionally version
	KeyName string `mapstructure:"key_name"`
}

func NewGCP() (*GCP, error) {
	return &GCP{config: &gcpConfig{}}, nil
}

func (driver *GCP) Configure(config map[string]interface{}) error {
	err := mapstructure.Decode(config, driver.config)
	if err != nil {
		return fmt.Errorf("failed to decode gcp driver config: %w", err)
	}

	// The KMS client will intepret the key name, no need to parse it, just check if set
	if len(driver.config.KeyName) == 0 {
		return fmt.Errorf("gcp kms key name has to be set")
	}

	// As with all GCP libraries, credentials are derived from GOOGLE_APPLICATION_CREDENTIALS
	// The SDK expects us to give the client the same context as to the later requests, not sure how to handle that here as we Configure the client before making the calls
	driver.client, err = kms.NewKeyManagementClient(context.TODO())
	if err != nil {
		return err
	}

	return nil
}

func (driver *GCP) Encrypt(input string) (string, error) {
	// Convert the message into bytes. Cryptographic plaintexts and
	// ciphertexts are always byte arrays.
	plaintext := []byte(input)

	// Compute plaintext's CRC32C.
	crc32c := func(data []byte) uint32 {
		t := crc32.MakeTable(crc32.Castagnoli)
		return crc32.Checksum(data, t)
	}
	plaintextCRC32C := crc32c(plaintext)

	// Data gets encrpyted using the GOOGLE_SYMMETRIC_ENCRYPTION algorithm, which uses AES-256. You cannot choose another algorihtm
	req := &kmspb.EncryptRequest{
		Name:            driver.config.KeyName,
		Plaintext:       plaintext,
		PlaintextCrc32C: wrapperspb.Int64(int64(plaintextCRC32C)),
	}

	res, err := driver.client.Encrypt(context.TODO(), req)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt: %v", err)
	}

	// Perform integrity verification on result
	if res.VerifiedPlaintextCrc32C == false {
		return "", fmt.Errorf("Encrypt: request corrupted in-transit")
	}
	if int64(crc32c(res.Ciphertext)) != res.CiphertextCrc32C.Value {
		return "", fmt.Errorf("Encrypt: response corrupted in-transit")
	}

	return base64.StdEncoding.EncodeToString(res.Ciphertext), nil
}

func (driver *GCP) Decrypt(input string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return "", err
	}

	// Compute decoded's CRC32C
	crc32c := func(data []byte) uint32 {
		t := crc32.MakeTable(crc32.Castagnoli)
		return crc32.Checksum(data, t)
	}
	decodedCRC32C := crc32c(decoded)

	// In case the key name contains the key version, we need to omit it for the decrypt request
	// GCP will choose the correct key version for decryption
	keyName := strings.Split(driver.config.KeyName, "/cryptoKeyVersions/")[0]

	req := &kmspb.DecryptRequest{
		Name:             keyName,
		Ciphertext:       decoded,
		CiphertextCrc32C: wrapperspb.Int64(int64(decodedCRC32C)),
	}

	res, err := driver.client.Decrypt(context.TODO(), req)

	if err != nil {
		return "", fmt.Errorf("failed to decrypt input: %v", err)
	}

	if int64(crc32c(res.Plaintext)) != res.PlaintextCrc32C.Value {
		return "", fmt.Errorf("Decrypt: response corrupted in-transit")
	}

	return string(res.Plaintext), nil
}

func (driver *GCP) Type() string {
	return "gcpkms"
}
