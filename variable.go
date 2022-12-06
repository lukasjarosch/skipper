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

func (v Variable) IsPredefined(predefinedVariables map[string]any) bool {
	for name := range predefinedVariables {
		if strings.EqualFold(v.Name, name) {
			return true
		}
	}
	return false
}

func (v *Variable) Replace(data Data, predefinedVariables map[string]any, classFiles []*Class) error {

	// Variable referencing shadowed value
	//
	// Find out whether the variables path is valid in the data AND one class-data.
	// In that case, Skipper would attempt to replace the variable with the target data
	// which results in the variable attempting to replace itself instead of
	// with the data inside the class (shadowing).
	// This is different from recursive variables because there is an actual
	// value which would be correct, it is only shadowed.
	//
	// Example:
	// ```
	// ---
	// myclass:
	//	self_reference: "bar"
	// ---
	// target:
	// 	myclass:
	//		self_reference: ${myclass:self_reference}
	// ```
	// Here, `myclass:self_reference` is a valid identifier for two keys.
	// Once on a class scope (myclass) and once on the main target scope.
	// Because the main target scope is evaluated first, before attempting a local reference,
	// Skipper would attempt to replace `${myclass:self_reference}` with `${myclass:self_reference}`
	// instead of replacing it with the value `bar`.
	//
	// If a variable is referencing a value in a class, which would be overwritten by
	// the target, we need to ensure that the variable is replaced with the class-data.
	isReferencingShadowedValue := func(variable Variable) (*Class, bool) {
		for _, class := range classFiles {

			classData, err := class.Data().GetPath(variable.Identifier...)
			if err != nil {
				continue
			}

			variableTargetData, err := class.Data().GetPath(variable.NameAsIdentifier()...)
			if err != nil {
				continue
			}

			if classData != nil && variableTargetData != nil {
				return class, true
			}
		}
		return nil, false
	}

	// Find recursive variables. A variable is recursive if it points to itself
	// and there is no class-key which also resolves to the path of the variable.
	//
	// Example:
	// ```
	// ---
	// target:
	// 	myclass:
	//		self_reference: ${myclass:self_reference} <- RECURIVE VARIABLE
	// ```
	// Note that there is no `myclass` which might provide the correct value.
	isRecursiveVariable := func(variable *Variable) bool {

		// a variable can only be recursive, if it is NOT referencing a shadowed value
		if _, shadowed := isReferencingShadowedValue(*variable); shadowed {
			return false
		}

		variableNameAsPath := strings.ReplaceAll(variable.Name, ":", ".")

		// if either the variableName has the variable Path as prefix or vice versa, we can be sure
		// that the variable is recursive, because we know that the variable does not reference a class.
		if strings.HasPrefix(variableNameAsPath, variable.Path()) || strings.HasPrefix(variable.Path(), variableNameAsPath) {
			return true
		}

		return false
	}

	// targetValue is the value with which is assigned to the given variable
	// all occurrences of `variable` will be replaced with its `targetValue`.
	// Given the below example:
	///
	/// ```
	/// ---
	/// myclass:
	///		something:
	///			foo: "bar"
	/// ---
	/// anotherclass:
	///		test: "something_${myclass:something:foo}"
	/// ```
	/// then targetValue would be `bar`
	var targetValue interface{}

	var err error
	if class, ok := isReferencingShadowedValue(*v); ok {
		targetValue, err = class.Data().GetPath(v.NameAsIdentifier()...)
		if err != nil {
			return err
		}
	} else {

		// variable could still be recursive
		if isRecursiveVariable(v) {
			return fmt.Errorf("recursive variable usage found at path '%s': %s", v.Path(), v.FullName())
		}

		// variable is not referencing a shadowed class value, default case
		targetValue, err = v.GetTargetValue(predefinedVariables, data, classFiles)
		if err != nil {
			return err
		}
	}

	// sourceValue is the value where the variable is used. It needs to be replaced with the 'targetValue'
	// If the variable us used in a conext like this:
	///
	/// ```
	/// ---
	/// myclass:
	///		something:
	///			foo: "bar"
	/// ---
	/// anotherclass:
	///		test: "something_${myclass:something:foo}"
	/// ```
	///
	/// then the sourceValue would be `something_${myclass:something:foo}`.
	sourceValue, err := data.GetPath(v.Identifier...)
	if err != nil {
		return err
	}

	// an inline variable is a variable which occurs with a context and not alone
	// inline variable: "foo ${my_variable} bar"
	// not inline: "${my_variable}"
	isInlineVariable := func() bool {
		return v.FullName() != sourceValue
	}

	// If the variable is not 'inline', we are going to 'attach' whatever the variable points to
	// with the variable. This allows you to import a list from a different class for example.
	// If this differentiation is not made, one could not import 'sub-structures' from classes
	// because they would be converted to a string, and not valid yaml.
	//
	// class-file:
	// ```
	//	myclass:
	//		foo:
	//			- somewhere
	//			- over
	//			- the rainbow
	// ```
	//
	// target file:
	// ```
	// target:
	// 		something: ${myclass:foo} // <-- Non-Inline import of the list under `myclass.foo`
	// 		something_else: "hello ${myclass:foo:2}" // <-- inline variable which will be 'string replaced'
	// ```
	if isInlineVariable() {
		sourceValue = strings.ReplaceAll(fmt.Sprint(sourceValue), v.FullName(), fmt.Sprint(targetValue))
	} else {
		sourceValue = targetValue
	}

	// replace variable in Data
	data.SetPath(sourceValue, v.Identifier...)

	return nil
}

