package skipper_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/lukasjarosch/skipper"
	"github.com/lukasjarosch/skipper/codec"
	"github.com/lukasjarosch/skipper/data"
)

func TestValueManager_SetHooks(t *testing.T) {
	filePath := "testdata/references/class/valid.yaml"
	class, err := NewClass(filePath, codec.NewYamlCodec(), data.NewPath("valid"))
	assert.NoError(t, err)
	assert.NotNil(t, class)

	manager, err := NewValueReferenceManager(class)
	assert.NoError(t, err)
	assert.NotNil(t, manager)

	t.Run("expect no change", func(t *testing.T) {
		preReferenceCount := len(manager.AllReferences())
		err = class.Set("valid.new.test1", data.NewValue("ohai"))
		assert.NoError(t, err)
		posReferenceCount := len(manager.AllReferences())
		assert.Equal(t, preReferenceCount, posReferenceCount)
	})

	t.Run("add already known reference", func(t *testing.T) {
		preReferenceCount := len(manager.AllReferences())
		preUniqueReferenceCount := len(manager.ReferenceMap())
		err = class.Set("valid.new.test2", data.NewValue("${person:age}"))
		assert.NoError(t, err)
		assert.Equal(t, preReferenceCount+1, len(manager.AllReferences()), "expected one more reference in all references")
		assert.Equal(t, preUniqueReferenceCount, len(manager.ReferenceMap()), "expected no additional unique references")
	})

	// ${person:age} is an already known reference
	// if it is added somewhere else and then removed, it must still be known
	// and the total number of references must also not have changed
	t.Run("remove already known reference", func(t *testing.T) {
		preReferenceCount := len(manager.AllReferences())
		preUniqueReferenceCount := len(manager.ReferenceMap())

		err = class.Set("valid.new.test3", data.NewValue("${person:age}"))
		assert.NoError(t, err)
		assert.Equal(t, preReferenceCount+1, len(manager.AllReferences()), "expected one more reference in all references")
		assert.Equal(t, preUniqueReferenceCount, len(manager.ReferenceMap()), "expected no additional unique references")

		err = class.Set("valid.new.test3", data.NewValue("reference vanished *shocked*"))
		assert.NoError(t, err)
		assert.Equal(t, preReferenceCount, len(manager.AllReferences()), "expected initial reference count")
		assert.Equal(t, preUniqueReferenceCount, len(manager.ReferenceMap()), "expected no additional unique references")
	})

	// ${person:occupation} is not known in the testdata
	// introduce it and verify that the manager knows about it afterwards
	t.Run("introduce new reference", func(t *testing.T) {
		preReferenceCount := len(manager.AllReferences())
		preUniqueReferenceCount := len(manager.ReferenceMap())

		err = class.Set("valid.new.test4", data.NewValue("${person:occupation}"))
		assert.NoError(t, err)
		assert.Equal(t, preReferenceCount+1, len(manager.AllReferences()), "expected one more reference in all references")
		assert.Equal(t, preUniqueReferenceCount+1, len(manager.ReferenceMap()), "expected one more additional unique reference")
	})

	t.Run("adding an invalid reference must fail", func(t *testing.T) {
		err = class.Set("valid.new.test5", data.NewValue("${valid:this:is:invalid}"))
		assert.ErrorIs(t, err, ErrInvalidReferenceTargetPath)
	})

	t.Run("adding a reference to a new path must only work if the target path is added first", func(t *testing.T) {
		err = class.Set("valid.this.is.invalid", data.NewValue("yeah, no this works :)"))
		assert.NoError(t, err)
		err = class.Set("valid.new.test5", data.NewValue("${valid:this:is:invalid}"))
		assert.NoError(t, err)
	})
}

