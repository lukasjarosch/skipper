package skipper

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"
)

// valid variables: ${foo:bar} ${foo:bar:baz} ${something}
// invalid variables: ${foo:} ${bar::} ${:bar}
var variableRegex = regexp.MustCompile(`\$\{((\w*)(\:\w+)*)\}`)

// Variable is a keyword which self-references the Data map it is defined in.
// A Variable has the form ${key:key}.
type Variable struct {
	// Name of the variable is whatever string is between ${}.
	// + For dynamic variables, this can be a ':' separated string which points somewhere into the Data map.
	// 	 The reason we use ':' is to improve readability between curly braces.
	// + For predefined variables, this can be any string and must not be a path into the Data map.
	Name string
	// Identifier is the list of keys which point to the variable itself within the data set in which it is used.
	Identifier []interface{}
}

func (v Variable) FullName() string {
	return fmt.Sprintf("${%s}", v.Name)
}

func (v Variable) NameAsIdentifier() (id []interface{}) {
	tmp := strings.Split(v.Name, ":")
	id = make([]interface{}, len(tmp))

	for i := 0; i < len(tmp); i++ {
		id[i] = tmp[i]
	}
	return id
}

type VariableList []Variable

// FindVariables recursively iterates over the data to find any leaf values which match the variableRegex.
func FindVariables(data any) (variables VariableList) {

	// newPath is used to copy an existing []interface and hard-copy it.
	// This is required because Go wants to optimize slice usage by reusing memory.
	// Most of the time, this is totally find, but in this case it would mess up the slice
	// by rewriting variables already stored in the slice.
	newPath := func(path []interface{}, appendValue interface{}) []interface{} {
		tmp := make([]interface{}, len(path))
		copy(tmp, path)
		tmp = append(tmp, appendValue)
		return tmp
	}

	var walk func(reflect.Value, []interface{})
	walk = func(v reflect.Value, path []interface{}) {

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
			// Here we've arrived at actual values, hence we can check whether the value is a variable
			matches := variableRegex.FindAllStringSubmatch(v.String(), -1)
			if len(matches) > 0 {
				for _, variable := range matches {
					if len(variable) >= 2 {
						variables = append(variables, Variable{
							Name:       variable[1],
							Identifier: path,
						})
						log.Println("found variable", variable[0], "at", path)
					}
				}
			}
		}
	}
	walk(reflect.ValueOf(data), nil)

	for _, v := range variables {
		log.Println("loaded variable", v.FullName(), "at", v.Identifier)
	}

	return variables
}
