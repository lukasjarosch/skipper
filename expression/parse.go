package expression

import (
	"fmt"
	"strings"

	"github.com/davecgh/go-spew/spew"

	"github.com/lukasjarosch/skipper/data"
)

type Tree struct {
	root         *ListNode
	lex          *lexer
	input        string
	token        [3]Token // lookahead buffer
	inExpression bool
	peekCount    int
}

func Parse(text string) []*ExpressionNode {
	t := &Tree{}
	return t.Parse(text)
}

func ParsePath(segments []string, variables map[string]any) (*PathNode, error) {
	pathExpression := fmt.Sprintf("${%s}", strings.Join(segments, ":"))

	t := &Tree{}
	exprNodes := t.Parse(pathExpression)

	if len(exprNodes) == 0 {
		return nil, fmt.Errorf("parsing did not return a valid expression")
	}

	// there is at least one expression (there won't be more due to the input)
	switch node := exprNodes[0].Child.(type) {
	case *PathNode:
		for _, segment := range node.Segments {
			// skip identifier nodes
			if segment.Type() == NodeIdentifier {
				continue
			}

			// node is a variable
			varNode, ok := segment.(*VariableNode)
			if !ok {
				return nil, fmt.Errorf("unexpected node type: %s", segment.Type().String())
			}

			variableValue, exists := variables[varNode.Name]
			if !exists {
				return nil, fmt.Errorf("%w: %s", ErrUndefinedVariable, varNode.Name)
			}

			data.ToString(variableValue)
			spew.Dump("VAR", variableValue)
		}
		return node, nil
	default:
		return nil, fmt.Errorf("expected path node, got %q", node)
	}
}

func (t *Tree) Parse(input string) []*ExpressionNode {
	t.input = input
	t.lex = lex(input)
	t.parse()

	expr := []*ExpressionNode{}
	for _, node := range t.root.Nodes {
		expr = append(expr, node.(*ExpressionNode))
	}

	return expr
}

// expect consumes the next token and guarantees it has the required type.
func (t *Tree) expect(expected TokenType, context string) Token {
	token := t.next()
	if token.Type != expected {
		t.unexpected(token, context)
	}
	return token
}

// unexpected complains about the token and terminates processing.
func (t *Tree) unexpected(token Token, context string) {
	if token.Type == tError {
		t.errorfWithContext(token, "%s in %s: %s", TokenString(token.Type), context, token.Value)
	}
	t.errorfWithContext(token, "unexpected %s in %s", TokenString(token.Type), context)
}

// errorf formats the error and terminates processing.
func (t *Tree) errorf(format string, args ...any) {
	t.root = nil
	format = fmt.Sprintf("parse: %s at %d: %s", t.token[0].Value, t.token[0].Pos, format)
	panic(fmt.Errorf(format, args...))
}

// errorfWithContext calls 'errorf' but adds failure context for the user based on the token.
func (t *Tree) errorfWithContext(tok Token, format string, args ...interface{}) {
	line := func(pos int) (string, int) {
		// in case the input is multiline, extract just the line we're in
		if strings.Contains(t.input, "\n") {
			beforeNewLine := strings.LastIndex(t.input[:pos], "\n") + 1
			afterNewLine := strings.Index(t.input[pos:], "\n") + pos
			return strings.TrimSpace(t.input[beforeNewLine:afterNewLine]), pos - beforeNewLine - 1
		}

		return t.input, pos
	}

	context := "\nContext:"
	context += "\n|"

	contextLine, newPos := line(tok.Pos)
	context += fmt.Sprintf("\n| %s\n", contextLine)
	context += fmt.Sprintf("| %s^--HERE\n", strings.Repeat(" ", newPos-1))

	format += "\n%s"
	args = append(args, context)
	t.errorf(format, args...)
}

