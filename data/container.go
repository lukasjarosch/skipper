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

// NewContainer constructs a new container from the given data.

// A container must meet the following requirements:
//   - name cannot be empty
//   - data cannot be nil
//   - there can only be exactly one root key within the data
//   - the root key of the data must be the same as the container name
//
// If any of these conditions are not met, an error is returned.
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

// Get is used to retrieve Values from the container data.
//
// The given path can be absolute or relative.
// An absolute path starts with the container name whereas
// a relative path omits the name.
// For example, if the container is called 'foo' you can
// use the path 'foo.bar' or 'bar' to address the same value.
//
// If the path does not exist, a [ErrPathNotFound] error is returned.
func (container *Container) Get(path Path) (Value, error) {
	if len(path) == 0 {
		return NewValue(container.data), nil
	}

	// make sure the path is absolute and contains the container name as first segment (aka root key)
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

// Set is used to set values within the container data.
// It supports dynamic data creation when a given Path does not exist.
//
// It does not support changing existing data-types.
// So if the path 'foo.bar' points to a `map` type, setting
// 'foo.1' will not convert it to an array, but rather use
// the string representation "1" as map key.
//
// Additionally, if somewhere along the given Path a scalar value is
// found, an error is returned for the same reason; types are not changed.
// This means that if there is a scalar value at 'foo.bar.baz', you
// cannot set 'foo.bar.baz.qux'.
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

// Merge is used to merge in new data at a given path.
//
// The way this works ist that the given Path is fetched from the container.
// Then the resulting map is merged with the given input map.
// When merging, the given data has precedence over existing data and will
// overwrite any existing values.
// In case of slices, the new data is appended.
//
// Note: Only map types can be merged! So the given path must point to a map type.
// TODO: Should empty paths (aka full data merging) be allowed?
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

// AllPaths returns all existing paths within the Container data.
func (container *Container) AllPaths() []Path {
	return Paths(container.data, SelectAllPaths)
}

// LeafPaths returns all paths which point to a scalar value (aka leaf paths).
func (container *Container) LeafPaths() []Path {
	return Paths(container.data, SelectLeafPaths)
}

// MustGet is a wrapper around [Container.Get] which panics on error.
func (container *Container) MustGet(path Path) Value {
	val, err := container.Get(path)
	if err != nil {
		panic(err)
	}
	return val
}

// HasPath returns true if the given path exists, false otherwise.
func (container *Container) HasPath(path Path) bool {
	if _, err := container.Get(path); err != nil {
		return false
	}
	return true
}
