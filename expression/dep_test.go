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
	pathMap := expression.PathMap{
		"some.path":       expression.Parse(`${some:$target:path}`),
		"some.other.path": expression.Parse(`${yet:another:$variable_path}`),
	}
	variableMap := map[string]any{
		"target": "other",
	}
	expected := []*expression.ExpressionNode{}
	expected = append(expected, pathMap["some.other.path"]...)

	deps, err := expression.Dependencies(pathMap["some.path"][0], pathMap, variableMap)

	assert.NoError(t, err)
	assert.ElementsMatch(t, expected, deps)
}
