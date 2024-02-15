package skipper_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/lukasjarosch/skipper"
	"github.com/lukasjarosch/skipper/codec"
	"github.com/lukasjarosch/skipper/data"
)

var expectedReferencesSimple = []Reference{
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

func TestParseReferencesSimple(t *testing.T) {
	_, err := ParseReferences(nil)
	assert.ErrorIs(t, err, ErrReferenceSourceIsNil)

	class, err := NewClass("testdata/references/local/simple.yaml", codec.NewYamlCodec(), data.NewPathFromOsPath("simple"))
	assert.NoError(t, err)

	references, err := ParseReferences(class)
	assert.NoError(t, err)
	for _, reference := range references {
		assert.Contains(t, expectedReferencesSimple, reference)
	}
}

func TestResolveReferencesSimple(t *testing.T) {
	class, err := NewClass("testdata/references/local/simple.yaml", codec.NewYamlCodec(), data.NewPathFromOsPath("simple"))
	assert.NoError(t, err)

	// Test: resolve all valid references which have a direct TargetValue
	resolved, err := ResolveReferences(expectedReferencesSimple, class)
	assert.NoError(t, err)
	assert.Len(t, resolved, len(expectedReferencesSimple), "Every Reference should emit a ResolveReference")
	for _, resolved := range resolved {
		assert.Contains(t, expectedReferencesSimple, resolved, "ResolvedReference should be returned")
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
}

func TestReplaceReferencesSimple(t *testing.T) {
	class, err := NewClass("testdata/references/local/simple.yaml", codec.NewYamlCodec(), data.NewPathFromOsPath("simple"))
	assert.NoError(t, err)
	resolved, err := ResolveReferences(expectedReferencesSimple, class)
	assert.NoError(t, err)
	assert.Len(t, resolved, len(expectedReferencesSimple))

	expected := map[string]data.Value{
		"simple.departments.engineering.manager": data.NewValue("John Doe"),
		"simple.departments.analytics.manager":   data.NewValue("Jane Smith"),
		"simple.departments.marketing.manager":   data.NewValue("Michael Johnson"),
		"simple.projects.Project_X.department":   data.NewValue("Engineering"),
	}

	err = ReplaceReferences(resolved, class)
	assert.NoError(t, err)
	for targetPath, expectedValue := range expected {
		value, err := class.Get(targetPath)
		assert.NoError(t, err)
		assert.Equal(t, expectedValue, value)
	}

	// there must be no more references after replacing
	parsedReferences, err := ParseReferences(class)
	assert.NoError(t, err)
	assert.Len(t, parsedReferences, 0)
}

var expectedReferencesNested = []Reference{
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
	{
		Path:       data.NewPath("nested.target_nested_mixed"),
		TargetPath: data.NewPath("nested_mixed"),
	},
}

func TestParseReferencesNested(t *testing.T) {
	class, err := NewClass("testdata/references/local/nested.yaml", codec.NewYamlCodec(), data.NewPathFromOsPath("nested"))
	assert.NoError(t, err)

	references, err := ParseReferences(class)
	assert.NoError(t, err)
	assert.NotNil(t, references)
	for _, reference := range references {
		assert.Contains(t, expectedReferencesNested, reference)
	}
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
		{
			Path:       data.NewPath("nested.target_nested_mixed"),
			TargetPath: data.NewPath("nested_mixed"),
		},
	}
	assert.Len(t, resolved, len(expected))
	for _, res := range resolved {
		assert.Contains(t, expected, res)
	}
}

func TestReplaceReferencesNested(t *testing.T) {
	class, err := NewClass("testdata/references/local/nested.yaml", codec.NewYamlCodec(), data.NewPathFromOsPath("nested"))
	assert.NoError(t, err)

	references, err := ParseReferences(class)
	assert.NoError(t, err)
	assert.NotNil(t, references)

	resolved, err := ResolveReferences(references, class)
	assert.NoError(t, err)
	assert.Len(t, resolved, len(references))

	expected := map[string]data.Value{
		"nested.target": data.NewValue(map[string]interface{}{
			"foo": "bar",
			"bar": "baz",
		},
		),
		"nested.target_array": data.NewValue([]interface{}{
			"foo", "bar", "baz",
		}),
		"nested.target_nested_map": data.NewValue(map[string]interface{}{
			"foo": map[string]interface{}{
				"bar": map[string]interface{}{
					"baz": "qux",
				},
			},
		}),
		"nested.target_nested_mixed": data.NewValue(map[string]interface{}{
			"foo": []interface{}{
				map[string]interface{}{
					"bar": "baz",
				},
				"test",
				map[string]interface{}{
					"foo": map[string]interface{}{
						"bar": "baz",
					},
				},
				map[string]interface{}{
					"array": []interface{}{
						"one", "two", "three",
					},
				},
			},
		}),
	}

	err = ReplaceReferences(resolved, class)
	assert.NoError(t, err)
	for targetPath, expectedValue := range expected {
		value, err := class.Get(targetPath)
		assert.NoError(t, err)
		assert.Equal(t, expectedValue, value)
	}

	// there must be no more references after replacing
	parsedReferences, err := ParseReferences(class)
	assert.NoError(t, err)
	assert.Len(t, parsedReferences, 0)
}

var expectedReferencesChained = []Reference{
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

func TestParseReferencesChained(t *testing.T) {
	class, err := NewClass("testdata/references/local/chained.yaml", codec.NewYamlCodec(), data.NewPathFromOsPath("chained"))
	assert.NoError(t, err)

	references, err := ParseReferences(class)
	assert.NoError(t, err)
	assert.NotNil(t, references)
	for _, reference := range references {
		assert.Contains(t, expectedReferencesChained, reference)
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

	assert.Equal(t, resolved, expectedReferencesChained)
}

func TestReplaceReferencesChained(t *testing.T) {
	class, err := NewClass("testdata/references/local/chained.yaml", codec.NewYamlCodec(), data.NewPathFromOsPath("chained"))
	assert.NoError(t, err)

	references, err := ParseReferences(class)
	assert.NoError(t, err)
	assert.NotNil(t, references)

	resolved, err := ResolveReferences(references, class)
	assert.NoError(t, err)
	assert.Len(t, resolved, len(references))

	expected := map[string]data.Value{
		"chained.greeting":         data.NewValue("Hello, John"),
		"chained.gotcha":           data.NewValue("John"),
		"chained.first_name":       data.NewValue("John"),
		"chained.name_placeholder": data.NewValue("John"),
	}

	err = ReplaceReferences(resolved, class)
	assert.NoError(t, err)
	for targetPath, expectedValue := range expected {
		value, err := class.Get(targetPath)
		assert.NoError(t, err)
		assert.Equal(t, expectedValue, value)
	}

	// there must be no more references after replacing
	parsedReferences, err := ParseReferences(class)
	assert.NoError(t, err)
	assert.Len(t, parsedReferences, 0)
}

func TestParseReferencesCycle(t *testing.T) {
	class, err := NewClass("testdata/references/local/cycle.yaml", codec.NewYamlCodec(), data.NewPathFromOsPath("cycle"))
	assert.NoError(t, err)

	expected := []Reference{
		{
			Path:       data.NewPath("cycle.john"),
			TargetPath: data.NewPath("middle"),
		},
		{
			Path:       data.NewPath("cycle.name"),
			TargetPath: data.NewPath("john"),
		},
		{
			Path:       data.NewPath("cycle.middle"),
			TargetPath: data.NewPath("name"),
		},
	}

	references, err := ParseReferences(class)
	assert.NoError(t, err)
	assert.NotNil(t, references)
	for _, reference := range references {
		assert.Contains(t, expected, reference)
	}
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

var expectedReferencesMulti = []Reference{
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

func TestParseReferencesMulti(t *testing.T) {
	class, err := NewClass("testdata/references/local/multi.yaml", codec.NewYamlCodec(), data.NewPathFromOsPath("multi"))
	assert.NoError(t, err)

	references, err := ParseReferences(class)
	assert.NoError(t, err)
	assert.NotNil(t, references)
	for _, reference := range references {
		assert.Contains(t, expectedReferencesMulti, reference)
	}
}

func TestResolveReferencesMulti(t *testing.T) {
	class, err := NewClass("testdata/references/local/multi.yaml", codec.NewYamlCodec(), data.NewPathFromOsPath("multi"))
	assert.NoError(t, err)

	references, err := ParseReferences(class)
	assert.NoError(t, err)
	assert.NotNil(t, references)
	assert.Len(t, references, len(expectedReferencesMulti))

	resolved, err := ResolveReferences(references, class)
	assert.NoError(t, err)
	assert.Len(t, resolved, len(expectedReferencesMulti))

	// The TopologicalSort does always produce a correct
	// result, but the order still may change as there
	// is usually more than one solution.
	// Hence we only check that the sincle nested reference
	// is properly resolved. Here the 'multi.common.repo_url'
	// MUST be the first reference to be resolved.
	assert.Equal(t, resolved[0], expectedReferencesMulti[4])
}

func TestReplaceReferencesMulti(t *testing.T) {
	class, err := NewClass("testdata/references/local/multi.yaml", codec.NewYamlCodec(), data.NewPathFromOsPath("multi"))
	assert.NoError(t, err)

	references, err := ParseReferences(class)
	assert.NoError(t, err)
	assert.NotNil(t, references)

	resolved, err := ResolveReferences(references, class)
	assert.NoError(t, err)
	assert.Len(t, resolved, len(references))

	expected := map[string]data.Value{
		"multi.project.repo":        data.NewValue("github.com/lukasjarosch/skipper"),
		"multi.project.description": data.NewValue("The skipper project is very cool. Because skipper helps in working with the Infrastructure as Data concept. The project 'skipper' is hosted on github.com/lukasjarosch/skipper.\n"),
	}

	err = ReplaceReferences(resolved, class)
	assert.NoError(t, err)
	for targetPath, expectedValue := range expected {
		value, err := class.Get(targetPath)
		assert.NoError(t, err)
		assert.Equal(t, expectedValue, value)
	}

	// there must be no more references after replacing
	parsedReferences, err := ParseReferences(class)
	assert.NoError(t, err)
	assert.Len(t, parsedReferences, 0)
}
