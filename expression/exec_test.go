package expression_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lukasjarosch/skipper/data"
	"github.com/lukasjarosch/skipper/expression"
	mock "github.com/lukasjarosch/skipper/mocks/expression"
)

var (
	envVar      = "WHY_DID_YOU_SET_THIS"
	envVarValue = "lol"
)

func init() {
	os.Setenv(envVar, envVarValue)
}

func TestExecuteExpression(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		pathValues  map[string]interface{}
		variableMap map[string]interface{}
		funcMap     map[string]interface{}
		expected    interface{}
		errExpected error
	}{
		{
			name:  "single path without variables",
			input: `${foo:bar:baz}`,
			pathValues: map[string]interface{}{
				"foo.bar.baz": "HELLO",
			},
			expected: "HELLO",
		},
		{
			name:  "single path with single variable",
			input: `${foo:bar:$target_name}`,
			variableMap: map[string]interface{}{
				"target_name": "develop",
			},
			pathValues: map[string]interface{}{
				"foo.bar.develop": "HELLO",
			},
			expected: "HELLO",
		},
		{
			name:  "single path with multiple variables",
			input: `${foo:$name:$target_name}`,
			variableMap: map[string]interface{}{
				"name":        "bar",
				"target_name": "develop",
			},
			pathValues: map[string]interface{}{
				"foo.bar.develop": "HELLO",
			},
			expected: "HELLO",
		},
		{
			name:  "standalone variable expression",
			input: `${target_name}`,
			variableMap: map[string]interface{}{
				"target_name": "develop",
			},
			expected: "develop",
		},
		{
			name:  "standalone inline variable expression",
			input: `${$target_name}`,
			variableMap: map[string]interface{}{
				"target_name": "develop",
			},
			expected: "develop",
		},
		{
			name:        "undefined variable",
			input:       `${foo_bar}`,
			variableMap: map[string]interface{}{},
			pathValues:  map[string]interface{}{},
			funcMap:     map[string]interface{}{},
			errExpected: expression.ErrUndefinedVariable,
		},
		{
			name:        "undefined function",
			input:       `${say_hello()}`,
			variableMap: map[string]interface{}{},
			pathValues:  map[string]interface{}{},
			funcMap:     map[string]interface{}{},
			errExpected: expression.ErrFunctionNotDefined,
		},
		{
			name:        "call user func with too many args",
			input:       `${say_hello("foo", "bar")}`,
			variableMap: map[string]interface{}{},
			pathValues:  map[string]interface{}{},
			funcMap: map[string]interface{}{
				"say_hello": func() string { return "Hello there" },
			},
			errExpected: expression.ErrCallInvalidArgumentCount,
		},
		{
			name:        "call user func with wrong return types",
			input:       `${say_hello()}`,
			variableMap: map[string]interface{}{},
			pathValues:  map[string]interface{}{},
			funcMap: map[string]interface{}{
				"say_hello": func() (string, int) { return "", 0 },
			},
			errExpected: expression.ErrBadFuncSignature,
		},
		{
			name:        "not a function in funcMap",
			input:       `${say_hello()}`,
			variableMap: map[string]interface{}{},
			pathValues:  map[string]interface{}{},
			funcMap: map[string]interface{}{
				"say_hello": "i am invalid",
			},
			errExpected: expression.ErrNotAFunc,
		},
		{
			name:        "call user defined function with no args",
			input:       `${say_hello()}`,
			variableMap: map[string]interface{}{},
			pathValues:  map[string]interface{}{},
			funcMap: map[string]interface{}{
				"say_hello": func() string { return "hello" },
			},
			expected: "hello",
		},
		{
			name:        "call user defined function with string argument",
			input:       `${say_hello("john")}`,
			variableMap: map[string]interface{}{},
			pathValues:  map[string]interface{}{},
			funcMap: map[string]interface{}{
				"say_hello": func(name string) string { return fmt.Sprintf("hello, %s", name) },
			},
			expected: "hello, john",
		},
		{
			name:        "call user defined function with invalid argument type",
			input:       `${say_hello("1337")}`,
			variableMap: map[string]interface{}{},
			pathValues:  map[string]interface{}{},
			funcMap: map[string]interface{}{
				"say_hello": func(count int) string {
					return fmt.Sprintf("hello, %d", count)
				},
			},
			errExpected: fmt.Errorf("expected integer; found String"),
		},
		{
			name:        "get_env builtin with variable not set",
			input:       `${get_env("THIS_CANNOT_POSSIBLY_BE_SET_COME_ON")}`,
			variableMap: map[string]interface{}{},
			pathValues:  map[string]interface{}{},
			errExpected: fmt.Errorf("environment variable not set"),
		},
		{
			name:        "get_env builtin with variable set",
			input:       `${get_env("WHY_DID_YOU_SET_THIS")}`, // set via init()
			variableMap: map[string]interface{}{},
			pathValues:  map[string]interface{}{},
			expected:    envVarValue,
		},
		{
			name:     "get_env builtin with variable not set but set_env alternative expression",
			input:    `${get_env("THIS_CANNOT_POSSIBLY_BE_SET_COME_ON") || set_env("THIS_CANNOT_POSSIBLY_BE_SET_COME_ON", "lol")}`,
			expected: envVarValue,
		},
		{
			name:  "set_env with variable arguments",
			input: `${set_env($MY_VAR, $value)}`,
			variableMap: map[string]interface{}{
				"MY_VAR": "THIS_CANNOT_POSSIBLY_BE_SET_COME_ON_NOW_REALLY",
				"value":  envVarValue,
			},
			pathValues: map[string]interface{}{},
			expected:   envVarValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expressions, err := expression.Parse(tt.input)
			assert.NoError(t, err)
			assert.NotEmpty(t, expressions)

			pathValueProvider := mock.NewMockPathValueProvider(t)

			for _, expr := range expressions {

				for path, val := range tt.pathValues {
					pathValueProvider.EXPECT().GetPath(data.NewPath(path)).Return(val, nil)
				}

				val, err := expression.Execute(expr, pathValueProvider, tt.variableMap, tt.funcMap)

				if tt.errExpected != nil {
					assert.ErrorContains(t, err, tt.errExpected.Error())
					return
				}

				assert.NoError(t, err)
				assert.Equal(t, tt.expected, val.String())
			}
		})
	}
}
