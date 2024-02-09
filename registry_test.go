package skipper_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lukasjarosch/skipper"
	"github.com/lukasjarosch/skipper/codec"
	"github.com/lukasjarosch/skipper/data"
)

func makeNewRegistry(t *testing.T) *skipper.Registry {
	persons, pizza, hans, john := makeClasses(t)
	registry := skipper.NewRegistry()

	// Test: register multiple classes in two namespaces
	err := registry.RegisterClass(data.NewPath(persons.Name), persons)
	assert.NoError(t, err)
	err = registry.RegisterClass(data.NewPath(pizza.Name), pizza)
	assert.NoError(t, err)
	err = registry.RegisterClass(data.NewPathVar("people", hans.Name), hans)
	assert.NoError(t, err)
	err = registry.RegisterClass(data.NewPathVar("people", john.Name), john)
	assert.NoError(t, err)

	return registry
}

func TestNewRegistry(t *testing.T) {
	makeNewRegistry(t)
}

func TestRegistryRegisterClass(t *testing.T) {
	_, _, _, john := makeClasses(t)
	registry := makeNewRegistry(t)

	// Test: class already registered
	err := registry.RegisterClass(data.NewPathVar("people", john.Name), john)
	assert.ErrorIs(t, err, skipper.ErrClassAlreadyRegistered)

	// Test: class with empty namespace
	err = registry.RegisterClass(data.Path{}, john)
	assert.ErrorIs(t, err, skipper.ErrEmptyClassIdentifier)

	// Test: class name must be last segment of identifier
	err = registry.RegisterClass(data.NewPathVar("foo", "bar"), john)
	assert.ErrorIs(t, err, skipper.ErrInvalidClassIdentifier)

	// Test: register a class which introduces a duplicate path
	people, err := skipper.NewClass(peopleClassPath, codec.NewYamlCodec())
	assert.NoError(t, err)
	assert.NotNil(t, people)
	err = registry.RegisterClass(data.NewPath(people.Name), people)
	assert.ErrorIs(t, err, skipper.ErrDuplicatePath)
}

// func TestRegistryClassesInNamespace(t *testing.T) {
// 	registry := makeNewRegistry(t)
//
// 	// Test: namespace exists
// 	classes, err := registry.ClassesInNamespace(skipper.RootNamespace)
// 	assert.NoError(t, err)
// 	assert.Len(t, classes, 2)
//
// 	// Test: namespace does not exist
// 	classes, err = registry.ClassesInNamespace(skipper.RootNamespace.Append("something"))
// 	assert.ErrorIs(t, err, skipper.ErrNamespaceDoesNotExist)
// 	assert.Nil(t, classes)
// }

// func TestRegistryGetClass(t *testing.T) {
// 	persons, _, hans, _ := makeClasses(t)
// 	registry := makeNewRegistry(t)
//
// 	// Test: class in root namespace, exists
// 	class, err := registry.GetClass(persons.Name)
// 	assert.NoError(t, err)
// 	assert.Equal(t, persons, class)
//
// 	// Test: class not in root namespace, exists
// 	class, err = registry.GetClass(hans.Name)
// 	assert.NoError(t, err)
// 	assert.Equal(t, hans, class)
//
// 	// Test: class does not exist
// 	class, err = registry.GetClass("undefined")
// 	assert.ErrorIs(t, err, skipper.ErrClassDoesNotExist)
// }
//
// func TestRegistryResolveClass(t *testing.T) {
// 	persons, _, _, _ := makeClasses(t)
// 	registry := makeNewRegistry(t)
//
// 	// Test: resolve existing class in root namespace
// 	class, err := registry.ResolveClass(skipper.RootNamespace.Append(persons.Name))
// 	assert.NoError(t, err)
// 	assert.Equal(t, persons, class)
// }
