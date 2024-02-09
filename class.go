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
	FilePath string
	// Access to the underlying container is usually not advised.
	// The Class itself exposes all the functionality of the container anyway.
	container *data.Container
}

// NewClass attempts to create a new class given a filesystem path and a codec.
// The class will only be created if the file is readable, can be decoded and
// adheres to the constraints set by [data.Container].
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

// Get the value at the given Path.
// Wrapper for [data.Container#Get]
func (c Class) Get(path string) (data.Value, error) {
	return c.container.Get(data.NewPath(path))
}

// GetAll returns the whole data represented by this class.
// Wrapper for [data.Container#Get]
func (c Class) GetAll() data.Value {
	ret, _ := c.container.Get(data.NewPath(""))
	return ret
}

// Set will set the given value at the specified path.
// Wrapper for [data.Container#Set]
func (c *Class) Set(path string, value interface{}) error {
	// preSetHook
	return c.container.Set(data.NewPath(path), value)
	// postSetHook
}

func (c *Class) AllPaths() []data.Path {
	return c.container.AllPaths()
}

func (c *Class) Walk(walkFunc data.WalkContainerFunc) error {
	return c.container.Walk(walkFunc)
}
