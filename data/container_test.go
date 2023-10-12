package data_test

import (
	"fmt"
	"testing"

	"github.com/lukasjarosch/skipper/codec"
	. "github.com/lukasjarosch/skipper/data"
	dataMocks "github.com/lukasjarosch/skipper/mocks/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewRawContainer_EmptyContainerName(t *testing.T) {
	_, err := NewRawContainer("", nil, nil)
	assert.Error(t, err, ErrEmptyContainerName)
}

func TestNewRawContainer_NilData(t *testing.T) {
	_, err := NewRawContainer("name", nil, nil)
	assert.ErrorIs(t, err, ErrNilData)
}

func TestNewRawContainer_EmptyData(t *testing.T) {
	_, err := NewRawContainer("name", Map{}, codec.NewYamlCodec())
	assert.NoError(t, err)
}

func TestNewRawContainer_NilCodec(t *testing.T) {
	_, err := NewRawContainer("name", Map{}, nil)
	assert.ErrorIs(t, err, ErrNilCodec)
}

func TestNewRawContainer_MarshalError(t *testing.T) {
	expectedError := fmt.Errorf("an error")

	mockCodec := &dataMocks.MockFileCodec{}
	mockCodec.On("Marshal", mock.Anything).Return([]byte{}, expectedError)

	_, err := NewRawContainer("name", Map{}, mockCodec)
	assert.Error(t, err, expectedError)
	mockCodec.AssertExpectations(t)
}

func TestNewRawContainer_UnmarshalError(t *testing.T) {
	expectedError := fmt.Errorf("an error")

	mockCodec := &dataMocks.MockFileCodec{}
	mockCodec.On("Marshal", mock.Anything).Return([]byte{}, nil)
	mockCodec.On("Unmarshal", mock.Anything).Return(Map{}, expectedError)

	_, err := NewRawContainer("name", Map{}, mockCodec)
	assert.ErrorIs(t, err, expectedError)
	mockCodec.AssertExpectations(t)
}

func TestNewRawContainer_InvalidRootKey(t *testing.T) {
	expectedError := fmt.Errorf("invalid root key")
	data := Map{
		"invalidRootKey": Map{
			"foo": "bar",
		},
	}

	mockCodec := &dataMocks.MockFileCodec{}
	mockCodec.On("Marshal", mock.Anything).Return([]byte{}, nil)
	mockCodec.On("Unmarshal", mock.Anything).Return(data, nil)

	_, err := NewRawContainer("name", data, mockCodec)
	assert.ErrorContains(t, err, expectedError.Error())
	mockCodec.AssertExpectations(t)
}

func TestNewRawContainer_Valid(t *testing.T) {
	data := Map{
		"name": Map{
			"foo": "bar",
		},
	}

	mockCodec := &dataMocks.MockFileCodec{}
	mockCodec.On("Marshal", mock.Anything).Return([]byte{}, nil)
	mockCodec.On("Unmarshal", mock.Anything).Return(data, nil)

	rawContainer, err := NewRawContainer("name", data, mockCodec)
	assert.NoError(t, err)
	assert.NotNil(t, rawContainer)
	mockCodec.AssertExpectations(t)
}