func TestValueManager_SetHooks_Registry(t *testing.T) {
	person, err := NewClass("testdata/references/registry/person.yaml", codec.NewYamlCodec(), data.NewPath("person"))
	assert.NoError(t, err)
	greeting, err := NewClass("testdata/references/registry/greeting.yaml", codec.NewYamlCodec(), data.NewPath("greeting"))
	assert.NoError(t, err)

	registry := NewRegistry()
	err = registry.RegisterClass(person)
	assert.NoError(t, err)
	err = registry.RegisterClass(greeting)
	assert.NoError(t, err)

	manager, err := NewValueReferenceManager(registry)
	assert.NoError(t, err)

	// These are the same tests as above, just to be sure that also works.
	t.Run("expect no change if no reference is added/removed", func(t *testing.T) {
		preReferenceCount := len(manager.AllReferences())
		err = person.Set("test1", data.NewValue("hello there"))
		assert.NoError(t, err)
		posReferenceCount := len(manager.AllReferences())
		assert.Equal(t, preReferenceCount, posReferenceCount)
	})

	t.Run("adding an invalid reference must fail", func(t *testing.T) {
		err = person.Set("valid.new.test5", data.NewValue("${valid:this:is:invalid}"))
		assert.ErrorIs(t, err, ErrInvalidReferenceTargetPath)
	})

	t.Run("adding a reference to a new path must only work if the target path is added first", func(t *testing.T) {
		err = person.Set("valid.this.is.invalid", data.NewValue("yeah, no this works :)"))
		assert.NoError(t, err)
		err = person.Set("valid.new.test5", data.NewValue("${valid:this:is:invalid}"))
		assert.NoError(t, err)
	})

	// These tests are registry-specific as they use the postRegisterClass hook
	t.Run("register a new class with only valid references", func(t *testing.T) {
		common, err := NewClass("testdata/references/registry/common.yaml", codec.NewYamlCodec(), data.NewPath("common"))
		assert.NoError(t, err)

		err = registry.RegisterClass(common)
		assert.NoError(t, err)
	})
	t.Run("registering a new class with invalid references must fail", func(t *testing.T) {
		registry := NewRegistry()
		_, _ = NewValueReferenceManager(registry)

		// the common class is the only class, hence some references are invalid
		common, err := NewClass("testdata/references/registry/common.yaml", codec.NewYamlCodec(), data.NewPath("common"))
		assert.NoError(t, err)

		err = registry.RegisterClass(common)
		assert.ErrorIs(t, err, ErrInvalidReferenceTargetPath)
	})
}

func TestValueManager_SetHooks_Inventory(t *testing.T) {
	person, err := NewClass("testdata/references/inventory/data/person.yaml", codec.NewYamlCodec(), data.NewPath("person"))
	assert.NoError(t, err)
	greeting, err := NewClass("testdata/references/inventory/data/greeting.yaml", codec.NewYamlCodec(), data.NewPath("greeting"))
	assert.NoError(t, err)

	dataScope := NewRegistry()
	err = dataScope.RegisterClass(person)
	assert.NoError(t, err)
	err = dataScope.RegisterClass(greeting)
	assert.NoError(t, err)

	inventory, err := NewInventory()
	assert.NoError(t, err)
	err = inventory.RegisterScope(DataScope, dataScope)
	assert.NoError(t, err)

	manager, err := NewValueReferenceManager(inventory)
	assert.NoError(t, err)

	// These are the same tests as above, just to be sure that also works.
	t.Run("expect no change if no reference is added/removed", func(t *testing.T) {
		preReferenceCount := len(manager.AllReferences())
		err = person.Set("test1", data.NewValue("hello there"))
		assert.NoError(t, err)
		posReferenceCount := len(manager.AllReferences())
		assert.Equal(t, preReferenceCount, posReferenceCount)
	})

	t.Run("adding an invalid reference must fail", func(t *testing.T) {
		err = person.Set("valid.new.test5", data.NewValue("${valid:this:is:invalid}"))
		assert.ErrorIs(t, err, ErrInvalidReferenceTargetPath)
	})

	t.Run("adding a reference to a new path must only work if the target path is added first", func(t *testing.T) {
		err = person.Set("valid.this.is.invalid", data.NewValue("yeah, no this works :)"))
		assert.NoError(t, err)
		err = person.Set("valid.new.test5", data.NewValue("${valid:this:is:invalid}"))
		assert.NoError(t, err)
	})

	// Inventory specific tests
	t.Run("register scope which introduces valid references", func(t *testing.T) {
		testTarget, err := NewClass("testdata/references/inventory/targets/test.yaml", codec.NewYamlCodec(), data.NewPath("test"))
		assert.NoError(t, err)

		targetScope := NewRegistry()
		err = targetScope.RegisterClass(testTarget)
		assert.NoError(t, err)

		err = inventory.RegisterScope(TargetsScope, targetScope)
		assert.NoError(t, err)
	})

	t.Run("register scope which introduces an invalid references on a class after RegisterScope", func(t *testing.T) {
		inventory, err := NewInventory()
		assert.NoError(t, err)
		err = inventory.RegisterScope(DataScope, dataScope)
		assert.NoError(t, err)
		_, _ = NewValueReferenceManager(inventory)

		targetScope := NewRegistry()
		assert.NoError(t, err)

		testTarget, err := NewClass("testdata/references/inventory/targets/test.yaml", codec.NewYamlCodec(), data.NewPath("test"))
		assert.NoError(t, err)
		err = targetScope.RegisterClass(testTarget)
		assert.NoError(t, err)

		err = inventory.RegisterScope(TargetsScope, targetScope)
		assert.NoError(t, err)

		err = testTarget.Set("invalid.ref", data.NewValue("${data:does:not:exist}"))
		assert.ErrorIs(t, err, ErrInvalidReferenceTargetPath)
	})

	t.Run("register scope which introduces invalid references on RegisterScope", func(t *testing.T) {
		inventory, err := NewInventory()
		assert.NoError(t, err)
		err = inventory.RegisterScope(DataScope, dataScope)
		assert.NoError(t, err)
		_, _ = NewValueReferenceManager(inventory)

		targetScope := NewRegistry()
		assert.NoError(t, err)

		testTarget, err := NewClass("testdata/references/inventory/targets/test.yaml", codec.NewYamlCodec(), data.NewPath("test"))
		assert.NoError(t, err)
		err = targetScope.RegisterClass(testTarget)
		assert.NoError(t, err)

		err = testTarget.Set("invalid.ref", data.NewValue("${data:does:not:exist}"))
		assert.NoError(t, err)

		err = inventory.RegisterScope(TargetsScope, targetScope)
		assert.ErrorIs(t, err, ErrInvalidReferenceTargetPath)
	})
}

