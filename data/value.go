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