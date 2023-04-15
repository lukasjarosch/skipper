package skipper

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/dominikbraun/graph"
)

var (
	ErrEmptyPath = errors.New("empty path")
)

// PathSeparator is the separator used for string representations of [Path].
const PathSeparator = "."

// Path is the central way of indexing data in Skipper.
// Paths are used to traverse [Data] and to create namespaces in [Class]es
type Path []string

// P is a helper function to quickly create a [Path] from a string
//
// It takes a string representing a Path and returns a normalized Path.
// A normalized path is a Path without leading or trailing PathSeparator characters,
// and with empty segments removed.
// You can still use
//
//	skipper.Path{"foo", "bar"}
//
// but using
//
//	skipper.P("foo.bar")
//
// is usually way more convenient.
func P(path string) Path {
	path = strings.TrimSpace(path)
	path = strings.Trim(path, PathSeparator)
	pathSegments := strings.Split(path, PathSeparator)

	// remove empty segments
	var p Path
	for _, segment := range pathSegments {
		if len(segment) == 0 {
			continue
		}
		p = append(p, segment)
	}

	// if now no more segments are left, return a completely empty path
	if len(p) == 0 {
		return Path{}
	}

	return p
}

// String returns a string representation of a Path.
//
// In case the Path was constructed using
//
//	Path{"foo"}
//
// we need to normalize the string
// as it is not guaranteed that the Path looks like what we want it to.
func (p Path) String() string {
	pathString := strings.Join(p, PathSeparator)
	path := P(pathString)

	return strings.Join(path, PathSeparator)
}

var (
	ErrPathAlreadyRegistered   error = errors.New("path already registered")
	ErrDependencyAlreadyExists error = errors.New("dependency already exists")
)

type Resolver struct {
	graph graph.Graph[string, string]
}

func NewResolver() *Resolver {
	resolver := &Resolver{
		graph: graph.New(graph.StringHash, graph.Directed(), graph.Acyclic()),
	}

	return resolver
}

func (r *Resolver) RegisterPath(path Path) error {
	err := r.graph.AddVertex(path.String())
	if err != nil {
		if errors.Is(err, graph.ErrVertexAlreadyExists) {
			return fmt.Errorf("%w: %s", ErrPathAlreadyRegistered, path)
		}
		return fmt.Errorf("resolve error: %w", err)
	}

	log.Println("registered path:", path)

	return nil
}

func (r *Resolver) DependsOn(parent, child Path) error {
	err := r.graph.AddEdge(parent.String(), child.String())
	if err != nil {
		if errors.Is(err, graph.ErrEdgeAlreadyExists) {
			return fmt.Errorf("%w: %s --> %s", ErrDependencyAlreadyExists, parent, child)
		}
		return fmt.Errorf("resolve error: %w", err)
	}

	log.Printf("added dependency %s --> %s", parent, child)

	return nil
}

func (r *Resolver) TopologicalSort() ([]Path, error) {
	order, err := graph.TopologicalSort(r.graph)
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
