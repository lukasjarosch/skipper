package skipper_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/lukasjarosch/skipper"
	"github.com/lukasjarosch/skipper/codec"
	"github.com/lukasjarosch/skipper/data"
)

func TestValueManager_preSetHook(t *testing.T) {
	filePath := "testdata/references/class/valid.yaml"
	class, err := NewClass(filePath, codec.NewYamlCodec(), data.NewPath("valid"))
	assert.NoError(t, err)
	assert.NotNil(t, class)

	manager, err := NewValueReferenceManager(class)
	assert.NoError(t, err)
	assert.NotNil(t, manager)

	// TEST: setting a path to a value without reference should not change anything
	preReferenceCount := len(manager.AllReferences())
	err = class.Set("valid.new.test", data.NewValue("ohai"))
	assert.NoError(t, err)
	posReferenceCount := len(manager.AllReferences())
	assert.Equal(t, preReferenceCount, posReferenceCount)

	// TEST: introducing an additional, but existing, value reference
	preReferenceCount = len(manager.AllReferences())
	preUniqueReferenceCount := len(manager.ReferenceMap())
	err = class.Set("valid.new.test", data.NewValue("${person:age}"))
	assert.NoError(t, err)
	assert.Equal(t, preReferenceCount+1, len(manager.AllReferences()), "expected one more reference in all references")
	assert.Equal(t, preUniqueReferenceCount, len(manager.ReferenceMap()), "expected no additional unique references")

	// TEST: removing the previous, valid but known, reference
	preReferenceCount = len(manager.AllReferences())
	preUniqueReferenceCount = len(manager.ReferenceMap())
	err = class.Set("valid.new.test", data.NewValue("reference vanished *shocked*"))
	assert.NoError(t, err)
	assert.Equal(t, preReferenceCount-1, len(manager.AllReferences()), "expected one more reference in all references")
	assert.Equal(t, preUniqueReferenceCount, len(manager.ReferenceMap()), "expected no additional unique references")
}
