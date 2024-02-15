package skipper_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/lukasjarosch/skipper"
	"github.com/lukasjarosch/skipper/codec"
	"github.com/lukasjarosch/skipper/data"
)

var localReferences = []Reference{
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
	assert.Len(t, resolved, len(localReferences), "Every Reference should emit a ResolveReference")
	for _, resolved := range resolved {
		assert.Contains(t, localReferences, resolved, "ResolvedReference should be returned")
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
	assert.ErrorIs(t, err, ErrUndefinedReferenceTarget)
	assert.Nil(t, resolved)

	// TODO: reference to reference
	// TODO: cycle
}

func TestResolveReferencesNested(t *testing.T) {
	class, err := NewClass("testdata/references/local/nested.yaml", codec.NewYamlCodec(), data.NewPathFromOsPath("nested"))
	assert.NoError(t, err)

	references, err := ParseReferences(class)
	assert.NoError(t, err)
	assert.NotNil(t, references)

	resolved, err := ResolveReferences(references, class)
	assert.NoError(t, err)
	assert.Len(t, resolved, len(references))

	expected := []Reference{
		{
			Path:       data.NewPath("nested.target"),
			TargetPath: data.NewPath("source"),
		},
		{
			Path:       data.NewPath("nested.target_array"),
			TargetPath: data.NewPath("source_array"),
		},
		{
			Path:       data.NewPath("nested.target_nested_map"),
			TargetPath: data.NewPath("nested_map"),
		},
	}
	assert.Len(t, resolved, len(expected))
	for _, res := range resolved {
		assert.Contains(t, expected, res)
	}
}

func TestResolveReferencesChained(t *testing.T) {
	class, err := NewClass("testdata/references/local/chained.yaml", codec.NewYamlCodec(), data.NewPathFromOsPath("chained"))
	assert.NoError(t, err)

	references, err := ParseReferences(class)
	assert.NoError(t, err)
	assert.NotNil(t, references)

	resolved, err := ResolveReferences(references, class)
	assert.NoError(t, err)
	assert.Len(t, resolved, len(references))

	expected := []Reference{
		{
			Path:       data.NewPath("chained.gotcha"),
			TargetPath: data.NewPath("chained.john.first_name"),
		},
		{
			Path:       data.NewPath("chained.name_placeholder"),
			TargetPath: data.NewPath("gotcha"),
		},
		{
			Path:       data.NewPath("chained.first_name"),
			TargetPath: data.NewPath("name_placeholder"),
		},
		{
			Path:       data.NewPath("chained.greeting"),
			TargetPath: data.NewPath("first_name"),
		},
	}

	assert.Equal(t, resolved, expected)
}

func TestResolveReferencesCycle(t *testing.T) {
	class, err := NewClass("testdata/references/local/cycle.yaml", codec.NewYamlCodec(), data.NewPathFromOsPath("cycle"))
	assert.NoError(t, err)

	references, err := ParseReferences(class)
	assert.NoError(t, err)
	assert.NotNil(t, references)

	resolved, err := ResolveReferences(references, class)
	assert.ErrorIs(t, err, ErrReferenceCycle)
	assert.Len(t, resolved, 0)
}

func TestResolveReferencesMulti(t *testing.T) {
	class, err := NewClass("testdata/references/local/multi.yaml", codec.NewYamlCodec(), data.NewPathFromOsPath("multi"))
	assert.NoError(t, err)

	expected := []Reference{
		{
			Path:       data.NewPath("multi.project.description"),
			TargetPath: data.NewPath("project.name"),
		},
		{
			Path:       data.NewPath("multi.project.description"),
			TargetPath: data.NewPath("multi.project.name"),
		},
		{
			Path:       data.NewPath("multi.project.description"),
			TargetPath: data.NewPath("project.name"),
		},
		{
			Path:       data.NewPath("multi.project.description"),
			TargetPath: data.NewPath("multi.project.repo"),
		},
		{
			Path:       data.NewPath("multi.project.repo"),
			TargetPath: data.NewPath("multi.common.repo_url"),
		},
	}

	references, err := ParseReferences(class)
	assert.NoError(t, err)
	assert.NotNil(t, references)
	assert.Len(t, references, len(expected))

	resolved, err := ResolveReferences(references, class)
	assert.NoError(t, err)
	assert.Len(t, resolved, len(expected))

	// The TopologicalSort does always produce a correct
	// result, but the order still may change as there
	// is usually more than one solution.
	// Hence we only check that the sincle nested reference
	// is properly resolved. Here the 'multi.common.repo_url'
	// MUST be the first reference to be resolved.
	assert.Equal(t, resolved[0], expected[4])
}