func TestRawContainer_Get(t *testing.T) {
	containerName := "test"
	defaultMap := Map{
		containerName: Map{
			"foo": Map{
				"bar": Map{
					"baz": "hello",
				},
			},
		},
	}

	tests := []struct {
		test          string
		data          Map
		path          Path
		errExpected   bool
		err           error
		valueExpected Value
	}{
		{
			test:        "empty path",
			data:        defaultMap,
			path:        NewPath(""),
			errExpected: true,
			err:         ErrEmptyPath,
		},
		{
			test:          "single wildcard path",
			data:          defaultMap,
			path:          NewPathVar(WildcardIdentifier),
			errExpected:   false,
			valueExpected: NewValue(defaultMap),
		},
		{
			test:          "wildcard path with container name",
			data:          defaultMap,
			path:          NewPathVar(containerName, WildcardIdentifier),
			errExpected:   false,
			valueExpected: NewValue(defaultMap[containerName]),
		},
		{
			test:          "nested wildcard path without container name",
			data:          defaultMap,
			path:          NewPathVar("foo", WildcardIdentifier),
			errExpected:   false,
			valueExpected: NewValue(defaultMap[containerName].(Map)["foo"]),
		},
		{
			test:          "nested wildcard path with container name",
			data:          defaultMap,
			path:          NewPathVar(containerName, "foo", WildcardIdentifier),
			errExpected:   false,
			valueExpected: NewValue(defaultMap[containerName].(Map)["foo"]),
		},
		{
			test:          "inline wildcard path",
			data:          defaultMap,
			path:          NewPathVar(containerName, WildcardIdentifier, "bar"),
			errExpected:   true,
			err:           ErrInlineWildcard,
			valueExpected: Value{},
		},
		{
			test:          "invalid wildcard path",
			data:          defaultMap,
			path:          NewPathVar(containerName, "invalidKey", WildcardIdentifier),
			errExpected:   true,
			err:           ErrPathNotFound{Path: NewPathVar(containerName, "invalidKey", WildcardIdentifier)},
			valueExpected: Value{},
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
			valueExpected: NewValue(defaultMap[containerName].(Map)["foo"]),
		},
		{
			test:          "valid path without container name",
			data:          defaultMap,
			path:          NewPathVar("foo"),
			errExpected:   false,
			valueExpected: NewValue(defaultMap[containerName].(Map)["foo"]),
		},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {

			// setup container
			mockCodec := &dataMocks.MockFileCodec{}
			mockCodec.On("Marshal", mock.Anything).Return([]byte{}, nil)
			mockCodec.On("Unmarshal", mock.Anything).Return(defaultMap, nil)
			container, err := NewRawContainer(containerName, defaultMap, mockCodec)
			assert.NoError(t, err)
			mockCodec.AssertExpectations(t)

			// test
			value, err := container.Get(tt.path)
			if tt.errExpected {
				assert.ErrorContains(t, err, tt.err.Error())
				assert.Equal(t, Value{}, value)
			}
			assert.Equal(t, tt.valueExpected, value)
		})
	}
}

func TestRawContainer_MustGet(t *testing.T) {
	containerName := "test"
	defaultMap := Map{
		containerName: Map{
			"foo": Map{
				"bar": Map{
					"baz": "hello",
				},
			},
		},
	}

	tests := []struct {
		test          string
		data          Map
		path          Path
		panicExpected bool
		valueExpected Value
	}{
		{
			test:          "empty path",
			data:          defaultMap,
			path:          NewPath(""),
			panicExpected: true,
		},
		{
			test:          "single wildcard path",
			data:          defaultMap,
			path:          NewPathVar(WildcardIdentifier),
			panicExpected: false,
			valueExpected: NewValue(defaultMap),
		},
		{
			test:          "wildcard path with container name",
			data:          defaultMap,
			path:          NewPathVar(containerName, WildcardIdentifier),
			panicExpected: false,
			valueExpected: NewValue(defaultMap[containerName]),
		},
		{
			test:          "nested wildcard path without container name",
			data:          defaultMap,
			path:          NewPathVar("foo", WildcardIdentifier),
			panicExpected: false,
			valueExpected: NewValue(defaultMap[containerName].(Map)["foo"]),
		},
		{
			test:          "nested wildcard path with container name",
			data:          defaultMap,
			path:          NewPathVar(containerName, "foo", WildcardIdentifier),
			panicExpected: false,
			valueExpected: NewValue(defaultMap[containerName].(Map)["foo"]),
		},
		{
			test:          "inline wildcard path",
			data:          defaultMap,
			path:          NewPathVar(containerName, WildcardIdentifier, "bar"),
			panicExpected: true,
			valueExpected: Value{},
		},
		{
			test:          "invalid wildcard path",
			data:          defaultMap,
			path:          NewPathVar(containerName, "invalidKey", WildcardIdentifier),
			panicExpected: true,
			valueExpected: Value{},
		},
		{
			test:          "invalid path",
			data:          defaultMap,
			path:          NewPathVar(containerName, "invalidKey"),
			panicExpected: true,
			valueExpected: Value{},
		},
		{
			test:          "valid path with container name",
			data:          defaultMap,
			path:          NewPathVar(containerName, "foo"),
			panicExpected: false,
			valueExpected: NewValue(defaultMap[containerName].(Map)["foo"]),
		},
		{
			test:          "valid path without container name",
			data:          defaultMap,
			path:          NewPathVar("foo"),
			panicExpected: false,
			valueExpected: NewValue(defaultMap[containerName].(Map)["foo"]),
		},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			// setup container
			mockCodec := &dataMocks.MockFileCodec{}
			mockCodec.On("Marshal", mock.Anything).Return([]byte{}, nil)
			mockCodec.On("Unmarshal", mock.Anything).Return(defaultMap, nil)
			container, err := NewRawContainer(containerName, defaultMap, mockCodec)
			assert.NoError(t, err)
			mockCodec.AssertExpectations(t)

			// test
			if tt.panicExpected {
				assert.Panics(t, func() {
					container.MustGet(tt.path)
				})
				return
			}
			value := container.MustGet(tt.path)
			assert.Equal(t, tt.valueExpected, value)
		})
	}
}

