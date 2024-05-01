package reference_test

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/lukasjarosch/skipper/v1/data"
	mocks "github.com/lukasjarosch/skipper/v1/mocks/reference"
	. "github.com/lukasjarosch/skipper/v1/reference"
)

func TestReplaceValueReferences(t *testing.T) {
	first_name := ValueReference{
		Path:               data.NewPath("foo.bar"),
		TargetPath:         data.NewPath("first_name"),
		AbsoluteTargetPath: data.NewPath("person.first_name"),
	}
	last_name := ValueReference{
		Path:               data.NewPath("foo.baz"),
		TargetPath:         data.NewPath("last_name"),
		AbsoluteTargetPath: data.NewPath("person.last_name"),
	}
	age := ValueReference{
		Path:               data.NewPath("foo.qux"),
		TargetPath:         data.NewPath("age"),
		AbsoluteTargetPath: data.NewPath("person.age"),
	}

	tests := []struct {
		name         string
		references   []ValueReference
		targetValues map[string]data.Value
		sourceValues map[string]data.Value
		errExpected  error
	}{
		{
			name:        "without any references, nothing should be done and no error returned",
			references:  nil,
			errExpected: nil,
		},
		{
			name:       "single reference replacement",
			references: []ValueReference{first_name},
			targetValues: map[string]data.Value{
				"person.first_name": data.NewValue("john"),
			},
			sourceValues: map[string]data.Value{
				"foo.bar": data.NewValue("${last_name}"),
			},
			errExpected: nil,
		},
		{
			name:       "multiple reference replacement",
			references: []ValueReference{first_name, last_name, age},
			targetValues: map[string]data.Value{
				"person.first_name": data.NewValue("john"),
				"person.last_name":  data.NewValue("doe"),
			},
			sourceValues: map[string]data.Value{
				"foo.bar": data.NewValue("${last_name}"),
				"foo.baz": data.NewValue("${first_name}"),
				"foo.qux": data.NewValue("${age}"),
			},
			errExpected: nil,
		},
		{
			name:       "multiple embedded reference replacement",
			references: []ValueReference{first_name, last_name, age},
			targetValues: map[string]data.Value{
				"person.first_name": data.NewValue("john"),
				"person.last_name":  data.NewValue("doe"),
			},
			sourceValues: map[string]data.Value{
				"foo.bar": data.NewValue("First name is: ${last_name}"),
				"foo.baz": data.NewValue("Last name is: ${first_name}"),
				"foo.qux": data.NewValue("Age is: ${age}"),
			},
			errExpected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := mocks.NewMockValueTarget(t)
			for _, ref := range tt.references {
				target.EXPECT().GetPath(ref.AbsoluteTargetPath).Return(tt.targetValues[ref.AbsoluteTargetPath.String()], nil)
				target.EXPECT().GetPath(ref.Path).Return(tt.sourceValues[ref.Path.String()], nil)
				target.EXPECT().SetPath(ref.Path, mock.Anything).Return(nil)
			}
			_, _, _ = first_name, last_name, age

			err := ReplaceValues(target, tt.references)
			if tt.errExpected != nil {
				assert.ErrorIs(t, err, tt.errExpected)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestReorderValueReferences(t *testing.T) {
	first_name := ValueReference{
		Path:               data.NewPath("foo.bar"),
		TargetPath:         data.NewPath("first_name"),
		AbsoluteTargetPath: data.NewPath("person.first_name"),
	}
	last_name := ValueReference{
		Path:               data.NewPath("foo.baz"),
		TargetPath:         data.NewPath("last_name"),
		AbsoluteTargetPath: data.NewPath("person.last_name"),
	}
	age := ValueReference{
		Path:               data.NewPath("foo.qux"),
		TargetPath:         data.NewPath("age"),
		AbsoluteTargetPath: data.NewPath("person.age"),
	}

	tests := []struct {
		name          string
		order         []ValueReference
		allReferences []ValueReference
		expected      []ValueReference
	}{
		{
			name:          "empty order must return the same unordered references",
			order:         []ValueReference{},
			allReferences: []ValueReference{first_name, last_name, age},
			expected:      []ValueReference{first_name, last_name, age},
		},
		{
			name:          "nil order must return the same unordered references",
			order:         nil,
			allReferences: []ValueReference{first_name, last_name, age},
			expected:      []ValueReference{first_name, last_name, age},
		},
		{
			name:          "empty allReferences must return nil",
			order:         []ValueReference{first_name, last_name, age},
			allReferences: []ValueReference{},
			expected:      []ValueReference(nil),
		},
		{
			name:          "nil allReferences must return nil",
			order:         []ValueReference{first_name, last_name, age},
			allReferences: nil,
			expected:      []ValueReference(nil),
		},
		{
			name:          "ordered, non-duplicate, allReferences must not be altered",
			order:         []ValueReference{first_name, last_name, age},
			allReferences: []ValueReference{first_name, last_name, age},
			expected:      []ValueReference{first_name, last_name, age},
		},
		{
			name:          "unordered, non-duplicate, allReferences must be ordered",
			order:         []ValueReference{first_name, last_name, age},
			allReferences: []ValueReference{age, first_name, last_name},
			expected:      []ValueReference{first_name, last_name, age},
		},
		{
			name:          "unordered, duplicate, allReferences must be ordered",
			order:         []ValueReference{first_name, last_name, age},
			allReferences: []ValueReference{age, first_name, age, age, last_name, first_name, last_name},
			expected:      []ValueReference{first_name, first_name, last_name, last_name, age, age, age},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := ReorderValueReferences(tt.order, tt.allReferences)
			assert.Equal(t, tt.expected, res)
		})
	}
}

func TestCalculateReplacementOrder(t *testing.T) {
	tests := []struct {
		name          string
		references    []ValueReference
		errExpected   error
		expectedOrder []ValueReference
	}{
		{
			name: "non dependent references",
			references: []ValueReference{
				{
					Path:               data.NewPath("foo.bar"),
					TargetPath:         data.NewPath("name"),
					AbsoluteTargetPath: data.NewPath("person.name"),
				},
				{
					Path:               data.NewPath("foo.baz"),
					TargetPath:         data.NewPath("age"),
					AbsoluteTargetPath: data.NewPath("person.age"),
				},
			},
			expectedOrder: []ValueReference{
				{
					Path:               data.NewPath("foo.bar"),
					TargetPath:         data.NewPath("name"),
					AbsoluteTargetPath: data.NewPath("person.name"),
				},
				{
					Path:               data.NewPath("foo.baz"),
					TargetPath:         data.NewPath("age"),
					AbsoluteTargetPath: data.NewPath("person.age"),
				},
			},
		},
		{
			name: "long dependency chain",
			references: []ValueReference{
				{
					Path:               data.NewPath("foo.bar"),
					TargetPath:         data.NewPath("bar"),
					AbsoluteTargetPath: data.NewPath("common.bar"),
				},
				{
					Path:               data.NewPath("common.bar"),
					TargetPath:         data.NewPath("name"),
					AbsoluteTargetPath: data.NewPath("common.name"),
				},
				{
					Path:               data.NewPath("common.name"),
					TargetPath:         data.NewPath("peter"),
					AbsoluteTargetPath: data.NewPath("common.peter"),
				},
				{
					Path:               data.NewPath("common.peter"),
					TargetPath:         data.NewPath("hans"),
					AbsoluteTargetPath: data.NewPath("common.hans"),
				},
			},
			expectedOrder: []ValueReference{
				{
					Path:               data.NewPath("common.peter"),
					TargetPath:         data.NewPath("hans"),
					AbsoluteTargetPath: data.NewPath("common.hans"),
				},
				{
					Path:               data.NewPath("common.name"),
					TargetPath:         data.NewPath("peter"),
					AbsoluteTargetPath: data.NewPath("common.peter"),
				},
				{
					Path:               data.NewPath("common.bar"),
					TargetPath:         data.NewPath("name"),
					AbsoluteTargetPath: data.NewPath("common.name"),
				},
				{
					Path:               data.NewPath("foo.bar"),
					TargetPath:         data.NewPath("bar"),
					AbsoluteTargetPath: data.NewPath("common.bar"),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			graph, err := ValueDependencyGraph(tt.references)
			assert.NoError(t, err)
			assert.NotNil(t, graph)

			order, err := ValueReplacementOrder(graph)

			if tt.errExpected != nil {
				assert.ErrorIs(t, err, tt.errExpected)
				assert.Nil(t, order)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOrder, order)
			}
		})
	}
}

func TestBuildDependencyGraph(t *testing.T) {
	tests := []struct {
		name        string
		references  []ValueReference
		errExpected error
	}{
		{
			name: "non dependent references",
			references: []ValueReference{
				{
					Path:               data.NewPath("foo.bar"),
					TargetPath:         data.NewPath("name"),
					AbsoluteTargetPath: data.NewPath("person.name"),
				},
				{
					Path:               data.NewPath("foo.baz"),
					TargetPath:         data.NewPath("age"),
					AbsoluteTargetPath: data.NewPath("person.age"),
				},
			},
		},
		{
			name: "self-referencing references are not allowed",
			references: []ValueReference{
				{
					Path:               data.NewPath("foo.bar"),
					TargetPath:         data.NewPath("bar"),
					AbsoluteTargetPath: data.NewPath("foo.bar"),
				},
				{
					Path:               data.NewPath("foo.baz"),
					TargetPath:         data.NewPath("baz"),
					AbsoluteTargetPath: data.NewPath("foo.baz"),
				},
			},
			errExpected: ErrSelfReferencingReference,
		},
		{
			name: "dependency cycles are not allowed",
			references: []ValueReference{
				{
					Path:               data.NewPath("foo.bar"),
					TargetPath:         data.NewPath("name"),
					AbsoluteTargetPath: data.NewPath("person.name"),
				},
				{
					Path:               data.NewPath("person.name"),
					TargetPath:         data.NewPath("some_name"),
					AbsoluteTargetPath: data.NewPath("person.some_name"),
				},
				{
					Path:               data.NewPath("person.some_name"),
					TargetPath:         data.NewPath("foo.bar"),
					AbsoluteTargetPath: data.NewPath("foo.bar"),
				},
			},
			errExpected: ErrCyclicReference,
		},
		{
			name: "multiple dependencies",
			references: []ValueReference{
				{
					Path:               data.NewPath("foo.bar"),
					TargetPath:         data.NewPath("bar"),
					AbsoluteTargetPath: data.NewPath("common.bar"),
				},
				{
					Path:               data.NewPath("common.bar"),
					TargetPath:         data.NewPath("name"),
					AbsoluteTargetPath: data.NewPath("common.name"),
				},
				{
					Path:               data.NewPath("foo.bar"),
					TargetPath:         data.NewPath("baz"),
					AbsoluteTargetPath: data.NewPath("common.baz"),
				},
				{
					Path:               data.NewPath("common.baz"),
					TargetPath:         data.NewPath("ohai"),
					AbsoluteTargetPath: data.NewPath("common.ohai"),
				},
			},
		},
		{
			name: "long dependency chain",
			references: []ValueReference{
				{
					Path:               data.NewPath("foo.bar"),
					TargetPath:         data.NewPath("bar"),
					AbsoluteTargetPath: data.NewPath("common.bar"),
				},
				{
					Path:               data.NewPath("common.bar"),
					TargetPath:         data.NewPath("name"),
					AbsoluteTargetPath: data.NewPath("common.name"),
				},
				{
					Path:               data.NewPath("common.name"),
					TargetPath:         data.NewPath("peter"),
					AbsoluteTargetPath: data.NewPath("common.peter"),
				},
				{
					Path:               data.NewPath("common.peter"),
					TargetPath:         data.NewPath("hans"),
					AbsoluteTargetPath: data.NewPath("common.hans"),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dependencyGraph, err := ValueDependencyGraph(tt.references)

			if tt.errExpected != nil {
				assert.ErrorIs(t, err, tt.errExpected)
				assert.Nil(t, dependencyGraph)
			} else {
				assert.NotNil(t, dependencyGraph)
				assert.Nil(t, err)
			}
		})
	}
}

func TestResolveDependantValueReferences(t *testing.T) {
	tests := []struct {
		name          string
		reference     ValueReference
		allReferences []ValueReference
		expected      []ValueReference
	}{
		{
			name: "no dependencies",
			reference: ValueReference{
				Path:               data.NewPath("foo.bar"),
				TargetPath:         data.NewPath("name"),
				AbsoluteTargetPath: data.NewPath("person.name"),
			},
			allReferences: []ValueReference{
				{
					Path:               data.NewPath("foo.bar"),
					TargetPath:         data.NewPath("name"),
					AbsoluteTargetPath: data.NewPath("person.name"),
				},
			},
			expected: nil,
		},
		{
			name: "empty dependency list will always return nil",
			reference: ValueReference{
				Path:               data.NewPath("foo.bar"),
				TargetPath:         data.NewPath("name"),
				AbsoluteTargetPath: data.NewPath("person.name"),
			},
			allReferences: []ValueReference{},
			expected:      nil,
		},
		{
			name: "one dependency",
			reference: ValueReference{
				Path:               data.NewPath("foo.bar"),
				TargetPath:         data.NewPath("name"),
				AbsoluteTargetPath: data.NewPath("person.name"),
			},
			allReferences: []ValueReference{
				{
					Path:               data.NewPath("foo.bar"),
					TargetPath:         data.NewPath("name"),
					AbsoluteTargetPath: data.NewPath("person.name"),
				},
				{
					Path:               data.NewPath("person.name"),
					TargetPath:         data.NewPath("another.name"),
					AbsoluteTargetPath: data.NewPath("person..another.name"),
				},
			},
			expected: []ValueReference{
				{
					Path:               data.NewPath("person.name"),
					TargetPath:         data.NewPath("another.name"),
					AbsoluteTargetPath: data.NewPath("person..another.name"),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValueDependencies(tt.reference, tt.allReferences)

			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.ElementsMatch(t, tt.expected, result)
			}
		})
	}
}

func TestFindValueReferences(t *testing.T) {
	type sourceData struct {
		// maps paths to values of said paths
		values map[string]data.Value
		// maps the relative path to the expected absolute path
		absPaths map[string]string
	}

	tests := []struct {
		name     string
		input    sourceData
		expected []ValueReference
	}{
		{
			name: "simple references",
			input: sourceData{
				values: map[string]data.Value{
					"foo":        data.NewValue("${bar}"),
					"bar":        data.NewValue("bar"),
					"name":       data.NewValue("${peter:name}"),
					"peter.name": data.NewValue("Bob, lol"),
				},
				absPaths: map[string]string{
					"bar":        "bar",
					"peter.name": "persons.peter.name",
				},
			},
			expected: []ValueReference{
				{
					Path:               data.NewPath("foo"),
					TargetPath:         data.NewPath("bar"),
					AbsoluteTargetPath: data.NewPath("bar"),
				},
				{
					Path:               data.NewPath("name"),
					TargetPath:         data.NewPath("peter.name"),
					AbsoluteTargetPath: data.NewPath("persons.peter.name"),
				},
			},
		},
		{
			name: "multiple references in value",
			input: sourceData{
				values: map[string]data.Value{
					"foo": data.NewValue("${bar} ${baz} ${qux}"),
					"bar": data.NewValue("bar"),
					"baz": data.NewValue("baz"),
					"qux": data.NewValue("qux"),
				},
				absPaths: map[string]string{
					"bar": "foo.bar",
					"baz": "foo.baz",
					"qux": "foo.qux",
				},
			},

			expected: []ValueReference{
				{
					Path:               data.NewPath("foo"),
					TargetPath:         data.NewPath("bar"),
					AbsoluteTargetPath: data.NewPath("foo.bar"),
				},
				{
					Path:               data.NewPath("foo"),
					TargetPath:         data.NewPath("baz"),
					AbsoluteTargetPath: data.NewPath("foo.baz"),
				},
				{
					Path:               data.NewPath("foo"),
					TargetPath:         data.NewPath("qux"),
					AbsoluteTargetPath: data.NewPath("foo.qux"),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := mocks.NewMockValueSource(t)
			source.EXPECT().Values().Return(tt.input.values)

			for rel, abs := range tt.input.absPaths {
				source.EXPECT().AbsolutePath(data.NewPath(rel), mock.AnythingOfType("data.Path")).Return(data.NewPath(abs), nil)
			}

			references, err := FindAllValueReferences(source)
			assert.NoError(t, err)
			assert.ElementsMatch(t, tt.expected, references)
		})
	}
}

func TestFindReferenceTargetPaths(t *testing.T) {
	tests := []struct {
		name          string
		value         data.Value
		expected      []data.Path
		panicExpected bool
		regex         *regexp.Regexp
	}{
		{
			name:     "nil value must not return any references",
			value:    data.NilValue,
			regex:    ValueReferenceRegex,
			expected: nil,
		},
		{
			name:     "empty value must not return any references",
			value:    data.NewValue(""),
			regex:    ValueReferenceRegex,
			expected: nil,
		},
		{
			name:          "empty regex must result in panic",
			value:         data.NewValue("${foo}"),
			regex:         regexp.MustCompile(``),
			expected:      nil,
			panicExpected: true,
		},
		{
			name:  "single reference in value",
			value: data.NewValue("${foo:bar}"),
			regex: ValueReferenceRegex,
			expected: []data.Path{
				data.NewPath("foo.bar"),
			},
		},
		{
			name:  "multiple references in value",
			value: data.NewValue("${foo} and ${foo:bar}, well but ${bar:baz:another:deep:path}"),
			regex: ValueReferenceRegex,
			expected: []data.Path{
				data.NewPath("foo"),
				data.NewPath("foo.bar"),
				data.NewPath("bar.baz.another.deep.path"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.panicExpected {
				assert.Panics(t, func() {
					FindReferenceTargetPaths(tt.regex, tt.value)
				})
				return
			}

			targetPaths := FindReferenceTargetPaths(tt.regex, tt.value)
			assert.ElementsMatch(t, tt.expected, targetPaths)
		})
	}
}
