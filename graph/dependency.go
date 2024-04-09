package graph

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dominikbraun/graph"
)

type dependencyGraph graph.Graph[string, string]

var (
	ErrSelfReferencingDependency = fmt.Errorf("self-referencing dependency")
	ErrCyclicReference           = fmt.Errorf("cyclic dependency")
)

type DependencyGraph struct {
	graph dependencyGraph
}

func NewDependencyGraph() *DependencyGraph {
	g := graph.New(graph.StringHash, graph.Acyclic(), graph.Directed(), graph.PreventCycles())
	return &DependencyGraph{graph: g}
}

func (dg *DependencyGraph) AddVertex(vertexHash string) error {
	err := dg.graph.AddVertex(vertexHash)
	if err != nil {
		// ignore duplicate vertex errors
		if !errors.Is(err, graph.ErrVertexAlreadyExists) {
			return fmt.Errorf("cannot add vertex: %w", err)
		}
	}

	return nil
}

func (dg *DependencyGraph) RegisterDependencies(dependerVertexHash string, dependeeVertexHashes []string) error {
	for _, dependency := range dependeeVertexHashes {
		dependeeVertex, err := dg.graph.Vertex(dependency)
		if err != nil {
			return fmt.Errorf("unexpectedly could not fetch dependency vertex %s: %w", dependency, err)
		}

		// prevent self-referencing references
		if strings.EqualFold(dependerVertexHash, dependeeVertex) {
			return fmt.Errorf("%s: %w", dependerVertexHash, ErrSelfReferencingDependency)
		}

		err = dg.graph.AddEdge(dependerVertexHash, dependeeVertex)
		if err != nil {
			if errors.Is(err, graph.ErrEdgeCreatesCycle) {
				return fmt.Errorf("%s -> %s: %w", dependerVertexHash, dependeeVertex, ErrCyclicReference)
			}
			return fmt.Errorf("failed to register dependency %s: %w", dependency, err)
		}
	}

	return nil
}

func (dg *DependencyGraph) RemoveVertex(vertexHash string) error {
	edges, err := dg.graph.Edges()
	if err != nil {
		return err
	}

	// Find all edges with either originate from or point to the reference and remove them.
	for _, edge := range edges {
		if edge.Source == vertexHash {
			err = dg.graph.RemoveEdge(edge.Source, edge.Target)
			if err != nil {
				return err
			}
			continue
		}
		if edge.Target == vertexHash {
			err = dg.graph.RemoveEdge(edge.Source, edge.Target)
			if err != nil {
				return err
			}
			continue
		}
	}

	return dg.graph.RemoveVertex(vertexHash)
}

func (dg *DependencyGraph) TopologicalSort() (vertexHashes []string, err error) {
	// Perform a stable topological sort of the dependency graph.
	// The returned orderedHashes is stable in that the hashes are sorted
	// by their length or alphabetically if they are the same length.
	// This eliminates the issue that the actual topological sorting algorithm usually
	// has multiple valid solutions.
	orderedHashes, err := graph.StableTopologicalSort[string, string](dg.graph, func(s1, s2 string) bool {
		// Strings are of different length, sort by length
		if len(s1) != len(s2) {
			return len(s1) < len(s2)
		}
		// Otherwise, sort alphabetically
		return s1 > s2
	})
	if err != nil {
		return nil, fmt.Errorf("failed to sort reference graph: %w", err)
	}

	// The result of the topological sorting is reverse to what we want.
	// We expect the vertex without any dependencies to be at the top.
	for i := len(orderedHashes) - 1; i >= 0; i-- {
		ref, err := dg.graph.Vertex(orderedHashes[i])
		if err != nil {
			return nil, err
		}
		vertexHashes = append(vertexHashes, ref)
	}

	return
}
