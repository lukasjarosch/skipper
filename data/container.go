package data

import (
	"fmt"
	"strings"
)

var (
	ErrEmptyContainerName = fmt.Errorf("container name empty")
	ErrNilData            = fmt.Errorf("data is nil")
	ErrNoRootKey          = fmt.Errorf("data has no root key (empty)")
	ErrMultipleRootKeys   = fmt.Errorf("multiple root keys")
	ErrInvalidRootKey     = fmt.Errorf("invalid root key")
	ErrNestedArrayPath    = fmt.Errorf("nested array paths are currently not supported")
)

// Container holds the raw data and provides access to it.
// The data can have only one key, which must match the container name.
// So a container named 'foo' must have data like 'map[string]interface{"foo": ....}'
type Container struct {
	name string
	data map[string]interface{}
}

func NewContainer(name string, data map[string]interface{}) (*Container, error) {
	if name == "" {
		return nil, ErrEmptyContainerName
	}
	if data == nil {
		return nil, ErrNilData
	}

	dataRootKeys := Keys(data)

	// empty data is not allowed, it must have ONE root key
	if len(dataRootKeys) <= 0 {
		return nil, ErrNoRootKey
	}

	// there can only be one root key in data
	if len(dataRootKeys) > 1 {
		return nil, ErrMultipleRootKeys
	}

	// the root key must be the same as the container name
	if !strings.EqualFold(fmt.Sprint(dataRootKeys[0]), name) {
		return nil, fmt.Errorf("expected root key '%s': %w", name, ErrInvalidRootKey)
	}

	// the name must exist as root key within data
	rootKey, err := Get(data, name)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrNoRootKey, err)
	}
	if rootKey == nil {
		return nil, ErrNoRootKey
	}

	c := &Container{
		name: name,
		data: data,
	}

	return c, nil
}

func (container *Container) Get(path Path) (Value, error) {
	if len(path) == 0 {
		return NewValue(container.data), nil
	}

	// make sure the path is abolute and contains the container name as first segment (aka root key)
	if path.First() != container.name {
		path = path.Prepend(container.name)
	}

	ret, err := DeepGet(container.data, []string(path))
	if err != nil {
		return Value{}, fmt.Errorf("%s: %w", ErrPathNotFound{Path: path}.Error(), err)
	}
	if ret == nil {
		return Value{}, ErrPathNotFound{Path: path}
	}

	return NewValue(ret), nil
}

func (container *Container) Set(path Path, value interface{}) error {
	if len(path) == 0 {
		return ErrEmptyPath
	}

	// ensure path is absolute
	if path.First() != container.name {
		path = path.Prepend(container.name)
	}

	ret, err := DeepSet(container.data, path, value)
	if err != nil {
		return fmt.Errorf("failed to set data: %w", err)
	}

	if _, ok := ret.(map[string]interface{}); !ok {
		return fmt.Errorf("DeepSet did not return a map, but '%T'", ret)
	}

	container.data = ret.(map[string]interface{})

	return nil
}

func (container *Container) Merge(path Path, data map[string]interface{}) error {
	inputData, err := container.Get(path)
	if err != nil {
		return err
	}

	inputDataMap, err := inputData.Map()
	if err != nil {
		return fmt.Errorf("input data is not a map: %w", err)
	}

	replaced := Merge(inputDataMap, data)

	err = container.Set(path, NewValue(replaced))
	if err != nil {
		return err
	}

	return nil
}

func (container *Container) AllPaths() []Path {
	return Paths(container.data, SelectAllPaths)
}

func (container *Container) LeafPaths() []Path {
	return Paths(container.data, SelectLeafPaths)
}

func (container *Container) MustGet(path Path) Value {
	val, err := container.Get(path)
	if err != nil {
		panic(err)
	}
	return val
}

func (container *Container) HasPath(path Path) bool {
	if _, err := container.Get(path); err != nil {
		return false
	}
	return true
}
