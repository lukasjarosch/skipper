package secret

import (
	"encoding/base64"
	"fmt"
)

// Base64 is the most basic secret driver.
// It does not expect any encryption and will just return any data which exists in the secret files.
type Base64 struct {
	*driver
	data map[string]interface{}
}

// NewBase64 returns a newly initialized Plain driver.
func NewBase64() (*Base64, error) {
	driver := Base64{
		driver: &driver{
			initialized: true, // the driver does not require initialization
		},
	}

	return &driver, nil
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

func (p *Base64) Type() string {
	return "base64"
}
