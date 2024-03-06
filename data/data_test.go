package data_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lukasjarosch/skipper/data"
)

func TestWalk(t *testing.T) {
	tests := []struct {
		name                   string
		input                  interface{}
		expectedTraversedPaths []data.Path
		leafsOnly              bool
	}{
		{
			name:                   "empty map",
			input:                  map[string]interface{}{},
			expectedTraversedPaths: []data.Path{},
		},
		{
			name: "simple map",
			input: map[string]interface{}{
				"foo": "bar",
				"baz": "qux",
			},
			expectedTraversedPaths: []data.Path{
				data.NewPath("foo"),
				data.NewPath("baz"),
			},
		},
		{
			name: "nested map",
			input: map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": map[string]interface{}{
						"baz": map[string]interface{}{
							"qux": "hello",
						},
					},
				},
			},
			expectedTraversedPaths: []data.Path{
				data.NewPath("foo"),
				data.NewPath("foo.bar"),
				data.NewPath("foo.bar.baz"),
				data.NewPath("foo.bar.baz.qux"),
			},
		},
		{
			name: "nested map with slices",
			input: map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": map[string]interface{}{
						"baz": []interface{}{
							"one",
							"two",
							"three",
						},
					},
				},
			},
			expectedTraversedPaths: []data.Path{
				data.NewPath("foo"),
				data.NewPath("foo.bar"),
				data.NewPath("foo.bar.baz"),
				data.NewPath("foo.bar.baz.0"),
				data.NewPath("foo.bar.baz.1"),
				data.NewPath("foo.bar.baz.2"),
			},
		},
		{
			name: "map with nested slices",
			input: map[string]interface{}{
				"foo": []interface{}{
					[]interface{}{
						"one",
					},
					[]interface{}{
						[]interface{}{
							"two",
						},
					},
					[]interface{}{
						[]interface{}{
							[]interface{}{
								"three",
							},
						},
					},
				},
			},
			expectedTraversedPaths: []data.Path{
				data.NewPath("foo"),
				data.NewPath("foo.0"),
				data.NewPath("foo.0.0"),
				data.NewPath("foo.1"),
				data.NewPath("foo.1.0"),
				data.NewPath("foo.1.0.0"),
				data.NewPath("foo.2"),
				data.NewPath("foo.2.0"),
				data.NewPath("foo.2.0.0"),
				data.NewPath("foo.2.0.0.0"),
			},
		},
		{
			name: "map with slices containing maps",
			input: map[string]interface{}{
				"foo": []interface{}{
					[]interface{}{
						map[string]interface{}{
							"one": "hello",
						},
					},
					[]interface{}{
						[]interface{}{
							map[string]interface{}{
								"two": "ohai",
							},
						},
					},
					[]interface{}{
						[]interface{}{
							[]interface{}{
								map[string]interface{}{
									"three": "welcome",
								},
							},
						},
					},
				},
			},
			expectedTraversedPaths: []data.Path{
				data.NewPath("foo"),
				data.NewPath("foo.0"),
				data.NewPath("foo.0.0"),
				data.NewPath("foo.0.0.one"),
				data.NewPath("foo.1"),
				data.NewPath("foo.1.0"),
				data.NewPath("foo.1.0.0"),
				data.NewPath("foo.1.0.0.two"),
				data.NewPath("foo.2"),
				data.NewPath("foo.2.0"),
				data.NewPath("foo.2.0.0"),
				data.NewPath("foo.2.0.0.0"),
				data.NewPath("foo.2.0.0.0.three"),
			},
		},
		{
			name:      "only leafs of map with slices containing maps",
			leafsOnly: true,
			input: map[string]interface{}{
				"foo": []interface{}{
					[]interface{}{
						map[string]interface{}{
							"one": "hello",
						},
					},
					[]interface{}{
						[]interface{}{
							map[string]interface{}{
								"two": "ohai",
							},
						},
					},
					[]interface{}{
						[]interface{}{
							[]interface{}{
								map[string]interface{}{
									"three": "welcome",
								},
							},
						},
					},
				},
			},
			expectedTraversedPaths: []data.Path{
				data.NewPath("foo.0.0.one"),
				data.NewPath("foo.1.0.0.two"),
				data.NewPath("foo.2.0.0.0.three"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// we need to store the paths in a map to prevent the 'append' behavior
			// where the last element is overwritten by the next
			pathMap := make(map[string]bool)
			err := data.Walk(tt.input, func(path data.Path, data interface{}, isLeaf bool) error {
				if tt.leafsOnly && !isLeaf {
					return nil
				}
				pathMap[path.String()] = true
				return nil
			})

			// convert back to slice
			paths := []data.Path{}
			for p := range pathMap {
				paths = append(paths, data.NewPath(p))
			}

			data.SortPaths(tt.expectedTraversedPaths)
			data.SortPaths(paths)

			assert.True(t, reflect.DeepEqual(tt.expectedTraversedPaths, paths))
			assert.NoError(t, err)
		})
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name        string
		data        interface{}
		key         string
		expected    interface{}
		errExpected bool
		err         error
	}{
		{
			name:        "nil data",
			data:        nil,
			key:         "foo",
			expected:    nil,
			errExpected: true,
			err:         data.ErrNilData,
		},
		{
			name:        "empty map",
			data:        map[string]interface{}{},
			key:         "foo",
			expected:    nil,
			errExpected: true,
			err:         data.ErrKeyNotFound,
		},
		{
			name:        "empty slice",
			data:        []interface{}{},
			key:         "0",
			expected:    nil,
			errExpected: true,
			err:         data.ErrKeyNotFound,
		},
		{
			name: "simple map",
			data: map[string]interface{}{
				"foo": "bar",
			},
			key:      "foo",
			expected: "bar",
		},
		{
			name: "simple map invalid key",
			data: map[string]interface{}{
				"foo": "bar",
			},
			key:         "invalid",
			expected:    nil,
			errExpected: true,
			err:         data.ErrKeyNotFound,
		},
		{
			name: "nested map",
			data: map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": map[string]interface{}{
						"baz": "hello",
					},
				},
			},
			key: "foo",
			expected: map[string]interface{}{
				"bar": map[string]interface{}{
					"baz": "hello",
				},
			},
		},
		{
			name: "nested map with numeric key",
			data: map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": map[string]interface{}{
						"baz": "hello",
					},
				},
			},
			key:         "1",
			expected:    nil,
			errExpected: true,
			err:         data.ErrKeyNotFound,
		},
		{
			name: "simple slice",
			data: []interface{}{
				1, 2, 3,
			},
			key:      "1",
			expected: 2,
		},
		{
			name: "simple slice key too large",
			data: []interface{}{
				1, 2, 3,
			},
			key:         "10",
			expected:    nil,
			errExpected: true,
			err:         data.ErrKeyNotFound,
		},
		{
			name: "simple slice key negative",
			data: []interface{}{
				1, 2, 3,
			},
			key:         "-10",
			expected:    nil,
			errExpected: true,
			err:         data.ErrKeyNotFound,
		},
		{
			name: "simple slice with non numeric key",
			data: []interface{}{
				1, 2, 3,
			},
			key:         "two",
			expected:    nil,
			errExpected: true,
			err:         data.ErrExpectedNumericArrayIndex,
		},
		{
			name: "nested slice",
			data: []interface{}{
				[]interface{}{1},
				[]interface{}{2},
				[]interface{}{3},
			},
			key:      "1",
			expected: []interface{}{2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			returned, err := data.Get(tt.data, tt.key)

			if tt.errExpected {
				assert.ErrorContains(t, err, tt.err.Error())
				assert.Nil(t, returned)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, returned)
		})
	}
}

