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

func TestData_GetPath(t *testing.T) {
	data := skipper.Data{
		"foo": skipper.Data{
			"bar": "baz",
		},
		"baz": []interface{}{
			"qux",
			skipper.Data{
				"quux": "corge",
			},
			map[string]interface{}{
				"hello": "world",
			},
		},
	}

	tests := []struct {
		path        string
		expected    interface{}
		errExpected bool
	}{
		{path: "foo.bar", expected: "baz", errExpected: false},
		{path: "baz.0", expected: "qux", errExpected: false},
		{path: "baz.1.quux", expected: "corge", errExpected: false},
		{path: "baz.2.hello", expected: "world", errExpected: false},
		{path: "baz.-2.hello", expected: nil, errExpected: true},
		{path: "baz.99f.hello", expected: nil, errExpected: true},
		{path: "baz.100.hello", expected: nil, errExpected: true},
		{path: "baz.2.notfound", expected: nil, errExpected: true},
		{path: "baz.1.notfound", expected: nil, errExpected: true},
		{path: "foo", expected: data["foo"], errExpected: false},
		{path: "notfound", expected: nil, errExpected: true},
		{path: "", expected: data, errExpected: false},
	}

	for _, tt := range tests {
		actual, err := data.GetPath(skipper.P(tt.path))

		if tt.errExpected {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}

		assert.Equal(t, tt.expected, actual)
	}
}

func TestData_UnmarshalPath(t *testing.T) {
	data := skipper.Data{
		"foo": skipper.Data{
			"bar": "baz",
		},
		"baz": []interface{}{
			"qux",
			skipper.Data{
				"quux": "corge",
			},
			map[string]interface{}{
				"hello": "world",
			},
		},
	}

	type MyStruct struct {
		Something string `yaml:"bar"`
	}

	tests := []struct {
		name        string
		path        string
		target      interface{}
		expected    interface{}
		errExpected bool
	}{
		{
			name:        "valid path and target",
			path:        "foo",
			target:      &MyStruct{},
			expected:    &MyStruct{Something: "baz"},
			errExpected: false,
		},
		{
			name:        "target not a pointer",
			path:        "foo",
			target:      MyStruct{},
			expected:    MyStruct{},
			errExpected: true,
		},
		{
			name:        "invalid path",
			path:        "notfound",
			target:      &MyStruct{},
			expected:    &MyStruct{},
			errExpected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := data.UnmarshalPath(skipper.P(tt.path), tt.target)

			if tt.errExpected {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expected, tt.target)
		})
	}
}
