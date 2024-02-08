package skipper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"github.com/lukasjarosch/skipper"
	"github.com/lukasjarosch/skipper/codec"
)

func TestNewClass(t *testing.T) {
	path := "/tmp/foo.yaml"

	class, err := skipper.NewClass(path, codec.NewYamlCodec())
	assert.NoError(t, err)
	assert.NotNil(t, class)

	fmt.Println(class.Get("foo.bar"))
	class.Set("foo.baz", "changed")
	err = class.Set("foo.new.bar.baz.foo.test.ohai", map[string]interface{}{
		"hello": "world",
	})
	err = class.Set("foo.ohai.35.test", map[string]interface{}{
		"hello": "world",
	})
	assert.NoError(t, err)
	fmt.Println(class.Get("foo.ohai.23"))

	d := class.GetAll()
	o, _ := yaml.Marshal(d.Raw)
	fmt.Println(string(o))
}
