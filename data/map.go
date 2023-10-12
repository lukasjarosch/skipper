package data

import (
	"fmt"
	"reflect"
	"strconv"
)

type Map map[string]interface{}

// Get allows path based indexing into Data.
// A path is a slice of interfaces which are used as keys in order.
// Supports array indexing (arrays start at 0).
// If the path does not exist, an error will be returned.
func (data Map) Get(path Path) (tree interface{}, err error) {
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

// CanSet determines whether the given Path can be used within [Set]
// to write to [Map].
// We currently only allow setting one more segment than any existing paths.
// This has to do with the tremendous amount of edge-cases if one allows more.
// So if one wants to set `foo.bar.baz`, the path `foo.bar` must already exist.
// Additionally, if the path to be set is a parent of a leaf path (aka value path),
// the path cannot be set (can't create a child of a leaf).
//
// This restriction is totally synthetic and may be lifted in the future.
// It all depends on [Map.Set] to be able to dynamically create new types within [Map].
func (data Map) CanSet(path Path) error {
	// Get the parent node of the path (all but the last segment)
	parentPath := path[:len(path)-1]

	// Set will not create any paths except for the last segment (which is either a map key or an array index).
	// So if the parentPath does not exist, we cannot continue.
	if !data.HasPath(parentPath) {
		return fmt.Errorf("cannot set path which creates more than one new path segment")
	}

	// in case the parentPath is already a value path (e.g. points to a scalar value)
	// we cannot add another path segment
	if data.IsValuePath(parentPath) {
		return fmt.Errorf("cannot set path which creates a child segment on an existing value path")
	}

	return nil
}

// Set sets a value at the specified path within the Map.
func (data Map) Set(path Path, value interface{}) error {
	if len(path) == 0 {
		return ErrEmptyPath
	}

	if err := data.CanSet(path); err != nil {
		return err
	}

	// value gatekeeper
	if value != nil {
		kind := reflect.TypeOf(value).Kind()

		switch kind {
		case reflect.Func:
			return fmt.Errorf("cannot set function as value")
		case reflect.Struct:
			return fmt.Errorf("cannot set struct as value")
		}
	}

	// Get the parent node of the path (all but the last segment)
	parentPath := path[:len(path)-1]

	// fetch the parent (we know it exists becase [CanSet] verifies that)
	parent, _ := data.Get(parentPath)

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

// IsValuePath returns true if the given Path points to a scalar value (aka leaf node)
func (data Map) IsValuePath(path Path) bool {
	for _, p := range data.ValuePaths() {
		if p.String() == path.String() {
			return true
		}
	}
	return false
}

// HasPath returns true if calling [Map.Get] on the given path returns no error.
func (data Map) HasPath(path Path) bool {
	_, err := data.Get(path)
	return err == nil
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

// ValuePaths traverses over all data of the [Map] and returns all Paths to values (aka leaf nodes).
func (m Map) ValuePaths() []Path {
	// we're using a map to avoid the pescy append slice behaviour
	pathMap := make(map[string]bool)
	m.Walk(func(value interface{}, path Path) error {
		pathMap[path.String()] = true
		return nil
	})

	var paths []Path
	for p := range pathMap {
		paths = append(paths, NewPath(p))
	}

	return paths
}

func (m Map) AllPaths() []Path {
	return allPaths(m, NewPath(""))
}

// AllPaths traverses over every existing path in the given [Map].
func allPaths(m Map, pathPrefix Path) []Path {
	paths := make([]Path, 0)

	for key, value := range m {
		// Build the current path for the current key
		currentPath := pathPrefix.Append(key)

		switch v := value.(type) {
		case Map:
			// If the value is another map, recursively traverse it
			paths = append(paths, currentPath)
			paths = append(paths, allPaths(v, currentPath)...)
		case []interface{}:
			// If the value is a slice, iterate over elements and include numeric indices in the path
			for i, elem := range v {
				elemPath := currentPath.Append(strconv.Itoa(i))
				paths = append(paths, elemPath)
				if reflect.TypeOf(elem).Kind() == reflect.Map {
					paths = append(paths, allPaths(elem.(Map), elemPath)...)
				}
			}
		default:
			// If the value is a leaf node, add the current path to the list of paths
			paths = append(paths, currentPath)
		}
	}

	return paths
}