func TestDeepGet(t *testing.T) {
	tests := []struct {
		name        string
		data        interface{}
		path        data.Path
		expected    interface{}
		errExpected bool
		err         error
	}{
		{
			name:        "nil data",
			data:        nil,
			path:        data.NewPath("foo.bar.baz"),
			expected:    nil,
			errExpected: true,
			err:         data.ErrInvalidValue,
		},
		{
			name: "empty path",
			data: map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": map[string]interface{}{
						"baz": "qux",
					},
				},
			},
			path: data.NewPath(""),
			expected: map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": map[string]interface{}{
						"baz": "qux",
					},
				},
			},
		},
		{
			name: "nested map",
			data: map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": map[string]interface{}{
						"baz": "qux",
					},
				},
			},
			path:     data.NewPath("foo.bar.baz"),
			expected: "qux",
		},
		{
			name: "nested map with slices",
			data: map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": map[string]interface{}{
						"baz": "qux",
					},
					"array": []interface{}{
						"ohai",
						map[string]interface{}{
							"chicken": "pizza",
						},
					},
				},
			},
			path:     data.NewPath("foo.array.1.chicken"),
			expected: "pizza",
		},
		{
			name: "deeply nested slice",
			data: []interface{}{
				[]interface{}{1},
				[]interface{}{
					[]interface{}{2},
				},
				[]interface{}{
					[]interface{}{
						[]interface{}{3},
					},
				},
			},
			path:     data.NewPath("2.0.0.0"),
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			returned, err := data.DeepGet(tt.data, tt.path)

			if tt.errExpected {
				assert.ErrorContains(t, err, tt.err.Error())
				assert.Nil(t, returned)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, returned)
		})
	}
}

