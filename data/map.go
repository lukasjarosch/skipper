package data

import (
	"fmt"
	"reflect"
	"strconv"
)

type Map map[string]interface{}

type WalkFunc func(value interface{}, path Path) error

// Walk over the data and call the [WalkFunc] on every leaf node.
func (d Map) Walk(valueFunc WalkFunc) (err error) {
	return walk(reflect.ValueOf(d), Path{}, valueFunc)
}

// walk does a dfs on the [Map] and calls [WalkFunc] on every value
func walk(v reflect.Value, path Path, walkFunc WalkFunc) error {
	// fix indirects through pointers and interfaces
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			err := walk(v.Index(i), path.Append(fmt.Sprint(i)), walkFunc)
			if err != nil {
				return err
			}
		}
	case reflect.Map:
		for _, key := range v.MapKeys() {
			if v.MapIndex(key).IsNil() {
				continue
			}

			err := walk(v.MapIndex(key), path.Append(key.String()), walkFunc)
			if err != nil {
				return err
			}
		}
	default:
		// at this point we have found a value and give off control to the given valueFunc
		err := walkFunc(v, path)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetPath allows path based indexing into Data.
// A path is a slice of interfaces which are used as keys in order.
// Supports array indexing (arrays start at 0).
// If the path does not exist, an error will be returned.
func (data Map) GetPath(path Path) (tree interface{}, err error) {
	tree = data

	for i, pathSegment := range path {
		switch node := tree.(type) {
		case Map:
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
				return nil, fmt.Errorf("unexpected path segment '%v' for node %T", pathSegment, node)
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

// SetPath sets a value at the specified path within the Map.
func (data Map) SetPath(path Path, value interface{}) error {
	if len(path) == 0 {
		return ErrEmptyPath
	}

	// Get the parent node of the path (all but the last segment)
	parentPath := path[:len(path)-1]

	// SetPath will not create any paths except for the last segment (which is either a map key or an array index).
	// So if the parentPath does not exist, we cannot continue.
	parent, err := data.GetPath(parentPath)
	if err != nil {
		return fmt.Errorf("cannot set path: %s, parent path does not exist: %s", path, parentPath)
	}

	// Check the type of the parent node
	switch node := parent.(type) {
	case Map:
		// If it's a Map, set the value at the last segment
		node[path.Last()] = value
		return nil

	case map[string]interface{}:
		// If it's a map with interface{} values, set the value
		node[path.Last()] = value
		return nil

	case []interface{}:
		// If it's a slice, parse the last segment as an index and set the value
		index, err := strconv.Atoi(path.Last())
		if err != nil {
			return fmt.Errorf("unexpected path segment '%v' for node %T", path.Last(), node)
		}

		if index < 0 || index >= len(node) {
			return fmt.Errorf("path index out of range: %d", index)
		}

		node[index] = value
		return nil

	default:
		return fmt.Errorf("unexpected node type %T at path %v", parent, parentPath)
	}
}

// MergeReplace merges the existing [Map] with the given input [Map].
// If a path segment already exists, the passed map data has precedence and it's value will be used.
func (data Map) MergeReplace(input Map) Map {
	out := make(Map, len(data))
	for k, v := range data {
		out[k] = v
	}
	for k, v := range input {
		if v, ok := v.(Map); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(Map); ok {
					out[k] = bv.MergeReplace(v)
					continue
				}
			}
		}
		if v, ok := v.([]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.([]interface{}); ok {
					out[k] = append(bv, v...)
					continue
				}
			}
		}
		out[k] = v
	}

	return out
}

func (data Map) HasPath(path Path) bool {
	_, err := data.GetPath(path)
	return err == nil
}
