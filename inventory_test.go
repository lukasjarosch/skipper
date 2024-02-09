package skipper_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lukasjarosch/skipper"
	"github.com/lukasjarosch/skipper/codec"
	"github.com/lukasjarosch/skipper/data"
)

var (
	personClassPath = "testdata/classes/person.yaml"
	pizzaClassPath  = "testdata/classes/pizza.yaml"
	hansClassPath   = "testdata/classes/people/hans.yaml"
	johnClassPath   = "testdata/classes/people/john.yaml"
	peopleClassPath = "testdata/classes/people.yaml"
	stripPrefix     = data.NewPathFromOsPath("testdata/classes")
)

func makeClasses(t *testing.T) (*skipper.Class, *skipper.Class, *skipper.Class, *skipper.Class) {
	person, err := skipper.NewClass(personClassPath, codec.NewYamlCodec())
	assert.NoError(t, err)
	assert.NotNil(t, person)

	pizza, err := skipper.NewClass(pizzaClassPath, codec.NewYamlCodec())
	assert.NoError(t, err)
	assert.NotNil(t, person)

	hans, err := skipper.NewClass(hansClassPath, codec.NewYamlCodec())
	assert.NoError(t, err)
	assert.NotNil(t, person)

	john, err := skipper.NewClass(johnClassPath, codec.NewYamlCodec())
	assert.NoError(t, err)
	assert.NotNil(t, person)

	return person, pizza, hans, john
}

func makeInventory(t *testing.T) *skipper.Inventory {
	person, pizza, hans, john := makeClasses(t)

	inventory, err := skipper.NewInventory()
	assert.NoError(t, err)
	assert.NotNil(t, inventory)

	identifier := data.NewPathFromOsPath(person.FilePath).StripPrefix(stripPrefix)
	err = inventory.RegisterClass(identifier, person)
	assert.NoError(t, err)

	identifier = data.NewPathFromOsPath(pizza.FilePath).StripPrefix(stripPrefix)
	err = inventory.RegisterClassWithScope("food", identifier, pizza)
	assert.NoError(t, err)

	identifier = data.NewPathFromOsPath(hans.FilePath).StripPrefix(stripPrefix)
	err = inventory.RegisterClass(identifier, hans)
	assert.NoError(t, err)

	identifier = data.NewPathFromOsPath(john.FilePath).StripPrefix(stripPrefix)
	err = inventory.RegisterClass(identifier, john)
	assert.NoError(t, err)

	return inventory
}

func TestNewInventory(t *testing.T) {
	inventory, err := skipper.NewInventory()
	assert.NoError(t, err)
	assert.NotNil(t, inventory)
}

func TestInventoryRegisterClass(t *testing.T) {
	person, pizza, _, _ := makeClasses(t)
	inventory := makeInventory(t)

	// Test case: Class already exists in the default scope
	identifier := data.NewPathFromOsPath(person.FilePath).StripPrefix(stripPrefix)
	err := inventory.RegisterClass(identifier, person)
	assert.ErrorIs(t, err, skipper.ErrClassAlreadyRegistered)

	// Test case: Register class again in a different namespace must work
	identifier = data.NewPathFromOsPath(pizza.FilePath).StripPrefix(stripPrefix)
	err = inventory.RegisterClassWithScope("new_scope", identifier, pizza)
	assert.NoError(t, err)
}

func TestInventoryGet(t *testing.T) {
	inventory := makeInventory(t)

	// Test case: Full path; path exists
	ret, err := inventory.Get("default.people.hans.name")
	assert.NoError(t, err)
	assert.Equal(t, ret.String(), "Hans")

	// Test case: Path without 'default' scope; path exists
	ret, err = inventory.Get("people.hans.name")
	assert.NoError(t, err)
	assert.Equal(t, ret.String(), "Hans")

	// Test case: Path with a different scope; path exists
	ret, err = inventory.Get("food.pizza.description")
	assert.NoError(t, err)
	assert.NotEmpty(t, ret)

	// Test case: Invalid path
	ret, err = inventory.Get("invalid.path")
	assert.Error(t, err, skipper.ErrCannotResolvePath)
	assert.Nil(t, ret.Raw)
}
