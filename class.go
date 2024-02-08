package skipper

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/lukasjarosch/skipper/data"
)

var ErrEmptyFilePath = fmt.Errorf("file path cannot be empty")

type Codec interface {
	Unmarshal([]byte) (map[string]interface{}, error)
}

type Class struct {
	filePath  string
	container *data.Container
}

func NewClass(filePath string, codec Codec) (*Class, error) {
	if len(filePath) == 0 {
		return nil, ErrEmptyFilePath
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to open file: %w", err)
	}

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("unable to read file: %w", err)
	}

	fileData, err := codec.Unmarshal(fileBytes)
	if err != nil {
		return nil, fmt.Errorf("unable to decode class data: %w", err)
	}

	// the name of the container is the filename without file extension
	containerName := filepath.Base(filePath)
	containerName = strings.TrimSuffix(containerName, filepath.Ext(containerName))

	container, err := data.NewContainer(containerName, fileData)
	if err != nil {
		return nil, fmt.Errorf("invalid container: %w", err)
	}

	return &Class{
		container: container,
		filePath:  filePath,
	}, nil
}

func (c Class) Get(path string) (data.Value, error) {
	return c.container.Get(data.NewPath(path))
}

func (c Class) GetAll() data.Value {
	ret, _ := c.container.Get(data.NewPath(""))
	return ret
}

func (c *Class) Set(path string, value interface{}) error {
	return c.container.Set(data.NewPath(path), value)
}
