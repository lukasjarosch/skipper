package secret

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

type Aes struct {
	*driver
}

func NewAes() (*Aes, error) {
	return &Aes{driver: &driver{initialized: true}}, nil
}

func (driver *Aes) Decrypt(encrypted string) (string, error) {
	if !driver.initialized {
		return "", ErrDriverNotInitialized
	}
	if !driver.isKeySet() {
		return "", ErrDriverKeyNotSet
	}

	decrypted, err := driver.decrypt([]byte(driver.key), encrypted)
	if err != nil {
		return "", err
	}

	return decrypted, nil
}

func (driver *Aes) Encrypt(input string) (string, error) {

	if !driver.initialized {
		return "", fmt.Errorf("%s: %w", driver.Type(), ErrDriverNotInitialized)
	}
	if !driver.isKeySet() {
		return "", fmt.Errorf("%s: %w", driver.Type(), ErrDriverKeyEmpty)
	}

	encrypted, err := driver.encrypt([]byte(driver.key), input)
	if err != nil {
		return "", err
	}

	return encrypted, nil
}

func (driver *Aes) encrypt(key []byte, message string) (encoded string, err error) {
	plainText := []byte(message)

	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	// make the cipher text a byte array of size BlockSize + the length of the message
	cipherText := make([]byte, aes.BlockSize+len(plainText))

	// iv is the ciphertext up to the blocksize (16)
	iv := cipherText[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	// encrypt the data
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], plainText)

	// return string encoded in base64
	return base64.RawStdEncoding.EncodeToString(cipherText), err
}

func (driver *Aes) decrypt(key []byte, secure string) (decoded string, err error) {
	cipherText, err := base64.RawStdEncoding.DecodeString(secure)
	if err != nil {
		return
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	if len(cipherText) < aes.BlockSize {
		return "", fmt.Errorf("ciphertext block size is too short; want=%d, have=%d", aes.BlockSize, len(cipherText))
	}

	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	// decrypt()
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherText, cipherText)

	return string(cipherText), err
}

func (driver *Aes) SetKey(key string) error {
	if len(key) != 32 {
		return fmt.Errorf("AES key must be exactly 32 Byte long")
	}

	driver.key = key

	return nil
}

func (driver *Aes) Initialize(config map[string]interface{}) error {
	return nil
}

func (driver *Aes) Type() string {
	return "aes"
}
