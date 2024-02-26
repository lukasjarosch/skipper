package skipper

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/lukasjarosch/skipper/data"
)

var (
	ErrEmptyFilePath          = fmt.Errorf("file path cannot be empty")
	ErrEmptyClassId           = fmt.Errorf("empty class id")
	ErrEmptyClassIdentifier   = fmt.Errorf("class identifier cannot be empty")
	ErrInvalidClassIdentifier = fmt.Errorf("invalid class identifier")
)

// Codec is used to de-/encode the data which make up the Class.
type Codec interface {
	Unmarshal([]byte) (map[string]interface{}, error)
}

type (
	// SetHookFunc can be registered as either preSetHook or postSetHook
	// and will then be called respectively.
	SetHookFunc func(class Class, path data.Path, value data.Value) error
)

type DataGetter interface {
	GetPath(data.Path) (data.Value, error)
}

type DataSetter interface {
	SetPath(data.Path, interface{}) error
}

type DataSetterGetter interface {
	DataGetter
	DataSetter
}

type DataWalker interface {
	Walk(func(data.Path, data.Value, bool) error) error
}

type AbsolutePathMaker interface {
	// AbsolutePath resolves the given path to an absolute path, within the given context.
	// The context is required to resolve the Class to which the given path is relative to.
	// If the path is 'foo.bar' and the context is 'foo.bar.baz.key' the latter
	// can be used to uniquely identify which Class is the context of the path.
	AbsolutePath(path data.Path, context data.Path) (data.Path, error)
}

type DataContainer interface {
	DataSetterGetter
	DataWalker
	AbsolutePathMaker
}

// Class defines the main file-data abstraction used by skipper.
// Every file with hierarchical data can be represented by a Class.
type Class struct {
	// Name is the common name of the class.
	// It is derived from the filename which this class represents.
	Name string
	// Identifier
	Identifier data.Path
	// FilePath is the path to the underlying file on the filesystem.
	FilePath string
	// Access to the underlying container is usually not advised.
	// The Class itself exposes all the functionality of the container anyway.
	container DataContainer
	// The class allows hooks to be registered to monitor each call to Set
	preSetHooks  []SetHookFunc
	postSetHooks []SetHookFunc
}

