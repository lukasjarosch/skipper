package skipper

import (
	"fmt"
	"reflect"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Data is an arbitrary map of values which makes up the inventory.
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

// String returns the string result of `yaml.Marshal`.
// Can be useful for debugging or just dumping the inventory.
func (d Data) String() string {
	out, _ := yaml.Marshal(d)
	return string(out)
}

// Bytes returns a `[]byte` representation of Data
func (d Data) Bytes() []byte {
	return []byte(d.String())
}

// HasKey returns true if Data[key] exists.
// Note that his function does not support paths like `HasKey("foo.bar.baz")`.
// For that you can use [GetPath]
func (d Data) HasKey(key string) bool {
	if _, ok := d[key]; ok {
		return true
	}
	return false
}

// Get returns the value at `Data[key]` as [Data].
// Note that his function does not support paths like `HasKey("foo.bar.baz")`.
// For that you can use [GetPath]
func (d Data) Get(key string) Data {
	if d[key] == nil {
		return nil
	}
	return d[key].(Data)
}

// GetPath allows path based indexing into Data.
// A path is a slice of interfaces which are used as keys in order.
// Supports array indexing (arrays start at 0)
// Examples of valid paths:
//	- ["foo", "bar"]
//	- ["foo", "bar", 0]
func (d Data) GetPath(path ...interface{}) (tree interface{}, err error) {
	tree = d

	for i, el := range path {
		switch node := tree.(type) {
		case Data:
			key, ok := el.(string)
			if !ok {
				return nil, fmt.Errorf("unexpected string key in map[string]interface '%T' at index %d", el, i)
			}
			tree, ok = node[key]
			if !ok {
				return nil, fmt.Errorf("key not found: %v", el)
			}

		case map[interface{}]interface{}:
			var ok bool
			tree, ok = node[el]
			if !ok {
				return nil, fmt.Errorf("key not found: %v", el)
			}

		case []interface{}:
			index, ok := el.(int)
			if !ok {
				index, err = strconv.Atoi(fmt.Sprint(el))
				if err != nil {
					return nil, fmt.Errorf("unexpected integer path element '%v' (%T)", el, el)
				}
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

// SetPath uses the same path slices as [GetPath], only that it can set the value at the given path.
// Supports array indexing (arrays start at 0)
func (d *Data) SetPath(value interface{}, path ...interface{}) (err error) {
	var tree interface{}
	tree = (*d)

	if len(path) == 0 {
		return fmt.Errorf("path cannot be empty")
	}

	i := len(path) - 1
	if len(path) > 1 {
		var tmp interface{}
		tmp, err = tree.(Data).GetPath(path[:i]...)
		if err != nil {
			return err
		}
		tree = tmp
	}

	element := path[i]

	switch node := tree.(type) {
	case Data:
		key, ok := element.(string)
		if !ok {
			return fmt.Errorf("unexpected string key in map[string]interface '%T' at index %d", element, i)
		}
		node[key] = value

	case []interface{}:
		index, ok := element.(int)
		if !ok {
			index, err = strconv.Atoi(fmt.Sprint(element))
			if err != nil {
				return fmt.Errorf("unexpected integer path element '%v (%T)'", element, element)
			}
		}
		if index < 0 || index >= len(node) {
			return fmt.Errorf("path index out of range: %d", index)
		}
		node[index] = value

	default:
		return fmt.Errorf("unexpected node type %T at index %d", node, i)
	}

	return nil
}

// MergeReplace merges the existing Data with the given.
// If a key already exists, the passed data has precedence and it's value will be used.
func (d Data) MergeReplace(data Data) Data {
	out := make(Data, len(d))
	for k, v := range d {
		out[k] = v
	}
	for k, v := range data {
		if v, ok := v.(Data); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(Data); ok {
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

// FindValueFunc is a callback used to find values inside a Data map.
// `value` is the actual found value; `path` are the path segments which point to that value
// The function returns the extracted value and an error (if any).
type FindValueFunc func(value string, path []interface{}) (interface{}, error)

// FindValues can be used to find specific 'leaf' nodes, aka values.
// The Data is iterated recursively and once a plain value is found, the given FindValueFunc is called.
// It's the responsibility of the FindValueFunc to determine if the value is what is searched for.
// The FindValueFunc can return any data, which is aggregated and written into the passed `*[]interface{}`.
// The callee is then responsible of handling the returned value and ensuring the correct types were returned.
func (d Data) FindValues(valueFunc FindValueFunc, target *[]interface{}) (err error) {
	// newPath is used to copy an existing []interface and hard-copy it.
	// This is required because Go wants to optimize slice usage by reusing memory.
	// Most of the time, this is totally fine, but in this case it would mess up the slice
	// by changing the path []interface of already found secrets.
	newPath := func(path []interface{}, appendValue interface{}) []interface{} {
		tmp := make([]interface{}, len(path))
		copy(tmp, path)
		tmp = append(tmp, appendValue)
		return tmp
	}

	var walk func(reflect.Value, []interface{}) error
	walk = func(v reflect.Value, path []interface{}) error {

		// fix indirects through pointers and interfaces
		for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
			v = v.Elem()
		}

		switch v.Kind() {
		case reflect.Array, reflect.Slice:
			for i := 0; i < v.Len(); i++ {
				walk(v.Index(i), newPath(path, i))
			}
		case reflect.Map:
			for _, key := range v.MapKeys() {
				if v.MapIndex(key).IsNil() {
					break
				}

				walk(v.MapIndex(key), newPath(path, key.String()))
			}
		default:
			// at this point we have found a value and give off control to the given valueFunc
			value, err := valueFunc(v.String(), path)
			if err != nil {
				return err
			}
			(*target) = append(*target, value)
		}
		return nil
	}
	err = walk(reflect.ValueOf(d), nil)

	return err
}