func TestRawContainer_HasPath(t *testing.T) {
	containerName := "test"
	defaultMap := Map{
		containerName: Map{
			"foo": Map{
				"bar": Map{
					"baz": "hello",
				},
			},
		},
	}

	tests := []struct {
		test    string
		data    Map
		path    Path
		hasPath bool
	}{
		{
			test:    "empty path",
			data:    defaultMap,
			path:    NewPath(""),
			hasPath: false,
		},
		{
			test:    "single wildcard path",
			data:    defaultMap,
			path:    NewPathVar(WildcardIdentifier),
			hasPath: true,
		},
		{
			test:    "wildcard path with container name",
			data:    defaultMap,
			path:    NewPathVar(containerName, WildcardIdentifier),
			hasPath: true,
		},
		{
			test:    "nested wildcard path without container name",
			data:    defaultMap,
			path:    NewPathVar("foo", WildcardIdentifier),
			hasPath: true,
		},
		{
			test:    "nested wildcard path with container name",
			data:    defaultMap,
			path:    NewPathVar(containerName, "foo", WildcardIdentifier),
			hasPath: true,
		},
		{
			test:    "inline wildcard path",
			data:    defaultMap,
			path:    NewPathVar(containerName, WildcardIdentifier, "bar"),
			hasPath: false,
		},
		{
			test:    "invalid wildcard path",
			data:    defaultMap,
			path:    NewPathVar(containerName, "invalidKey", WildcardIdentifier),
			hasPath: false,
		},
		{
			test:    "invalid path",
			data:    defaultMap,
			path:    NewPathVar(containerName, "invalidKey"),
			hasPath: false,
		},
		{
			test:    "valid path with container name",
			data:    defaultMap,
			path:    NewPathVar(containerName, "foo"),
			hasPath: true,
		},
		{
			test:    "valid path without container name",
			data:    defaultMap,
			path:    NewPathVar("foo"),
			hasPath: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			// setup container
			mockCodec := &dataMocks.MockFileCodec{}
			mockCodec.On("Marshal", mock.Anything).Return([]byte{}, nil)
			mockCodec.On("Unmarshal", mock.Anything).Return(defaultMap, nil)
			container, err := NewRawContainer(containerName, defaultMap, mockCodec)
			assert.NoError(t, err)
			mockCodec.AssertExpectations(t)

			// test
			assert.Equal(t, tt.hasPath, container.HasPath(tt.path))
		})
	}
}

