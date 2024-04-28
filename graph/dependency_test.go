package graph

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/dominikbraun/graph"
	"github.com/stretchr/testify/assert"
)

func TestDependencyGraph_AddVertex(t *testing.T) {
	tests := []struct {
		name         string
		vertexHashes []string
		errExpected  error
	}{
		{
			name:         "Cannot add empty vertex hash",
			vertexHashes: []string{""},
			errExpected:  ErrEmptyVertexHash,
		},
		{
			name:         "Can add duplicate hashes without error",
			vertexHashes: []string{"foo", "foo"},
		},
		{
			name:         "Can add arbitrary hashes",
			vertexHashes: []string{"foo", "bar", "baz", "qux"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewDependencyGraph()

			var err error

			for _, hash := range tt.vertexHashes {
				err = g.AddVertex(hash)
			}

			if tt.errExpected != nil {
				assert.ErrorIs(t, err, tt.errExpected)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestDependencyGraph_RegisterDependencies(t *testing.T) {
	tests := []struct {
		name           string
		errExpected    error
		dependerHash   string
		dependeeHashes []string
	}{
		{
			name:           "Simple dependency",
			dependerHash:   "foo",
			dependeeHashes: []string{"bar"},
		},
		{
			name:           "Many dependencies",
			dependerHash:   "foo",
			dependeeHashes: []string{"bar", "baz", "qux", "bob"},
		},
		{
			name:           "Self references are not allowed",
			dependerHash:   "foo",
			dependeeHashes: []string{"foo"},
			errExpected:    ErrSelfReferencingDependency,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewDependencyGraph()

			// register all vertecies
			err := g.AddVertex(tt.dependerHash)
			assert.NoError(t, err)
			for _, hash := range tt.dependeeHashes {
				err = g.AddVertex(hash)
				assert.NoError(t, err)
			}

			err = g.RegisterDependencies(tt.dependerHash, tt.dependeeHashes)

			if tt.errExpected != nil {
				assert.ErrorIs(t, err, tt.errExpected)
				return
			}
			assert.NoError(t, err)

			for _, depHash := range tt.dependeeHashes {
				assert.True(t, g.HasEdge(tt.dependerHash, depHash))
			}
		})
	}
}

func TestDependencyGraph_RemoveVertex(t *testing.T) {
	tests := []struct {
		name              string
		vertexToDelete    string
		incomingEdgesFrom []string
		outgoingEdgesTo   []string
		errExpected       error
	}{
		{
			name:           "Vertex without edges",
			vertexToDelete: "foo",
		},
		{
			name:              "Vertex with only incoming edges",
			vertexToDelete:    "foo",
			incomingEdgesFrom: []string{"bar", "baz", "qux"},
		},
		{
			name:              "Vertex with incoming and outgoing edges",
			vertexToDelete:    "foo",
			incomingEdgesFrom: []string{"bar", "baz", "qux"},
			outgoingEdgesTo:   []string{"john", "doe", "peter"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewDependencyGraph()

			// Register all vertecies
			err := g.AddVertex(tt.vertexToDelete)
			assert.NoError(t, err)
			for _, vertex := range tt.incomingEdgesFrom {
				err := g.AddVertex(vertex)
				assert.NoError(t, err)
			}
			for _, vertex := range tt.outgoingEdgesTo {
				err = g.AddVertex(vertex)
				assert.NoError(t, err)
			}

			// Create edges
			for _, sourceVertex := range tt.incomingEdgesFrom {
				err = g.RegisterDependencies(sourceVertex, []string{tt.vertexToDelete})
				assert.NoError(t, err)
			}
			err = g.RegisterDependencies(tt.vertexToDelete, tt.outgoingEdgesTo)
			assert.NoError(t, err)

			// remove the vertex
			// this should remove all edges to and from it as well as the vertex
			err = g.RemoveVertex(tt.vertexToDelete)
			if tt.errExpected != nil {
				assert.ErrorIs(t, err, tt.errExpected)
				spew.Dump("ERR")
				return
			}

			// assert that vertex is gone
			vertex, err := g.graph.Vertex(tt.vertexToDelete)
			assert.Empty(t, vertex)
			assert.ErrorIs(t, err, graph.ErrVertexNotFound)

			// assert that all edges are gone as well
			edges, err := g.graph.Edges()
			assert.NoError(t, err)
			assert.Len(t, edges, 0)
		})
	}
}

func TestDependencyGraph_Subgraph(t *testing.T) {
	tests := []struct {
		name        string
		errExpected error
		edges       map[string][]string
	}{
		{
			name: "TestDependencyGraph_Subgraph",
			edges: map[string][]string{
				"a": {"b", "c"},
				"b": {"c", "d"},
				"c": {},
				"d": {"e"},
				"e": {},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewDependencyGraph()

			// register all vertecies and edges
			for sourceVertex, targetVertecies := range tt.edges {
				err := g.AddVertex(sourceVertex)
				assert.NoError(t, err)

				for _, targetVertex := range targetVertecies {
					err = g.AddVertex(targetVertex)
					assert.NoError(t, err)
				}
				err = g.RegisterDependencies(sourceVertex, targetVertecies)
				assert.NoError(t, err)

			}

			subgraph, err := g.Subgraph("d")

			if tt.errExpected != nil {
				assert.ErrorIs(t, err, tt.errExpected)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, subgraph)
			// TODO: test 'ret'
		})
	}
}
