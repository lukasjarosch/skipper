package skipper

import (
	"fmt"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Data is an arbitrary map of values which makes up the inventory.
type Data map[string]interface{}

// NewData attempts to convert any given interface{} into [Data].
// This is done by first using `yaml.Marshal` and then `yaml.Unmarshal`.
// If the given interface is compatible with [Data], these steps will succeed.
//
// Note that the process of marshalling will convert struct fields (which become keys in the Data map) to lowercase.
// Given the following type:
//
//	type MyStruct struct {
//		Name string
//	}
//
// 	example := MyStruct{ Name: "HelloWorld" }
//	data := NewData(example)
//
// The above data will look as follows:
//
//	Data {
//		"name": "HelloWorld"
//	}
//
// For the most part, this is something which will not have a big impact.
// But if you wish to change the keys, you can use the yaml structtags.
// The above example:
//
//	type MyStruct struct {
//		Name string `yaml:"Name"`
//	}
//
// Will result in the following data, just as expected:
//
//	Data {
//		"Name": "HelloWorld"
//	}
func NewData(input interface{}) (Data, error) {
	if input == nil {
		return make(Data), nil
	}

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
//
// Note that if the value key points to cannot be converted to [Data], nil is returned
func (d Data) Get(key string) Data {
	if d[key] == nil {
		return nil
	}

	if ret, ok := d[key].(Data); ok {
		return ret
	}

	return nil
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
				return nil, fmt.Errorf("unexpected key type (%T) for Data at index %d", el, i)
			}
			tree, ok = node[key]
			if !ok {
				return nil, fmt.Errorf("key not found: %v", el)
			}

		case []interface{}:
			index, ok := el.(int)
			if !ok {
				index, err = strconv.Atoi(fmt.Sprint(el))
				if err != nil {
					return nil, fmt.Errorf("unexpected key type (%T) for []interface{} at index %d", el, i)
				}
			}
			if index < 0 || index >= len(node) {
				return nil, fmt.Errorf("index out of range: %d", index)
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