func TestRawContainer_Set(t *testing.T) {
	containerName := "test"

	tests := []struct {
		name        string
		data        Map
		path        Path
		value       Value
		errExpected bool
		err         error
	}{
		{
			name: "empty path",
			data: Map{
				"test": Map{
					"foo": Map{
						"bar": Map{
							"baz": "hello",
						},
					},
				},
			},
			path:        NewPath(""),
			value:       NewValue("changed"),
			errExpected: true,
			err:         ErrEmptyPath,
		},
		{
			name: "nil value",
			data: Map{
				"test": Map{
					"foo": Map{
						"bar": Map{
							"baz": "hello",
						},
					},
				},
			},
			path:        NewPath("foo.bar.baz"),
			value:       NewValue(nil),
			errExpected: false,
		},
		{
			name: "empty value",
			data: Map{
				"test": Map{
					"foo": Map{
						"bar": Map{
							"baz": "hello",
						},
					},
				},
			},
			path:        NewPath("foo.bar.baz"),
			value:       Value{},
			errExpected: false,
		},
		{
			name: "invalid path",
			data: Map{
				"test": Map{
					"foo": Map{
						"bar": Map{
							"baz": "hello",
						},
					},
				},
			},
			path:        NewPath("invalid.path"),
			value:       NewValue("foo"),
			errExpected: true,
			err:         fmt.Errorf("cannot set path which creates more than one new path"),
		},
		{
			name: "overwrite existing path",
			data: Map{
				"test": Map{
					"foo": Map{
						"bar": Map{
							"baz": "hello",
						},
					},
				},
			},
			path:        NewPath("foo.bar.baz"),
			value:       NewValue("changed"),
			errExpected: false,
		},
		{
			name: "add one path segment",
			data: Map{
				"test": Map{
					"foo": Map{
						"bar": Map{
							"baz": "hello",
						},
					},
				},
			},
			path:        NewPath("foo.bar.qux"),
			value:       NewValue("new"),
			errExpected: false,
		},
		{
			name: "wildcard only path",
			data: Map{
				"test": Map{
					"foo": Map{
						"bar": Map{
							"baz": "hello",
						},
					},
				},
			},
			path:        NewPath(WildcardIdentifier),
			value:       NewValue("changed"),
			errExpected: true,
			err:         ErrCannotSetRootKey,
		},
		{
			name: "add path segment to value path",
			data: Map{
				"test": Map{
					"foo": Map{
						"bar": Map{
							"baz": "hello",
						},
					},
				},
			},
			path:        NewPath("foo.bar.baz.new"),
			value:       NewValue("new"),
			errExpected: true,
			err:         fmt.Errorf("cannot set path which creates a child segment on an existing value path"),
		},
		{
			name: "add slice to new path",
			data: Map{
				"test": Map{
					"foo": Map{
						"bar": Map{
							"baz": "hello",
						},
					},
				},
			},
			path:        NewPath("foo.bar.baz"),
			value:       NewValue([]string{"hello", "world"}),
			errExpected: false,
		},
		{
			name: "overwrite value with slice",
			data: Map{
				"test": Map{
					"foo": Map{
						"bar": Map{
							"baz": "hello",
						},
					},
				},
			},
			path:        NewPath("foo.bar.baz"),
			value:       NewValue([]string{"hello", "world"}),
			errExpected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// setup container
			mockCodec := &dataMocks.MockFileCodec{}
			mockCodec.On("Marshal", tt.data).Return([]byte("mocked"), nil)
			mockCodec.On("Unmarshal", []byte("mocked")).Return(tt.data, nil)
			container, err := NewRawContainer(containerName, tt.data, mockCodec)
			assert.NoError(t, err)
			mockCodec.AssertExpectations(t)

			// test

			// attemptEncode will only call marshal for non nil values
			// additionally, we are not testing complex types so we let
			// marshal return an error which causes attemptEncode to return immediately
			if tt.value.Raw != nil {
				mockCodec.On("Marshal", tt.value.Raw).Return(nil, fmt.Errorf("no marshal"))
			}

			err = container.Set(tt.path, tt.value)

			if tt.errExpected {
				assert.ErrorContains(t, err, tt.err.Error())
				return
			}
			mockCodec.AssertExpectations(t)
			assert.NoError(t, err)

			afterValue, err := container.Get(tt.path)
			assert.NotNil(t, afterValue)
			assert.NoError(t, err)
			assert.Equal(t, tt.value.Raw, afterValue.Raw)
		})
	}
}

