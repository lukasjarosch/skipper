package skipper_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lukasjarosch/skipper"
	. "github.com/lukasjarosch/skipper"
	"github.com/lukasjarosch/skipper/data"
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
	assert.NoError(t, err)
	assert.NotNil(t, val.Raw)

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

func TestInventoryAbsolutePath(t *testing.T) {
	inv, _ := NewInventory()
	registry := makeNewRegistry(t)

	err := inv.RegisterScope(DataScope, registry)
	assert.NoError(t, err)

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
			name:     "context does not have a valid scope prefix",
			path:     data.NewPath("name"),
			context:  data.NewPath("unknown.foo.bar"),
			expected: nil,
			err:      ErrPathNotFound,
		},
		{
			name:     "path is not valid within the given context",
			path:     data.NewPath("unknown"),
			context:  data.NewPath("data.person.gender"),
			expected: nil,
			err:      data.ErrCannotResolveAbsolutePath,
		},
		{
			name:     "context is only scope name",
			path:     data.NewPath("name"),
			context:  data.NewPath("data"),
			expected: nil,
			err:      data.ErrEmptyPath,
		},
		{
			name:     "context is scope and classIdentifier",
			path:     data.NewPath("name"),
			context:  data.NewPath("data.person"),
			expected: data.NewPath("data.person.name"),
			err:      nil,
		},
		{
			name:     "context is value path",
			path:     data.NewPath("name"),
			context:  data.NewPath("data.person.age"),
			expected: data.NewPath("data.person.name"),
			err:      nil,
		},
		{
			name:     "context is non-value path",
			path:     data.NewPath("name"),
			context:  data.NewPath("data.person.interests"),
			expected: data.NewPath("data.person.name"),
			err:      nil,
		},
		{
			name:     "path is already inventory-absolute",
			path:     data.NewPath("data.person.name"),
			context:  data.NewPath("data.person.interests"),
			expected: data.NewPath("data.person.name"),
			err:      nil,
		},
		{
			name:     "path is registry-absolute",
			path:     data.NewPath("person.name"),
			context:  data.NewPath("data.person.interests"),
			expected: data.NewPath("data.person.name"),
			err:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			abs, err := inv.AbsolutePath(tt.path, tt.context)

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
