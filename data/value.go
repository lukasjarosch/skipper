package data

import (
	"fmt"
	"time"
)

// Value represents a generic value that originates fom [Data]
type Value struct {
	Raw interface{}
}

// NewValue creates a new Value instance with the given raw value.
func NewValue(raw interface{}) Value {
	return Value{Raw: raw}
}

// String returns a string representation of the value.
func (val Value) String() string {
	return fmt.Sprint(val.Raw)
}

// Map attempts to convert the value to a Map. Returns an error if conversion is not possible.
func (val Value) Map() (Map, error) {
	m, ok := val.Raw.(Map)
	if !ok {
		return nil, fmt.Errorf("cannot convert value to Map: %s", val)
	}
	return m, nil
}

// Duration attempts to convert the value to a time.Duration. Returns an error if conversion is not possible.
func (val Value) Duration() (time.Duration, error) {
	dur, err := time.ParseDuration(val.String())
	if err != nil {
		return 0, err
	}
	return dur, nil
}

// Int attempts to convert the value to an int. Returns an error if conversion is not possible.
func (val Value) Int() (int, error) {
	i, ok := val.Raw.(int)
	if !ok {
		return 0, fmt.Errorf("cannot convert value to int: %s", val)
	}
	return i, nil
}

// Int32 attempts to convert the value to an int32. Returns an error if conversion is not possible.
func (val Value) Int32() (int32, error) {
	i, ok := val.Raw.(int32)
	if !ok {
		return 0, fmt.Errorf("cannot convert value to int32: %s", val)
	}
	return i, nil
}

// Int64 attempts to convert the value to an int64. Returns an error if conversion is not possible.
func (val Value) Int64() (int64, error) {
	i, ok := val.Raw.(int64)
	if !ok {
		return 0, fmt.Errorf("cannot convert value to int64: %s", val)
	}
	return i, nil
}

// InventoryValue represents any value within the [Inventory]
// It has a [ValueScope] which determines the scope in which the value is defined.
type InventoryValue struct {
	Value
	Scope ValueScope
}

// ValueScope defines a scope in which a value is valid / defined.
// It can also be used to retrieve the value again.
type ValueScope struct {
	// Namespace in which the value resides
	Namespace Path
	// The path within the container which points to the value
	ContainerPath Path
	// The actual container
	Container Container
}

// AbsolutePath returns the absolute path to this value.
// This is the namespace + ContainerPath.
// The AbsolutePath can be used to retrieve the value from the [Inventory]
//
// Note that the absolute path does not work with a scoped Inventory, only with the root Inventory.
// If you want to use it in a scoped inventory you need to strip the prefix yourself.
func (scope ValueScope) AbsolutePath() Path {
	return scope.Namespace.AppendPath(scope.ContainerPath)
}