// peek returns the next token without consuming it
func (t *Tree) peek() Token {
	if t.peekCount > 0 {
		return t.token[t.peekCount-1]
	}
	t.peekCount = 1
	t.token[0] = t.lex.nextToken()
	return t.token[0]
}

// backup backs the input stream up one token.
func (t *Tree) backup() {
	t.peekCount++
}

// backup2 backs the input stream up two tokens.
// The zeroth token is already there.
func (t *Tree) backup2(t1 Token) {
	t.token[1] = t1
	t.peekCount = 2
}

// next returns the next token.
func (t *Tree) next() Token {
	if t.peekCount > 0 {
		t.peekCount--
	} else {
		t.token[0] = t.lex.nextToken()
	}
	return t.token[t.peekCount]
}

// parse starts parsing the input
//
// grammar ::= '${' expression '}'
func (t *Tree) parse() {
	t.root = t.newList(Pos(t.peek().Pos))
	for t.peek().Type != tEOF {

		// consume the next token if its a left delimiter
		// otherwise backup
		if tok := t.next(); tok.Type == tLeftDelim {
			n := t.parseExpression()
			if n != nil {
				t.root.append(n)
			}
			continue
		}
		t.backup()

		// In case we're already inside an expression, go back in.
		if t.inExpression {
			t.parseExpression()
			continue
		}
	}
}

// parseExpression
//
// expression ::= standalone_variable | inline_variable | path | call
//
// NOTE: The left delimiter is already consumed at this point.
func (t *Tree) parseExpression() (expr *ExpressionNode) {
	t.inExpression = true

	switch tok := t.peek(); tok.Type {
	// inline variable (with dollar prefix)
	case tDollar:
		return t.newExpression(Pos(tok.Pos), t.parseInlineVariable())
	case tIdent:
		ident := t.next() // swallow identifier to peek at the next token

		switch tok := t.peek(); tok.Type {
		// identifier followed by '(' -> Call
		case tLeftParen:
			t.backup2(ident) // restore identifier
			return t.newExpression(Pos(tok.Pos), t.parseCall())

		// identifier followed by tPathSep -> Path
		case tPathSep:
			t.backup2(ident) // restore identifier
			return t.newExpression(Pos(tok.Pos), t.parsePath())

		// only an identifier, then a right-delimiter -> standalone_variable
		case tRightDelim:
			t.backup2(ident)
			return t.newExpression(Pos(ident.Pos), t.parseStandaloneVariable())

		default:
			t.errorfWithContext(tok, "unexpected %s after identifier", TokenString(tok.Type))
		}

	// expression ends
	case tRightDelim:
		t.next()
		t.inExpression = false
		return nil // nothing to return, all expression nodes were already emitted above

	case tError:
		t.errorfWithContext(tok, "lexer error")
	default:
		t.unexpected(tok, "parseExpression")
	}

	return
}

// parseInlineVariable
//
// inline_variable ::= '$' standalone_variable
func (t *Tree) parseInlineVariable() *VariableNode {
	t.expect(tDollar, "parseVariable")

	return t.parseStandaloneVariable()
}

// parseStandaloneVariable
//
// standalone_variable ::= identifier
func (t *Tree) parseStandaloneVariable() *VariableNode {
	tok := t.next()
	if tok.Type != tIdent {
		t.unexpected(tok, "parseVariable")
	}

	return t.newVariable(Pos(tok.Pos), tok.Value)
}

// parseCall
//
// call 					::= identifier "(" argument_list ")" alternative_expression
// alternative_expression 	::= "||" expression
func (t *Tree) parseCall() Node {
	ident := t.parseIdentifier()
	call := t.newCall(ident.Pos, ident)

	t.expect(tLeftParen, "parseCall")

	// argument list surrounded by parentheses
	for _, arg := range t.parseCallArgumentList() {
		call.appendArgument(arg)
	}

	// closing parentheses
	t.expect(tRightParen, "parseCall")

	// alternative exepression
	if t.peek().Type == tDoublePipe {
		t.expect(tDoublePipe, "parseCall")
		call.AlternativeExpr = t.parseExpression()
	}

	return call
}

