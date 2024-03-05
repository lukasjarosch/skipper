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
		// NOTE: at this point in the test, all age references have been replaced with '35'
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
