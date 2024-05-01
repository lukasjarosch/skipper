package expression

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/lukasjarosch/skipper/v1/data"
)

type state struct {
	node          Node // the current node
	expression    *ExpressionNode
	valueProvider PathValueProvider
	variableMap   map[string]any
	funcMap       map[string]any
}

var (
	zero             reflect.Value
	reflectValueType = reflect.TypeFor[reflect.Value]()

	ErrUndefinedVariable        = fmt.Errorf("undefined variable")
	ErrFunctionNotDefined       = fmt.Errorf("function not defined")
	ErrCallInvalidArgumentCount = fmt.Errorf("invalid argument count")
	ErrNotAFunc                 = fmt.Errorf("not a function")
	ErrBadFuncSignature         = fmt.Errorf("bad function signature")
	ErrIncompatibleArgType      = fmt.Errorf("incompatible argument type")
)

// VariablesInExpression returns a list of all variable names which are used within the expression.

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

// ResolveVariablePath will attempt to convert all VariableNodes into IdentifierNodes.
// Not all VariableNodes are convertible because they may resolve to complex values (e.g. a map).
// Only variables which can also be used within a path (identifiers) can be converted.
// If that translation is possible, the returned path will only contain IdentifierNode as segments.
func ResolveVariablePath(path PathNode, varMap map[string]any) (newPathNode *PathNode, err error) {
	if varMap == nil {
		varMap = make(map[string]any)
	}
	pathSegments := VariableNodesInPathNode(&path)
	if len(pathSegments) == 0 {
		return &path, nil
	}

	newSegmentStrings := make([]string, len(path.Segments))

	for i, node := range pathSegments {
		if node == nil {
			newSegmentStrings[i] = path.Segments[i].(*IdentifierNode).Value
			continue
		}

		varValue, err := ResolveVariable(node, varMap)
		if err != nil {
			return nil, err
		}

		// exclude complex types which cannot be converted to string
		varValueString, err := data.ToString(varValue)
		if err != nil {
			return nil, fmt.Errorf("%s: variable resolves to complex type: %w", node.Name, err)
		}

		if varValueString == "" {
			return nil, fmt.Errorf("%s: variables in path cannot be empty", node.Name)
		}

		newSegmentStrings[i] = varValueString
	}

	// use all string values to build a new expression to parse
	// TODO: what about the positions?
	pathExpr := fmt.Sprintf("${%s}", strings.Join(newSegmentStrings, ":"))

	// handle parser or lexer panics gracefully
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = fmt.Errorf("unexpected parsing error: %w", e)
			} else {
				err = fmt.Errorf("unexpected panic: %v", r)
			}
			newPathNode = nil
		}
	}()

	expressions := Parse(pathExpr)
	if len(expressions) < 1 || len(expressions) > 1 {
		return nil, fmt.Errorf("parse should return one expression but returned %d; this is a bug", len(expressions))
	}
	expr := expressions[0]

	newPathNode, ok := expr.Child.(*PathNode)
	if !ok {
		return nil, fmt.Errorf("invalid expression child, expected PathNode got %T", newPathNode)
	}

	return newPathNode, nil
}

func ResolveVariable(varNode *VariableNode, varMap map[string]any) (interface{}, error) {
	if varNode == nil {
		return nil, fmt.Errorf("nil VariableNode")
	}
	value, exists := varMap[varNode.Name]
	if !exists {
		return zero, fmt.Errorf("%w: %s", ErrUndefinedVariable, varNode.Name)
	}
	return value, nil
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
				s.error(err)
			}
			segments = append(segments, val.String())
		default:
			s.errorf("invalid path segment node: %q", seg)
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

func (s *state) evalString(string *StringNode) reflect.Value {
	s.at(string)
	return reflect.ValueOf(string.Value)
}

func (s *state) evalNumber(num *NumberNode) (reflect.Value, error) {
	s.at(num)
	if num.IsInt {
		return reflect.ValueOf(num.Int64), nil
	}
	if num.IsUint {
		return reflect.ValueOf(num.Uint64), nil
	}
	if num.IsFloat {
		return reflect.ValueOf(num.Float64), nil
	}

	return zero, fmt.Errorf("invalid number")
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

	unwrap := func(v reflect.Value) reflect.Value {
		if v.Kind() != reflect.Interface {
			return v
		}
		if v.IsNil() {
			return reflect.Value{}
		}
		return v.Elem()
	}

	// make argument list
	argv := make([]reflect.Value, numInArgs)
	for i := 0; i < numInArgs; i++ {
		var argValue reflect.Value

		// first, evaluate the value of the argument node
		switch node := call.Arguments[i].(type) {
		case *PathNode:
			val, err := s.evalPath(node)
			if err != nil {
				s.error(err)
			}
			argValue = val
		case *VariableNode:
			val, err := s.evalVariable(node)
			if err != nil {
				s.error(err)
			}
			argValue = val
		case *CallNode:
			val, err := s.evalCall(node)
			if err != nil {
				s.error(err)
			}
			argValue = val
		case *StringNode:
			argValue = s.evalString(node)

		case *NumberNode:
			val, err := s.evalNumber(node)
			if err != nil {
				s.error(err)
			}
			argValue = val
		}

		argType := typ.In(i)

		// argType and argValue need to be of the same type
		// except in cases where argType is interface{}
		if argType.Kind() != reflect.Interface {
			if argType.Kind() != argValue.Kind() {

				// it may very well be that, at this point, the argValue is a `data.Value`
				// because it is the result of an expression which was executed prior to this one
				dataValue, ok := argValue.Interface().(data.Value)
				if !ok {
					s.errorf("%s expected %s; got %s", ErrIncompatibleArgType, argType.Kind(), argValue.Kind())
				}
				if reflect.TypeOf(dataValue.Raw) == argType {
					argValue = reflect.ValueOf(dataValue.Raw)
				} else {
					s.errorf("%s (data.Value) expected %s; got %s", ErrIncompatibleArgType, argType.Kind(), argValue.Kind())
				}

			}
		}

		argv[i] = s.validateType(argValue, argType)
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

	val, err := safeCall(ident, fn, argv)
	_ = unwrap
	return unwrap(val), err
}

// canBeNil reports whether an untyped nil can be assigned to the type. See reflect.Zero.
func canBeNil(typ reflect.Type) bool {
	switch typ.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return true
	case reflect.Struct:
		return typ == reflectValueType
	}
	return false
}

