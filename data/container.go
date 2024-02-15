package data

import (
	"fmt"
	"strings"
)

var (
	ErrEmptyRootKeyName = fmt.Errorf("empty root key name")
	ErrNilData          = fmt.Errorf("data is nil")
	ErrNoRootKey        = fmt.Errorf("data has no root key (empty)")
	ErrMultipleRootKeys = fmt.Errorf("multiple root keys")
	ErrInvalidRootKey   = fmt.Errorf("invalid root key")
	ErrNestedArrayPath  = fmt.Errorf("nested array paths are currently not supported")
)

// Container holds the raw data and provides access to it.
// The data can have only one root key.
// So a container with root key 'foo' must have data like 'map[string]interface{"foo": ....}'
type Container struct {
	rootKey string
	data    map[string]interface{}
}

// NewContainer constructs a new container from the given data.

// A container must meet the following requirements:
//   - rootKeyName cannot be empty
//   - data cannot be nil
//   - there can only be exactly one root key within the data
//   - the root key of the data must be the same as the rootKeyName
//
// If any of these conditions are not met, an error is returned.
func NewContainer(rootKeyName string, data map[string]interface{}) (*Container, error) {
	if rootKeyName == "" {
		return nil, ErrEmptyRootKeyName
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

	// the root key must be the same as the rootKeyName
	if !strings.EqualFold(fmt.Sprint(dataRootKeys[0]), rootKeyName) {
		return nil, fmt.Errorf("expected root key '%s': %w", rootKeyName, ErrInvalidRootKey)
	}

	// the name must exist as root key within data
	rootKey, err := Get(data, rootKeyName)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrNoRootKey, err)
	}
	if rootKey == nil {
		return nil, ErrNoRootKey
	}

	c := &Container{
		rootKey: rootKeyName,
		data:    data,
	}

	return c, nil
}

// Get is used to retrieve Values from the container data.
//
// The given path can be absolute or relative.
// An absolute path starts with the root key whereas
// a relative path omits the name.
// For example, if the container is called 'foo' you can
// use the path 'foo.bar' or 'bar' to address the same value.
//
// If the path does not exist, a [ErrPathNotFound] error is returned.
func (container *Container) Get(path Path) (Value, error) {
	if len(path) == 0 {
		return NewValue(container.data), nil
	}

	// make sure the path is absolute and contains the root key as first segment
	if path.First() != container.rootKey {
		path = path.Prepend(container.rootKey)
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

	path = container.AbsolutePath(path)

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
// The way this works is that the given Path is fetched from the container.
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

type WalkContainerFunc func(path Path, value Value, isLeaf bool) error

// Walk is the container implementation of the general [Walk] function.
// The only difference is that it uses [Value] types instead of arbitrary interfaces.
func (container *Container) Walk(walkContainerFunc WalkContainerFunc) error {
	return Walk(container.data, func(path Path, data interface{}, isLeaf bool) error {
		return walkContainerFunc(container.AbsolutePath(path), NewValue(data), isLeaf)
	})
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

// AbsolutePath ensures that the given path is absolute.
func (container *Container) AbsolutePath(path Path) Path {
	if path.First() != container.rootKey {
		path = path.Prepend(container.rootKey)
	}
	return path
}