package skipper

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// callRegex matches everything between `%{....}`
	callRegex = regexp.MustCompile(`\%\{(.+)\}`)

	// callActionRegex matches the actual call syntax `function:param`
	callActionRegex = regexp.MustCompile(`(\w+)(\:(\w+))?`)

	ErrEmptyFunctionName error = fmt.Errorf("empty function name")
)

type Call struct {
	// Identifier points to wherever the call is used in the [Data] map
	Identifier   []interface{}
	FunctionName string
	Param        string
}

func NewCall(functionName string, param string, path []interface{}) (*Call, error) {
	if functionName == "" {
		return nil, ErrEmptyFunctionName
	}

	return &Call{
		Identifier:   path,
		FunctionName: functionName,
		Param:        param,
	}, nil
}

func FindCalls(data Data) ([]*Call, error) {
	var foundValues []interface{}
	err := data.FindValues(findCallFunc(), &foundValues)
	if err != nil {
		return nil, err
	}

	var foundCalls []*Call
	for _, val := range foundValues {
		calls, ok := val.([]*Call)
		if !ok {
			return nil, fmt.Errorf("unexpected error during call detection, file a bug report")
		}
		foundCalls = append(foundCalls, calls...)
	}

	return foundCalls, nil
}

func findCallFunc() FindValueFunc {
	return func(value string, path []interface{}) (interface{}, error) {
		var calls []*Call
		matches := callRegex.FindAllStringSubmatch(value, -1)
		for _, match := range matches {
			// matches should be a slice with two values. we're interested in the second
			if len(match[0]) > 1 {

				// now we can use the second regex to extract the desired parts of the call
				segments := callActionRegex.FindAllStringSubmatch(match[1], -1)

				// if len of the matches is not at least 1, we did not match and can continue
				for _, call := range segments {
					function := call[1]
					param := call[3]

					newFunc, err := NewCall(function, param, path)
					if err != nil {
						return nil, err
					}

					calls = append(calls, newFunc)
				}
			}

		}

		return calls, nil
	}
}

func (c Call) Path() string {
	var segments []string
	for _, seg := range c.Identifier {
		segments = append(segments, fmt.Sprint(seg))
	}
	return strings.Join(segments, ".")
}
