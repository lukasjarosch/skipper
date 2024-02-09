package skipper_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/lukasjarosch/skipper"
	"github.com/lukasjarosch/skipper/codec"
	"github.com/lukasjarosch/skipper/data"
)

var (
	personClassPath     = "testdata/classes/person.yaml"
	pizzaClassPath      = "testdata/classes/pizza.yaml"
	foodCommonClassPath = "testdata/classes/food/common.yaml"
	foodClassPath       = "testdata/classes/food.yaml"
	hansClassPath       = "testdata/classes/people/hans.yaml"
	johnClassPath       = "testdata/classes/people/john.yaml"
	janeClassPath       = "testdata/classes/people/jane.yaml"
	peopleClassPath     = "testdata/classes/people.yaml"

	stripPrefix = data.NewPathFromOsPath("testdata/classes")

	personClass,
	pizzaClass,
	foodCommonClass,
	foodClass,
	hansClass,
	johnClass,
	janeClass,
	peopleClass *Class
)

func makeClasses(t *testing.T) {
	var err error

	personClass, err = NewClass(personClassPath, codec.NewYamlCodec())
	assert.NoError(t, err)
	assert.NotNil(t, personClass)

	pizzaClass, err = NewClass(pizzaClassPath, codec.NewYamlCodec())
	assert.NoError(t, err)
	assert.NotNil(t, pizzaClass)

	foodCommonClass, err = NewClass(foodCommonClassPath, codec.NewYamlCodec())
	assert.NoError(t, err)
	assert.NotNil(t, foodCommonClass)

	foodClass, err = NewClass(foodClassPath, codec.NewYamlCodec())
	assert.NoError(t, err)
	assert.NotNil(t, foodClass)

	hansClass, err = NewClass(hansClassPath, codec.NewYamlCodec())
	assert.NoError(t, err)
	assert.NotNil(t, hansClass)

	johnClass, err = NewClass(johnClassPath, codec.NewYamlCodec())
	assert.NoError(t, err)
	assert.NotNil(t, johnClass)

	janeClass, err = NewClass(janeClassPath, codec.NewYamlCodec())
	assert.NoError(t, err)
	assert.NotNil(t, janeClass)

	peopleClass, err = NewClass(peopleClassPath, codec.NewYamlCodec())
	assert.NoError(t, err)
	assert.NotNil(t, peopleClass)
}

func makeNewRegistry(t *testing.T) *Registry {
	makeClasses(t)
	registry := NewRegistry()

	err := registry.RegisterClass(data.NewPath(personClass.Name), personClass)
	assert.NoError(t, err)
	err = registry.RegisterClass(data.NewPath(pizzaClass.Name), pizzaClass)
	assert.NoError(t, err)
	err = registry.RegisterClass(data.NewPathVar("people", hansClass.Name), hansClass)
	assert.NoError(t, err)
	err = registry.RegisterClass(data.NewPathVar("people", johnClass.Name), johnClass)
	assert.NoError(t, err)
	err = registry.RegisterClass(data.NewPathVar("food", foodCommonClass.Name), foodCommonClass)
	assert.NoError(t, err)

	return registry
}

func TestNewRegistry(t *testing.T) {
	makeNewRegistry(t)
}

func TestRegistryRegisterClass(t *testing.T) {
	registry := makeNewRegistry(t)

	// Test: class already registered
	err := registry.RegisterClass(data.NewPathVar("people", johnClass.Name), johnClass)
	assert.ErrorIs(t, err, ErrClassAlreadyRegistered)

	// Test: class with empty namespace
	err = registry.RegisterClass(data.Path{}, johnClass)
	assert.ErrorIs(t, err, ErrEmptyClassIdentifier)

	// Test: class name must be last segment of identifier
	err = registry.RegisterClass(data.NewPathVar("foo", "bar"), johnClass)
	assert.ErrorIs(t, err, ErrInvalidClassIdentifier)

	// Test: register a class which introduces a duplicate path
	err = registry.RegisterClass(data.NewPath(peopleClass.Name), peopleClass)
	assert.ErrorIs(t, err, ErrDuplicatePath)

	// Test: class already has a pre-set hook
	jane, err := NewClass(janeClassPath, codec.NewYamlCodec())
	assert.NoError(t, err)
	assert.NotNil(t, jane)
	jane.SetPreSetHook(func(class Class, path data.Path, value data.Value) error { return nil })
	err = registry.RegisterClass(data.NewPathVar("people", jane.Name), jane)
	assert.ErrorIs(t, err, ErrCannotOverwriteHook)

	// Test: class already has a post-set hook
	jane, err = NewClass(janeClassPath, codec.NewYamlCodec())
	assert.NoError(t, err)
	assert.NotNil(t, jane)
	jane.SetPostSetHook(func(class Class, path data.Path, value data.Value) error { return nil })
	err = registry.RegisterClass(data.NewPathVar("people", jane.Name), jane)
	assert.ErrorIs(t, err, ErrCannotOverwriteHook)
}

func TestRegistryPreSetHook(t *testing.T) {
	registry := makeNewRegistry(t)

	// Register a 'food' class which can cause issues with the 'food.common' class
	food, err := NewClass(foodClassPath, codec.NewYamlCodec())
	assert.NoError(t, err)
	assert.NotNil(t, food)
	err = registry.RegisterClass(data.NewPath(food.Name), food)
	assert.NoError(t, err)

	// Test: Attempt to set existing path from different class
	// The path 'food.common.should_be_tasty' is already registered by the 'food.common' class
	// and must not be registered again by the 'food' class.
	err = food.Set("food.common.should_be_tasty", false)
	assert.ErrorIs(t, err, ErrDuplicatePath)

	// Test: Attempting to set a non-existing path must work
	err = food.Set("food.common.eaten", false)
	assert.NoError(t, err)
}

func TestRegistryPostSetHook(t *testing.T) {
	registry := makeNewRegistry(t)

	// Register a 'food' class
	food, err := NewClass(foodClassPath, codec.NewYamlCodec())
	assert.NoError(t, err)
	assert.NotNil(t, food)
	err = registry.RegisterClass(data.NewPath(food.Name), food)
	assert.NoError(t, err)

	// Test: new path must be resolvable by the registry afterwards
	err = food.Set("food.burger", "very_juicy")
	assert.NoError(t, err)
	val, err := registry.Get("food.burger")
	assert.NoError(t, err)
	assert.NotNil(t, val.Raw)
	assert.Equal(t, val.Raw, "very_juicy")
}
