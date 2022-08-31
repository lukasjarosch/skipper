package skipper

import (
	"fmt"
	"log"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/spf13/afero"
)

// secretRegex match pattern
// ?{driver:path/to/file|ifNotExistsAction}
var secretRegex = regexp.MustCompile(`\?\{(\w*)\:([\w\/\-\.\_]+)(\|([\w\-\_\.]+))?\}`)

type SecretDriver interface {
	Type() string
	Value(data map[string]interface{}) (string, error)
}

// SecretDriverFactory is any function which is capable of returning a driver, given the type.
type SecretDriverFactory func(driverName string) (SecretDriver, error)

type Secret struct {
	*YamlFile
	Driver            SecretDriver
	DriverName        string
	AlternativeAction string
}

func NewSecret(driver, file, alternative string) (*Secret, error) {
	f, err := NewFile(file)
	if err != nil {
		return nil, err
	}

	s := &Secret{
		DriverName:        driver,
		AlternativeAction: alternative,
		YamlFile:          f,
	}

	return s, nil
}

// Load is used to initialize the driver and use it to check the secret.
// Load does NOT load the actual value, it just ensures that it could be loaded using the driver.Value() call.
func (s *Secret) Load(fs afero.Fs, factory SecretDriverFactory) error {
	if err := s.YamlFile.Load(fs); err != nil {
		return fmt.Errorf("failed to load secret file: %w", err)
	}

	// every secret needs to have a 'data' key in which the actual secret is stored (encrypted on in plaintext depends on the type)
	if _, exists := s.Data["data"]; !exists {
		return fmt.Errorf("missing 'data' key in secret file: %s", s.Path)
	}

	// the 'type' key also needs to exist and it must be a string. It tells us which driver to load.
	typ, exists := s.Data["type"]
	if !exists {
		return fmt.Errorf("missing 'type' key in secret file: %s", s.Path)
	}
	driverType, ok := typ.(string)
	if !ok {
		return fmt.Errorf("secret field 'type' must be of type string, got %T", typ)
	}

	// attempt to load the driver with the given factory
	// FIXME: currently we're loading the driver for every secret. This will become ineffective once we're using actual clients and needs to be addressed.
	driver, err := factory(s.DriverName)
	if err != nil {
		return fmt.Errorf("cannot initialize driver '%s': %w", s.DriverName, err)
	}
	s.Driver = driver

	// driverType in the secret file must match the type of the loaded driver
	if !strings.EqualFold(driverType, driver.Type()) {
		return fmt.Errorf("secret driver mismatch, data uses dirver '%s', was loaded with driver '%s'", typ, driver.Type())
	}

	return nil
}

// Value returns the actual secret value.
func (s *Secret) Value() (string, error) {
	return s.Driver.Value(s.Data)
}

// FullName returns the full secret name as it would be expected to ocurr in a class/target.
func (s Secret) FullName() string {
	if len(s.AlternativeAction) > 0 {
		return fmt.Sprintf("?{%s:%s|%s}", s.DriverName, s.Path, s.AlternativeAction)
	} else {
		return fmt.Sprintf("?{%s:%s}", s.DriverName, s.Path)
	}
}

type SecretList []*Secret

// TODO eventually combine FindVariables and FindSecrets
func FindSecrets(secretPath string, data any) (secrets SecretList) {

	// newPath is used to copy an existing []interface and hard-copy it.
	// This is required because Go wants to optimize slice usage by reusing memory.
	// Most of the time, this is totally fine, but in this case it would mess up the slice
	// by changing the path []interface of already found secrets.
	newPath := func(path []interface{}, appendValue interface{}) []interface{} {
		tmp := make([]interface{}, len(path))
		copy(tmp, path)
		tmp = append(tmp, appendValue)
		return tmp
	}

	var walk func(reflect.Value, []interface{})
	walk = func(v reflect.Value, path []interface{}) {

		// fix indirects through pointers and interfaces
		for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
			v = v.Elem()
		}

		switch v.Kind() {
		case reflect.Array, reflect.Slice:
			for i := 0; i < v.Len(); i++ {
				walk(v.Index(i), newPath(path, i))
			}
		case reflect.Map:
			for _, key := range v.MapKeys() {
				if v.MapIndex(key).IsNil() {
					break
				}

				walk(v.MapIndex(key), newPath(path, key.String()))
			}
		default:
			// Here we've arrived at actual values, hence we can check whether the value is a secret
			matches := secretRegex.FindAllStringSubmatch(v.String(), -1)
			if len(matches) > 0 {
				for _, secret := range matches {
					if len(secret) >= 3 {
						// based on the regex, we're interested in capture group 1 (driver), 2 (file) and 4 (alternative)
						newSecret, err := NewSecret(secret[1], filepath.Join(secretPath, secret[2]), secret[4])
						if err != nil {
							log.Fatalln(fmt.Errorf("invalid secret detected: %w", err)) // this error is not recoverable, user error
						}
						secrets = append(secrets, newSecret)
					}
				}
			}
		}
	}
	walk(reflect.ValueOf(data), nil)

	return secrets
}
