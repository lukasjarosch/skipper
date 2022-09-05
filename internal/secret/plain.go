package secret

import (
	"fmt"
)

// Plain is the most basic secret driver.
// It does not expect any encryption and will just return any data which exists in the secret files.
type Plain struct {
	*driver
	data map[string]interface{}
}

// NewPlain returns a newly initialized Plain driver.
func NewPlain() (*Plain, error) {
	driver := Plain{
		driver: &driver{
			initialized: true, // the driver does not require initialization
		},
	}

	return &driver, nil
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

func (p *Plain) Type() string {
	return "plain"
}
