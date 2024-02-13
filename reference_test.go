package skipper_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/lukasjarosch/skipper"
	"github.com/lukasjarosch/skipper/codec"
	"github.com/lukasjarosch/skipper/data"
)

// func makeInventory(t *testing.T) *Inventory {
// 	dataFiles, err := DiscoverFiles("testdata/references/data", codec.YamlPathSelector)
// 	assert.NoError(t, err)
// 	dataRegistry, err := NewRegistryFromFiles(dataFiles, func(filePaths []string) ([]*Class, error) {
// 		return ClassLoader("testdata/references/data", filePaths, codec.NewYamlCodec())
// 	})
// 	assert.NoError(t, err)
//
// 	targetFiles, err := DiscoverFiles("testdata/references/targets", codec.YamlPathSelector)
// 	assert.NoError(t, err)
// 	targetsRegistry, err := NewRegistryFromFiles(targetFiles, func(filePaths []string) ([]*Class, error) {
// 		return ClassLoader("testdata/references/targets", filePaths, codec.NewYamlCodec())
// 	})
// 	assert.NoError(t, err)
//
// 	inventory, err := NewInventory()
// 	assert.NoError(t, err)
//
// 	err = inventory.RegisterScope(DataScope, dataRegistry)
// 	assert.NoError(t, err)
// 	err = inventory.RegisterScope(TargetsScope, targetsRegistry)
// 	assert.NoError(t, err)
//
// 	err = inventory.SetDefaultScope(DataScope)
// 	assert.NoError(t, err)
//
// 	return inventory
// }

var (
	localReferences = []Reference{
		{
			Path:       data.NewPath("simple.departments.engineering.manager"),
			TargetPath: data.NewPath("employees.0.name"),
		},
		{
			Path:       data.NewPath("simple.departments.analytics.manager"),
			TargetPath: data.NewPath("simple.employees.1.name"),
		},
		{
			Path:       data.NewPath("simple.departments.marketing.manager"),
			TargetPath: data.NewPath("simple.employees.2.name"),
		},
		{
			Path:       data.NewPath("simple.projects.Project_X.department"),
			TargetPath: data.NewPath("simple.departments.engineering.name"),
		},
	}
	localResolvedReferences = []ResolvedReference{
		{
			Reference: Reference{
				Path:       data.NewPath("simple.departments.engineering.manager"),
				TargetPath: data.NewPath("employees.0.name"),
			},
			TargetValue:     data.NewValue("John Doe"),
			TargetReference: nil,
		},
		{
			Reference: Reference{
				Path:       data.NewPath("simple.departments.analytics.manager"),
				TargetPath: data.NewPath("simple.employees.1.name"),
			},
			TargetValue:     data.NewValue("Jane Smith"),
			TargetReference: nil,
		},
		{
			Reference: Reference{
				Path:       data.NewPath("simple.departments.marketing.manager"),
				TargetPath: data.NewPath("simple.employees.2.name"),
			},
			TargetValue:     data.NewValue("Michael Johnson"),
			TargetReference: nil,
		},
		{
			Reference: Reference{
				Path:       data.NewPath("simple.projects.Project_X.department"),
				TargetPath: data.NewPath("simple.departments.engineering.name"),
			},
			TargetValue:     data.NewValue("Engineering"),
			TargetReference: nil,
		},
	}
)

func TestParseReferences(t *testing.T) {
	_, err := ParseReferences(nil)
	assert.ErrorIs(t, err, ErrReferenceSourceIsNil)

	class, err := NewClass("testdata/references/local/simple.yaml", codec.NewYamlCodec(), data.NewPathFromOsPath("simple"))
	assert.NoError(t, err)

	references, err := ParseReferences(class)
	assert.NoError(t, err)
	for _, reference := range references {
		assert.Contains(t, localReferences, reference)
	}
}

func TestResolveReferencesSimple(t *testing.T) {
	class, err := NewClass("testdata/references/local/simple.yaml", codec.NewYamlCodec(), data.NewPathFromOsPath("simple"))
	assert.NoError(t, err)

	// Test: resolve all valid references which have a direct TargetValue
	resolved, err := ResolveReferences(localReferences, class)
	assert.NoError(t, err)
	assert.Len(t, resolved, len(localResolvedReferences), "Every Reference should emit a ResolveReference")
	for _, resolved := range resolved {
		assert.Contains(t, localResolvedReferences, resolved, "ResolvedReference should be returned")
		assert.Nil(t, resolved.TargetReference)
	}

	// Test: references with invalid TargetPath
	invalidReferences := []Reference{
		{
			Path:       data.NewPath("simple.departments.marketing.name"),
			TargetPath: data.NewPath("invalid.path"),
		},
		{
			Path:       data.NewPath("simple.departments.marketing.name"),
			TargetPath: data.NewPath("another.invalid.path"),
		},
	}
	resolved, err = ResolveReferences(invalidReferences, class)
	assert.ErrorIs(t, err, ErrUndefinedReference)
	assert.Nil(t, resolved)

	// TODO: reference to reference
	// TODO: cycle
}

func TestResolveReferencesMap(t *testing.T) {
	class, err := NewClass("testdata/references/local/nested.yaml", codec.NewYamlCodec(), data.NewPathFromOsPath("nested"))
	assert.NoError(t, err)

	references, err := ParseReferences(class)
	assert.NoError(t, err)
	assert.NotNil(t, references)

	resolved, err := ResolveReferences(references, class)
	assert.NoError(t, err)
	assert.Len(t, resolved, len(references))

	expected := []ResolvedReference{
		{
			Reference: Reference{
				Path:       data.NewPath("nested.target"),
				TargetPath: data.NewPath("source"),
			},
			TargetReference: nil,
			TargetValue: data.NewValue(map[string]interface{}{
				"foo": "bar",
				"bar": "baz",
			}),
		},
		{
			Reference: Reference{
				Path:       data.NewPath("nested.target_array"),
				TargetPath: data.NewPath("source_array"),
			},
			TargetReference: nil,
			TargetValue:     data.NewValue([]interface{}{"foo", "bar", "baz"}),
		},
		{
			Reference: Reference{
				Path:       data.NewPath("nested.target_nested_map"),
				TargetPath: data.NewPath("nested_map"),
			},
			TargetReference: nil,
			TargetValue: data.NewValue(map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": map[string]interface{}{
						"baz": "qux",
					},
				},
			}),
		},
	}
	assert.Len(t, resolved, len(expected))
	for _, res := range resolved {
		assert.Contains(t, expected, res)
		assert.Nil(t, res.TargetReference)
	}
}
