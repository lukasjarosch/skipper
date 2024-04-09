package expression

import (
	"errors"
	"fmt"

	"github.com/dominikbraun/graph"
)

// expressionNodeHash returns a string which allows the graph to
// distinguish expression nodes.
func expressionNodeHash(expr *ExpressionNode) string {
	return expr.Text()
}

var dependencyGraph graph.Graph[string, *ExpressionNode] = graph.New(expressionNodeHash, graph.Acyclic(), graph.Directed(), graph.PreventCycles())

func InitializeDependencyGraph(expressions []*ExpressionNode) error {
	for _, expr := range expressions {
		err := dependencyGraph.AddVertex(expr)
		if err != nil {
			// ignore duplicate vertex errors
			if !errors.Is(err, graph.ErrVertexAlreadyExists) {
				return fmt.Errorf("cannot add expression to graph: %w", err)
			}
		}
	}

	return nil
}

// PathMap maps path strings (dot-separated) and maps it to the expressions which occur at the path.
type PathMap map[string][]*ExpressionNode

// Dependencies takes an expression node and a [PathMap]
// and returns those on which the passed expression depends on.
//
// An ExpressionNode is dependent on another if
// it has a PathNode at which's value another expression exists.
func Dependencies(expr *ExpressionNode, allExpressions PathMap, variables map[string]interface{}) ([]*ExpressionNode, error) {
	pathNodes := PathsInExpression(expr)
	resolvedPathNodes := make([]*PathNode, len(pathNodes))

	// Resolve variables within the pathNodes, leaving us with PathNodes with only IdentifierNodes as Segments.
	// Only those paths can possibly be valid skipper paths.
	for i, pathNode := range pathNodes {
		res, err := ResolveVariablePath(*pathNode, variables)
		if err != nil {
			return nil, err
		}
		resolvedPathNodes[i] = res
	}

	// If the skipper path occurs as key in allExpressions,
	// then all those expressions are direct dependencies to the current one.
	var dependingOnExpressions []*ExpressionNode
	for _, path := range resolvedPathNodes {
		if expr, ok := allExpressions[path.SkipperPath().String()]; ok {
			dependingOnExpressions = append(dependingOnExpressions, expr...)
		}
	}

	return dependingOnExpressions, nil
}

// PathsInExpression returns all PathNodes which occur in the expression
func PathsInExpression(expr *ExpressionNode) (paths []*PathNode) {
	pathsInCall := func(call *CallNode) (paths []*PathNode) {
		for _, argNode := range call.Arguments {
			switch node := argNode.(type) {
			case *PathNode:
				paths = append(paths, node)
			default:
				continue
			}
		}
		if call.AlternativeExpr != nil {
			paths = append(paths, PathsInExpression(call.AlternativeExpr)...)
		}
		return
	}

	switch node := expr.Child.(type) {
	case *PathNode:
		paths = append(paths, node)
	case *CallNode:
		paths = append(paths, pathsInCall(node)...)
	default:
		// fallthrough
	}
	return
}

func VariablesInPathNode(path *PathNode) (vars []string) {
	for _, segNode := range path.Segments {
		switch node := segNode.(type) {
		case *VariableNode:
			vars = append(vars, node.Name)
		default:
			continue
		}
	}
	return
}

func VariableNodesInPathNode(path *PathNode) (vars []*VariableNode) {
	vars = make([]*VariableNode, len(path.Segments))

	for i, segNode := range path.Segments {
		switch node := segNode.(type) {
		case *VariableNode:
			vars[i] = node
		default:
			continue
		}
	}
	return
}

func VariablesInExpression(expr *ExpressionNode) (variableNames []string) {
	var variablesInCall func(*CallNode) []string
	variablesInCall = func(call *CallNode) (vars []string) {
		for _, argNode := range call.Arguments {
			switch node := argNode.(type) {
			case *VariableNode:
				vars = append(vars, node.Name)
			case *PathNode:
				vars = append(vars, VariablesInPathNode(node)...)
			case *CallNode:
				vars = append(vars, variablesInCall(node)...)
			}
		}
		if call.AlternativeExpr != nil {
			vars = append(vars, VariablesInExpression(call.AlternativeExpr)...)
		}
		return
	}

	switch node := expr.Child.(type) {
	case *PathNode:
		variableNames = append(variableNames, VariablesInPathNode(node)...)
	case *CallNode:
		variableNames = append(variableNames, variablesInCall(node)...)
	case *VariableNode:
		variableNames = append(variableNames, node.Name)
	}

	return
}