func (v Variable) GetTargetValue(predefinedVariables map[string]any, data Data, classFiles []*Class) (targetValue interface{}, err error) {
	// in case the variable is in the predefinedVariables map,
	// getting it's targetValue is straight-forward.
	if v.IsPredefined(predefinedVariables) {
		targetValue = predefinedVariables[v.Name]
	} else {
		// targetValue is the value on which the variable points to.
		// This is the value we need to replace the variable with
		// In case the targetValue cannot be loaded (GetPath returns an error), it is very likely
		// that the variable is 'local referencing' a class which is why we will first try to
		// replace it as local variable, before actually returning an error.
		targetValue, err = data.GetPath(v.NameAsIdentifier()...)
		if err != nil {
			targetValue, err = v.getClassScopedTargetValue(data, classFiles)
			if err != nil {
				return nil, err
			}
		}
	}
	return targetValue, nil
}

// getClassScopedTargetValue returns the targetValue of `v` by assuming
// the variable identifier is a class-scoped identifier.
// Variables can be absolutely identified by referencing from the target-data `Data` map.
// But variables can also be scoped on a class-data `Data` map.
//
// ```
//	---
//	myclass:
//		foo:
//			bar: "baz"
//		class_scoped_local_variable: ${foo:bar}
// 	---
// 	anotherclass:
// 		absolute_variable: ${myclass:foo:bar}
// ```
func (v Variable) getClassScopedTargetValue(data Data, classFiles []*Class) (interface{}, error) {
	// Local referencing variable handling
	//
	// at this point we have failed to resolve the variable using 'absolute' paths
	// but the variable may be only locally defined which means we need to change the lookup path.
	// We iterate over all classes and attempt to resolve the variable within that limited scope.
	var targetValue interface{}
	for i, class := range classFiles {

		// if the value to which the variable points is valid inside the class scope, we just need to add the class identifier
		// if the combination works this means we have found ourselves a local variable and we can set the targetValue
		fullPath := []interface{}{}
		fullPath = append(fullPath, class.NameAsIdentifier()...)

		// edge case: the class root key is 'foo', and the variable used references it like ${foo:bar:baz}
		// this would result in the full path being 'foo foo bar baz', hence we need to strip the class name from the variable reference.
		if strings.EqualFold(class.RootKey(), v.NameAsIdentifier()[0].(string)) {
			fullPath = append(fullPath, v.NameAsIdentifier()[1:]...)
		} else {
			// default case: the class root key is not used in the variable, we can add the full variable identifier
			fullPath = append(fullPath, v.NameAsIdentifier()...)
		}

		targetValue, err := data.GetPath(fullPath...)
		if err != nil {
			return nil, err
		}

		// as long as not all classes have been checked, we cannot be sure that the variable is undefined (aka. key not found error)
		if targetValue == nil &&
			i < len(classFiles) &&
			strings.Contains(err.Error(), "key not found") {
			continue
		}

		// the local variable is really not defined at this point
		if err != nil {
			return nil, fmt.Errorf("reference to invalid variable '%s': %w", v.FullName(), err)
		}

		break
	}
	return targetValue, nil
}

// ReplaceVariables searches and replaces variables defined in data.
// The classFiles are used for local referencing variables (class internal references).
// predefinedVariables can be used to provide global user-defined variables.
func ReplaceVariables(data Data, classFiles []*Class, predefinedVariables map[string]interface{}) (err error) {

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

			err = variable.Replace(data, predefinedVariables, classFiles)
			if err != nil {
				return err
			}
		}
	}

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