// parseCallArgumentList
//
// argument_list 		::= {(quoted_string | inline_variable | path | call)} {argument_list_tail}
// argument_list_tail 	::= {"," (quoted_string | inline_variable | path | call)}
func (t *Tree) parseCallArgumentList() (args []Node) {
	for t.peek().Type != tRightParen {
		tok := t.peek()

		switch tok.Type {

		// variable argument
		case tDollar:
			args = append(args, t.parseInlineVariable())
			continue

		// path or call argument
		case tIdent:
			ident := t.next()

			switch tok := t.peek(); tok.Type {
			// path is argument
			case tPathSep:
				t.backup2(ident)
				args = append(args, t.parsePath())

			// call is argument
			case tLeftParen:
				t.backup2(ident)
				args = append(args, t.parseCall())

			default:
				t.errorfWithContext(tok, "expected path separator or left parentheses after identifier, got %s", TokenString(tok.Type))
			}
			continue

		// quoted string argument
		case tString:
			args = append(args, t.parseString())
			continue
		case tNumber:
			args = append(args, t.parseNumber())
			continue
		case tError:
			t.errorfWithContext(tok, "lexer error")
		}

		// if there are more args, there must be a comma
		// anything other is a syntax error
		if tok.Type != tRightParen {
			switch tok.Type {
			case tComma:
				t.next() // consume comma, and continue parsing args
				continue
			default:
				t.errorfWithContext(t.peek(), "unexpected %s in argument list", TokenString(t.peek().Type))
			}
		}
	}

	return
}

// parseIdentifier expects the next token to be tIdent.
func (t *Tree) parseIdentifier() *IdentifierNode {
	tok := t.next()
	if tok.Type != tIdent {
		t.unexpected(tok, "parseIdentifier")
	}

	return t.newIdentifier(Pos(tok.Pos), tok.Value)
}

// parseString expects the next token to be tString
func (t *Tree) parseString() *StringNode {
	tok := t.next()
	if tok.Type != tString {
		t.unexpected(tok, "parseString")
	}
	return t.newString(Pos(tok.Pos), tok.Value)
}

func (t *Tree) parseNumber() *NumberNode {
	tok := t.next()
	if tok.Type != tNumber {
		t.unexpected(tok, "parseNumber")
	}

	node, err := t.newNumber(Pos(tok.Pos), tok.Value)
	if err != nil {
		t.errorfWithContext(tok, err.Error())
		return nil
	}

	return node
}

// parsePath
//
// path 		::= identifier, path_tail
// path_tail 	::= {":" (identifier | inline_variable)}+
//
// Thus, a path MUST start with an identifier and must have at lest
// one path_tail segment.
func (t *Tree) parsePath() Node {
	// a path can have only 256 path segments
	const maxLength = 256

	path := t.newPath(Pos(t.peek().Pos))

	// Every second (uneven) token must be a path identifier
	// There are maxLength-1 identifiers in a path of maxLength
	// If this loop terminates, the path is too long.
	for i := 0; i <= (maxLength + (maxLength - 1)); i++ {
		// The first segment of a path must be an identifier.
		if i == 0 {
			path.appendSegment(t.parseIdentifier())
			continue
		}

		// Intermediate segments may be either identifiers or variables
		switch tok := t.peek(); tok.Type {
		case tIdent:
			path.appendSegment(t.parseIdentifier())
			continue
		case tDollar:
			path.appendSegment(t.parseInlineVariable())
			continue
		case tRightDelim, tRightParen, tComma:
			return path
		}

		// Every second token must be a separator
		if i%2 == 1 {
			t.expect(tPathSep, "parsePath")
			continue
		}
	}

	t.errorf("path is too long, max length is %d segments", maxLength)
	return nil
}
