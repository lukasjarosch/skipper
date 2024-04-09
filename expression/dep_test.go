package expression_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lukasjarosch/skipper/expression"
)

func TestInitializeDependencyGraph(t *testing.T) {
	input := `${some:path} and ${another:path} and again ${some:path}`
	expressions := expression.Parse(input)

	err := expression.InitializeDependencyGraph(expressions)
	assert.NoError(t, err)
}

func TestDependencies(t *testing.T) {
	tests := []struct {
		name        string
		errExpected error
		subject     *expression.ExpressionNode
		pathMap     expression.PathMap
		variableMap expression.VariableMap
		expected    []string
	}{
		{
			name:    "One dependency",
			subject: expression.Parse("${foo:bar}")[0], // <- note that
			pathMap: expression.PathMap{
				"foo.bar": expression.Parse("${bar:baz}"),
				"bar.baz": expression.Parse("hello"), // not a dependency, scalar value
			},
			expected: []string{"foo.bar"},
		},
		{
			name:    "Multiple dependencies",
			subject: expression.Parse("${default(foo:bar, bar:baz) || set_env(foo:qux)}")[0],
			pathMap: expression.PathMap{
				"foo.bar":    expression.Parse("${bar:baz}"),
				"bar.baz":    expression.Parse("${foo:qux}"),
				"foo.qux":    expression.Parse("${ohai:there}"),
				"ohai.there": expression.Parse("ohai"),
			},
			expected: []string{"foo.bar", "bar.baz", "foo.qux"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ret, err := expression.Dependencies(tt.subject, tt.pathMap, tt.variableMap)

			if tt.errExpected != nil {
				assert.ErrorIs(t, err, tt.errExpected)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, ret, tt.expected)

			return
		})
	}
}
