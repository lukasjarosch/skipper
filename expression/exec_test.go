package expression_test

import (
	"fmt"
	"os"
	"reflect"
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
			errExpected: expression.ErrIncompatibleArgType,
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
		{
			name:        "path as call arg",
			input:       `${to_upper(some:value)}`,
			variableMap: map[string]interface{}{},
			pathValues: map[string]interface{}{
				"some.value": "hello",
			},
			expected: "HELLO",
		},
		{
			name:        "call nesting",
			input:       `${to_screaming_kebab(to_snake(to_upper(some:value)))}`,
			variableMap: map[string]interface{}{},
			pathValues: map[string]interface{}{
				"some.value": "hello there",
			},
			expected: "HELLO-THERE",
		},
		{
			name:        "default func with empty int",
			input:       `${default(some:empty:path, 123)}`,
			variableMap: map[string]interface{}{},
			pathValues: map[string]interface{}{
				"some.empty.path": 0,
			},
			expected: 123,
		},
		{
			name:        "default func with empty float",
			input:       `${default(some:empty:path, 123)}`,
			variableMap: map[string]interface{}{},
			pathValues: map[string]interface{}{
				"some.empty.path": 0.0,
			},
			expected: 123,
		},
		{
			name:        "default func with empty string",
			input:       `${default(some:empty:path, 123)}`,
			variableMap: map[string]interface{}{},
			pathValues: map[string]interface{}{
				"some.empty.path": "",
			},
			expected: 123,
		},
		{
			name:        "default func with path as default value",
			input:       `${default(some:empty:path, some:not:empty:path)}`,
			variableMap: map[string]interface{}{},
			pathValues: map[string]interface{}{
				"some.empty.path":     "",
				"some.not.empty.path": "hello",
			},
			expected: "hello",
		},
		{
			name:  "default func with variable as default value",
			input: `${default(some:empty:path, $default)}`,
			variableMap: map[string]interface{}{
				"default": "henlo",
			},
			pathValues: map[string]interface{}{
				"some.empty.path": "",
			},
			expected: "henlo",
		},
		{
			name:  "default func with call as default value",
			input: `${default(some:empty:path, set_env("FOO_BAR", "henlo"))}`,
			variableMap: map[string]interface{}{
				"default": "henlo",
			},
			pathValues: map[string]interface{}{
				"some.empty.path": "",
			},
			expected: "henlo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expressions := expression.Parse(tt.input)
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
				assert.EqualValues(t, reflect.ValueOf(tt.expected).Interface(), val.Interface(), "values should be equal, have %v got %v", val, reflect.ValueOf(tt.expected))
			}
		})
	}
}

func makeIdentifierNode(value string) *expression.IdentifierNode {
	return &expression.IdentifierNode{
		Pos:      0,
		NodeType: expression.NodeIdentifier,
		Value:    value,
	}
}

func makeVariableNode(name string) *expression.VariableNode {
	return &expression.VariableNode{
		Pos:      0,
		NodeType: expression.NodeVariable,
		Name:     name,
	}
}

func makePathNode(segments []expression.Node) *expression.PathNode {
	return &expression.PathNode{
		Pos:      0,
		NodeType: expression.NodePath,
		Segments: segments,
	}
}

func TestResolveVariablePath(t *testing.T) {
	tests := []struct {
		name         string
		inputPath    *expression.PathNode
		inputVarMap  map[string]any
		expectedPath string
		errExpected  error
	}{
		{
			name:         "Path with only identifiers",
			inputPath:    makePathNode([]expression.Node{makeIdentifierNode("foo"), makeIdentifierNode("bar"), makeIdentifierNode("baz")}),
			expectedPath: "foo:bar:baz",
		},
		{
			name:         "Path with only one variable",
			inputPath:    makePathNode([]expression.Node{makeIdentifierNode("foo"), makeVariableNode("bar"), makeIdentifierNode("baz")}),
			expectedPath: "foo:bar:baz",
			inputVarMap: map[string]any{
				"bar": "bar",
			},
		},
		{
			name:      "Path with only complex variable",
			inputPath: makePathNode([]expression.Node{makeIdentifierNode("foo"), makeVariableNode("bar"), makeIdentifierNode("baz")}),
			inputVarMap: map[string]any{
				"bar": []struct{}{},
			},
			errExpected: fmt.Errorf("variable resolves to complex type"),
		},
		{
			name:      "Path with empty string variable",
			inputPath: makePathNode([]expression.Node{makeIdentifierNode("foo"), makeVariableNode("bar"), makeIdentifierNode("baz")}),
			inputVarMap: map[string]any{
				"bar": "",
			},
			errExpected: fmt.Errorf("variables in path cannot be empty"),
		},
		{
			name:      "Path with nil variable",
			inputPath: makePathNode([]expression.Node{makeIdentifierNode("foo"), makeVariableNode("bar"), makeIdentifierNode("baz")}),
			inputVarMap: map[string]any{
				"bar": nil,
			},
			errExpected: fmt.Errorf("variables in path cannot be empty"),
		},
		{
			name:      "Path with number variable",
			inputPath: makePathNode([]expression.Node{makeIdentifierNode("foo"), makeVariableNode("bar"), makeIdentifierNode("baz")}),
			inputVarMap: map[string]any{
				"bar": 12,
			},
			errExpected: fmt.Errorf("unexpected parsing error"),
		},
		{
			name:      "Path with float variable",
			inputPath: makePathNode([]expression.Node{makeIdentifierNode("foo"), makeVariableNode("bar"), makeIdentifierNode("baz")}),
			inputVarMap: map[string]any{
				"bar": 12.12,
			},
			errExpected: fmt.Errorf("unexpected parsing error"),
		},
		{
			name:      "Path with illegal characters in variable",
			inputPath: makePathNode([]expression.Node{makeIdentifierNode("foo"), makeVariableNode("bar"), makeIdentifierNode("baz")}),
			inputVarMap: map[string]any{
				"bar": ":*(",
			},
			errExpected: fmt.Errorf("unexpected parsing error"),
		},
		{
			name:      "Path with only variables",
			inputPath: makePathNode([]expression.Node{makeVariableNode("foo"), makeVariableNode("bar"), makeVariableNode("baz")}),
			inputVarMap: map[string]any{
				"foo": "foo",
				"bar": "bar",
				"baz": "ohai",
			},
			expectedPath: "foo:bar:ohai",
		},
		{
			name:         "Empty path",
			inputPath:    makePathNode([]expression.Node{}),
			expectedPath: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ret, err := expression.ResolveVariablePath(*tt.inputPath, tt.inputVarMap)

			if tt.errExpected != nil {
				assert.ErrorContains(t, err, tt.errExpected.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedPath, ret.Text())
		})
	}
}
