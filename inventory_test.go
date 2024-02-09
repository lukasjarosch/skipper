package skipper_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lukasjarosch/skipper"
	. "github.com/lukasjarosch/skipper"
)

func TestNewInventory(t *testing.T) {
	// TODO: implement
}

func TestInventoryRegisterScope(t *testing.T) {
	inv, _ := NewInventory()

	// Test: empty scope
	err := inv.RegisterScope("", NewRegistry())
	assert.ErrorIs(t, err, ErrEmptyScope)

	// Test: nil registry
	err = inv.RegisterScope(DataScope, nil)
	assert.ErrorIs(t, err, ErrNilRegistry)

	// Test: duplicate scope registered
	err = inv.RegisterScope(DataScope, NewRegistry())
	assert.NoError(t, err)
	err = inv.RegisterScope(DataScope, NewRegistry())
	assert.ErrorIs(t, err, ErrScopeAlreadyRegistered)
}

func TestInventoryGet(t *testing.T) {
	inv, _ := NewInventory()
	registry := makeNewRegistry(t)

	err := inv.RegisterScope(DataScope, registry)
	assert.NoError(t, err)

	// Test: get existing with scope prefix
	val, err := inv.Get("data.pizza.description")
	assert.NoError(t, err)
	assert.NotNil(t, val.Raw)

	// Test: get existing path without scope prefix and without default scope
	val, err = inv.Get("pizza.description")
	assert.ErrorIs(t, err, ErrPathNotFound)
	assert.Nil(t, val.Raw)

	// Test: get existing path without scope prefix and with default scope
	err = inv.SetDefaultScope(DataScope)
	assert.NoError(t, err)
	val, err = inv.Get("pizza.description")
	assert.NoError(t, err)
	assert.NotNil(t, val.Raw)
}

func TestInventorySetDefaultScope(t *testing.T) {
	inv, _ := NewInventory()

	err := inv.RegisterScope(DataScope, skipper.NewRegistry())
	assert.NoError(t, err)

	// Test: empty scope
	err = inv.SetDefaultScope("")
	assert.ErrorIs(t, err, ErrEmptyScope)

	// Test: not existing scope
	err = inv.SetDefaultScope("something")
	assert.ErrorIs(t, err, ErrScopeDoesNotExist)

	// Test: valid, existing scope
	err = inv.SetDefaultScope(DataScope)
	assert.NoError(t, err)
}
