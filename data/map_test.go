package data_test

import (
	"reflect"
	"testing"

	. "github.com/lukasjarosch/skipper/data"
	"github.com/stretchr/testify/assert"
)

func TestMapWalk(t *testing.T) {
	testCases := []struct {
		name         string
		input        Map
		expectedKeys []string
	}{
		{
			name: "Complex Map",
			input: Map{
				"foo": 1,
				"bar": Map{
					"baz": 2,
				},
				"list": []interface{}{3, 4, Map{"qux": 5}},
			},
			expectedKeys: []string{
				"foo",
				"bar.baz",
				"list.0",
				"list.1",
				"list.2.qux",
			},
		},
		{
			name:         "Empty Map",
			input:        Map{},
			expectedKeys: []string{},
		},
		{
			name: "Map with Nested Empty Maps",
			input: Map{
				"foo": Map{},
				"bar": Map{
					"baz": Map{},
				},
			},
			expectedKeys: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var keys []string
			err := tc.input.Walk(func(value interface{}, path Path) error {
				keys = append(keys, path.String())
				return nil
			})

			if err != nil {
				t.Errorf("Walk returned an error: %v", err)
			}

			assert.ElementsMatch(t, keys, tc.expectedKeys, "Expected keys %v, but got %v", tc.expectedKeys, keys)
		})
	}
}

func TestGetPath(t *testing.T) {
	data := Map{
		"foo": 1,
		"bar": Map{
			"baz": 2,
		},
		"list": []interface{}{3, 4, Map{"qux": 5}},
	}

	tests := []struct {
		path     Path
		expected interface{}
	}{
		{NewPath("foo"), 1},
		{NewPath("bar.baz"), 2},
		{NewPath("list.0"), 3},
		{NewPath("list.2.qux"), 5},
		{NewPath("notfound"), nil},
		{NewPath("bar.notfound"), nil},
		{NewPath("list.3"), nil},
	}

	for _, test := range tests {
		t.Run(test.path.String(), func(t *testing.T) {
			actual, err := data.Get(test.path)
			if err != nil {
				if test.expected != nil {
					t.Errorf("Expected %v, but got an error: %v", test.expected, err)
				}
				return
			}
			if !reflect.DeepEqual(actual, test.expected) {
				t.Errorf("Expected %v, but got %v", test.expected, actual)
			}
		})
	}
}

func TestSetPath(t *testing.T) {
	testCases := []struct {
		name                string
		inputMap            Map
		path                Path
		value               interface{}
		expectedMap         Map
		errExpected         bool
		errExpectedContains string
	}{
		{
			name:        "Set a value at an existing path",
			inputMap:    Map{"foo": Map{"bar": "old"}},
			path:        NewPathVar("foo", "bar"),
			value:       "new",
			expectedMap: Map{"foo": Map{"bar": "new"}},
			errExpected: false,
		},
		{
			name:        "Set a value at a non-existing path",
			inputMap:    Map{"foo": Map{}},
			path:        NewPathVar("foo", "bar"),
			value:       "new",
			expectedMap: Map{"foo": Map{"bar": "new"}},
			errExpected: false,
		},
		{
			name:                "Attempt to set a value at an invalid path",
			inputMap:            Map{"foo": "value"},
			path:                NewPathVar("foo", "bar"),
			value:               "new",
			expectedMap:         Map{"foo": "value"},
			errExpected:         true,
			errExpectedContains: "cannot set path which creates a child segment on an existing value path",
		},
		{
			name:        "Set a value at a deeply nested path",
			inputMap:    Map{"a": Map{"b": Map{"c": Map{"d": "old"}}}},
			path:        NewPathVar("a", "b", "c", "d"),
			value:       "new",
			expectedMap: Map{"a": Map{"b": Map{"c": Map{"d": "new"}}}},
			errExpected: false,
		},
		{
			name:                "Attempt to set a value at an empty path",
			inputMap:            Map{"foo": "old"},
			path:                NewPath(""),
			value:               "new",
			expectedMap:         Map{"foo": "old"},
			errExpected:         true,
			errExpectedContains: "empty path",
		},
		{
			name:        "Attempt to set a value with an empty path segment",
			inputMap:    Map{"foo": "old"},
			path:        NewPathVar("foo", ""),
			value:       "new",
			expectedMap: Map{"foo": "new"},
			errExpected: false,
		},
		{
			name:                "Attempt to set a value at a non-existent path segment",
			inputMap:            Map{"foo": Map{}},
			path:                NewPathVar("foo", "bar", "baz"),
			value:               "new",
			expectedMap:         Map{"foo": Map{}},
			errExpected:         true,
			errExpectedContains: "cannot set path which creates more than one new path segment",
		},
		{
			name:                "Attempt to set a value at an invalid path segment with error",
			inputMap:            Map{"foo": "value"},
			path:                NewPathVar("foo", "bar"),
			value:               "new",
			expectedMap:         Map{"foo": "value"},
			errExpected:         true,
			errExpectedContains: "cannot set path which creates a child segment on an existing value path",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.inputMap.Set(tc.path, tc.value)

			if !assert.Equal(t, tc.expectedMap, tc.inputMap) {
				t.Errorf("Expected map %v, got %v", tc.expectedMap, tc.inputMap)
			}

			if !tc.errExpected {
				assert.NoError(t, err)
				return
			}

			assert.Error(t, err)
			if !assert.ErrorContains(t, err, tc.errExpectedContains) {
				t.Errorf("Expected error %v, got %v", tc.errExpectedContains, err)
			}
		})
	}
}
