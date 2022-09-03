package secret

import (
	"encoding/base64"
	"fmt"
)

// Base64 is the most basic secret driver.
// It does not expect any encryption and will just return any data which exists in the secret files.
type Base64 struct {
	initialized bool
	data        map[string]interface{}
}

// NewBase64 returns a newly initialized Plain driver.
func NewBase64() (*Base64, error) {
	driver := Base64{
		initialized: true, // the driver does not require initialization
	}

	return &driver, nil
}

// Value returns the secret value for the plaintext driver.
// The value is whatever is given in the 'data' field of the secret file.
// As long as the key exists and is a string, it is returned. Otherwise an error is returned.
func (p *Base64) Value(data map[string]interface{}) (string, error) {
	if !p.initialized {
		return "", fmt.Errorf("driver not initialized: %s", p.Type())
	}
	secretData, exists := data["data"]
	if !exists {
		return "", fmt.Errorf("secret data missing 'data' key")
	}
	secretValue, ok := secretData.(string)
	if !ok {
		return "", fmt.Errorf("secret field 'data' must be of type string, got %T", data["data"])
	}

	return secretValue, nil
}

func (p *Base64) Decrypt(encrypted string) (string, error) {
	if !p.initialized {
		return "", fmt.Errorf("driver not initialized: %s", p.Type())
	}

	out, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func (p *Base64) Encrypt(input string) (string, error) {
	if !p.initialized {
		return "", fmt.Errorf("driver not initialized: %s", p.Type())
	}

	return base64.StdEncoding.EncodeToString([]byte(input)), nil
}

func (p *Base64) Initialize(config map[string]interface{}) error {
	p.initialized = true
	return nil
}

func (p *Base64) Type() string {
	return "base64"
}