func TestRawContainer_Set_complex(t *testing.T) {
	containerName := "test"

	type TestData struct {
		Name     string
		Location string
	}

	type TestDataComplex struct {
		TestData
		Something string
		Else      []string
	}

	tests := []struct {
		name               string
		data               Map
		path               Path
		value              Value
		pathExistsAfter    Path
		valueExpectedAfter Value
		errExpected        bool
		err                error
	}{
		{
			name: "set struct without codec modify",
			data: Map{
				"test": Map{
					"foo": Map{
						"bar": Map{
							"baz": "hello",
						},
					},
				},
			},
			path:               NewPath("foo.bar.baz"),
			value:              NewValue(TestData{Name: "John", Location: "Home"}),
			valueExpectedAfter: NewValue(Map{"Name": "John", "Location": "Home"}),
			errExpected:        false,
		},
		{
			name: "set struct with codec modify",
			data: Map{
				"test": Map{
					"foo": Map{
						"bar": Map{
							"baz": "hello",
						},
					},
				},
			},
			path:               NewPath("foo.bar.baz"),
			value:              NewValue(TestData{Name: "John", Location: "Home"}),
			valueExpectedAfter: NewValue(Map{"name": "John", "location": "Home"}),
			errExpected:        false,
		},
		{
			name: "set complex struct without codec modify",
			data: Map{
				"test": Map{
					"foo": Map{
						"bar": Map{
							"baz": "hello",
						},
					},
				},
			},
			path: NewPath("foo.bar.baz"),
			value: NewValue(TestDataComplex{
				TestData: TestData{
					Name:     "John",
					Location: "Home",
				},
				Something: "Hello",
				Else:      []string{"foo", "bar", "baz"},
			}),
			valueExpectedAfter: NewValue(Map{
				"TestData": Map{
					"Name":     "John",
					"Location": "Home",
				},
				"Something": "Hello",
				"Else":      []string{"foo", "bar", "baz"},
			}),
			errExpected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// create new container
			mockCodec := &dataMocks.MockFileCodec{}
			mockCodec.On("Marshal", tt.data).Return([]byte("initial"), nil)
			mockCodec.On("Unmarshal", []byte("initial")).Return(tt.data, nil)
			container, err := NewRawContainer(containerName, tt.data, mockCodec)
			mockCodec.AssertExpectations(t)
			assert.NoError(t, err)

			// set the data
			mockCodec.On("Marshal", tt.value.Raw).Return([]byte("attemptEncode"), nil)
			mockCodec.On("Unmarshal", []byte("attemptEncode")).Return(tt.valueExpectedAfter.Raw, nil)
			err = container.Set(tt.path, tt.value)
			if tt.errExpected {
				assert.ErrorContains(t, err, tt.err.Error())
				mockCodec.AssertExpectations(t)
				return
			}
			assert.NoError(t, err)
			mockCodec.AssertExpectations(t)

			// assert changed data
			afterValue, err := container.Get(tt.path)
			assert.NotNil(t, afterValue)
			assert.NoError(t, err)
			assert.Equal(t, tt.valueExpectedAfter.Raw, afterValue.Raw)
		})
	}
}

