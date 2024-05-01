package expression_test

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"

	"github.com/lukasjarosch/skipper/v1/expression"
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
			expressions := expression.Parse(tt.input)
			assert.ElementsMatch(t, tt.expressions, expressions)

			// TODO: write own ElementsMatch which only compares values and not positions
		})
	}
}

func TestParsePath(t *testing.T) {
	tests := []struct {
		name        string
		segments    []string
		variables   map[string]any
		errExpected error
	}{
		{
			name:        "path without variables",
			segments:    []string{"foo", "bar"},
			variables:   make(map[string]any),
			errExpected: nil,
		},
		{
			name:     "path with string variable",
			segments: []string{"foo", "bar", "$baz"},
			variables: map[string]any{
				"baz": "baz",
			},
			errExpected: nil,
		},
		{
			name:     "path with float variable",
			segments: []string{"foo", "bar", "$baz"},
			variables: map[string]any{
				"baz": 123.123,
			},
			errExpected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := expression.ParsePath(tt.segments, tt.variables)

			if tt.errExpected != nil {
				assert.ErrorIs(t, err, tt.errExpected)
				return
			}

			spew.Dump(path.Text())

			assert.NoError(t, err)
			assert.Equal(t, len(tt.segments), len(path.Segments))
		})
	}
}
