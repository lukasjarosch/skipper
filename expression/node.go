package expression

import (
	"fmt"
	"strconv"
	"strings"
)

type Node interface {
	Type() NodeType
	Position() Pos
	Text() string
}

// NodeType identifies the type of a parse tree node.
type NodeType int

// Pos represents a byte position in the original input text
type Pos int

func (p Pos) Position() Pos {
	return p
}

// Type returns itself and provides an easy default implementation
// for embedding in a Node. Embedded in all non-trivial Nodes.
func (t NodeType) Type() NodeType {
	return t
}

func (t NodeType) Text() string {
	return ""
}

func (t NodeType) String() string {
	switch t {
	case NodeExpression:
		return "Expression"
	case NodeList:
		return "List"
	case NodePath:
		return "Path"
	case NodeIdentifier:
		return "Identifier"
	case NodeVariable:
		return "Variable"
	case NodeCall:
		return "Call"
	case NodeString:
		return "String"
	case NodeNumber:
		return "Number"
	default:
		return "UNKNOWN NODE TYPE"
	}
}

const (
	NodeExpression NodeType = iota
	NodeList
	NodePath
	NodeIdentifier
	NodeVariable
	NodeCall
	NodeString
	NodeNumber
)

// ListNode holds a sequence of nodes.
type ListNode struct {
	NodeType
	Pos
	Nodes []Node // The element nodes in lexical order.
}

func (t *Tree) newList(pos Pos) *ListNode {
	return &ListNode{NodeType: NodeList, Pos: pos}
}

func (l *ListNode) append(n Node) {
	l.Nodes = append(l.Nodes, n)
}

type ExpressionNode struct {
	NodeType
	Pos
	Child Node
}

func (n ExpressionNode) Text() string {
	return fmt.Sprintf("${%s}", n.Child.Text())
}

func (n ExpressionNode) ErrorContext(node Node) string {
	underline := func(a, b string) string {
		return strings.Repeat(" ", strings.Index(a, b)) + strings.Repeat("^", len(b))
	}

	context := fmt.Sprintln("Context:")
	// context += fmt.Sprintln("|")
	context += fmt.Sprintln("|", n.Text())
	context += fmt.Sprintln("|", underline(n.Text(), node.Text()), "-- HERE")

	return context
}

func (t *Tree) newExpression(pos Pos, child Node) *ExpressionNode {
	return &ExpressionNode{Pos: pos, NodeType: NodeExpression, Child: child}
}

type VariableNode struct {
	NodeType
	Pos
	Name string
}

func (t *Tree) newVariable(pos Pos, name string) *VariableNode {
	return &VariableNode{Pos: pos, Name: name, NodeType: NodeVariable}
}

func (v *VariableNode) Text() string {
	return "$" + v.Name
}

type CallNode struct {
	NodeType
	Pos
	Identifier      *IdentifierNode
	Arguments       []Node
	AlternativeExpr *ExpressionNode
}

func (t *Tree) newCall(pos Pos, ident *IdentifierNode) *CallNode {
	return &CallNode{Pos: pos, Identifier: ident, NodeType: NodeCall}
}

func (n *CallNode) Text() string {
	args := []string{}
	for _, a := range n.Arguments {
		args = append(args, a.Text())
	}

	if n.AlternativeExpr != nil {
		return fmt.Sprintf("%s(%s) || %s", n.Identifier.Text(), strings.Join(args, ", "), n.AlternativeExpr.Text())
	}

	return fmt.Sprintf("%s(%s)", n.Identifier.Text(), strings.Join(args, ", "))
}

func (n *CallNode) appendArgument(arg Node) {
	n.Arguments = append(n.Arguments, arg)
}

type PathNode struct {
	NodeType
	Pos
	Segments []Node // path segments from left to right, without separators
}

func (t *Tree) newPath(pos Pos) *PathNode {
	return &PathNode{Pos: pos, NodeType: NodePath}
}

func (n *PathNode) appendSegment(node Node) {
	n.Segments = append(n.Segments, node)
}

func (n *PathNode) Text() string {
	segments := []string{}
	for _, seg := range n.Segments {
		segments = append(segments, seg.Text())
	}
	return strings.Join(segments, ":")
}

type IdentifierNode struct {
	NodeType
	Pos
	Value string
}

func (t *Tree) newIdentifier(pos Pos, value string) *IdentifierNode {
	return &IdentifierNode{Pos: pos, NodeType: NodeIdentifier, Value: value}
}

func (i *IdentifierNode) Text() string {
	return i.Value
}

type StringNode struct {
	NodeType
	Pos
	Value string
}

func (t *Tree) newString(pos Pos, value string) *StringNode {
	return &StringNode{Pos: pos, NodeType: NodeString, Value: value}
}

func (s *StringNode) Text() string {
	return fmt.Sprintf("\"%s\"", s.Value)
}

type NumberNode struct {
	NodeType
	Pos
	IsInt   bool    // Number has an integral value.
	IsUint  bool    // Number has an unsigned integral value.
	IsFloat bool    // Number has a floating-point value.
	Int64   int64   // The signed integer value.
	Uint64  uint64  // The unsigned integer value.
	Float64 float64 // The floating-point value.
	Value   string
}

func (n *NumberNode) Text() string {
	return n.Value
}

func (t *Tree) newNumber(pos Pos, value string) (*NumberNode, error) {
	n := &NumberNode{Pos: pos, Value: value, NodeType: NodeNumber}

	u, err := strconv.ParseUint(value, 0, 64)
	if err == nil {
		n.IsUint = true
		n.Uint64 = u
	}
	i, err := strconv.ParseInt(value, 0, 64)
	if err == nil {
		n.IsInt = true
		n.Int64 = i
	}

	if n.IsInt {
		n.IsFloat = true
		n.Float64 = float64(n.Int64)
	} else if n.IsUint {
		n.IsFloat = true
		n.Float64 = float64(n.Uint64)
	} else {
		f, err := strconv.ParseFloat(value, 64)
		if err == nil {
			n.IsFloat = true
			n.Float64 = f

			// a float may also be a valid integer
			if !n.IsInt && float64(int64(f)) == f {
				n.IsInt = true
				n.Int64 = int64(f)
			}
			if !n.IsUint && float64(uint64(f)) == f {
				n.IsUint = true
				n.Uint64 = uint64(f)
			}
		}
	}

	if !n.IsInt && !n.IsUint && !n.IsFloat {
		return nil, fmt.Errorf("illegal number syntax: %q", value)
	}

	return n, nil
}
