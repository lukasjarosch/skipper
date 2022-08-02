package skipper

import (
	"reflect"
	"regexp"
	"strings"
)

var variableRegex = regexp.MustCompile(`\$\{((\w*)(\:\w+)*)\}`)

type Data map[string]interface{}

func (d Data) HasKey(k string) bool {
	if _, ok := d[k]; ok {
		return true
	}
	return false
}

func (d Data) Get(k string) Data {
	return d[k].(Data)
}

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

// Variable is a keyword which self-references the Data map it is defined in.
// A Variable has the form ${key:key}.
type Variable struct {
	// FullName is the whole variable name, including the delimiters (${foo:bar})
	Name string
	// Identifier is the actual variable identifier (foo:bar) which points to the referenced value
	Identifier string
}

// Variables recursively iterates over the data to find any leaf values which match the variableRegex
// Any value matching the regex is extracted and returned as Variable
func (d Data) Variables() (variables []Variable) {
	var walk func(reflect.Value)
	walk = func(v reflect.Value) {
		// fix indirects through pointers and interfaces
		for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
			v = v.Elem()
		}

		switch v.Kind() {
		case reflect.Array, reflect.Slice:
			for i := 0; i < v.Len(); i++ {
				walk(v.Index(i))
			}
		case reflect.Map:
			for _, key := range v.MapKeys() {
				if v.MapIndex(key).IsNil() {
					break
				}
				walk(v.MapIndex(key))
			}
		default:
			matches := variableRegex.FindAllStringSubmatch(v.String(), -1)
			if len(matches) > 0 {
				for _, variable := range matches {
					if len(variable) >= 2 {
						variables = append(variables, Variable{
							Name:       variable[0],
							Identifier: variable[1],
						})
					}
				}
			}
		}
	}
	walk(reflect.ValueOf(d))

	return variables
}

// VariableValue attempts to fetch the variable identifier (foo:bar) from Data and returns it as interface type.
func (d Data) VariableValue(variable Variable) interface{} {
	if len(variable.Identifier) == 0 {
		return nil
	}

	variableKeys := strings.Split(variable.Identifier, ":")
	data := d

	for i, val := range variableKeys {
		if i == len(variableKeys)-1 {
			return data[val]
		}
		tmp := make(Data)

		if _, ok := data[val]; !ok {
			return nil
		}

		m, ok := data[val].(Data)
		if !ok {
			return nil
		}

		for key, value := range m {
			tmp[key] = value
		}
		data = tmp
	}

	return nil
}
