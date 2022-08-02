package skipper

import (
	"reflect"
	"regexp"
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

type Variable struct {
	FullName string
	Name     string
}

func (d Data) FindVariables() (variables []Variable) {
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
							FullName: variable[0],
							Name:     variable[1],
						})
					}
				}
			}
		}
	}
	walk(reflect.ValueOf(d))

	return variables
}
