package secret

import (
	"fmt"
)

var (
	ErrDriverNotInitialized error = fmt.Errorf("secret driver not initialized")
	ErrDriverKeyNotSet error = fmt.Errorf("secret driver key not set")
	ErrDriverKeyEmpty error = fmt.Errorf("secret driver key cannot be empty")
)

type driver struct {
	initialized bool
	key         string
}

func (drv *driver) Initialize(config map[string]interface{}) error {
	drv.initialized = true
	return nil
}

func (drv *driver) SetKey(key string) error {
	if key == "" {
		return ErrDriverKeyEmpty
	}
	drv.key = key
	return nil
}

func (drv *driver) isKeySet() bool {
	if drv.key == "" {
		return false
	}
	return true
}
