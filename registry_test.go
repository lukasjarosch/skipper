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

	personClass, err = NewClass(personClassPath, codec.NewYamlCodec(), data.NewPath("person"))
	assert.NoError(t, err)
	assert.NotNil(t, personClass)

	pizzaClass, err = NewClass(pizzaClassPath, codec.NewYamlCodec(), data.NewPath("pizza"))
	assert.NoError(t, err)
	assert.NotNil(t, pizzaClass)

	foodCommonClass, err = NewClass(foodCommonClassPath, codec.NewYamlCodec(), data.NewPath("food.common"))
	assert.NoError(t, err)
	assert.NotNil(t, foodCommonClass)

	foodClass, err = NewClass(foodClassPath, codec.NewYamlCodec(), data.NewPath("food"))
	assert.NoError(t, err)
	assert.NotNil(t, foodClass)

	hansClass, err = NewClass(hansClassPath, codec.NewYamlCodec(), data.NewPath("people.hans"))
	assert.NoError(t, err)
	assert.NotNil(t, hansClass)

	johnClass, err = NewClass(johnClassPath, codec.NewYamlCodec(), data.NewPath("people.john"))
	assert.NoError(t, err)
	assert.NotNil(t, johnClass)

	janeClass, err = NewClass(janeClassPath, codec.NewYamlCodec(), data.NewPath("people.jane"))
	assert.NoError(t, err)
	assert.NotNil(t, janeClass)

	peopleClass, err = NewClass(peopleClassPath, codec.NewYamlCodec(), data.NewPath("people"))
	assert.NoError(t, err)
	assert.NotNil(t, peopleClass)
}

func makeNewRegistry(t *testing.T) *Registry {
	makeClasses(t)
	registry := NewRegistry()

	err := registry.RegisterClass(personClass)
	assert.NoError(t, err)
	err = registry.RegisterClass(pizzaClass)
	assert.NoError(t, err)
	err = registry.RegisterClass(hansClass)
	assert.NoError(t, err)
	err = registry.RegisterClass(johnClass)
	assert.NoError(t, err)
	err = registry.RegisterClass(foodCommonClass)
	assert.NoError(t, err)

	return registry
}

func TestNewRegistry(t *testing.T) {
	makeNewRegistry(t)
}

func TestNewRegistryFromFiles(t *testing.T) {
	rootPath := "testdata/references"

	files, err := DiscoverFiles(rootPath, codec.YamlPathSelector)
	assert.NoError(t, err)

	reg, err := NewRegistryFromFiles(files, func(filePaths []string) ([]*Class, error) {
		return ClassLoader("testdata/references", filePaths, codec.NewYamlCodec())
	})
	assert.NoError(t, err)
	_ = reg
}

func TestRegistryRegisterClass(t *testing.T) {
	registry := makeNewRegistry(t)

	// Test: class already registered
	err := registry.RegisterClass(johnClass)
	assert.ErrorIs(t, err, ErrClassAlreadyRegistered)

	// Test: register a class which introduces a duplicate path
	err = registry.RegisterClass(peopleClass)
	assert.ErrorIs(t, err, ErrDuplicatePath)

	// Test: class already has a pre-set hook
	jane, err := NewClass(janeClassPath, codec.NewYamlCodec(), data.NewPath("people.jane"))
	assert.NoError(t, err)
	assert.NotNil(t, jane)
	jane.SetPreSetHook(func(class Class, path data.Path, value data.Value) error { return nil })
	err = registry.RegisterClass(jane)
	assert.ErrorIs(t, err, ErrCannotOverwriteHook)

	// Test: class already has a post-set hook
	jane, err = NewClass(janeClassPath, codec.NewYamlCodec(), data.NewPath("people.jane"))
	assert.NoError(t, err)
	assert.NotNil(t, jane)
	jane.SetPostSetHook(func(class Class, path data.Path, value data.Value) error { return nil })
	err = registry.RegisterClass(jane)
	assert.ErrorIs(t, err, ErrCannotOverwriteHook)
}

func TestRegistryPreSetHook(t *testing.T) {
	registry := makeNewRegistry(t)

	// Register a 'food' class which can cause issues with the 'food.common' class
	food, err := NewClass(foodClassPath, codec.NewYamlCodec(), data.NewPath("food"))
	assert.NoError(t, err)
	assert.NotNil(t, food)
	err = registry.RegisterClass(food)
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
	food, err := NewClass(foodClassPath, codec.NewYamlCodec(), data.NewPath("food"))
	assert.NoError(t, err)
	assert.NotNil(t, food)
	err = registry.RegisterClass(food)
	assert.NoError(t, err)

	// Test: new path must be resolvable by the registry afterwards
	err = food.Set("food.burger", "very_juicy")
	assert.NoError(t, err)
	val, err := registry.Get("food.burger")
	assert.NoError(t, err)
	assert.NotNil(t, val.Raw)
	assert.Equal(t, val.Raw, "very_juicy")
}

func TestRegistryAbsolutePath(t *testing.T) {
	tests := []struct {
		name     string
		path     data.Path
		context  data.Path
		expected data.Path
		err      error
	}{
		{
			name:     "path is nil",
			path:     nil,
			context:  data.NewPath("person.name"),
			expected: nil,
			err:      data.ErrEmptyPath,
		},
		{
			name:     "path is empty",
			path:     data.Path{},
			context:  data.NewPath("person.name"),
			expected: nil,
			err:      data.ErrEmptyPath,
		},
		{
			name:     "context is nil",
			path:     data.NewPath("name"),
			context:  nil,
			expected: nil,
			err:      data.ErrEmptyPath,
		},
		{
			name:     "context is empty",
			path:     data.NewPath("name"),
			context:  data.Path{},
			expected: nil,
			err:      data.ErrEmptyPath,
		},
		{
			name:     "context is classIdentifier",
			path:     data.NewPath("name"),
			context:  data.NewPath("person"),
			expected: data.NewPath("person.name"),
			err:      nil,
		},
		{
			name:     "context is value path",
			path:     data.NewPath("name"),
			context:  data.NewPath("person.contact.email"),
			expected: data.NewPath("person.name"),
			err:      nil,
		},
		{
			name:     "context path is unknown",
			path:     data.NewPath("name"),
			context:  data.NewPath("unknown.path"),
			expected: nil,
			err:      ErrPathNotFound,
		},
		{
			name:     "context and path are value paths",
			path:     data.NewPath("address.city"),
			context:  data.NewPath("person.contact.phone"),
			expected: data.NewPath("person.address.city"),
			err:      nil,
		},
		{
			name:     "context and path are relative paths",
			path:     data.NewPath("address.city"),
			context:  data.NewPath("contact.phone"),
			expected: nil,
			err:      ErrPathNotFound,
		},
		{
			name:     "path is already absolute",
			path:     data.NewPath("person.address.city"),
			context:  data.NewPath("person.contact.phone"),
			expected: data.NewPath("person.address.city"),
			err:      nil,
		},
	}

	registry := makeNewRegistry(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			abs, err := registry.AbsolutePath(tt.path, tt.context)

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
