package data_test

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	. "github.com/lukasjarosch/skipper/data"
	"github.com/stretchr/testify/assert"
)

func TestNewContainer(t *testing.T) {
	containerName := "test"
	d := map[string]interface{}{
		containerName: map[string]interface{}{
			"foo": map[string]interface{}{
				"bar": map[string]interface{}{
					"baz": "hello",
				},
			},
		},
	}

	container, err := NewContainer(containerName, d)
	assert.NoError(t, err)
	assert.NotNil(t, container)
}

func TestContainer_Get(t *testing.T) {
	containerName := "test"
	d := map[string]interface{}{
		containerName: map[string]interface{}{
			"foo": map[string]interface{}{
				"bar": map[string]interface{}{
					"baz": "hello",
				},
			},
		},
	}
	container, err := NewContainer(containerName, d)
	assert.NoError(t, err)
	assert.NotNil(t, container)

	ret, err := container.Get(NewPathVar(containerName, "foo"))

	spew.Dump(ret, err)
}

func TestContainer_LeafPaths(t *testing.T) {
	containerName := "test"
	d := map[string]interface{}{
		containerName: map[string]interface{}{
			"foo": map[string]interface{}{
				"bar": map[string]interface{}{
					"baz": "hello",
					"qux": "ohai",
				},
			},
		},
	}
	container, err := NewContainer(containerName, d)
	assert.NoError(t, err)
	assert.NotNil(t, container)

	paths := container.LeafPaths()

	spew.Dump(paths, err)
}

func TestContainer_Set(t *testing.T) {
	containerName := "test"
	d := map[string]interface{}{
		containerName: map[string]interface{}{
			"foo": map[string]interface{}{
				"bar": map[string]interface{}{
					"baz": "hello",
					"qux": "ohai",
				},
			},
			"array": []interface{}{
				[]interface{}{"one"},
				[]interface{}{
					[]interface{}{
						"two",
					},
				},
			},
		},
	}
	container, err := NewContainer(containerName, d)
	assert.NoError(t, err)
	assert.NotNil(t, container)

	err = container.Set(NewPath("test.array.1.0"), []interface{}{1, 2, 3})
}

func TestContainer_Merge(t *testing.T) {
	containerName := "test"
	d := map[string]interface{}{
		containerName: map[string]interface{}{
			"foo": map[string]interface{}{
				"bar": map[string]interface{}{
					"baz": "hello",
					"qux": "ohai",
				},
			},
			"array": []interface{}{
				[]interface{}{"one"},
				[]interface{}{
					[]interface{}{
						"two",
					},
				},
			},
		},
	}
	container, err := NewContainer(containerName, d)
	assert.NoError(t, err)
	assert.NotNil(t, container)

	mergeData := map[string]interface{}{
		"array": []interface{}{
			[]interface{}{"one", "two"},
			[]interface{}{
				[]interface{}{
					"three", 4, 5,
				},
			},
		},
	}

	_ = mergeData
	err = container.Merge(NewPath("test"), mergeData)
	if err != nil {
		panic(err)
	}

}
