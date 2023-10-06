package data

import (
	"fmt"
	"time"
)

type Value struct {
	Raw interface{}
}

func NewValue(raw interface{}) Value {
	return Value{Raw: raw}
}

func (val Value) String() string {
	return fmt.Sprint(val.Raw)
}

func (val Value) Map() (Map, error) {
	m, ok := val.Raw.(Map)
	if !ok {
		return nil, fmt.Errorf("cannot convert value to Map: %s", val)
	}
	return m, nil
}

func (val Value) Duration() (time.Duration, error) {
	dur, err := time.ParseDuration(val.String())
	if err != nil {
		return 0, err
	}
	return dur, nil
}

func (val Value) Int() (int, error) {
	i, ok := val.Raw.(int)
	if !ok {
		return 0, fmt.Errorf("cannot convert value to int: %s", val)
	}
	return i, nil
}

func (val Value) Int32() (int32, error) {
	i, ok := val.Raw.(int32)
	if !ok {
		return 0, fmt.Errorf("cannot convert value to int32: %s", val)
	}
	return i, nil
}

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
