package skipper

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/lukasjarosch/skipper/data"
	"github.com/lukasjarosch/skipper/expression"
	"github.com/lukasjarosch/skipper/graph"
)

// expressionRegex is used to find the offsets within the context in which an expression occurred
var expressionRegex = regexp.MustCompile(`\$\{[^$.]+\}`)

type ExpressionRegistry map[string][]*expression.ExpressionNode

func (reg ExpressionRegistry) Expressions(path string) []*expression.ExpressionNode {
	ret, ok := reg[path]
	if !ok {
		return nil
	}
	return ret
}

// DependentPaths returns a list of paths on which the given expression is dependent on.
func DependentPaths(expr *expression.ExpressionNode, reg ExpressionRegistry, varSource VariableSource) ([]string, error) {
	pathNodes := expression.PathsInExpression(expr)
	resolvedPathNodes := make([]*expression.PathNode, len(pathNodes))

	// Resolve variables within the pathNodes, leaving us with PathNodes with only IdentifierNodes as Segments.
	// Only those paths can possibly be valid skipper paths.
	for i, pathNode := range pathNodes {
		res, err := expression.ResolveVariablePath(*pathNode, map[string]any(varSource.GetAll()))
		if err != nil {
			return nil, err
		}
		resolvedPathNodes[i] = res
	}

	// If the skipper path occurs as key in the registry,
	// then all those expressions are direct dependencies on the current expression.
	var dependingOnPaths []string
	for _, pathNode := range resolvedPathNodes {
		pathMapKey := pathNode.SkipperPath().String()
		if _, ok := reg[pathMapKey]; ok {
			dependingOnPaths = append(dependingOnPaths, pathMapKey)
		}
	}

	return dependingOnPaths, nil
}

type VariableSource interface {
	GetAll() map[string]any
	GetValue(string) (data.Value, error)
}

// PathValueSource is anything (class, registry, inventory) which provides values, given paths.
type PathValueSource interface {
	Values() map[string]data.Value
	GetPath(data.Path) (data.Value, error)
	SetPath(data.Path, interface{}) error
}

var (
	ErrNilPathValueSource = fmt.Errorf("nil PathValueSource")
	ErrNilVariableSource  = fmt.Errorf("nil VariableSource")
)

type ExpressionManager struct {
	registry     ExpressionRegistry
	dependencies *graph.DependencyGraph
	variables    VariableSource
	source       PathValueSource
}

func NewExpressionManager(source PathValueSource, varSource VariableSource) (*ExpressionManager, error) {
	if source == nil {
		return nil, ErrNilPathValueSource
	}
	if varSource == nil {
		return nil, ErrNilVariableSource
	}

	manager := &ExpressionManager{
		source:       source,
		variables:    varSource,
		dependencies: graph.NewDependencyGraph(),
		registry:     make(ExpressionRegistry),
	}

	// populate registry with all paths and expressions
	for path, val := range manager.source.Values() {
		manager.registry[path] = append(manager.registry[path], expression.Parse(val.String())...)
	}

	if err := manager.resetDependencyGraph(manager.dependencies); err != nil {
		return nil, err
	}

	// TODO: register hooks

	return manager, nil
}

func (m *ExpressionManager) executeExpression(expr *expression.ExpressionNode) (data.Value, error) {
	// TODO: check if the expression is already part of the registry
	// if so, it is already a vertex within the depGraph and all dependencies are resolved
	// otherwise, add the expression into the registry temporarily and register its dependencies

	// TODO: build subgraph from expression and all its dependencies

	// TODO: topological sort the subgraph

	// TODO: execute all dependencies and the target expression in order

	return data.NilValue, nil
}

type pathValueProvider map[string]interface{}

func (pvp pathValueProvider) GetPath(path data.Path) (interface{}, error) {
	res, ok := pvp[path.String()]
	if !ok {
		return nil, fmt.Errorf("path does not exist: %s", path)
	}
	return res, nil
}

func (pvp pathValueProvider) GetPathValue(path data.Path) (data.Value, error) {
	res, ok := pvp[path.String()]
	if !ok {
		return data.NilValue, fmt.Errorf("path does not exist: %s", path)
	}
	return data.NewValue(res), nil
}

func (pvp pathValueProvider) Add(path data.Path, value interface{}) {
	pvp[path.String()] = value
}

