package skipper

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	// callRegex matches everything between `%{....}`
	callRegex = regexp.MustCompile(`\%\{(.+)\}`)

	// callActionRegex matches the actual call syntax `function:param`
	callActionRegex = regexp.MustCompile(`(\w+)(\:(\w+))?`)

	callFuncMap = map[string]CallFunc{
		"env": func(param string) string {
			out := os.Getenv(param)
			if len(out) == 0 {
				return "UNDEFINED"
			}
			return out
		},
		"randomstring": func(param string) string {
			const defaultLength = 32

			var length int
			if param == "" {
				length = defaultLength
			}

			length, err := strconv.Atoi(param)
			if err != nil {
				length = defaultLength
			}

			const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-_"
			ret := make([]byte, length)
			for i := 0; i < length; i++ {
				num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
				if err != nil {
					return err.Error()
				}
				ret[i] = letters[num.Int64()]
			}

			return string(ret)
		},
	}

	ErrEmptyFunctionName error = fmt.Errorf("empty function name")
)

type CallFunc func(param string) string

type Call struct {
	// Identifier points to wherever the call is used in the [Data] map
	Identifier   []interface{}
	FunctionName string
	Param        string
	callback     CallFunc
}

func NewCall(functionName string, param string, path []interface{}) (*Call, error) {
	if functionName == "" {
		return nil, ErrEmptyFunctionName
	}

	if !validCallFunc(functionName) {
		return nil, fmt.Errorf("invalid call function '%s'", functionName)
	}

	return &Call{
		Identifier:   path,
		FunctionName: functionName,
		Param:        param,
		callback:     callFuncMap[strings.ToLower(functionName)],
	}, nil
}

func NewRawCall(callString string) (*Call, bool, error) {
	// now we can use the second regex to extract the desired parts of the call
	segments := callActionRegex.FindAllStringSubmatch(callString, -1)

	// if len of the matches is not at least 1, we did not match and can continue
	for _, call := range segments {
		function := call[1]
		param := call[3]

		if !validCallFunc(function) {
			return nil, false, fmt.Errorf("invalid call function '%s'", function)
		}

		return &Call{
			Identifier:   nil,
			FunctionName: function,
			Param:        param,
			callback:     callFuncMap[strings.ToLower(function)],
		}, true, nil
	}

	return nil, false, nil
}

func (c *Call) RawString() string {
	if len(c.Param) == 0 {
		return c.FunctionName
	}
	return fmt.Sprintf("%s:%s", c.FunctionName, c.Param)
}

func (c *Call) Execute() string {
	return c.callback(c.Param)
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

func (c Call) FullName() string {
	if len(c.Param) == 0 {
		return "%" + fmt.Sprintf("{%s}", c.FunctionName)
	}
	return "%" + fmt.Sprintf("{%s:%s}", c.FunctionName, c.Param)
}

func (c Call) Path() string {
	var segments []string
	for _, seg := range c.Identifier {
		segments = append(segments, fmt.Sprint(seg))
	}
	return strings.Join(segments, ".")
}

func validCallFunc(funcName string) bool {
	if _, exists := callFuncMap[strings.ToLower(funcName)]; exists {
		return true
	}
	return false
}
