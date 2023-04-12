package skipper_test

import (
	"testing"

	"github.com/lukasjarosch/skipper"
	"github.com/stretchr/testify/assert"
)

func TestNewData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       interface{}
		expected    skipper.Data
		errExpected bool
	}{
		{
			name:  "map[string]interface{}",
			input: map[string]interface{}{"foo": "bar"},
			expected: skipper.Data{
				"foo": "bar",
			},
			errExpected: false,
		},
		{
			name:  "map[interface{}]interface{}",
			input: map[interface{}]interface{}{"foo": "bar"},
			expected: skipper.Data{
				"foo": "bar",
			},
			errExpected: false,
		},
		{
			name:  "struct",
			input: struct{ Foo string }{"bar"},
			expected: skipper.Data{
				"foo": "bar",
			},
			errExpected: false,
		},
		{
			name:        "invalid input",
			input:       8,
			expected:    nil,
			errExpected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := skipper.NewData(tt.input)

			if tt.errExpected {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expected, actual)
		})
	}
}
