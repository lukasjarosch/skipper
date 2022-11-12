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

// ReplaceVariables searches and replaces variables defined in data.
// The classFiles are used for local referencing variables (class internal references).
// predefinedVariables can be used to provide global user-defined variables.
func ReplaceVariables(data Data, classFiles []*Class, predefinedVariables map[string]interface{}) (err error) {
	isPredefinedVariable := func(variable Variable) bool {
		for name := range predefinedVariables {
			if strings.EqualFold(variable.Name, name) {
				return true
			}
		}
		return false
	}

	// TODO: gosh, make this a standalone function already
	replaceVariable := func(variable Variable) error {
		var targetValue interface{}
		if isPredefinedVariable(variable) {
			targetValue = predefinedVariables[variable.Name]
		} else {
			// targetValue is the value on which the variable points to.
			// This is the value we need to replace the variable with
			targetValue, err = data.GetPath(variable.NameAsIdentifier()...)
			if err != nil {

				// for any other error than a 'key not found' there is nothing we can do
				if !strings.Contains(err.Error(), "key not found") {
					return fmt.Errorf("reference to invalid variable '%s': %w", variable.FullName(), err)
				}

				// Local variable handling
				//
				// at this point we have failed to resolve the variable using 'absolute' paths
				// but the variable may be only locally defined which means we need to change the lookup path.
				// We iterate over all classes and attempt to resolve the variable within that limited scope.
				for i, class := range classFiles {

					// if the value to which the variable points is valid inside the class scope, we just need to add the class identifier
					// if the combination works this means we have found ourselves a local variable and we can set the targetValue
					fullPath := []interface{}{}
					fullPath = append(fullPath, class.NameAsIdentifier()...)

					// edge case: the class root key is 'foo', and the variable used references it like ${foo:bar:baz}
					// this would result in the full path being 'foo foo bar baz', hence we need to strip the class name from the variable reference.
					if strings.EqualFold(class.RootKey(), variable.NameAsIdentifier()[0].(string)) {
						fullPath = append(fullPath, variable.NameAsIdentifier()[1:]...)
					} else {
						// default case: the class root key is not used in the variable, we can add the full variable identifier
						fullPath = append(fullPath, variable.NameAsIdentifier()...)
					}

					targetValue, err = data.GetPath(fullPath...)

					// as long as not all classes have been checked, we cannot be sure that the variable is undefined (aka. key not found error)
					if targetValue == nil &&
						i < len(classFiles) &&
						strings.Contains(err.Error(), "key not found") {
						continue
					}

					// the local variable is really not defined at this point
					if err != nil {
						return fmt.Errorf("reference to invalid variable '%s': %w", variable.FullName(), err)
					}

					break
				}
			}
		}

		// sourceValue is the value where the variable is located. It needs to be replaced with the 'targetValue'
		sourceValue, err := data.GetPath(variable.Identifier...)
		if err != nil {
			return err
		}

		// Replace the full variable name (${variable}) with the targetValue
		sourceValue = strings.ReplaceAll(fmt.Sprint(sourceValue), variable.FullName(), fmt.Sprint(targetValue))
		data.SetPath(sourceValue, variable.Identifier...)

		return nil
	}

	// Replace variables in an undefined amount of iterations.
	// This needs to be done because one variable can be replaced by another, which will only be replaced in the next iteration.
	var variables []Variable
	for {
		variables, err = FindVariables(data)
		if err != nil {
			return err
		}

		if len(variables) == 0 {
			break
		}

		for _, variable := range variables {
			err = replaceVariable(variable)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func replaceVariable(data Data, variable Variable, classFiles []*Class, predefinedVariables map[string]interface{}) (err error) {
	isPredefinedVariable := func(variable Variable) bool {
		for name := range predefinedVariables {
			if strings.EqualFold(variable.Name, name) {
				return true
			}
		}
		return false
	}

	var targetValue interface{}
	if isPredefinedVariable(variable) {
		targetValue = predefinedVariables[variable.Name]
		return nil
	} else {
		// targetValue is the value on which the variable points to.
		// This is the value we need to replace the variable with
		targetValue, err = data.GetPath(variable.NameAsIdentifier()...)
		if err != nil {

			// for any other error than a 'key not found' there is nothing we can do
			if !strings.Contains(err.Error(), "key not found") {
				return fmt.Errorf("reference to invalid variable '%s': %w", variable.FullName(), err)
			}

			// Local variable handling
			//
			// at this point we have failed to resolve the variable using 'absolute' paths
			// but the variable may be only locally defined which means we need to change the lookup path.
			// We iterate over all classes and attempt to resolve the variable within that limited scope.
			for i, class := range classFiles {

				// if the value to which the variable points is valid inside the class scope, we just need to add the class identifier
				// if the combination works this means we have found ourselves a local variable and we can set the targetValue
				fullPath := []interface{}{}
				fullPath = append(fullPath, class.NameAsIdentifier()...)

				// edge case: the class root key is 'foo', and the variable used references it like ${foo:bar:baz}
				// this would result in the full path being 'foo foo bar baz', hence we need to strip the class name from the variable reference.
				if strings.EqualFold(class.RootKey(), variable.NameAsIdentifier()[0].(string)) {
					fullPath = append(fullPath, variable.NameAsIdentifier()[1:]...)
				} else {
					// default case: the class root key is not used in the variable, we can add the full variable identifier
					fullPath = append(fullPath, variable.NameAsIdentifier()...)
				}

				targetValue, err = data.GetPath(fullPath...)

				// as long as not all classes have been checked, we cannot be sure that the variable is undefined (aka. key not found error)
				if targetValue == nil &&
					i < len(classFiles) &&
					strings.Contains(err.Error(), "key not found") {
					continue
				}

				// the local variable is really not defined at this point
				if err != nil {
					return fmt.Errorf("reference to invalid variable '%s': %w", variable.FullName(), err)
				}

				break
			}
		}
	}

	// sourceValue is the value where the variable is located. It needs to be replaced with the 'targetValue'
	sourceValue, err := data.GetPath(variable.Identifier...)
	if err != nil {
		return err
	}

	// Replace the full variable name (${variable}) with the targetValue
	sourceValue = strings.ReplaceAll(fmt.Sprint(sourceValue), variable.FullName(), fmt.Sprint(targetValue))
	data.SetPath(sourceValue, variable.Identifier...)

	return nil
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
