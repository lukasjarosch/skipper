package skipper

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"
)

var variableRegex = regexp.MustCompile(`\$\{((\w*)(\:\w+)*)\}`)

// Variable is a keyword which self-references the Data map it is defined in.
// A Variable has the form ${key:key}.
type Variable struct {
	// FullName is the whole variable name, including the delimiters (${foo:bar})
	Name string
	// Identifier is the actual variable identifier (foo:bar) which points to the referenced value
	Identifier string
}

type VariableList []Variable

func (vs VariableList) Deduplicate() VariableList {
	found := make(map[string]bool)
	newList := VariableList{}

	for _, v := range vs {
		if _, exists := found[v.Identifier]; !exists {
			found[v.Identifier] = true
			newList = append(newList, v)
		}
	}

	return newList
}

// Variables recursively iterates over the data to find any leaf values which match the variableRegex
// Any value matching the regex is extracted and returned as Variable
func FindVariables(data any) (variables VariableList) {

	// TODO: check if we've already loaded the variables, if so, return the cache

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
			// Here we've arrived at actual values, hence we can check whether the value is a variable
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
	walk(reflect.ValueOf(data))

	return variables
}

func (variable Variable) FindIdentifiersNew(data Data) []string {
	var identifiers [][]string
	_ = identifiers

	obj := data

	for k, v := range obj {
		switch v.(type) {
		case Data:
			log.Println("DATA", k, v)
			obj = Data(obj[k].(Data))

		default:
			log.Println("LEAF", k, v)
		}
	}

	return nil
}

func (variable Variable) FindIdentifiers(data any) []string {

	log.Println("==> FINDING IDENTIFIERS FOR", variable.Identifier)

	variableIdentifiers := [][]string{}

	var walk func(reflect.Value, *[][]string, int)
	walk = func(v reflect.Value, identifier *[][]string, index int) {
		// fix indirects through pointers and interfaces
		for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
			v = v.Elem()
		}

		switch v.Kind() {
		case reflect.Array, reflect.Slice:
			for i := 0; i < v.Len(); i++ {
				if identifier == nil {
					break
				}
				(*identifier)[index] = append((*identifier)[index], fmt.Sprint(i))
				walk(v.Index(i), identifier, index+1)
			}
		case reflect.Map:
			for _, key := range v.MapKeys() {
				if len((*identifier)) == 0 {
					(*identifier) = append((*identifier), []string{})
				}

				// for every discovered map key, append a new identifier slice with it's value up to the current index
				// if we're at 'foo.bar' a new key (index+1) with 'foo.bar' is added to the identifier array
				if len(*identifier) <= index+1 {
					(*identifier) = append((*identifier), (*identifier)[index])
				}

				(*identifier)[index+1] = append((*identifier)[index+1], key.String())
				walk(v.MapIndex(key), identifier, index+1)

			}
		default:
			// Here we've arrived at actual values, hence we can check whether the value is a variable
			if strings.Contains(v.String(), variable.Name) {
				log.Println(index, (*identifier)[index])
				//(*identifier)[index] = append((*identifier)[index], strings.Join((*identifier)[index], "."))
				index += 1
				//(*identifier) = append((*identifier), []string{})
				log.Println("DONE")
				return
			} else {
				(*identifier)[index] = []string{} // reset identifier, leaf node is not a variable
				index = 0
			}
		}
	}

	log.Println("NEW")
	walk(reflect.ValueOf(data), &variableIdentifiers, 0)

	for _, id := range variableIdentifiers {
		log.Println("DISCOVERED:", strings.Join(id, "."))
	}

	return nil
}

/*
func ReplaceVariable(variable Variable) error {
	data := d

	var walk func(data Data) Data
	walk = func(data Data) Data {
		val := reflect.ValueOf(data)

		// fix indirects through pointers and interfaces
		for val.Kind() == reflect.Ptr || val.Kind() == reflect.Interface {
			val = val.Elem()
		}

		switch val.Kind() {
		case reflect.Array, reflect.Slice:
			log.Println("ARRAY")
		case reflect.Map:
			for _, key := range val.MapKeys() {
				if data[val.MapIndex(key).String()].(Data) == nil {
					break
				}
				log.Println("ABASFASFD:", val.MapIndex(key).Elem().Interface().(Data))
				walk(data[val.MapIndex(key).String()].(Data))
			}

		default:
			log.Println("CANSET", val.CanSet())
		}

			switch v.Elem().Kind() {
			case reflect.Array, reflect.Slice:
				for i := 0; i < v.Elem().Len(); i++ {
					walk(reflect.Indirect(v).Index(i))
				}
			case reflect.Map:
				for _, key := range v.Elem().MapKeys() {
					if v.Elem().MapIndex(key).IsNil() {
						break
					}
					walk(v.Elem().MapIndex(key))
				}
			default:
				v.SetString("abc")
				log.Println(v.Elem())
				if v.CanSet() {
					log.Println("LEAF", v)

				}
			}
		return data
	}
	*d = walk(*data)
	return nil
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
*/