func (s *state) validateType(value reflect.Value, typ reflect.Type) reflect.Value {
	if !value.IsValid() {
		if typ == nil {
			// An untyped nil interface{}. Accept as proper nil value.
			return reflect.ValueOf(nil)
		}
		if canBeNil(typ) {
			// Like above, but use the zero value of the non-nil type.
			return reflect.Zero(typ)
		}
		if typ == reflectValueType && value.Type() != typ {
			return reflect.ValueOf(value)
		}
		if typ != nil && !value.Type().AssignableTo(typ) {
			if value.Kind() == reflect.Interface && !value.IsNil() {
				value = value.Elem()
				if value.Type().AssignableTo(typ) {
					return value
				}
				// fallthrough
			}
			// Does one dereference or indirection work? We could do more, as we
			// do with method receivers, but that gets messy and method receivers
			// are much more constrained, so it makes more sense there than here.
			// Besides, one is almost always all you need.
			switch {
			case value.Kind() == reflect.Pointer && value.Type().Elem().AssignableTo(typ):
				value = value.Elem()
				if !value.IsValid() {
					s.errorf("dereference of nil pointer of type %s", typ)
				}
			case reflect.PointerTo(value.Type()).AssignableTo(typ) && value.CanAddr():
				value = value.Addr()
			default:
				s.errorf("wrong type for value; expected %s; got %s", typ, value.Type())
			}
		}
	}

	return value
}

// func (s *state) evalCallArg(typ reflect.Type, value reflect.Value) reflect.Value {
// 	return value.Convert(typ)
// switch typ.Kind() {
// case reflect.Interface:
// 	// TODO: convert to interface
// 	return value.Convert(typ)
// case reflect.String:
// 	return value.Convert(typ)
// case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
// 	// TODO: convert to int
// 	return value
// case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
// 	// TODO: convert to uint
// 	return value
// case reflect.Float32, reflect.Float64:
// 	// TODO: convert to float
// 	return value
// }
// s.at(node)
// switch typ.Kind() {
// case reflect.Interface:
// 	switch n := node.(type) {
// 	case *PathNode:
// 		val, err := s.evalPath(n)
// 		if err != nil {
// 			s.error(err)
// 		}
// 		return val
// 	case *StringNode:
// 		val := s.evalString(reflect.TypeOf(""), n)
// 		return val
// 	}
// 	spew.Dump(typ, node)
// case reflect.String:
// 	switch n := node.(type) {
// 	case *VariableNode:
// 		val, err := s.evalVariable(n)
// 		if err != nil {
// 			s.error(err)
// 		}
// 		if typ != val.Type() {
// 			s.errorf("argument must be of type %s, but variable value is %q %s", typ.String(), val, val.String())
// 		}
// 		return val
// 	case *CallNode:
// 		val, err := s.evalCall(n)
// 		if err != nil {
// 			s.error(err)
// 		}
// 		if typ != val.Type() {
// 			s.errorf("argument must be of type %s, but call value is %q %s", typ.String(), val, val.String())
// 		}
// 		return val
// 	case *PathNode:
// 		val, err := s.evalPath(n)
// 		if err != nil {
// 			s.error(err)
// 		}
// 		if typ != val.Type() {
// 			s.errorf("argument must be of type %s, but path value is %q %s", typ.String(), val, val.String())
// 		}
// 		return val
// 	default:
// 		return s.evalString(typ, node)
// 	}
// case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
// 	// TODO: handle variable and call
// 	return s.evalInteger(typ, node)
// }

// TODO: handle floats
// TODO: handle boolean

// return zero
// }

// func (s *state) evalString(typ reflect.Type, n Node) reflect.Value {
// 	s.at(n)
//
// 	if n, ok := n.(*StringNode); ok {
// 		value := reflect.New(typ).Elem()
// 		value.SetString(n.Value)
// 		return value
// 	}
// 	s.errorf("expected string, found %s", n)
// 	panic("not reached")
// }
//
// func (s *state) evalInteger(typ reflect.Type, n Node) reflect.Value {
// 	s.at(n)
//
// 	if n, ok := n.(*NumberNode); ok && n.IsInt {
// 		value := reflect.New(typ).Elem()
// 		value.SetInt(n.Int64)
// 		return value
// 	}
// 	s.errorf("expected integer; found %s", n)
// 	panic("not reached")
// }
