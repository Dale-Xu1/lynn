package lynn

import (
	"fmt"
	"regexp"
	"strings"
)

type Parser struct {
	lexer *Lexer
}

// Returns new parser struct.
func NewParser(lexer *Lexer) *Parser {
	return &Parser { lexer }
}

func (p *Parser) Parse() {
	// p.lexer.PrintTokenStream()
	fmt.Println(p.parseExpression())
}

func (p *Parser) parseExpression() AST {
	switch {
	default: return p.parsePrimary()
	}
}

func (p *Parser) parsePrimary() AST {
	switch token := p.lexer.Token; {
	case p.lexer.Match(L_PAREN):
		expr := p.parseExpression()
		if !p.lexer.Expect(R_PAREN) { return nil }
		return expr

	case p.lexer.Match(DOT): return &AnyNode { }
	case p.lexer.Match(IDENTIFIER): return &IdentifierNode { token.Value }
	case p.lexer.Match(STRING):
		value := token.Value[1:len(token.Value) - 1]
		value = strings.ReplaceAll(value, "\\\\", "\\")
		value = strings.ReplaceAll(value, "\\t", "\t")
		value = strings.ReplaceAll(value, "\\n", "\n")
		value = strings.ReplaceAll(value, "\\r", "\r")
		value = strings.ReplaceAll(value, "\\b", "\b")
		value = strings.ReplaceAll(value, "\\f", "\f")
		value = strings.ReplaceAll(value, "\\0", "\x00")
		value = strings.ReplaceAll(value, "\\\"", "\"")

		re := regexp.MustCompile(`\\(.)`)
		value = re.ReplaceAllString(value, "$1")
		return &StringNode { value }
	case p.lexer.Match(CLASS):
		value := token.Value[1:len(token.Value) - 1]
		negated := value[0] == '^'
		if negated { value = value[1:] }

		value = strings.ReplaceAll(value, "\\\\", "\\")
		value = strings.ReplaceAll(value, "\\t", "\t")
		value = strings.ReplaceAll(value, "\\n", "\n")
		value = strings.ReplaceAll(value, "\\r", "\r")
		value = strings.ReplaceAll(value, "\\b", "\b")
		value = strings.ReplaceAll(value, "\\f", "\f")
		value = strings.ReplaceAll(value, "\\0", "\x00")
		value = strings.ReplaceAll(value, "\\]", "]")

		re := regexp.MustCompile(`\\(.)`)
		value = re.ReplaceAllString(value, "$1")
		return &ClassNode { []rune(value), negated }
	default: return nil
	}
}

type AST any

type AnyNode struct { }
type IdentifierNode struct { Name string }
type StringNode struct { Value string }
type ClassNode struct {
	Chars   []rune
	Negated bool
}
