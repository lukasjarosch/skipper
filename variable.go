package skipper

import (
	"fmt"
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
	// Name is the ':' separated identifier which points to the value, we consider this the name of the variable.
	// The reason we use ':' is to improve readability between curly braces.
	Name string
	// Identifier is the list of keys which point to the variable itself within the data set it is defined.
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

// Variables recursively iterates over the data to find any leaf values which match the variableRegex.
func FindVariables(data any) (variables VariableList) {

	var walk func(reflect.Value, []interface{})
	walk = func(v reflect.Value, path []interface{}) {

		// fix indirects through pointers and interfaces
		for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
			v = v.Elem()
		}

		switch v.Kind() {
		case reflect.Array, reflect.Slice:
			for i := 0; i < v.Len(); i++ {
				walk(v.Index(i), append(path, i))
			}
		case reflect.Map:
			for _, key := range v.MapKeys() {
				if v.MapIndex(key).IsNil() {
					break
				}
				walk(v.MapIndex(key), append(path, key.String()))
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
					}
				}
			}
		}
	}
	walk(reflect.ValueOf(data), nil)

	return variables
}
