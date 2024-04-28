package skipper_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lukasjarosch/skipper"
	"github.com/lukasjarosch/skipper/data"
	mocks "github.com/lukasjarosch/skipper/mocks/skipper"
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
			name:  "ExpressionManager_ExecuteInput",
			input: "${john:name}",
			pathValues: map[string]data.Value{
				"john.name":  data.NewValue("${john:first} ${john:last}"),
				"john.first": data.NewValue("john"),
				"john.last":  data.NewValue("doe"),
			},
			expected: data.NewValue("john doe"),
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
				assert.ErrorIs(t, err, tt.errExpected)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected.Raw, ret.Raw)
		})
	}
}