func TestSet(t *testing.T) {
	tests := []struct {
		name          string
		input         interface{}
		key           string
		value         interface{}
		expectedAfter interface{}
		errExpected   bool
		err           error
	}{
		{
			name:          "nil value",
			input:         nil,
			key:           "foo",
			value:         "coolValue",
			expectedAfter: nil,
			errExpected:   true,
			err:           data.ErrNilData,
		},
		{
			name:  "empty map",
			input: map[string]interface{}{},
			key:   "foo",
			value: "coolValue",
			expectedAfter: map[string]interface{}{
				"foo": "coolValue",
			},
		},
		{
			name:  "empty slice",
			input: []interface{}{},
			key:   "2",
			value: "coolValue",
			expectedAfter: []interface{}{
				nil,
				nil,
				"coolValue",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			returned, err := data.Set(tt.input, tt.key, tt.value)

			if tt.errExpected {
				assert.ErrorContains(t, err, tt.err.Error())
				assert.Equal(t, tt.input, tt.input)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedAfter, returned)
		})
	}
}

func TestDeepSet(t *testing.T) {
	tests := []struct {
		name          string
		input         interface{}
		path          data.Path
		value         interface{}
		expectedAfter interface{}
		errExpected   bool
		err           error
	}{
		{
			name:  "nil value",
			input: nil,
			path:  data.NewPath("foo"),
			value: "coolValue",
			expectedAfter: map[string]interface{}{
				"foo": "coolValue",
			},
		},
		{
			name:        "scalar input",
			input:       5,
			path:        data.NewPath("foo"),
			value:       "coolValue",
			errExpected: true,
			err:         data.ErrTypeChange,
		},
		{
			name:        "struct input",
			input:       struct{}{},
			path:        data.NewPath("foo"),
			value:       "coolValue",
			errExpected: true,
			err:         data.ErrUnsupportedDataType,
		},
		{
			name:  "empty map",
			input: map[string]interface{}{},
			path:  data.NewPath("foo"),
			value: "coolValue",
			expectedAfter: map[string]interface{}{
				"foo": "coolValue",
			},
		},
		{
			name:  "empty slice",
			input: []interface{}{},
			path:  data.NewPath("2"),
			value: "coolValue",
			expectedAfter: []interface{}{
				nil,
				nil,
				"coolValue",
			},
		},
		{
			name: "simple map with nil",
			input: map[string]interface{}{
				"foo": nil,
			},
			path:  data.NewPath("foo.bar"),
			value: "coolValue",
			expectedAfter: map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": "coolValue",
				},
			},
		},
		{
			name: "nested map with nil scalar",
			input: map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": map[string]interface{}{
						"baz": nil,
					},
				},
			},
			path:  data.NewPath("foo.bar.baz"),
			value: "coolValue",
			expectedAfter: map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": map[string]interface{}{
						"baz": "coolValue",
					},
				},
			},
		},
		{
			name:  "empty map with nested simple path",
			input: map[string]interface{}{},
			path:  data.NewPath("foo.bar.baz"),
			value: "coolValue",
			expectedAfter: map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": map[string]interface{}{
						"baz": "coolValue",
					},
				},
			},
		},
		{
			name:  "empty map with array path",
			input: map[string]interface{}{},
			path:  data.NewPath("foo.1.bar.2.baz.3"),
			value: "coolValue",
			expectedAfter: map[string]interface{}{
				"foo": []interface{}{
					nil,
					map[string]interface{}{
						"bar": []interface{}{
							nil,
							nil,
							map[string]interface{}{
								"baz": []interface{}{
									nil,
									nil,
									nil,
									"coolValue",
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "empty map with nested array path",
			input: map[string]interface{}{},
			path:  data.NewPath("foo.1.bar.2.3.4.baz"),
			value: "coolValue",
			expectedAfter: map[string]interface{}{
				"foo": []interface{}{
					nil,
					map[string]interface{}{
						"bar": []interface{}{
							nil,
							nil,
							[]interface{}{
								nil,
								nil,
								nil,
								[]interface{}{
									nil,
									nil,
									nil,
									nil,
									map[string]interface{}{
										"baz": "coolValue",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "empty nested slice",
			input: []interface{}{[]interface{}{}},
			path:  data.NewPath("2.3"),
			value: "coolValue",
			expectedAfter: []interface{}{
				[]interface{}{},
				nil,
				[]interface{}{
					nil,
					nil,
					nil,
					"coolValue",
				},
			},
		},
		{
			name:          "nil input and empty path",
			input:         nil,
			path:          data.NewPath(""),
			value:         "coolValue",
			expectedAfter: nil,
			errExpected:   true,
			err:           data.ErrEmptyPath,
		},
		{
			name:  "non-empty path on nil input",
			input: nil,
			path:  data.NewPath("foo.bar"),
			value: "coolValue",
			expectedAfter: map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": "coolValue",
				},
			},
		},
		{
			name:          "empty slice with negative index",
			input:         []interface{}{},
			path:          data.NewPath("-1"),
			value:         "coolValue",
			expectedAfter: nil,
			errExpected:   true,
			err:           data.ErrNegativeIndex,
		},
		{
			name:          "empty map with negative path",
			input:         map[string]interface{}{},
			path:          data.NewPath("foo.-1.bar"),
			value:         "coolValue",
			expectedAfter: nil,
			errExpected:   true,
			err:           data.ErrNegativeIndex,
		},
		{
			name:          "nested array with negative path",
			input:         map[string]interface{}{},
			path:          data.NewPath("foo.1.bar.-2.baz"),
			value:         "coolValue",
			expectedAfter: nil,
			errExpected:   true,
			err:           data.ErrNegativeIndex,
		},
		{
			name:  "deep array nesting",
			input: []interface{}{},
			path:  data.NewPath("0.1.2.3.4.5"),
			value: "coolValue",
			expectedAfter: []interface{}{
				[]interface{}{
					nil,
					[]interface{}{
						nil,
						nil,
						[]interface{}{
							nil,
							nil,
							nil,
							[]interface{}{
								nil,
								nil,
								nil,
								nil,
								[]interface{}{
									nil,
									nil,
									nil,
									nil,
									nil,
									"coolValue",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "mixed data types",
			input: map[string]interface{}{
				"stringKey": "stringData",
				"intKey":    42,
				"mapKey": map[string]interface{}{
					"nestedKey": "nestedData",
				},
			},
			path:        data.NewPath("intKey.nestedKey"),
			value:       100,
			errExpected: true,
			err:         data.ErrTypeChange,
		},
		{
			name: "scalar to map change",
			input: map[string]interface{}{
				"foo": "oldValue",
			},
			path:  data.NewPath("foo"),
			value: "newValue",
			expectedAfter: map[string]interface{}{
				"foo": "newValue",
			},
		},
		{
			name: "array index at map",
			input: map[string]interface{}{
				"foo": "oldValue",
			},
			path:  data.NewPath("0"),
			value: "newValue",
			expectedAfter: map[string]interface{}{
				"foo": "oldValue",
				"0":   "newValue",
			},
		},
		{
			name: "map key at array",
			input: map[string]interface{}{
				"foo": []interface{}{
					nil,
					nil,
				},
			},
			path:        data.NewPath("foo.bar"),
			value:       "newValue",
			errExpected: true,
			err:         data.ErrTypeChange,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			returned, err := data.DeepSet(tt.input, tt.path, tt.value)

			if tt.errExpected {
				assert.ErrorContains(t, err, tt.err.Error())
				assert.Equal(t, tt.input, tt.input)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedAfter, returned)
		})
	}
}

func TestPaths_SelectAll(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name     string
		input    map[string]interface{}
		expected []data.Path
	}{
		{
			name: "SingleLevelMap",
			input: map[string]interface{}{
				"foo": "bar",
				"baz": "qux",
			},
			expected: []data.Path{
				data.NewPath("foo"),
				data.NewPath("baz"),
			},
		},
		{
			name: "NestedMap",
			input: map[string]interface{}{
				"parent": map[string]interface{}{
					"foo": "bar",
					"baz": "qux",
				},
			},
			expected: []data.Path{
				data.NewPath("parent"),
				data.NewPath("parent.foo"),
				data.NewPath("parent.baz"),
			},
		},
		{
			name: "ArrayWithNestedMap",
			input: map[string]interface{}{
				"items": []interface{}{
					map[string]interface{}{"name": "item1"},
					map[string]interface{}{"name": "item2"},
				},
			},
			expected: []data.Path{
				data.NewPath("items"),
				data.NewPath("items.0"),
				data.NewPath("items.0.name"),
				data.NewPath("items.1"),
				data.NewPath("items.1.name"),
			},
		},
		{
			name: "SliceOfSlices",
			input: map[string]interface{}{
				"slice": []interface{}{
					[]interface{}{
						"one", "two", "three",
					},
					[]interface{}{
						"four", "five", "six",
					},
				},
			},
			expected: []data.Path{
				data.NewPath("slice"),
				data.NewPath("slice.0"),
				data.NewPath("slice.0.0"),
				data.NewPath("slice.0.1"),
				data.NewPath("slice.0.2"),
				data.NewPath("slice.1"),
				data.NewPath("slice.1.0"),
				data.NewPath("slice.1.1"),
				data.NewPath("slice.1.2"),
			},
		},
		{
			name: "MapWithNestedArray",
			input: map[string]interface{}{
				"items": map[string]interface{}{
					"hello": map[string]interface{}{
						"slice": []interface{}{
							"one",
							map[string]interface{}{
								"foo": "bar",
							},
							[]interface{}{
								1, 2, 3,
							},
						},
					},
				},
			},
			expected: []data.Path{
				data.NewPath("items"),
				data.NewPath("items.hello"),
				data.NewPath("items.hello.slice"),
				data.NewPath("items.hello.slice.0"),
				data.NewPath("items.hello.slice.1"),
				data.NewPath("items.hello.slice.1.foo"),
				data.NewPath("items.hello.slice.2"),
				data.NewPath("items.hello.slice.2.0"),
				data.NewPath("items.hello.slice.2.1"),
				data.NewPath("items.hello.slice.2.2"),
			},
		},
		{
			name:     "EmptyMap",
			input:    map[string]interface{}{},
			expected: []data.Path{},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result := data.Paths(tt.input, data.SelectAllPaths)

			data.SortPaths(result)
			data.SortPaths(tt.expected)

			assert.True(t, reflect.DeepEqual(tt.expected, result))
		})
	}
}

func TestPaths_SelectLeaves(t *testing.T) {
	testCases := []struct {
		name     string
		input    map[string]interface{}
		expected []data.Path
	}{
		{
			name: "SingleLevelMap",
			input: map[string]interface{}{
				"foo": "bar",
				"baz": "qux",
			},
			expected: []data.Path{
				data.NewPath("foo"),
				data.NewPath("baz"),
			},
		},
		{
			name: "NestedMap",
			input: map[string]interface{}{
				"parent": map[string]interface{}{
					"foo": "bar",
					"baz": "qux",
				},
			},
			expected: []data.Path{
				data.NewPath("parent.foo"),
				data.NewPath("parent.baz"),
			},
		},
		{
			name: "ArrayWithNestedMap",
			input: map[string]interface{}{
				"items": []interface{}{
					map[string]interface{}{"name": "item1"},
					map[string]interface{}{"name": "item2"},
				},
			},
			expected: []data.Path{
				data.NewPath("items.0.name"),
				data.NewPath("items.1.name"),
			},
		},
		{
			name: "SliceOfSlices",
			input: map[string]interface{}{
				"slice": []interface{}{
					[]interface{}{
						"one", "two", "three",
					},
					[]interface{}{
						"four", "five", "six",
					},
				},
			},
			expected: []data.Path{
				data.NewPath("slice.0.0"),
				data.NewPath("slice.0.1"),
				data.NewPath("slice.0.2"),
				data.NewPath("slice.1.0"),
				data.NewPath("slice.1.1"),
				data.NewPath("slice.1.2"),
			},
		},
		{
			name: "MapWithNestedArray",
			input: map[string]interface{}{
				"items": map[string]interface{}{
					"hello": map[string]interface{}{
						"slice": []interface{}{
							"one",
							map[string]interface{}{
								"foo": "bar",
							},
							[]interface{}{
								1, 2, 3,
							},
						},
					},
				},
			},
			expected: []data.Path{
				data.NewPath("items.hello.slice.0"),
				data.NewPath("items.hello.slice.1.foo"),
				data.NewPath("items.hello.slice.2.0"),
				data.NewPath("items.hello.slice.2.1"),
				data.NewPath("items.hello.slice.2.2"),
			},
		},
		{
			name:     "EmptyMap",
			input:    map[string]interface{}{},
			expected: []data.Path{},
		},
	}

	// Run test cases
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result := data.Paths(tt.input, data.SelectLeafPaths)

			data.SortPaths(result)
			data.SortPaths(tt.expected)

			assert.True(t, reflect.DeepEqual(tt.expected, result))
		})
	}
}

func TestMerge(t *testing.T) {
	tests := []struct {
		name     string
		initial  map[string]interface{}
		input    map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name:     "merge empty map",
			initial:  map[string]interface{}{"foo": "bar"},
			input:    map[string]interface{}{},
			expected: map[string]interface{}{"foo": "bar"},
		},
		{
			name:     "overwrite existing map key",
			initial:  map[string]interface{}{"foo": "bar"},
			input:    map[string]interface{}{"foo": "new_value"},
			expected: map[string]interface{}{"foo": "new_value"},
		},
		{
			name:     "add map key",
			initial:  map[string]interface{}{"foo": "bar"},
			input:    map[string]interface{}{"baz": "qux"},
			expected: map[string]interface{}{"foo": "bar", "baz": "qux"},
		},
		{
			name:     "simple nested map",
			initial:  map[string]interface{}{"parent": map[string]interface{}{"foo": "bar"}},
			input:    map[string]interface{}{"parent": map[string]interface{}{"baz": "qux"}},
			expected: map[string]interface{}{"parent": map[string]interface{}{"foo": "bar", "baz": "qux"}},
		},
		{
			name:     "append slice",
			initial:  map[string]interface{}{"foo": []interface{}{1, 2, 3}},
			input:    map[string]interface{}{"foo": []interface{}{4, 5, 6}},
			expected: map[string]interface{}{"foo": []interface{}{1, 2, 3, 4, 5, 6}},
		},
		{
			name:     "append slice with mixed types",
			initial:  map[string]interface{}{"foo": []interface{}{1, 2, 3}},
			input:    map[string]interface{}{"foo": []interface{}{"four", 5.0, "six"}},
			expected: map[string]interface{}{"foo": []interface{}{1, 2, 3, "four", 5.0, "six"}},
		},
		{
			name:     "overwrite slice key with map",
			initial:  map[string]interface{}{"foo": []interface{}{1, 2, 3}},
			input:    map[string]interface{}{"foo": map[string]interface{}{"changed": "value"}},
			expected: map[string]interface{}{"foo": map[string]interface{}{"changed": "value"}},
		},
		{
			name:     "overwrite map key with slice",
			initial:  map[string]interface{}{"foo": map[string]interface{}{"changed": "value"}},
			input:    map[string]interface{}{"foo": []interface{}{1, 2, 3}},
			expected: map[string]interface{}{"foo": []interface{}{1, 2, 3}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := data.Merge(tt.initial, tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
