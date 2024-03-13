package skipper_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lukasjarosch/skipper"
	"github.com/lukasjarosch/skipper/codec"
)

func TestPluginManager_LoadPlugin(t *testing.T) {
	pm := skipper.NewPluginManager()

	path := "./plugins/output/copy/plugin.so"
	err := pm.LoadPlugin(path)
	assert.NoError(t, err)
}

func TestPluginManager_InitializyPlugin(t *testing.T) {
	pm := skipper.NewPluginManager()
	path := "./plugins/output/copy/plugin.so"
	err := pm.LoadPlugin(path)
	assert.NoError(t, err)

	rootPath := "testdata/plugins/output/inventory/config"
	registry, err := skipper.LoadRegistry(rootPath, codec.NewYamlCodec(), codec.YamlPathSelector)
	assert.NoError(t, err)
	inv, err := skipper.NewInventory()
	assert.NoError(t, err)
	err = inv.RegisterScope(skipper.ConfigScope, registry)
	assert.NoError(t, err)

	config, err := inv.Get("config.plugins.output.copy")
	assert.NoError(t, err)
	err = pm.ConfigurePlugin(skipper.OutputPlugin, "copy", config)
	assert.NoError(t, err)
}

// func TestLoadPlugin(t *testing.T) {
// 	plugin, err := skipper.LoadPlugin("./plugins/output/copy/plugin.so")
// 	assert.NoError(t, err)
// 	assert.NotNil(t, plugin)
// }
//
// func TestInitializePlugin(t *testing.T) {
// 	rootPath := "testdata/plugins/output/inventory/config"
// 	registry, err := skipper.LoadRegistry(rootPath, codec.NewYamlCodec(), codec.YamlPathSelector)
// 	assert.NoError(t, err)
// 	inv, err := skipper.NewInventory()
// 	assert.NoError(t, err)
// 	err = inv.RegisterScope(skipper.ConfigScope, registry)
// 	assert.NoError(t, err)
// }
