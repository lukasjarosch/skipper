package skipper

import (
	"fmt"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Data map[string]interface{}

func (d Data) String() string {
	out, _ := yaml.Marshal(d)
	return string(out)
}

func (d Data) HasKey(k string) bool {
	if _, ok := d[k]; ok {
		return true
	}
	return false
}

func (d Data) Get(k string) Data {
	if d[k] == nil {
		return nil
	}
	return d[k].(Data)
}

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
