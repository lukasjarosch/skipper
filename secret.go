package skipper

import (
	"fmt"
	"log"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/afero"
)

// secretRegex match pattern
// ?{driver:path/to/file|ifNotExistsAction}
var secretRegex = regexp.MustCompile(`\?\{(\w+)\:([\w\/\-\.\_]+)(\|\|([\w\-\_\.\:]+))?\}`)

type Secret struct {
	*SecretFile
	Driver            SecretDriver
	DriverName        string
	AlternativeAction AlternativeAction
	Identifier        []interface{}
}

func NewSecret(secretFile *SecretFile, driver, alternative string, path []interface{}) (*Secret, error) {
	action, err := NewAlternativeAction(alternative)
	if err != nil {
		return nil, err
	}

	return &Secret{
		SecretFile:        secretFile,
		Driver:            nil,
		DriverName:        driver,
		AlternativeAction: action,
		Identifier:        path,
	}, nil
}

// Load is used to initialize the driver and use it to check the secret.
// Load does NOT load the actual value, it just ensures that it could be loaded using the driver.Value() call.
func (s *Secret) Load(fs afero.Fs) error {
	if err := s.YamlFile.Load(fs); err != nil {
		return fmt.Errorf("failed to load secret file: %w", err)
	}

	// every secret needs to have a 'data' key in which the actual secret is stored (encrypted on in plaintext depends on the type)
	if _, exists := s.Data["data"]; !exists {
		return fmt.Errorf("missing 'data' key in secret file: %s", s.Path())
	}

	// the 'type' key also needs to exist and it must be a string. It tells us which driver to load.
	typ, exists := s.Data["type"]
	if !exists {
		return fmt.Errorf("missing 'type' key in secret file: %s", s.Path())
	}
	driverType, ok := typ.(string)
	if !ok {
		return fmt.Errorf("secret field 'type' must be of type string, got %T", typ)
	}

	// attempt to load the secret driver
	driver, err := SecretDriverFactory(s.DriverName)
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
	return s.Driver.Decrypt(s.Data["data"].(string))
}

// FullName returns the full secret name as it would be expected to ocurr in a class/target.
func (s Secret) FullName() string {
	if s.AlternativeAction.IsSet() {
		return fmt.Sprintf("?{%s:%s||%s}", s.DriverName, s.SecretFile.RelativePath, s.AlternativeAction.String())
	} else {
		return fmt.Sprintf("?{%s:%s}", s.DriverName, s.SecretFile.RelativePath)
	}
}

func (s Secret) Path() string {
	var segments []string
	for _, seg := range s.Identifier {
		segments = append(segments, fmt.Sprint(seg))
	}
	return strings.Join(segments, ".")
}

// FindSecrets will leverage the `FindValues` function of [Data] to recursively search for secrets.
// All returned values are converted to *Secret and then returned as []*Secret.
func FindOrCreateSecrets(data Data, secretFiles SecretFileList, secretPath string, fs afero.Fs) ([]*Secret, error) {
	var foundValues []interface{}
	err := data.FindValues(secretFindValueFunc(secretFiles), &foundValues)
	if err != nil {
		return nil, err
	}

	var foundSecrets []*Secret
	for _, val := range foundValues {
		// secretFindValueFunc returns []*Variable so we need to ensure that matches
		vars, ok := val.([]*Secret)
		if !ok {
			return nil, fmt.Errorf("unexpected error during secret detection, file a bug report")
		}

		for _, secret := range vars {

			// make sure every secret has a reference to its driver
			driver, err := SecretDriverFactory(secret.DriverName)
			if err != nil {
				return nil, fmt.Errorf("cannot initialize driver '%s' on secret '%s': %w", secret.DriverName, secret.FullName(), err)
			}
			secret.Driver = driver

			// secrets which do not have a file associated are candidates for automatic creation
			if secret.SecretFile.YamlFile == nil {

				// TODO: create actual file (if possible)
				// TODO: add file to SecretFileList

				// if the secret does not have an alternative action, it is considered invalid and we cannot continue because we require the secret file to exist
				if !secret.AlternativeAction.IsSet() {
					return nil, fmt.Errorf("found secret without secret file and no alternative action: %s in '%s'", secret.FullName(), secret.Path())
				}

				// call the given alternative action function to get the target output
				output, err := secret.AlternativeAction.Call()
				if err != nil {
					return nil, fmt.Errorf("failed to call alternative action: %w", err)
				}

				encryptedData, err := secret.Driver.Encrypt(output)
				if err != nil {
					return nil, fmt.Errorf("data encryption failed: %w", err)
				}

				// TODO: ugly
				type secretFileData struct {
					Data string `yaml:"data"`
					Type string `yaml:"type"`
				}
				secretData := secretFileData{
					Data: encryptedData,
					Type: secret.Driver.Type(),
				}

				tmp, err := NewData(secretData)
				if err != nil {
					return nil, err
				}

				// create the secret file with the data from the alternative action
				secretFile, err := CreateNewFile(fs, filepath.Join(secretPath, secret.RelativePath), tmp.Bytes())
				if err != nil {
					return nil, err
				}
				secret.YamlFile = secretFile

				log.Println("CREATED SECRET", secret.SecretFile.Path)
			}
		}

		foundSecrets = append(foundSecrets, vars...)
	}

	return foundSecrets, nil
}

// secretFindValueFunc implements the [FindValueFunc] and searches for secrets inside [Data].
// Secrets can be found by matching any value to the [secretRegex].
// All found secrets are initialized, matched agains the SecretFileList to ensure they exist and added to the output.
// The function returns `[]*String` which needs to be restored afterwards.
func secretFindValueFunc(secretFiles SecretFileList) FindValueFunc {
	return func(value string, path []interface{}) (val interface{}, err error) {
		var secrets []*Secret

		matches := secretRegex.FindAllStringSubmatch(value, -1)
		if len(matches) > 0 {
			for _, secret := range matches {
				if len(secret) >= 3 {

					secretDriver := secret[1]
					secretRelativePath := secret[2]
					secretAlternativeAction := secret[4]

					// in case the secret file does not (yet) exist, the secretFile will be nil
					secretFile := secretFiles.GetSecretFile(secretRelativePath)

					// if the secretFile is nil, the secret does not (yet) exist.
					// we will need to create it further on, but store the relative path by creating an empty [SecretFile]
					if secretFile == nil {
						secretFile, err = NewSecretFile(nil, secretRelativePath)
						if err != nil {
							return nil, err
						}
					}

					newSecret, err := NewSecret(secretFile, secretDriver, secretAlternativeAction, path)
					if err != nil {
						return nil, fmt.Errorf("invalid secret %s: %w", secret[0], err)
					}
					secrets = append(secrets, newSecret)
				}
			}
		}
		return secrets, nil
	}
}
