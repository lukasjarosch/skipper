package expression_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lukasjarosch/skipper/expression"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expressions []*expression.ExpressionNode
	}{
		{
			name:  "single path expr with only identifiers",
			input: `${foo:bar:baz}`,
			expressions: []*expression.ExpressionNode{
				{
					Child: &expression.PathNode{
						Segments: []expression.Node{
							&expression.IdentifierNode{
								Value: "foo",
							},
							&expression.IdentifierNode{
								Value: "bar",
							},
							&expression.IdentifierNode{
								Value: "baz",
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expressions, err := expression.Parse(tt.input)
			assert.NoError(t, err)
			assert.ElementsMatch(t, tt.expressions, expressions)

			// TODO: write own ElementsMatch which only compares values and not positions
		})
	}
}
