package skipper

import (
	"fmt"
	"reflect"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Data map[string]interface{}

// NewData attempts to convert any given interface{} into [Data].
// This is done by first using `yaml.Marshal` and then `yaml.Unmarshal`.
// If the given interface is compatible with [Data], these steps will succeed.
func NewData(input interface{}) (Data, error) {
	outBytes, err := yaml.Marshal(input)
	if err != nil {
		return nil, err
	}

	var data Data
	err = yaml.Unmarshal(outBytes, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// GetPath allows path based indexing into Data.
// A path is a slice of interfaces which are used as keys in order.
// Supports array indexing (arrays start at 0)
// Examples of valid paths:
//   - ["foo", "bar"]
//   - ["foo", "bar", 0]
func (data Data) GetPath(path Path) (tree interface{}, err error) {
	tree = data

	for i, pathSegment := range path {
		switch node := tree.(type) {
		case Data:
			var ok bool
			tree, ok = node[pathSegment]
			if !ok {
				return nil, fmt.Errorf("path segment not found: %s", pathSegment)
			}

		case map[string]interface{}:
			var ok bool
			tree, ok = node[pathSegment]
			if !ok {
				return nil, fmt.Errorf("path segment not found: %s", pathSegment)
			}

		case map[interface{}]interface{}:
			var ok bool
			tree, ok = node[pathSegment]
			if !ok {
				return nil, fmt.Errorf("path segment not found: %v", pathSegment)
			}

		case []interface{}:
			index, err := strconv.Atoi(fmt.Sprint(pathSegment))
			if err != nil {
				return nil, fmt.Errorf("unexpected integer path segment '%v' (%T)", pathSegment, pathSegment)
			}
			if index < 0 || index >= len(node) {
				return nil, fmt.Errorf("path index out of range: %d", index)
			}
			tree = node[index]
		default:
			return nil, fmt.Errorf("unexpected node type %T at index %d", node, i)
		}
	}

	return tree, nil
}

// HasPath returns true if `GetPath` does not return an error for the given Path
func (data Data) HasPath(path Path) bool {
	_, err := data.GetPath(path)
	return err == nil
}

// UnmarshalPath can be used to YAML unmarshal a path into the given target.
// The target must be a pointer.
// Since `GetPath` returns an interface, the data is first marshalled into YAML and
// then unmarshalled into the target interface.
func (data Data) UnmarshalPath(path Path, target interface{}) error {
	if reflect.ValueOf(target).Kind() != reflect.Ptr {
		return fmt.Errorf("cannot unmarshal path '%s': target is not a pointer", path.String())
	}

	tmp, err := data.GetPath(path)
	if err != nil {
		return fmt.Errorf("cannot get data at path '%s': %w", path.String(), err)
	}

	tmpBytes, err := yaml.Marshal(tmp)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	err = yaml.Unmarshal(tmpBytes, target)
	if err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return nil
}
