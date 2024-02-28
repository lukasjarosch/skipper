package data_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/lukasjarosch/skipper/data"
)

func TestNewContainer_Valid(t *testing.T) {
	containerName := "test"
	d := map[string]interface{}{
		containerName: map[string]interface{}{
			"foo": map[string]interface{}{
				"bar": map[string]interface{}{
					"baz": "hello",
				},
			},
		},
	}

	container, err := NewContainer(containerName, d)
	assert.NoError(t, err)
	assert.NotNil(t, container)
}

func TestNewContainer_EmptyContainerName(t *testing.T) {
	_, err := NewContainer("", nil)
	assert.Error(t, err, ErrEmptyRootKeyName)
}

func TestNewContainer_NilData(t *testing.T) {
	_, err := NewContainer("name", nil)
	assert.ErrorIs(t, err, ErrNilData)
}

func TestNewContainer_EmptyData(t *testing.T) {
	_, err := NewContainer("name", map[string]interface{}{})
	assert.ErrorIs(t, err, ErrNoRootKey)
}

func TestNewContainer_MultipleRootKeys(t *testing.T) {
	_, err := NewContainer("name", map[string]interface{}{
		"foo": map[string]interface{}{},
		"bar": map[string]interface{}{},
		"baz": map[string]interface{}{},
	})
	assert.ErrorIs(t, err, ErrMultipleRootKeys)
}

func TestNewContainer_InvalidRootKey(t *testing.T) {
	expectedError := fmt.Errorf("invalid root key")
	data := map[string]interface{}{
		"invalidRootKey": map[string]interface{}{
			"foo": "bar",
		},
	}

	_, err := NewContainer("name", data)
	assert.ErrorContains(t, err, expectedError.Error())
}

func TestContainer_Get(t *testing.T) {
	containerName := "test"
	defaultMap := map[string]interface{}{
		containerName: map[string]interface{}{
			"foo": map[string]interface{}{
				"bar": map[string]interface{}{
					"baz": "hello",
				},
			},
		},
	}

	tests := []struct {
		test          string
		data          map[string]interface{}
		path          Path
		errExpected   bool
		err           error
		valueExpected Value
	}{
		{
			test:          "empty path",
			data:          defaultMap,
			path:          NewPath(""),
			valueExpected: NewValue(defaultMap),
		},
		{
			test:          "invalid path",
			data:          defaultMap,
			path:          NewPathVar(containerName, "invalidKey"),
			errExpected:   true,
			err:           ErrPathNotFound{Path: NewPathVar(containerName, "invalidKey")},
			valueExpected: Value{},
		},
		{
			test:          "valid path with container name",
			data:          defaultMap,
			path:          NewPathVar(containerName, "foo"),
			errExpected:   false,
			valueExpected: NewValue(defaultMap[containerName].(map[string]interface{})["foo"]),
		},
		{
			test:          "valid path without container name",
			data:          defaultMap,
			path:          NewPathVar("foo"),
			errExpected:   false,
			valueExpected: NewValue(defaultMap[containerName].(map[string]interface{})["foo"]),
		},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			container, err := NewContainer(containerName, defaultMap)
			assert.NoError(t, err)

			// test
			value, err := container.GetPath(tt.path)
			if tt.errExpected {
				assert.ErrorContains(t, err, tt.err.Error())
				assert.Equal(t, Value{}, value)
			}
			assert.Equal(t, tt.valueExpected, value)
		})
	}
}

func TestContainer_LeafPaths(t *testing.T) {
	containerName := "test"
	d := map[string]interface{}{
		containerName: map[string]interface{}{
			"foo": map[string]interface{}{
				"bar": map[string]interface{}{
					"baz": "hello",
					"qux": "ohai",
				},
			},
		},
	}
	container, err := NewContainer(containerName, d)
	assert.NoError(t, err)
	assert.NotNil(t, container)
}

func TestContainer_Set(t *testing.T) {
	containerName := "test"
	d := map[string]interface{}{
		containerName: map[string]interface{}{
			"foo": map[string]interface{}{
				"bar": map[string]interface{}{
					"baz": "hello",
					"qux": "ohai",
				},
			},
			"array": []interface{}{
				[]interface{}{"one"},
				[]interface{}{
					[]interface{}{
						"two",
					},
				},
			},
		},
	}
	container, err := NewContainer(containerName, d)
	assert.NoError(t, err)
	assert.NotNil(t, container)

	err = container.SetPath(NewPath("test.array.1.0"), []interface{}{1, 2, 3})
}

func TestContainer_Merge(t *testing.T) {
	containerName := "test"
	d := map[string]interface{}{
		containerName: map[string]interface{}{
			"foo": map[string]interface{}{
				"bar": map[string]interface{}{
					"baz": "hello",
					"qux": "ohai",
				},
			},
			"array": []interface{}{
				[]interface{}{"one"},
				[]interface{}{
					[]interface{}{
						"two",
					},
				},
			},
		},
	}
	container, err := NewContainer(containerName, d)
	assert.NoError(t, err)
	assert.NotNil(t, container)

	mergeData := map[string]interface{}{
		"array": []interface{}{
			[]interface{}{"one", "two"},
			[]interface{}{
				[]interface{}{
					"three", 4, 5,
				},
			},
		},
	}

	_ = mergeData
	err = container.Merge(NewPath("test"), mergeData)
	if err != nil {
		panic(err)
	}
}

func TestContainerAbsolutePath(t *testing.T) {
	containerName := "test"
	data := map[string]interface{}{
		containerName: map[string]interface{}{
			"foo": map[string]interface{}{
				"bar": map[string]interface{}{
					"baz": "hello",
					"qux": "ohai",
				},
			},
			"array": []interface{}{
				[]interface{}{"one"},
				[]interface{}{
					[]interface{}{
						"two",
					},
				},
			},
		},
	}

	tests := []struct {
		name     string
		input    Path
		expected Path
		err      error
	}{
		{
			name:     "nil path",
			input:    nil,
			expected: nil,
			err:      ErrEmptyPath,
		},
		{
			name:     "empty path",
			input:    Path{},
			expected: nil,
			err:      ErrEmptyPath,
		},
		{
			name:     "already absolute path",
			input:    NewPath("test.foo.bar.baz"),
			expected: NewPath("test.foo.bar.baz"),
			err:      nil,
		},
		{
			name:     "valid relative path",
			input:    NewPath("foo.bar.baz"),
			expected: NewPath("test.foo.bar.baz"),
			err:      nil,
		},
		{
			name:     "path is container name",
			input:    NewPath(containerName),
			expected: NewPath(containerName),
			err:      nil,
		},
	}

	container, err := NewContainer(containerName, data)
	assert.NoError(t, err)
	assert.NotNil(t, container)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			abs, err := container.AbsolutePath(tt.input)

			if tt.err != nil {
				assert.ErrorIs(t, err, tt.err)
				assert.Nil(t, abs)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, abs)
			}
		})
	}
}