func TestValueManager_ReplaceReferences(t *testing.T) {
	filePath := "testdata/references/class/valid.yaml"
	class, err := NewClass(filePath, codec.NewYamlCodec(), data.NewPath("valid"))
	assert.NoError(t, err)
	assert.NotNil(t, class)

	manager, err := NewValueReferenceManager(class)
	assert.NoError(t, err)
	assert.NotNil(t, manager)

	expected := map[string]data.Value{
		"valid.person.age":       data.NewValue(35),
		"valid.greetings.casual": data.NewValue("Hey John"),
		"valid.greetings.formal": data.NewValue("Welcome, John Doe"),
		"valid.greetings.age":    data.NewValue("You are 35 years old"),
	}

	t.Run("valid class without modifications", func(t *testing.T) {
		err := manager.ReplaceReferences()
		assert.NoError(t, err)

		for path, expectedValue := range expected {
			val, err := class.Get(path)
			assert.NoError(t, err)
			assert.Equal(t, expectedValue.Raw, val.Raw)
		}
	})

	t.Run("replacing should be idempotent", func(t *testing.T) {
		err := manager.ReplaceReferences()
		assert.NoError(t, err)
		for path, expectedValue := range expected {
			val, err := class.Get(path)
			assert.NoError(t, err)
			assert.Equal(t, expectedValue.Raw, val.Raw)
		}

		err = manager.ReplaceReferences()
		assert.NoError(t, err)
		for path, expectedValue := range expected {
			val, err := class.Get(path)
			assert.NoError(t, err)
			assert.Equal(t, expectedValue.Raw, val.Raw)
		}
	})

	t.Run("valid class but one reference is overwritten with a literal before replacement", func(t *testing.T) {
		// we need a clean class and manager in this case, otherwise the references have already been replaced
		filePath := "testdata/references/class/valid.yaml"
		class, err := NewClass(filePath, codec.NewYamlCodec(), data.NewPath("valid"))
		assert.NoError(t, err)
		assert.NotNil(t, class)

		manager, err := NewValueReferenceManager(class)
		assert.NoError(t, err)
		assert.NotNil(t, manager)

		err = class.Set("valid.person.age", "over 9000") // lol
		assert.NoError(t, err)
		err = manager.ReplaceReferences()
		assert.NoError(t, err)

		expected := map[string]data.Value{
			"valid.person.age":       data.NewValue("over 9000"),
			"valid.greetings.casual": data.NewValue("Hey John"),
			"valid.greetings.formal": data.NewValue("Welcome, John Doe"),
			"valid.greetings.age":    data.NewValue("You are over 9000 years old"),
		}

		for path, expectedValue := range expected {
			val, err := class.Get(path)
			assert.NoError(t, err)
			assert.Equal(t, expectedValue.Raw, val.Raw)
		}
	})
}

