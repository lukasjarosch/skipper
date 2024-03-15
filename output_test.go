package skipper_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lukasjarosch/skipper"
	"github.com/lukasjarosch/skipper/codec"
	"github.com/lukasjarosch/skipper/data"
	"github.com/lukasjarosch/skipper/output"
)

func TestOutputManager_RegisterOutput(t *testing.T) {
	om := skipper.NewOutputManager()

	err := om.RegisterOutput(output.CopyOutputType, output.NewCopyOutput())
	assert.NoError(t, err)
}

func TestOutputManager_ConfigureOutput(t *testing.T) {
	rootPath := "testdata/output/inventory/config"
	registry, err := skipper.LoadRegistry(rootPath, codec.NewYamlCodec(), codec.YamlPathSelector)
	assert.NoError(t, err)

	inv, err := skipper.NewInventory()
	assert.NoError(t, err)
	inv.RegisterGlobal("root_path", data.NewValue("FOO BAR"))
	err = inv.RegisterScope(skipper.ConfigScope, registry)
	assert.NoError(t, err)

	refManager, err := skipper.NewValueReferenceManager(inv)
	assert.NoError(t, err)
	err = refManager.ReplaceReferences()
	assert.NoError(t, err)

	om := skipper.NewOutputManager()
	err = om.RegisterOutput(output.CopyOutputType, output.NewCopyOutput())
	assert.NoError(t, err)

	config, err := inv.Get("config.output.copy")
	assert.NoError(t, err)
	err = om.ConfigureOutput(output.CopyOutputType, config)
	assert.NoError(t, err)

	err = om.RunAll()
	assert.NoError(t, err)
}
