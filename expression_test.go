package skipper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lukasjarosch/skipper/v1"
	"github.com/lukasjarosch/skipper/v1/data"
	mocks "github.com/lukasjarosch/skipper/v1/mocks/skipper"
)

func TestNewExpressionManager(t *testing.T) {
	tests := []struct {
		name        string
		pathValues  map[string]data.Value
		variables   map[string]any
		errExpected error
	}{
		{
			name:      "NewExpressionManager",
			variables: map[string]any{},
			pathValues: map[string]data.Value{
				"foo.bar":  data.NewValue("hello ${foo:name}"),
				"foo.name": data.NewValue("asdf"),
				"foo.baz":  data.NewValue("Some ${a:b}, ${c:d} and ${e:f}"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := mocks.NewMockPathValueSource(t)
			source.EXPECT().Values().Return(tt.pathValues)

			varSource := mocks.NewMockVariableSource(t)
			varSource.EXPECT().GetAll().Return(tt.variables)

			manager, err := skipper.NewExpressionManager(source, varSource)

			if tt.errExpected != nil {
				assert.ErrorIs(t, err, tt.errExpected)
				return
			}

			assert.NoError(t, err)
			_ = manager
		})
	}
}

func TestExpressionManager_ExecuteInput(t *testing.T) {
	tests := []struct {
		name        string
		pathValues  map[string]data.Value
		variables   map[string]any
		input       string
		expected    data.Value
		errExpected error
	}{
		{
			name:  "Simple valid path",
			input: "This is ${project:name}",
			pathValues: map[string]data.Value{
				"project.name": data.NewValue("skipper"),
			},
			expected: data.NewValue("This is skipper"),
		},
		{
			name:  "Simple path with dependencies",
			input: "Hello there, ${john:name}",
			pathValues: map[string]data.Value{
				"john.name":  data.NewValue("${to_upper(john:first)} ${to_lower(john:last)}"),
				"john.first": data.NewValue("john"),
				"john.last":  data.NewValue("doe"),
			},
			expected: data.NewValue("Hello there, JOHN doe"),
		},
		{
			name:  "Function call",
			input: "This is ${to_upper(project:name)}",
			pathValues: map[string]data.Value{
				"project.name": data.NewValue("skipper"),
			},
			expected: data.NewValue("This is SKIPPER"),
		},
		{
			name:  "Function call with two params",
			input: `This is ${replace(project:name, "skipper", "peter")}`,
			pathValues: map[string]data.Value{
				"project.name": data.NewValue("skipper"),
			},
			expected: data.NewValue("This is peter"),
		},
		{
			name:  "List function call with valid string list",
			input: `This is ${first(project:names)}`,
			pathValues: map[string]data.Value{
				"project.names": data.NewValue([]string{"skipper", "bob"}),
			},
			expected: data.NewValue("This is skipper"),
		},
		{
			name:  "List function call with valid int list",
			input: `This is ${first(project:names)}`,
			pathValues: map[string]data.Value{
				"project.names": data.NewValue([]int{1, 2, 3, 4}),
			},
			expected: data.NewValue("This is 1"),
		},
		{
			name:  "List function call with invalid list",
			input: `This is ${first(project:names)}`,
			pathValues: map[string]data.Value{
				"project.names": data.NewValue("def-not-a-list"),
			},
			errExpected: fmt.Errorf("failed to convert data.Value to list"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := mocks.NewMockPathValueSource(t)
			source.EXPECT().Values().Return(tt.pathValues)

			varSource := mocks.NewMockVariableSource(t)
			varSource.EXPECT().GetAll().Return(tt.variables)

			for path := range tt.pathValues {
				source.EXPECT().GetPath(data.NewPath(path)).Return(tt.pathValues[path], nil)
			}
			source.EXPECT().GetPath(data.NewPath("temporary-vertex-hash")).Return(data.NewValue(tt.input), nil)

			exprManager, err := skipper.NewExpressionManager(source, varSource)
			assert.NoError(t, err)

			ret, err := exprManager.ExecuteInput(tt.input)

			if tt.errExpected != nil {
				assert.ErrorContains(t, err, tt.errExpected.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected.Raw, ret.Raw)
		})
	}
}
