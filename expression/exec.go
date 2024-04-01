package expression

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/lukasjarosch/skipper/data"
)

type state struct {
	node          Node // the current node
	expression    *ExpressionNode
	valueProvider PathValueProvider
	variableMap   map[string]any
	funcMap       map[string]any
}

var (
	zero reflect.Value

	ErrUndefinedVariable        = fmt.Errorf("undefined variable")
	ErrFunctionNotDefined       = fmt.Errorf("function not defined")
	ErrCallInvalidArgumentCount = fmt.Errorf("invalid argument count")
	ErrNotAFunc                 = fmt.Errorf("not a function")
	ErrBadFuncSignature         = fmt.Errorf("bad function signature")
)

// UsedVariables returns a list of all variable names which are used within the expression.
func UsedVariables(expr *ExpressionNode) (variableNames []string) {
	variablesInPath := func(path *PathNode) (vars []string) {
		for _, segNode := range path.Segments {
			switch node := segNode.(type) {
			case *IdentifierNode:
				continue
			case *VariableNode:
				vars = append(vars, node.Name)
			}
		}
		return
	}
	var variablesInCall func(*CallNode) []string
	variablesInCall = func(call *CallNode) (vars []string) {
		for _, argNode := range call.Arguments {
			switch node := argNode.(type) {
			case *VariableNode:
				vars = append(vars, node.Name)
			case *PathNode:
				vars = append(vars, variablesInPath(node)...)
			case *CallNode:
				vars = append(vars, variablesInCall(node)...)
			}
		}
		if call.AlternativeExpr != nil {
			vars = append(vars, UsedVariables(call.AlternativeExpr)...)
		}
		return
	}

	switch node := expr.Child.(type) {
	case *PathNode:
		variableNames = append(variableNames, variablesInPath(node)...)
	case *CallNode:
		variableNames = append(variableNames, variablesInCall(node)...)
	case *VariableNode:
		variableNames = append(variableNames, node.Name)
	}

	return
}

type PathValueProvider interface {
	GetPath(data.Path) (interface{}, error)
}

func Execute(expr *ExpressionNode, valueProvider PathValueProvider, variableMap map[string]any, funcMap map[string]any) (val reflect.Value, err error) {
	state := &state{
		expression:    expr,
		valueProvider: valueProvider,
		variableMap:   variableMap,
		funcMap:       funcMap,
	}
	defer errRecover(&err)
	val, err = state.walkExpression(state.expression)
	return
}

// at marks the node as current node
func (s *state) at(node Node) {
	s.node = node
}

// errRecover is the handler that turns panics into returns from the top
// level of Parse.
func errRecover(errp *error) {
	e := recover()
	if e != nil {
		*errp = fmt.Errorf("%v", e)
	}
}

func (s *state) error(err error) {
	s.errorf("error: %w", err)
}

func (s *state) errorf(format string, args ...any) {
	if s.node != nil {
		panic(fmt.Errorf("%w\n%s", fmt.Errorf(format, args...), s.expression.ErrorContext(s.node)))
	}

	panic(fmt.Errorf(format, args...))
}

func (s *state) walkExpression(node *ExpressionNode) (reflect.Value, error) {
	s.at(node)
	switch node := node.Child.(type) {
	case *PathNode:
		return s.evalPath(node)
	case *VariableNode:
		return s.evalVariable(node)
	case *CallNode:
		return s.evalCall(node)
	}

	return reflect.ValueOf(nil), fmt.Errorf("unimplemented")
}

func (s *state) evalPath(path *PathNode) (reflect.Value, error) {
	s.at(path)

	segments := []string{}

	for _, seg := range path.Segments {
		switch segment := seg.(type) {
		case *IdentifierNode:
			segments = append(segments, segment.Value)
		case *VariableNode:
			val, err := s.evalVariable(segment)
			if err != nil {
				panic(err) // TODO: implement
			}
			segments = append(segments, val.String())
		default:
			panic("NOT IMPLEMENTED")
		}
	}

	val, err := s.valueProvider.GetPath(data.NewPathVar(segments...))
	if err != nil {
		s.error(err)
	}

	return reflect.ValueOf(val), nil
}

func (s *state) evalVariable(variable *VariableNode) (reflect.Value, error) {
	s.at(variable)
	value, exists := s.variableMap[variable.Name]
	if !exists {
		return zero, fmt.Errorf("%w: %s", ErrUndefinedVariable, variable.Name)
	}
	return reflect.ValueOf(value), nil
}

func (s *state) evalCall(call *CallNode) (reflect.Value, error) {
	s.at(call)

	ident := call.Identifier.Value
	fn, ok := findFunction(ident, s)
	if !ok {
		return zero, fmt.Errorf("%w: %s", ErrFunctionNotDefined, ident)
	}

	// make sure we're actually dealing with a func
	typ := fn.Type()
	if typ.Kind() != reflect.Func {
		return zero, fmt.Errorf("%w: %s", ErrNotAFunc, ident)
	}

	// number of argument nodes must match the parameter count of the func
	numInArgs := len(call.Arguments)
	if numInArgs != typ.NumIn() {
		return zero, fmt.Errorf("%w for %s: want %d got %d", ErrCallInvalidArgumentCount, ident, typ.NumIn(), numInArgs)
	}

	// assert that the function return values are valid
	if !goodFunc(typ) {
		outStr := []string{}
		for i := 0; i < typ.NumOut(); i++ {
			outStr = append(outStr, typ.Out(i).String())
		}
		return zero, fmt.Errorf("%w %s: does not meet the requirements: %s() (%s)", ErrBadFuncSignature, ident, ident, strings.Join(outStr, ", "))
	}

	// make argument list
	argv := make([]reflect.Value, numInArgs)
	for i := 0; i < numInArgs; i++ {
		argv[i] = s.evalCallArg(typ.In(i), call.Arguments[i])
	}

	// in case of an existing AlternativeExpr, perform the call and execute the AlternativeExpr in case of an error
	if call.AlternativeExpr != nil {
		val, err := safeCall(ident, fn, argv)
		if err != nil {
			return Execute(call.AlternativeExpr, s.valueProvider, s.variableMap, s.funcMap)
		}
		return val, nil
	}

	// TODO: handle variadic functions

	return safeCall(ident, fn, argv)
}

func (s *state) evalCallArg(typ reflect.Type, node Node) reflect.Value {
	s.at(node)
	switch typ.Kind() {
	case reflect.String:
		switch n := node.(type) {
		case *VariableNode:
			val, err := s.evalVariable(n)
			if err != nil {
				s.error(err)
			}
			return val
		case *CallNode:
			val, err := s.evalCall(n)
			if err != nil {
				s.error(err)
			}
			return val
		default:
			return s.evalString(typ, node)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// TODO: handle variable and call
		return s.evalInteger(typ, node)
	}

	// TODO: handle floats

	return zero
}

func (s *state) evalString(typ reflect.Type, n Node) reflect.Value {
	s.at(n)

	if n, ok := n.(*StringNode); ok {
		value := reflect.New(typ).Elem()
		value.SetString(n.Value)
		return value
	}
	s.errorf("expected string, found %s", n)
	panic("not reached")
}

func (s *state) evalInteger(typ reflect.Type, n Node) reflect.Value {
	s.at(n)

	if n, ok := n.(*NumberNode); ok && n.IsInt {
		value := reflect.New(typ).Elem()
		value.SetInt(n.Int64)
		return value
	}
	s.errorf("expected integer; found %s", n)
	panic("not reached")
}
