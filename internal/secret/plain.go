package secret

import (
	"fmt"
)

// Plain is the most basic secret driver.
// It does not expect any encryption and will just return any data which exists in the secret files.
type Plain struct {
	initialized bool
	data        map[string]interface{}
}

// NewPlain returns a newly initialized Plain driver.
func NewPlain() (*Plain, error) {
	driver := Plain{
		initialized: true, // the driver does not require initialization
	}

	return &driver, nil
}

// Value returns the secret value for the plaintext driver.
// The value is whatever is given in the 'data' field of the secret file.
// As long as the key exists and is a string, it is returned. Otherwise an error is returned.
func (p *Plain) Value(data map[string]interface{}) (string, error) {
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

func (p *Plain) Decrypt(encrypted string) (string, error) {
	if !p.initialized {
		return "", fmt.Errorf("driver not initialized: %s", p.Type())
	}

	// the plain driver does not do anything
	return encrypted, nil
}

func (p *Plain) Encrypt(input string) (string, error) {
	if !p.initialized {
		return "", fmt.Errorf("driver not initialized: %s", p.Type())
	}

	// the plain driver does not do anything
	return input, nil
}

func (p *Plain) Initialize(config map[string]interface{}) error {
	p.initialized = true
	return nil
}

func (p *Plain) Type() string {
	return "plain"
}
