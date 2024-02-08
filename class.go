package skipper

import (
	"fmt"
	"io"
	"os"

	"github.com/lukasjarosch/skipper/data"
)

var ErrEmptyFilePath = fmt.Errorf("file path cannot be empty")

type Codec interface {
	Unmarshal([]byte) (map[string]interface{}, error)
}

type Class struct {
	// Name is the common name of the class.
	Name string
	// FilePath is the path to the file which this class represents.
	FilePath  string
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

	container, err := data.NewContainer(PathFileBaseName(filePath), fileData)
	if err != nil {
		return nil, fmt.Errorf("invalid container: %w", err)
	}

	return &Class{
		container: container,
		FilePath:  filePath,
		Name:      PathFileBaseName(filePath),
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
