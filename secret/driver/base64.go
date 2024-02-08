package driver

import (
	"encoding/base64"
)

// Base64 is the most basic secret driver.
// It does not expect any encryption and will just return any data which exists in the secret files.
type Base64 struct {
	data map[string]interface{}
}

// NewBase64 returns a newly initialized Plain driver.
func NewBase64() (*Base64, error) {
	driver := Base64{}

	return &driver, nil
}

func (p *Base64) Decrypt(encrypted string, key string) (string, error) {
	// key is dismissed, as base64 isn't decrypting stuff

	out, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func (p *Base64) Encrypt(input string) (string, error) {
	return base64.StdEncoding.EncodeToString([]byte(input)), nil
}

func (p *Base64) Type() string {
	return "base64"
}

func (p *Base64) GetKey() string {
	return "base64DoesNotHaveAKey"
}
