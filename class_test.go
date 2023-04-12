package skipper_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/lukasjarosch/skipper"
	"github.com/lukasjarosch/skipper/mocks"
	"github.com/stretchr/testify/assert"
)

func TestNewClass_EmptyNamespace(t *testing.T) {
	class, err := skipper.NewClass(skipper.P(""), nil)
	assert.Nil(t, class)
	assert.ErrorIs(t, err, skipper.ErrEmptyNamespace)
}

func TestNewClass_NilDataProvider(t *testing.T) {
	class, err := skipper.NewClass(skipper.P("foo.bar"), nil)
	assert.Nil(t, class)
	assert.ErrorIs(t, err, skipper.ErrNilDataProvider)
}

func TestNewClass_InvalidRootKey(t *testing.T) {
	namespace := skipper.P("foo.bar")
	rootKey := namespace[len(namespace)-1]

	mock := mocks.NewDataProvider(t)
	mock.EXPECT().HasPath(skipper.P(rootKey)).Return(false)

	class, err := skipper.NewClass(skipper.P("foo.bar"), mock)
	assert.Nil(t, class)
	assert.ErrorContains(t, err, skipper.ErrInvalidRootKey.Error())
}

func TestNewClass_NoConfiguration(t *testing.T) {
	namespace := skipper.P("foo.bar")
	rootKey := namespace[len(namespace)-1]

	mock := mocks.NewDataProvider(t)
	mock.EXPECT().HasPath(skipper.P(rootKey)).Return(true)
	mock.EXPECT().HasPath(skipper.Path{rootKey, skipper.SkipperKey}).Return(false)

	class, err := skipper.NewClass(skipper.P("foo.bar"), mock)
	assert.NotNil(t, class)
	assert.NoError(t, err)
	assert.Nil(t, class.Configuration)
}

func TestNewClass_InvalidConfiguration(t *testing.T) {
	namespace := skipper.P("foo.bar")
	rootKey := namespace[len(namespace)-1]
	configPath := skipper.Path{rootKey, skipper.SkipperKey}
	target := new(skipper.ClassConfig)

	mock := mocks.NewDataProvider(t)
	mock.EXPECT().HasPath(skipper.P(rootKey)).Return(true)
	mock.EXPECT().HasPath(configPath).Return(true)
	mock.EXPECT().UnmarshalPath(configPath, target).Return(fmt.Errorf("an-error"))

	class, err := skipper.NewClass(skipper.P("foo.bar"), mock)
	assert.Nil(t, class)
	assert.Error(t, err)
}

func TestNewClass_ValidConfiguration(t *testing.T) {
	namespace := skipper.P("foo.bar")
	rootKey := namespace[len(namespace)-1]
	configPath := skipper.Path{rootKey, skipper.SkipperKey}
	target := new(skipper.ClassConfig)

	mock := mocks.NewDataProvider(t)
	mock.EXPECT().HasPath(skipper.P(rootKey)).Return(true)
	mock.EXPECT().HasPath(configPath).Return(true)
	mock.EXPECT().UnmarshalPath(configPath, target).Return(nil)

	class, err := skipper.NewClass(skipper.P("foo.bar"), mock)
	assert.NotNil(t, class)
	assert.NotNil(t, class.Configuration)
	assert.NoError(t, err)
}

func setupClass(mock *mocks.DataProvider) *skipper.Class {
	namespace := skipper.P("foo.bar")
	rootKey := namespace[len(namespace)-1]

	mock.On("HasPath", skipper.P(rootKey)).Return(true)
	mock.On("HasPath", skipper.Path{rootKey, skipper.SkipperKey}).Return(false)

	class, _ := skipper.NewClass(skipper.P("foo.bar"), mock)

	return class
}

func TestGet_PathExists(t *testing.T) {
	path := skipper.P("foo.bar")
	expectedValue := "hello"
	expectedExists := true

	mock := mocks.NewDataProvider(t)
	mock.EXPECT().GetPath(path).Once().Return(expectedValue, nil)

	class := setupClass(mock)

	val, exists := class.Get(path)
	assert.Equal(t, val, expectedValue)
	assert.Equal(t, exists, expectedExists)
}

func TestGet_PathDoesNotExist(t *testing.T) {
	path := skipper.P("foo.bar")
	expectedExists := false

	mock := mocks.NewDataProvider(t)
	mock.EXPECT().GetPath(path).Once().Return(nil, errors.New("error"))

	class := setupClass(mock)

	val, exists := class.Get(path)
	assert.Nil(t, val)
	assert.Equal(t, exists, expectedExists)
}
