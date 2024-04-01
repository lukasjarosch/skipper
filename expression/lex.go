package expression

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

type TokenType int

const (
	tError TokenType = iota
	tEOF
	tIdent
	tLeftDelim
	tRightDelim
	tLeftParen
	tRightParen
	tPathSep    // :
	tDoublePipe // ||
	tDollar
	tComma

	tString
	tNumber
)

func TokenString(t TokenType) string {
	switch t {
	case tEOF:
		return "EOF"
	case tError:
		return "ERROR"
	case tString:
		return "STRING"
	case tIdent:
		return "IDENTIFIER"
	case tPathSep:
		return "PATH_SEPARATOR"
	case tLeftDelim:
		return "LEFT_DELIMITER"
	case tRightDelim:
		return "RIGHT_DELIMITER"
	case tLeftParen:
		return "LEFT_PARENTHESES"
	case tRightParen:
		return "RIGHT_PARENTHESES"
	case tDoublePipe:
		return "DOUBLE_PIPE"
	case tDollar:
		return "DOLLAR"
	case tComma:
		return "COMMA"
	case tNumber:
		return "NUMBER"
	default:
		return "UNKNOWN"
	}
}

const eof = -1

type Token struct {
	Pos   int
	Type  TokenType
	Value string
}

type stateFn func(*lexer) stateFn

type lexer struct {
	input      string
	start      int
	pos        int
	width      int
	tokens     chan Token
	token      Token
	parenDepth int // nesting depth of '( )' expressions
	exprDepth  int // nesting depth of delimited expressions (e.g. ${foo:${bar}})
}

func lex(input string) *lexer {
	l := &lexer{
		input:  input,
		tokens: make(chan Token, 3),
	}
	go l.run()
	return l
}

// run starts the lexer
func (l *lexer) run() {
	for state := lexText; state != nil; {
		state = state(l)
	}
	close(l.tokens)
}

// nextToken returns the next Token from the input.
// Called by the parser, not the lexer!
func (l *lexer) nextToken() Token {
	for {
		select {
		case token, ok := <-l.tokens:
			if !ok {
				return Token{Type: tEOF, Value: "EOF"}
			}
			return token
		default:
		}
	}
}

func (l *lexer) emit(t TokenType) {
	value := l.current()
	l.tokens <- Token{
		Pos:   l.pos,
		Value: value,
		Type:  t,
	}
	l.updatePos()
}

func (l *lexer) next() rune {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = w
	l.pos += l.width
	return r
}

// accept consumes the next rune if it's in the valid set
func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set
func (l *lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

// acceptRegexRun consumes all runes which match the given regex
func (l *lexer) acceptRegexRun(valid *regexp.Regexp) {
	for valid.MatchString(string(l.next())) {
	}
	l.backup()
}

func (l *lexer) current() string {
	return l.input[l.start:l.pos]
}

func (l *lexer) updatePos() {
	l.start = l.pos
}

func (l *lexer) ignore() {
	l.start = l.pos
}

func (l *lexer) backup() {
	l.pos -= l.width
}

func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *lexer) atLeftDelim() bool {
	return strings.HasPrefix(l.input[l.pos:], leftDelim)
}

func (l *lexer) atRightDelim() bool {
	return strings.HasPrefix(l.input[l.pos:], rightDelim)
}

// isSpace reports whether r is a space character.
func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\r' || r == '\n'
}

// isAlphaNumeric reports whether r is an alphabetic, digit, or underscore.
func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

// ----- state transition functions

const (
	leftDelim  = "${"
	rightDelim = "}"
)

// errorf emits an error token and returns nil
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.tokens <- Token{
		Type:  tError,
		Pos:   l.pos,
		Value: fmt.Sprintf(format, args...),
	}
	return nil
}

func lexText(l *lexer) stateFn {
	for {
		if l.atLeftDelim() {
			if l.pos > l.start {
				l.ignore() // ignore any preceding text
			}
			l.pos += len(leftDelim) // skip leftDelim
			l.emit(tLeftDelim)
			l.exprDepth++
			return lexExpression
		}

		if l.next() == eof {
			break
		}
	}
	l.ignore() // drop any text
	l.emit(tEOF)
	return nil
}

func lexExpression(l *lexer) stateFn {
	if l.atRightDelim() {
		l.pos += len(rightDelim) // skip rightDelim
		l.emit(tRightDelim)
		l.exprDepth--
		if l.exprDepth == 0 {

			// if there are still unclosed function calls, the expression cannot be ending here
			if l.parenDepth > 0 {
				return l.errorf("missing right parentheses")
			}

			return lexText
		}
		if l.exprDepth < 0 {
			return l.errorf("unexpected right delimiter %s", rightDelim)
		}
		return lexExpression
	}

	switch r := l.next(); {
	case r == '+' || r == '-' || ('0' <= r && r <= '9'):
		l.backup()
		return lexNumber
	case isAlphaNumeric(r):
		l.backup()
		return lexIdentifier
	case r == ':':
		l.emit(tPathSep)
		return lexExpression
	case r == '$':
		// nested expression?
		if l.peek() == '{' {
			l.next()
			l.emit(tLeftDelim)
			l.exprDepth++
			return lexExpression
		}
		// dollars indicate a variable
		l.emit(tDollar)
		return lexExpression

	// start param list
	case r == '(':
		l.parenDepth++
		l.emit(tLeftParen)
		return lexExpression

	// end param list
	case r == ')':
		l.parenDepth--
		l.emit(tRightParen)
		return lexExpression

	// quoted strings
	case r == '\'':
		l.ignore()
		return lexQuotedString('\'')
	case r == '"':
		l.ignore()
		return lexQuotedString('"')

	// commas within parameter lists
	case r == ',':
		l.emit(tComma)
		return lexExpression

	// drop any spaces within the expression
	case isSpace(r):
		l.ignore()
		return lexExpression

	// alternate expressions
	case r == '|':
		if l.peek() == '|' {
			l.next()
			l.emit(tDoublePipe)
			return lexExpression
		}
		return l.errorf("invalid token %#U", r)

	// input cannot end before the rightDelim is found
	case r == eof:
		return l.errorf("unclosed expression, expected %s, got %s", rightDelim, TokenString(tEOF))

		// fail on any other rune
	default:
		return l.errorf("unrecognized rune in expression: %#U", r)
	}
}

func lexIdentifier(l *lexer) stateFn {
	l.acceptRegexRun(regexp.MustCompile(`\w+`))
	l.emit(tIdent)
	return lexExpression
}

func lexQuotedString(quote rune) stateFn {
	return func(l *lexer) stateFn {
	Loop:
		for {
			switch l.next() {
			case eof, '\n':
				return l.errorf("unterminated single quoted string")
			case quote:
				l.backup()
				break Loop
			}
		}

		l.emit(tString)

		// skip quote and return
		l.next()
		l.ignore()
		return lexExpression
	}
}

// lexNumber scans a number.
// This is very basic as there is no support for octal, hex, imaginary, etc.
// The only supported numbers are integers and floats including signs
func lexNumber(l *lexer) stateFn {
	// optional: leading sign
	l.accept("+-")
	digits := "0123456789"
	l.acceptRun(digits)
	if l.accept(".") {
		l.acceptRun(digits)
	}
	l.emit(tNumber)
	return lexExpression
}
