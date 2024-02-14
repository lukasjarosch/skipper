package driver

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

// the plain driver does not do anything
func (p *Plain) Decrypt(encrypted string, key string) (string, error) {
	// key is dismissed, as plain does not do anything
	return encrypted, nil
}

// the plain driver does not do anything
func (p *Plain) Encrypt(input string) (string, error) {
	return input, nil
}

func (p Plain) Type() string {
	return "plain"
}

func (p Plain) GetKey() string {
	return "plainDoesntHaveAKey"
}