func TestRawContainer_SetRaw(t *testing.T) {
	containerName := "test"

	type testInterface interface {
		Test()
	}

	tests := []struct {
		name        string
		data        Map
		path        Path
		value       Value
		errExpected bool
		err         error
	}{
		{
			name:        "empty path",
			data:        nil,
			path:        Path{},
			value:       Value{},
			errExpected: true,
			err:         ErrEmptyPath,
		},
		{
			name: "empty value",
			data: Map{
				"test": Map{
					"foo": Map{
						"bar": Map{
							"baz": "hello",
						},
					},
				},
			},
			path:        NewPath("foo.bar.baz"),
			value:       Value{},
			errExpected: false,
		},
		{
			name: "wildcard only path",
			data: Map{
				"test": Map{
					"foo": Map{
						"bar": Map{
							"baz": "hello",
						},
					},
				},
			},
			path:        NewPath(WildcardIdentifier),
			value:       NewValue("changed value"),
			errExpected: true,
			err:         ErrCannotSetRootKey,
		},
		{
			name: "string value",
			data: Map{
				"test": Map{
					"foo": Map{
						"bar": Map{
							"baz": "hello",
						},
					},
				},
			},
			path:        NewPath("foo.bar.baz"),
			value:       NewValue("changed value"),
			errExpected: false,
		},
		{
			name: "function definition",
			data: Map{
				"test": Map{
					"foo": Map{
						"bar": Map{
							"baz": "hello",
						},
					},
				},
			},
			path:        NewPath("foo.bar.baz"),
			value:       NewValue(func() error { return fmt.Errorf("whoops") }),
			errExpected: true,
			err:         fmt.Errorf("cannot set function as value"),
		},
		{
			name: "nested map",
			data: Map{
				"test": Map{
					"foo": Map{
						"bar": Map{
							"baz": "hello",
						},
					},
				},
			},
			path:        NewPath("foo.bar.baz"),
			value:       NewValue(Map{"hello": Map{"test": "chicken"}}),
			errExpected: false,
		},
		{
			name: "struct",
			data: Map{
				"test": Map{
					"foo": Map{
						"bar": Map{
							"baz": "hello",
						},
					},
				},
			},
			path:        NewPath("foo.bar.baz"),
			value:       NewValue(struct{ Name string }{Name: "John"}),
			errExpected: true,
			err:         fmt.Errorf("cannot set struct as value"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// create new container
			mockCodec := &dataMocks.MockFileCodec{}
			mockCodec.On("Marshal", tt.data).Return([]byte("initial"), nil)
			mockCodec.On("Unmarshal", []byte("initial")).Return(tt.data, nil)
			container, err := NewRawContainer(containerName, tt.data, mockCodec)
			mockCodec.AssertExpectations(t)
			assert.NoError(t, err)

			// set the data
			err = container.SetRaw(tt.path, tt.value)
			if tt.errExpected {
				assert.ErrorContains(t, err, tt.err.Error())
				return
			}
			assert.NoError(t, err)

			// assert changed data
			afterValue, err := container.Get(tt.path)
			assert.NotNil(t, afterValue)
			assert.NoError(t, err)
			assert.Equal(t, tt.value.Raw, afterValue.Raw)
		})
	}
}

