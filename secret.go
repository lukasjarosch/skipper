package skipper

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/afero"
)

// secretRegex match pattern
// ?{driver:path/to/file||ifNotExistsAction:actionParam}
var secretRegex = regexp.MustCompile(`\?\{(\w+)\:([\w\/\-\.\_]+)(\|\|([\w\-\_\.\:]+))?\}`)

type Secret struct {
	*SecretFile
	Driver          SecretDriver
	DriverName      string
	AlternativeCall *Call
	Identifier      []interface{}
}

func NewSecret(secretFile *SecretFile, driver string, alternative *Call, path []interface{}) (*Secret, error) {
	return &Secret{
		SecretFile:      secretFile,
		Driver:          nil,
		DriverName:      driver,
		Identifier:      path,
		AlternativeCall: alternative,
	}, nil
}

// SecretFileData describes the generic structure of secret files.
type SecretFileData struct {
	Data string `yaml:"data"`
	Type string `yaml:"type"`
}

// NewSecretData constructs a [Data] map as it is required for secrets.
func NewSecretData(data string, driver string) (*SecretFileData, error) {
	if data == "" {
		return nil, fmt.Errorf("secret data cannot be empty")
	}
	if driver == "" {
		return nil, fmt.Errorf("secret file cannot have an empty type")
	}

	return &SecretFileData{
		Data: data,
		Type: driver,
	}, nil
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
		// secretFindValueFunc returns []*Secret so we need to ensure that matches
		vars, ok := val.([]*Secret)
		if !ok {
			return nil, fmt.Errorf("unexpected error during secret detection, file a bug report")
		}

		for _, secret := range vars {

			// ensure that the driver is loaded and assigned to every secret
			driver, err := SecretDriverFactory(secret.DriverName)
			if err != nil {
				return nil, fmt.Errorf("cannot get secret driver '%s': %w", secret.DriverName, err)
			}
			secret.Driver = driver

			// secrets which do not have a file associated are candidates for automatic creation
			if secret.SecretFile.YamlFile == nil {
				err = secret.attemptCreate(fs, secretPath)
				if err != nil {
					return nil, fmt.Errorf("failed to auto-create secret: %w", err)
				}
			}
		}

		foundSecrets = append(foundSecrets, vars...)
	}

	return foundSecrets, nil
}

// Load is used to load the actual secret files and ensure that they are correctly formatted.
// Load does NOT load the actual value, it just ensures that it could be loaded using the secret.Value() call.
func (s *Secret) Load(fs afero.Fs) error {
	if err := s.LoadSecretFileData(fs); err != nil {
		return fmt.Errorf("failed to load secret file: %w", err)
	}

	return nil
}

// attemptCreate will attempt to use the AlternativeAction of a secret to create it and write the required secret file to the filesystem.
func (secret *Secret) attemptCreate(fs afero.Fs, secretPath string) error {
	// if the secret does not have an alternative call, it is considered invalid and we cannot continue because we require the secret file to exist
	if secret.AlternativeCall == nil {
		return fmt.Errorf("secret does not have an alternative cwll: %s in '%s'", secret.FullName(), secret.Path())
	}

	// call the given alternative call function to get the target output
	output := secret.AlternativeCall.Execute()

	// use the driver implementation to encrypt the secret data
	encryptedData, err := secret.Driver.Encrypt(output)
	if err != nil {
		return fmt.Errorf("data encryption failed: %w", err)
	}

	// create new Data map which can then be written into the secret file
	secretFileData, err := NewSecretData(encryptedData, secret.Driver.Type())
	if err != nil {
		return fmt.Errorf("could not create NewSecretData: %w", err)
	}

	fileData, err := NewData(secretFileData)
	if err != nil {
		return fmt.Errorf("malformed SecretFileData cannot be converted to Data: %w", err)
	}

	// create the secret file with the data from the alternative action
	secretFile, err := CreateNewFile(fs, filepath.Join(secretPath, secret.RelativePath), fileData.Bytes())
	if err != nil {
		return err
	}
	secret.YamlFile = secretFile

	return nil
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

					alternativeCall, valid, err := NewStandaloneCall(secretAlternativeAction)
					if err != nil {
						return nil, err
					}

					// the call is not going to be executed if it is nil, thus we nil it here
					if !valid {
						alternativeCall = nil
					}

					newSecret, err := NewSecret(secretFile, secretDriver, alternativeCall, path)
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

// Value returns the actual secret value.
func (s *Secret) Value() (string, error) {
	return s.Driver.Decrypt(s.Data.Data)
}

// FullName returns the full secret name as it would be expected to ocurr in a class/target.
func (s Secret) FullName() string {
	if s.AlternativeCall != nil {
		return fmt.Sprintf("?{%s:%s||%s}", s.DriverName, s.SecretFile.RelativePath, s.AlternativeCall.RawString())
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
