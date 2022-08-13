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
	// Identifier is the dot-notation which points to the variable itself (e.g. "classname.foo.bar")
	Identifier string
}

func (v Variable) NameAsIdentifier() string {
	return strings.ReplaceAll(v.Name, ":", ".")
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

// Variables recursively iterates over the data to find any leaf values which match the variableRegex.
func FindVariables(data any) (variables VariableList) {

	var walk func(reflect.Value, string)
	walk = func(v reflect.Value, identifier string) {

		// fix indirects through pointers and interfaces
		for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
			v = v.Elem()
		}

		switch v.Kind() {
		case reflect.Array, reflect.Slice:
			for i := 0; i < v.Len(); i++ {
				walk(v.Index(i), fmt.Sprintf("%s.%d", identifier, i))
			}
		case reflect.Map:
			for _, key := range v.MapKeys() {
				if v.MapIndex(key).IsNil() {
					break
				}
				walk(v.MapIndex(key), fmt.Sprintf("%s.%s", identifier, key))
			}
		default:
			// Here we've arrived at actual values, hence we can check whether the value is a variable
			matches := variableRegex.FindAllStringSubmatch(v.String(), -1)
			if len(matches) > 0 {
				for _, variable := range matches {
					if len(variable) >= 2 {
						variables = append(variables, Variable{
							Name:       variable[1],
							Identifier: strings.TrimLeft(identifier, "."),
						})
					}
				}
			}
		}
	}
	walk(reflect.ValueOf(data), "")

	return variables
}
