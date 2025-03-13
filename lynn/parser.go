package lynn

import "fmt"

type Parser struct {
	lexer *Lexer
}

// Returns new parser struct.
func NewParser(lexer *Lexer) *Parser {
	return &Parser { lexer }
}

func (p *Parser) Parse() {
	// p.lexer.PrintTokenStream()
	fmt.Println(p.parseLiteral())
}

func (p *Parser) parseLiteral() AST {
	token := p.lexer.Token
	switch {
	case p.lexer.Match(DOT): return &AnyNode { }
	case p.lexer.Match(IDENTIFIER): return &IdentifierNode { token.Value }
	case p.lexer.Match(STRING):
	case p.lexer.Match(CLASS):
	}

	return nil
}

type AST any

type AnyNode struct { }
type IdentifierNode struct { Name string }
type StringNode struct { Value string }
type ClassNode struct {
	Chars   []string
	Negated bool
}