func TestRawContainer_Merge(t *testing.T) {
	containerName := "test"
	tests := []struct {
		name                string
		path                Path
		containerData       Map
		mergeMap            Map
		mergedContainerData Map
		errExpected         bool
		err                 error
	}{
		{
			name:        "empty path",
			errExpected: true,
			err:         ErrEmptyPath,
		},
		{
			name: "empty merge data",
			containerData: Map{
				"test": Map{
					"foo": Map{
						"bar": Map{
							"baz": "hello",
						},
					},
				},
			},
			path:     NewPath("foo.bar"),
			mergeMap: Map{},
			mergedContainerData: Map{
				"test": Map{
					"foo": Map{
						"bar": Map{
							"baz": "hello",
						},
					},
				},
			},
		},
		{
			name: "invalid merge path",
			containerData: Map{
				"test": Map{
					"foo": Map{
						"bar": Map{
							"baz": "hello",
						},
					},
				},
			},
			path:        NewPath("this.path.does.not.exist"),
			errExpected: true,
			err:         ErrPathNotFound{},
		},
		{
			name: "single value replace",
			containerData: Map{
				"test": Map{
					"foo": Map{
						"bar": Map{
							"baz": "hello",
						},
					},
				},
			},
			path: NewPath("foo.bar"),
			mergeMap: Map{
				"baz": "CHANGED",
			},
			mergedContainerData: Map{
				"test": Map{
					"foo": Map{
						"bar": Map{
							"baz": "CHANGED",
						},
					},
				},
			},
		},
		{
			name: "wildcard path",
			containerData: Map{
				"test": Map{
					"foo": Map{
						"bar": Map{
							"baz": "hello",
						},
					},
				},
			},
			path: NewPath("foo.bar.*"),
			mergeMap: Map{
				"baz": "CHANGED",
			},
			mergedContainerData: Map{
				"test": Map{
					"foo": Map{
						"bar": Map{
							"baz": "CHANGED",
						},
					},
				},
			},
		},
		{
			name: "invalid wildcard path",
			containerData: Map{
				"test": Map{
					"foo": Map{
						"bar": Map{
							"baz": "hello",
						},
					},
				},
			},
			path:        NewPath("foo.bar.*.baz"),
			errExpected: true,
			err:         ErrInlineWildcard,
		},
		{
			name: "full container data replace",
			containerData: Map{
				"test": Map{
					"foo": Map{
						"bar": Map{
							"baz": "hello",
						},
					},
				},
			},
			path:        NewPath(WildcardIdentifier),
			errExpected: true,
			err:         ErrCannotSetRootKey,
		},
		{
			name: "nested value replace",
			containerData: Map{
				"test": Map{
					"foo": Map{
						"bar": Map{
							"baz": "hello",
						},
					},
				},
			},
			path: NewPath("test"),
			mergeMap: Map{
				"foo": Map{
					"bar": Map{
						"baz": "REPLACED",
						"qux": "ADDED",
					},
				},
			},
			mergedContainerData: Map{
				"test": Map{
					"foo": Map{
						"bar": Map{
							"baz": "REPLACED",
							"qux": "ADDED",
						},
					},
				},
			},
		},
		{
			name: "complex map replace",
			containerData: Map{
				"test": Map{
					"foo": Map{
						"bar": Map{
							"baz": "hello",
						},
					},
				},
			},
			path: NewPath("test"),
			mergeMap: Map{
				"foo": Map{
					"hello": []interface{}{
						"one", "two", "three",
					},
					"bar": Map{
						"baz": "REPLACED",
						"qux": Map{
							"hello": []interface{}{"just", "another", "one"},
						},
						"pizza": "rocks",
					},
				},
			},
			mergedContainerData: Map{
				"test": Map{
					"foo": Map{
						"hello": []interface{}{
							"one", "two", "three",
						},
						"bar": Map{
							"baz": "REPLACED",
							"qux": Map{
								"hello": []interface{}{"just", "another", "one"},
							},
							"pizza": "rocks",
						},
					},
				},
			},
		},
		{
			name: "slice append",
			containerData: Map{
				"test": Map{
					"foo": Map{
						"bar": []interface{}{
							"one",
							"two",
							"three",
						},
					},
				},
			},
			path: NewPath("test"),
			mergeMap: Map{
				"foo": Map{
					"bar": []interface{}{
						"four",
						"five",
						"six",
					},
				},
			},
			mergedContainerData: Map{
				"test": Map{
					"foo": Map{
						"bar": []interface{}{
							"one",
							"two",
							"three",
							"four",
							"five",
							"six",
						},
					},
				},
			},
		},
		{
			name: "slice replace with map",
			containerData: Map{
				"test": Map{
					"foo": Map{
						"bar": []interface{}{
							"one",
							"two",
							"three",
						},
					},
				},
			},
			path: NewPath("test"),
			mergeMap: Map{
				"foo": Map{
					"bar": Map{
						"nuked": "away",
					},
				},
			},
			mergedContainerData: Map{
				"test": Map{
					"foo": Map{
						"bar": Map{
							"nuked": "away",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create new container
			mockCodec := &dataMocks.MockFileCodec{}
			mockCodec.On("Marshal", tt.containerData).Return([]byte("initial"), nil)
			mockCodec.On("Unmarshal", []byte("initial")).Return(tt.containerData, nil)
			container, err := NewRawContainer(containerName, tt.containerData, mockCodec)
			mockCodec.AssertExpectations(t)
			assert.NoError(t, err)

			// merge
			err = container.Merge(tt.path, tt.mergeMap)
			if tt.errExpected {
				assert.ErrorContains(t, err, tt.err.Error())
				return
			}
			assert.NoError(t, err)

			afterData, err := container.MustGet(NewPath(WildcardIdentifier)).Map()
			assert.NoError(t, err)

			assert.Equal(t, tt.mergedContainerData, afterData)
		})
	}
}
