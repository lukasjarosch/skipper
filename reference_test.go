package skipper_test

import (
	"regexp"
	"testing"

	"github.com/dominikbraun/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	. "github.com/lukasjarosch/skipper"
	"github.com/lukasjarosch/skipper/data"
	mocks "github.com/lukasjarosch/skipper/mocks"
)

func TestNewValueReferenceManager(t *testing.T) {
	// TEST: source is nil
	manager, err := NewValueReferenceManager(nil)
	assert.ErrorIs(t, err, ErrReferenceSourceIsNil)
	assert.Nil(t, manager)
}

func TestCalculateReplacementOrder(t *testing.T) {
	tests := []struct {
		name          string
		references    map[string]ValueReference
		errExpected   error
		expectedOrder []ValueReference
	}{
		{
			name: "non dependent references",
			references: map[string]ValueReference{
				"ref1": {
					Path:               data.NewPath("foo.bar"),
					TargetPath:         data.NewPath("name"),
					AbsoluteTargetPath: data.NewPath("person.name"),
				},
				"ref2": {
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
			references: map[string]ValueReference{
				"ref1": {
					Path:               data.NewPath("foo.bar"),
					TargetPath:         data.NewPath("bar"),
					AbsoluteTargetPath: data.NewPath("common.bar"),
				},
				"ref2": {
					Path:               data.NewPath("common.bar"),
					TargetPath:         data.NewPath("name"),
					AbsoluteTargetPath: data.NewPath("common.name"),
				},
				"ref3": {
					Path:               data.NewPath("common.name"),
					TargetPath:         data.NewPath("peter"),
					AbsoluteTargetPath: data.NewPath("common.peter"),
				},
				"ref4": {
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
			graph, err := BuildDependencyGraph(tt.references)
			assert.NoError(t, err)
			assert.NotNil(t, graph)

			order, err := CalculateReplacementOrder(graph)

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
		references  map[string]ValueReference
		errExpected error
	}{
		{
			name: "non dependent references",
			references: map[string]ValueReference{
				"ref1": {
					Path:               data.NewPath("foo.bar"),
					TargetPath:         data.NewPath("name"),
					AbsoluteTargetPath: data.NewPath("person.name"),
				},
				"ref2": {
					Path:               data.NewPath("foo.baz"),
					TargetPath:         data.NewPath("age"),
					AbsoluteTargetPath: data.NewPath("person.age"),
				},
			},
		},
		{
			name: "references must be deduplicated",
			references: map[string]ValueReference{
				"ref1": {
					Path:               data.NewPath("foo.bar"),
					TargetPath:         data.NewPath("name"),
					AbsoluteTargetPath: data.NewPath("person.name"),
				},
				"ref2": {
					Path:               data.NewPath("foo.bar"),
					TargetPath:         data.NewPath("name"),
					AbsoluteTargetPath: data.NewPath("person.name"),
				},
			},
			errExpected: graph.ErrVertexAlreadyExists,
		},
		{
			name: "self-referencing references are not allowed",
			references: map[string]ValueReference{
				"ref1": {
					Path:               data.NewPath("foo.bar"),
					TargetPath:         data.NewPath("bar"),
					AbsoluteTargetPath: data.NewPath("foo.bar"),
				},
				"ref2": {
					Path:               data.NewPath("foo.baz"),
					TargetPath:         data.NewPath("baz"),
					AbsoluteTargetPath: data.NewPath("foo.baz"),
				},
			},
			errExpected: ErrSelfReferencingReference,
		},
		{
			name: "dependency cycles are not allowed",
			references: map[string]ValueReference{
				"ref1": {
					Path:               data.NewPath("foo.bar"),
					TargetPath:         data.NewPath("name"),
					AbsoluteTargetPath: data.NewPath("person.name"),
				},
				"ref2": {
					Path:               data.NewPath("person.name"),
					TargetPath:         data.NewPath("some_name"),
					AbsoluteTargetPath: data.NewPath("person.some_name"),
				},
				"ref3": {
					Path:               data.NewPath("person.some_name"),
					TargetPath:         data.NewPath("foo.bar"),
					AbsoluteTargetPath: data.NewPath("foo.bar"),
				},
			},
			errExpected: ErrCyclicReference,
		},
		{
			name: "multiple dependencies",
			references: map[string]ValueReference{
				"ref1": {
					Path:               data.NewPath("foo.bar"),
					TargetPath:         data.NewPath("bar"),
					AbsoluteTargetPath: data.NewPath("common.bar"),
				},
				"ref2": {
					Path:               data.NewPath("common.bar"),
					TargetPath:         data.NewPath("name"),
					AbsoluteTargetPath: data.NewPath("common.name"),
				},
				"ref3": {
					Path:               data.NewPath("foo.bar"),
					TargetPath:         data.NewPath("baz"),
					AbsoluteTargetPath: data.NewPath("common.baz"),
				},
				"ref4": {
					Path:               data.NewPath("common.baz"),
					TargetPath:         data.NewPath("ohai"),
					AbsoluteTargetPath: data.NewPath("common.ohai"),
				},
			},
		},
		{
			name: "long dependency chain",
			references: map[string]ValueReference{
				"ref1": {
					Path:               data.NewPath("foo.bar"),
					TargetPath:         data.NewPath("bar"),
					AbsoluteTargetPath: data.NewPath("common.bar"),
				},
				"ref2": {
					Path:               data.NewPath("common.bar"),
					TargetPath:         data.NewPath("name"),
					AbsoluteTargetPath: data.NewPath("common.name"),
				},
				"ref3": {
					Path:               data.NewPath("common.name"),
					TargetPath:         data.NewPath("peter"),
					AbsoluteTargetPath: data.NewPath("common.peter"),
				},
				"ref4": {
					Path:               data.NewPath("common.peter"),
					TargetPath:         data.NewPath("hans"),
					AbsoluteTargetPath: data.NewPath("common.hans"),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dependencyGraph, err := BuildDependencyGraph(tt.references)

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
			// allReferences needs to be a map
			allReferenceMap := make(map[string]ValueReference, len(tt.allReferences))
			for _, ref := range tt.allReferences {
				allReferenceMap[ref.Hash()] = ref
			}

			result := ResolveDependantValueReferences(tt.reference, allReferenceMap)

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
			source := mocks.NewMockReferenceValueSource(t)
			source.EXPECT().Values().Return(tt.input.values)

			for rel, abs := range tt.input.absPaths {
				source.EXPECT().AbsolutePath(data.NewPath(rel), mock.AnythingOfType("data.Path")).Return(data.NewPath(abs), nil)
			}

			references, err := FindValueReferences(source)
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
