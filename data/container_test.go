package data_test

import (
	"fmt"
	"testing"

	_ "github.com/charmbracelet/log"
	. "github.com/lukasjarosch/skipper/data"
	dataMocks "github.com/lukasjarosch/skipper/mocks/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewRawContainer_EmptyContainerName(t *testing.T) {
	_, err := NewRawContainer("", nil, nil)
	assert.Error(t, err, ErrEmptyContainerName)
}

func TestNewRawContainer_MarshalError(t *testing.T) {
	expectedError := fmt.Errorf("an error")

	mockCodec := &dataMocks.MockFileCodec{}
	mockCodec.On("Marshal", mock.Anything).Return([]byte{}, expectedError)

	_, err := NewRawContainer("name", nil, mockCodec)
	assert.Error(t, err, expectedError)
}

func TestNewRawContainer_UnmarshalError(t *testing.T) {
	expectedError := fmt.Errorf("an error")

	mockCodec := &dataMocks.MockFileCodec{}
	mockCodec.On("Marshal", mock.Anything).Return([]byte{}, nil)
	mockCodec.On("Unmarshal", mock.Anything).Return(Map{}, expectedError)

	_, err := NewRawContainer("name", nil, mockCodec)
	assert.ErrorIs(t, err, expectedError)
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
			mockCodec := &dataMocks.MockFileCodec{}
			mockCodec.On("Marshal", mock.Anything).Return([]byte{}, nil)
			mockCodec.On("Unmarshal", mock.Anything).Return(defaultMap, nil)

			container, err := NewRawContainer(containerName, defaultMap, mockCodec)
			assert.NoError(t, err)

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
			mockCodec := &dataMocks.MockFileCodec{}
			mockCodec.On("Marshal", mock.Anything).Return([]byte{}, nil)
			mockCodec.On("Unmarshal", mock.Anything).Return(defaultMap, nil)

			container, err := NewRawContainer(containerName, defaultMap, mockCodec)
			assert.NoError(t, err)

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
			mockCodec := &dataMocks.MockFileCodec{}
			mockCodec.On("Marshal", mock.Anything).Return([]byte{}, nil)
			mockCodec.On("Unmarshal", mock.Anything).Return(defaultMap, nil)

			container, err := NewRawContainer(containerName, defaultMap, mockCodec)
			assert.NoError(t, err)

			assert.Equal(t, tt.hasPath, container.HasPath(tt.path))
		})
	}
}
