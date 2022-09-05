package skipper

import (
	"fmt"
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

func (v Variable) Path() string {
	var segments []string
	for _, seg := range v.Identifier {
		segments = append(segments, fmt.Sprint(seg))
	}
	return strings.Join(segments, ".")
}

func (v Variable) NameAsIdentifier() (id []interface{}) {
	tmp := strings.Split(v.Name, ":")
	id = make([]interface{}, len(tmp))

	for i := 0; i < len(tmp); i++ {
		id[i] = tmp[i]
	}
	return id
}

// FindVariables leverages the [FindValues] function of the given Data to extract
// all variables by using the [variableFindValueFunc] as callback.
func FindVariables(data Data) ([]Variable, error) {
	var foundValues []interface{}
	err := data.FindValues(variableFindValueFunc(), &foundValues)
	if err != nil {
		return nil, err
	}

	var foundVariables []Variable
	for _, val := range foundValues {

		// variableFindValueFunc returns []Variable so we need to ensure that matches
		vars, ok := val.([]Variable)
		if !ok {
			return nil, fmt.Errorf("unexpected error during variable detection, file a bug report")
		}

		foundVariables = append(foundVariables, vars...)
	}

	return foundVariables, nil
}

// variableFindValueFunc implements the [FindValueFunc] and searches for variables inside [Data].
// Variables are extracted by matching the values to the [variableRegex].
// All found variables are initialized and added to the output.
// The function returns `[]Variable`.
func variableFindValueFunc() FindValueFunc {
	return func(value string, path []interface{}) (interface{}, error) {
		var variables []Variable

		matches := variableRegex.FindAllStringSubmatch(value, -1)
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
		return variables, nil
	}
}