// NewClass attempts to create a new class given a filesystem path and a codec.
// The class will only be created if the file is readable, can be decoded and
// adheres to the constraints set by [data.Container].
func NewClass(filePath string, codec Codec, identifier data.Path) (*Class, error) {
	if len(filePath) == 0 {
		return nil, ErrEmptyFilePath
	}
	if identifier.IsEmpty() {
		return nil, ErrEmptyClassIdentifier
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

	className := PathFileBaseName(filePath)
	if identifier.Last() != className {
		return nil, fmt.Errorf("class name '%s' must be last segment of classIdentifier '%s': %w", className, identifier, ErrInvalidClassIdentifier)
	}

	return &Class{
		container:    container,
		Identifier:   identifier,
		FilePath:     filePath,
		Name:         className,
		preSetHooks:  nil,
		postSetHooks: nil,
	}, nil
}

// Get the value at the given Path.
// Wrapper for [data.Container#Get]
func (c Class) Get(path string) (data.Value, error) {
	return c.container.GetPath(data.NewPath(path))
}

// GetPath is the same as [Class#Get], but it accepts a [data.Path]
func (c Class) GetPath(path data.Path) (data.Value, error) {
	return c.container.GetPath(path)
}

// HasPath returns true if the given path exists within the Class.
func (c Class) HasPath(path data.Path) bool {
	_, err := c.container.GetPath(path)
	if err != nil {
		return false
	}
	return true
}

// GetAll returns the whole data represented by this class.
// Wrapper for [data.Container#Get]
func (c Class) GetAll() data.Value {
	ret, _ := c.container.GetPath(data.NewPath(""))
	return ret
}

// Set will set the given value at the specified path.
// Wrapper for [data.Container#Set]
func (c *Class) Set(path string, value interface{}) error {
	absPath, err := c.container.AbsolutePath(data.NewPath(path), nil)
	if err != nil {
		return err
	}

	err = c.callPreSetHooks(absPath, data.NewValue(value))
	if err != nil {
		return err
	}

	err = c.container.SetPath(absPath, value)
	if err != nil {
		return err
	}

	err = c.callPostSetHooks(absPath, data.NewValue(value))
	if err != nil {
		return err
	}

	return nil
}

// SetPath the same as [Class#Set], but it accepts a [data.Path] instead.
func (c *Class) SetPath(path data.Path, value interface{}) error {
	return c.container.SetPath(path, value)
}

// RegisterPreSetHook registers a new pre-Set-hook.
// Hooks are called in the order they were registered.
func (c *Class) RegisterPreSetHook(preSetHookFunc SetHookFunc) {
	c.preSetHooks = append(c.preSetHooks, preSetHookFunc)
}

// RegisterPostSetHook registers a new post-Set-hook.
// Hooks are called in the order they were registered.
func (c *Class) RegisterPostSetHook(postSetHookFunc SetHookFunc) {
	c.postSetHooks = append(c.postSetHooks, postSetHookFunc)
}

// callPreSetHooks will call all registered preSetHooks in the order they were registered.
// If one hook returns an error, the execution is halted and the error is returned immediately.
func (c *Class) callPreSetHooks(path data.Path, value data.Value) error {
	for _, hook := range c.preSetHooks {
		err := hook(*c, path, value)
		if err != nil {
			return err
		}
	}
	return nil
}

// callPostSetHooks will call all registered postSetHooks in the order they were registered.
// If one hook returns an error, the execution is halted and the error is returned immediately.
func (c *Class) callPostSetHooks(path data.Path, value data.Value) error {
	for _, hook := range c.postSetHooks {
		err := hook(*c, path, value)
		if err != nil {
			return err
		}
	}
	return nil
}

// Walk allows traversing the underlying container data.
func (c *Class) Walk(walkFunc func(data.Path, data.Value, bool) error) error {
	return c.container.Walk(walkFunc)
}

// WalkValues is the same as [Walk] but it only traverses leaf paths
// hence only returns values as defined by the user.
// It also satisfies the [ReferenceSource] interface.
func (c *Class) WalkValues(walkFunc func(data.Path, data.Value) error) error {
	return c.Walk(func(path data.Path, value data.Value, isLeaf bool) error {
		if !isLeaf {
			return nil
		}
		return walkFunc(path, value)
	})
}

// Values returns a map of path -> data.Value where each value is a leaf value.
func (c *Class) Values() map[string]data.Value {
	valueMap := make(map[string]data.Value)

	c.WalkValues(func(p data.Path, v data.Value) error {
		valueMap[p.String()] = v
		return nil
	})

	return valueMap
}

// AbsolutePath ensures that the given path is absolute within the given context path.
// This function satisfies the [skipper.AbsolutePathMaker] interface.
// The second parameter is usually required to determine to which Class the path is relative to.
// In this case, that context is not needed as there is only this Class context.
// In case the path is empty or it is not valid within the given context, an error is returned.
func (c *Class) AbsolutePath(path data.Path, context data.Path) (data.Path, error) {
	return c.container.AbsolutePath(path, context)
}

// ClassLoader is a simple helper function which accepts a list of paths which will be loaded a Classes.
func ClassLoader(basePath string, files []string, codec Codec) ([]*Class, error) {
	var classes []*Class
	for _, file := range files {
		// The classIdentifier is derived from the path of the class and the 'basePath'.
		// The basePath, which is usually the path to the scope directory, is removed.
		strippedPath := strings.Replace(file, basePath, "", 1)
		classIdentifier := data.NewPathFromOsPath(strippedPath)

		class, err := NewClass(file, codec, classIdentifier)
		if err != nil {
			return nil, fmt.Errorf("cannot create class from file: %s: %w", file, err)
		}
		classes = append(classes, class)
	}

	return classes, nil
}
