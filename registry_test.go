package skipper_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lukasjarosch/skipper"
	"github.com/lukasjarosch/skipper/codec"
	"github.com/lukasjarosch/skipper/data"
)

func makeNewRegistry(t *testing.T) *skipper.Registry {
	persons, pizza, hans, john, foodCommon := makeClasses(t)
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
	err = registry.RegisterClass(data.NewPathVar("food", foodCommon.Name), foodCommon)
	assert.NoError(t, err)

	return registry
}

func TestNewRegistry(t *testing.T) {
	makeNewRegistry(t)
}

func TestRegistryRegisterClass(t *testing.T) {
	_, _, _, john, _ := makeClasses(t)
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

	// Test: class already has a pre-set hook
	jane, err := skipper.NewClass(janeClassPath, codec.NewYamlCodec())
	assert.NoError(t, err)
	assert.NotNil(t, jane)
	jane.SetPreSetHook(func(class skipper.Class, path data.Path, value data.Value) error { return nil })
	err = registry.RegisterClass(data.NewPathVar("people", jane.Name), jane)
	assert.ErrorIs(t, err, skipper.ErrCannotOverwriteHook)

	// Test: class already has a post-set hook
	jane, err = skipper.NewClass(janeClassPath, codec.NewYamlCodec())
	assert.NoError(t, err)
	assert.NotNil(t, jane)
	jane.SetPostSetHook(func(class skipper.Class, path data.Path, value data.Value) error { return nil })
	err = registry.RegisterClass(data.NewPathVar("people", jane.Name), jane)
	assert.ErrorIs(t, err, skipper.ErrCannotOverwriteHook)
}

func TestRegistryPreSetHook(t *testing.T) {
	registry := makeNewRegistry(t)

	// Register a 'food' class which can cause issues with the 'food.common' class
	food, err := skipper.NewClass(foodClassPath, codec.NewYamlCodec())
	assert.NoError(t, err)
	assert.NotNil(t, food)
	err = registry.RegisterClass(data.NewPath(food.Name), food)
	assert.NoError(t, err)

	// Test: Attempt to set existing path from different class
	// The path 'food.common.should_be_tasty' is already registered by the 'food.common' class
	// and must not be registered again by the 'food' class.
	err = food.Set("food.common.should_be_tasty", false)
	assert.ErrorIs(t, err, skipper.ErrDuplicatePath)

	// Test: Attempting to set a non-existing path must work
	err = food.Set("food.common.eaten", false)
	assert.NoError(t, err)
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
