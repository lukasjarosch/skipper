package secret

import (
	"fmt"
)

// Plain is the most basic secret driver.
// It does not expect any encryption and will just return any data which exists in the secret files.
type Plain struct {
	data map[string]interface{}
}

// NewPlain returns a newly initialized Plain driver.
func NewPlain() (*Plain, error) {
	driver := Plain{}

	return &driver, nil
}

// Value returns the secret value for the plaintext driver.
// The value is whatever is given in the 'data' field of the secret file.
// As long as the key exists and is a string, it is returned. Otherwise an error is returned.
func (p Plain) Value(data map[string]interface{}) (string, error) {
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

func (p Plain) Type() string {
	return "plain"
}