func TestValueManager_ReplaceReferences_Registry(t *testing.T) {
	person, err := NewClass("testdata/references/registry/person.yaml", codec.NewYamlCodec(), data.NewPath("person"))
	assert.NoError(t, err)
	greeting, err := NewClass("testdata/references/registry/greeting.yaml", codec.NewYamlCodec(), data.NewPath("greeting"))
	assert.NoError(t, err)

	registry := NewRegistry()
	err = registry.RegisterClass(person)
	assert.NoError(t, err)
	err = registry.RegisterClass(greeting)
	assert.NoError(t, err)

	manager, err := NewValueReferenceManager(registry)
	assert.NoError(t, err)

	expected := map[string]data.Value{
		"person.n":        data.NewValue(35),
		"greeting.casual": data.NewValue("Hey, John"),
	}

	t.Run("replace valid registry", func(t *testing.T) {
		err = manager.ReplaceReferences()
		assert.NoError(t, err)
		for path, expectedValue := range expected {
			val, err := registry.Get(path)
			assert.NoError(t, err)
			assert.Equal(t, expectedValue.Raw, val.Raw)
		}
	})

	t.Run("replacing should be idempotent", func(t *testing.T) {
		err := manager.ReplaceReferences()
		assert.NoError(t, err)
		for path, expectedValue := range expected {
			val, err := registry.Get(path)
			assert.NoError(t, err)
			assert.Equal(t, expectedValue.Raw, val.Raw)
		}

		err = manager.ReplaceReferences()
		assert.NoError(t, err)
		for path, expectedValue := range expected {
			val, err := registry.Get(path)
			assert.NoError(t, err)
			assert.Equal(t, expectedValue.Raw, val.Raw)
		}
	})

	t.Run("replace again after adding a new class", func(t *testing.T) {
		common, err := NewClass("testdata/references/registry/common.yaml", codec.NewYamlCodec(), data.NewPath("common"))
		assert.NoError(t, err)
		err = registry.RegisterClass(common)
		assert.NoError(t, err)

		// extend expected map
		expected["common.bar"] = data.NewValue("bar")
		expected["common.greeting"] = data.NewValue("Hey,")

		err = manager.ReplaceReferences()
		assert.NoError(t, err)
		for path, expectedValue := range expected {
			val, err := registry.Get(path)
			assert.NoError(t, err)
			assert.Equal(t, expectedValue.Raw, val.Raw)
		}
	})
}

func TestValueManager_ReplaceReferences_Inventory(t *testing.T) {
	person, err := NewClass("testdata/references/inventory/data/person.yaml", codec.NewYamlCodec(), data.NewPath("person"))
	assert.NoError(t, err)
	greeting, err := NewClass("testdata/references/inventory/data/greeting.yaml", codec.NewYamlCodec(), data.NewPath("greeting"))
	assert.NoError(t, err)

	// data scope
	dataScope := NewRegistry()
	err = dataScope.RegisterClass(person)
	assert.NoError(t, err)
	err = dataScope.RegisterClass(greeting)
	assert.NoError(t, err)

	inventory, err := NewInventory()
	assert.NoError(t, err)
	err = inventory.RegisterScope(DataScope, dataScope)
	assert.NoError(t, err)

	manager, err := NewValueReferenceManager(inventory)
	assert.NoError(t, err)

	expected := map[string]data.Value{
		"data.person.age": data.NewValue(35),
	}

	t.Run("replace valid inventory", func(t *testing.T) {
		err = manager.ReplaceReferences()
		assert.NoError(t, err)
		for path, expectedValue := range expected {
			val, err := inventory.Get(path)
			assert.NoError(t, err)
			assert.Equal(t, expectedValue.Raw, val.Raw)
		}
	})

	t.Run("replacing should be idempotent", func(t *testing.T) {
		err := manager.ReplaceReferences()
		assert.NoError(t, err)
		for path, expectedValue := range expected {
			val, err := inventory.Get(path)
			assert.NoError(t, err)
			assert.Equal(t, expectedValue.Raw, val.Raw)
		}

		err = manager.ReplaceReferences()
		assert.NoError(t, err)
		for path, expectedValue := range expected {
			val, err := inventory.Get(path)
			assert.NoError(t, err)
			assert.Equal(t, expectedValue.Raw, val.Raw)
		}
	})

	t.Run("replace again after adding a new scope", func(t *testing.T) {
		// target scope
		testTarget, err := NewClass("testdata/references/inventory/targets/test.yaml", codec.NewYamlCodec(), data.NewPath("test"))
		assert.NoError(t, err)
		targetScope := NewRegistry()
		assert.NoError(t, err)
		err = targetScope.RegisterClass(testTarget)
		assert.NoError(t, err)
		err = inventory.RegisterScope(TargetsScope, targetScope)
		assert.NoError(t, err)

		expected["targets.test.casual"] = data.NewValue("Hey, John")
		expected["targets.test.formal"] = data.NewValue("Hello, John Doe")

		err = manager.ReplaceReferences()
		assert.NoError(t, err)
		for path, expectedValue := range expected {
			val, err := inventory.Get(path)
			assert.NoError(t, err)
			assert.Equal(t, expectedValue.Raw, val.Raw)
		}
	})
}
