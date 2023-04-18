package skipper

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/dominikbraun/graph"
	"github.com/dominikbraun/graph/draw"
)

var (
	ErrClassAlreadyRegistered  error = errors.New("class already registered")
	ErrDependencyAlreadyExists error = errors.New("dependency already exists")

	defaultVertexAttributes []func(*graph.VertexProperties) = []func(*graph.VertexProperties){
		graph.VertexAttribute("colorscheme", "blues3"),
		graph.VertexAttribute("style", "filled"),
		graph.VertexAttribute("color", "2"),
		graph.VertexAttribute("fillcolor", "1"),
	}
)

type ClassResolver struct {
	Graph graph.Graph[string, string]
}

func NewClassResolver() *ClassResolver {
	resolver := &ClassResolver{
		Graph: graph.New(graph.StringHash, graph.Directed(), graph.Acyclic(), graph.Rooted()),
	}

	return resolver
}

func (r *ClassResolver) RegisterClass(class *Class) error {
	err := r.Graph.AddVertex(class.Namespace.String(), defaultVertexAttributes...)
	if err != nil {
		if errors.Is(err, graph.ErrVertexAlreadyExists) {
			return fmt.Errorf("%w: %s", ErrClassAlreadyRegistered, class.Namespace)
		}
		return fmt.Errorf("resolve error: %w", err)
	}

	log.Println("registered class:", class.Namespace)

	return nil
}

func (r *ClassResolver) DependsOn(parent, child Path) error {
	err := r.Graph.AddEdge(parent.String(), child.String())
	if err != nil {
		if errors.Is(err, graph.ErrEdgeAlreadyExists) {
			return fmt.Errorf("%w: %s --> %s", ErrDependencyAlreadyExists, parent, child)
		}
		return fmt.Errorf("resolve error: %w", err)
	}

	log.Printf("added dependency %s --> %s", parent, child)

	return nil
}

func (r *ClassResolver) ReduceAndSort() ([]Path, error) {
	reduced, err := graph.TransitiveReduction(r.Graph)
	if err != nil {
		return nil, err
	}

	order, err := graph.TopologicalSort(reduced)
	if err != nil {
		return nil, err
	}

	// convert order to [Path] while iterating in reverse
	// we want the path with no dependency to be the first in the slice
	// because thats the order we
	var orderList []Path
	for i := len(order) - 1; i >= 0; i-- {
		orderList = append(orderList, P(order[i]))
	}

	return orderList, nil
}

// AllPaths uses recursive DFS to find all possible paths between source and destination.
func (r *ClassResolver) AllPaths(source, destination string) [][]string {
	reduced, err := graph.TransitiveReduction(r.Graph)
	if err != nil {
		return nil
	}

	var paths [][]string

	visited := make(map[string]bool) // all visited noded to avoid cycles
	path := []string{source}         // current path being exploed

	// recursive dfs function to find all paths
	var dfs func(node string)
	dfs = func(node string) {
		if node == destination {
			// found a path to B, so copy the current path and add it to the list of paths
			newPath := make([]string, len(path))
			copy(newPath, path)
			paths = append(paths, newPath)
			return
		}

		visited[node] = true

		// explore all outgoing edges from this node
		adjMap, _ := reduced.AdjacencyMap()
		adjacent := adjMap[node] // get the next nodes from this node
		for _, edge := range adjacent {
			next := edge.Target
			if !visited[next] {
				// if the next node hasn't been visited yet, add it to the current path and explore it recursively
				path = append(path, next)
				dfs(next)
				path = path[:len(path)-1] // remove the last node from the path once we've finished exploring it
			}
		}

		visited[node] = false // unmark this node as visited to allow exploring it from other paths
	}

	dfs(source)

	return paths
}

func VisualizeGraph(graph graph.Graph[string, string], filePath string, label string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	err = draw.DOT(graph, file,
		draw.GraphAttribute("label", label),
		draw.GraphAttribute("overlap", "false"),
		draw.GraphAttribute("minlen", "5"))
	if err != nil {
		return err
	}
	return nil
}