// ExecuteInput attempts to execute the expression present within input.
// It uses all known variables and source paths as execution context.
func (m *ExpressionManager) ExecuteInput(input string) (data.Value, error) {
	inputExpressions := expression.Parse(input)
	if len(inputExpressions) == 0 {
		return data.NilValue, nil
	}

	depGraph := graph.NewDependencyGraph()
	m.resetDependencyGraph(depGraph)

	if len(inputExpressions) > 1 {
		return data.NilValue, fmt.Errorf("cannot execute more than one expression")
	}

	expr := inputExpressions[0]
	dependentPaths, err := DependentPaths(expr, m.registry, m.variables)
	if err != nil {
		return data.NilValue, err
	}

	// because this expression is not part of the registry, we need
	// to create a temporary vertex
	tmpVertexHash := "temporary-vertex-hash"
	err = depGraph.AddVertex(tmpVertexHash)
	if err != nil {
		return data.NilValue, err
	}

	err = depGraph.RegisterDependencies(tmpVertexHash, dependentPaths)
	if err != nil {
		return data.NilValue, fmt.Errorf("unable to register dependency: %w", err)
	}

	subGraph, err := depGraph.Subgraph(tmpVertexHash)
	if err != nil {
		return data.NilValue, err
	}

	vertexOrder, err := subGraph.TopologicalSort()
	if err != nil {
		return data.NilValue, err
	}

	valueProvider := make(pathValueProvider)

	for _, pathVertex := range vertexOrder {
		expressions := m.registry.Expressions(pathVertex)

		// in case the pathVertex is our temporary vertex it will not be part of the registry
		// thus we need to provide the expressions ourselves
		if pathVertex == tmpVertexHash {
			expressions = inputExpressions
		}

		// add source value
		pathValue, err := m.source.GetPath(data.NewPath(pathVertex))
		if err != nil {
			return data.NilValue, err
		}
		valueProvider.Add(data.NewPath(pathVertex), pathValue)

		// skip paths without expressions
		if len(expressions) == 0 {
			continue
		}

		// execute all expressions at the current path and store their results
		// the order of pathResults matters and must match the 'expressions' order
		pathResults := make([]interface{}, len(expressions))
		for i, expr := range expressions {
			result, err := expression.Execute(expr, valueProvider, m.variables.GetAll(), nil)
			if err != nil {
				return data.NilValue, fmt.Errorf("failed to execute expression at path '%s': %w", pathVertex, err)
			}
			pathResults[i] = result
		}

		for _, result := range pathResults {
			sourceValue, err := valueProvider.GetPathValue(data.NewPath(pathVertex))
			if err != nil {
				return data.NilValue, err
			}

			// fetch the start and end offset of this expression
			// and replace the old value with the result of the expression evaluation
			exprOffsets := expressionRegex.FindStringSubmatchIndex(sourceValue.String())
			if len(exprOffsets) == 0 {
				continue
			}
			oldValue := sourceValue.String()[exprOffsets[0]:exprOffsets[1]]
			newValue := strings.Replace(sourceValue.String(), oldValue, data.NewValue(result).String(), 1)

			// update the valueProvider with the new value
			valueProvider[pathVertex] = newValue
		}
	}

	return valueProvider.GetPathValue(data.NewPath(tmpVertexHash))
}

// ExecuteRegistry will execute all known expressions in the order determined by the DependencyGraph
func (m *ExpressionManager) ExecuteRegistry() error {
	return nil
}

// resetDependencyGraph resets the DependencyGraph with the current registry state.
func (m *ExpressionManager) resetDependencyGraph(depGraph *graph.DependencyGraph) error {
	// all known paths are vertecies in the graph
	for path := range m.registry {
		err := depGraph.AddVertex(path)
		if err != nil {
			return err
		}
	}

	// now that all paths are added, register the dependencies
	for path, expressions := range m.registry {
		for _, expr := range expressions {
			deps, err := DependentPaths(expr, m.registry, m.variables)
			if err != nil {
				return fmt.Errorf("failed to determine dependent paths: %w", err)
			}

			err = depGraph.RegisterDependencies(path, deps)
			if err != nil {
				return fmt.Errorf("failed to register dependencies: %w", err)
			}
		}
	}

	return nil
}
