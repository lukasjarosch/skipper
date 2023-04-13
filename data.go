package skipper

import (
	"fmt"
	"reflect"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Data is the core data abstraction used by Skipper.
// Skipper is based on YAML data, hence [Data] is just a
// raw representation of a yaml structure.
//
// Data allows accessing the map by path, without
// Skipper needing to know the underlying structure.
// This means we do not have to unmarshal the given [Data]
// into a specific type and still work with it.
//
// As long as the user knows the structure of its
// data, everything checks out.
type Data map[string]interface{}

// NewData attempts to convert any given interface{} into [Data].
// This is done by first using [yaml.Marshal] and then [yaml.Unmarshal].
//
// As long as the given input interface is compatible with yaml,
// these steps will succeed and a valid [Data] will be returned.
// In any other case, an error is returned.
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

// HasPath returns true if [Data.GetPath] does not return an error for the given Path.
func (data Data) HasPath(path Path) bool {
	_, err := data.GetPath(path)
	return err == nil
}

// UnmarshalPath can be used to YAML unmarshal a path into the given target.
// This is useful if you want to map from [Data] into an actual type you provide.
//
// The target must be a pointer, otherwise an error is returned.
//
// Using structtags in your type is preferable to control the mapping of values.
//
// Because [Data.GetPath] returns an interface, the data is first marshalled into YAML and
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
