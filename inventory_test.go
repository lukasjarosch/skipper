package skipper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/lukasjarosch/skipper"
	"github.com/lukasjarosch/skipper/codec"
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
	assert.ErrorIs(t, err, ErrPathNotFound)
	assert.Nil(t, val.Raw)
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

func TestInventory_Compile(t *testing.T) {
	dataRegistry := func(class string) *Registry {
		commonClass, err := NewClass(fmt.Sprintf("testdata/compile/data/%s.yaml", class), codec.NewYamlCodec(), data.NewPath(class))
		assert.NoError(t, err)
		dataRegistry := NewRegistry()
		err = dataRegistry.RegisterClass(commonClass)
		assert.NoError(t, err)

		return dataRegistry
	}

	targetRegistry := func(name string) *Registry {
		testTarget, err := NewClass(fmt.Sprintf("testdata/compile/targets/%s.yaml", name), codec.NewYamlCodec(), data.NewPath(name))
		assert.NoError(t, err)
		targetRegistry := NewRegistry()
		err = targetRegistry.RegisterClass(testTarget)
		assert.NoError(t, err)

		return targetRegistry
	}

	t.Run("valid target", func(t *testing.T) {
		inventory, _ := NewInventory()
		err := inventory.RegisterScope(DataScope, dataRegistry("common"))
		assert.NoError(t, err)
		err = inventory.RegisterScope(TargetsScope, targetRegistry("valid"))
		assert.NoError(t, err)

		err = inventory.Compile(data.NewPath("targets.valid"))
		assert.NoError(t, err)

		expected := map[string]data.Value{
			"data.common.name": data.NewValue("Jane"),
			"data.common.age":  data.NewValue("over 9000"),
		}

		for p, v := range expected {
			val, err := inventory.Get(p)
			assert.NoError(t, err)
			assert.Equal(t, v.Raw, val.Raw)
		}
	})

	t.Run("target cannot introduce paths in the inventory", func(t *testing.T) {
		inventory, _ := NewInventory()
		err := inventory.RegisterScope(DataScope, dataRegistry("common"))
		assert.NoError(t, err)
		err = inventory.RegisterScope(TargetsScope, targetRegistry("introduce_paths"))
		assert.NoError(t, err)

		err = inventory.Compile(data.NewPath("targets.introduce_paths"))
		assert.ErrorIs(t, err, ErrTargetCannotIntroducePaths)
	})

	t.Run("error if undefined value is not overwritten", func(t *testing.T) {
		inventory, _ := NewInventory()
		err := inventory.RegisterScope(DataScope, dataRegistry("undefined"))
		assert.NoError(t, err)
		err = inventory.RegisterScope(TargetsScope, targetRegistry("not_overwritten_undefined"))
		assert.NoError(t, err)

		err = inventory.Compile(data.NewPath("targets.not_overwritten_undefined"))
		assert.ErrorIs(t, err, ErrUndefinedValueNotOverwritten)
	})

	t.Run("no error if undefined value is overwritten", func(t *testing.T) {
		inventory, _ := NewInventory()
		err := inventory.RegisterScope(DataScope, dataRegistry("undefined"))
		assert.NoError(t, err)
		err = inventory.RegisterScope(TargetsScope, targetRegistry("overwritten_undefined"))
		assert.NoError(t, err)

		err = inventory.Compile(data.NewPath("targets.overwritten_undefined"))
		assert.NoError(t, err)

		val, err := inventory.Get("data.undefined.network_cidr")
		assert.NoError(t, err)
		assert.Equal(t, "10.0.0.0/8", val.String())
	})
}
