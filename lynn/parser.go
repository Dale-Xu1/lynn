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

// Reads tokens from stream and produces an abstract syntax tree representing the grammar.
func (p *Parser) Parse() {
	// p.lexer.PrintTokenStream()
	expr := p.parseExpression(UNION)
	fmt.Printf("%v", expr)
}

// Represents operation precedence as an enumerated integer.
type Precedence int
const (
	UNION Precedence = iota
	COMBINATION
	QUANTIFIER
)

func (p *Parser) parseExpression(precedence Precedence) AST {
	left := p.parsePrimary()
	main: for {
		// Determine precedence of current token in stream
		var next Precedence
		switch p.lexer.Token.Type {
		case BAR: next = UNION
		case L_PAREN, IDENTIFIER, STRING, CLASS, DOT: // All possible tokens for the beginning of a regular expression
			next = COMBINATION
		case PLUS, STAR, QUESTION: next = QUANTIFIER
		default: break main
		}
		if next < precedence { break main } // Stop parsing if precedence is too low

		// Continue parsing based on type of expression
		switch {
		case p.lexer.Match(BAR): left = &UnionNode { left, p.parseExpression(next + 1) }
		case p.lexer.Match(QUESTION): left = &OptionNode { left }
		case p.lexer.Match(STAR): left = &RepeatNode { left }
		case p.lexer.Match(PLUS): left = &RepeatOneNode { left }
		
		// This relies on the current token being a valid beginning of an expression since combinations have no delimiter
		default: left = &CombinationNode { left, p.parseExpression(next + 1) }
		}
	}
	return left
}

func (p *Parser) parsePrimary() AST {
	switch token := p.lexer.Token; {
	case p.lexer.Match(L_PAREN):
		expr := p.parseExpression(UNION)
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
		return &ClassNode { expandClass([]rune(value)), negated }
	default: return nil
	}
}

func expandClass(chars []rune) []rune {
	// TODO: Implementation
	return chars
}

// Interface for nodes of abstract syntax tree.
type AST any
// Node representing the entire grammar as a list of rules.
type GrammarNode struct {
	Rules []AST
}

// Node representing a grammar rule. Specifies the rule's identifier and regular expression.
type RuleNode struct {
	Identifier IdentifierNode
	Expression AST
}

// Node representing a token rule. Specifies the token's identifier and regular expression.
type TokenNode struct {
	Identifier IdentifierNode
	Expression AST
	Skip       bool
}

// Node representing a rule fragment. Specifies the fragment's identifier and regular expression.
// Fragments are used to repeat regular expressions in token rules.
type FragmentNode struct {
	Identifier IdentifierNode
	Expression AST
}

// Node representing an option quantifier. Allows zero or one occurrence of the given regular expression.
type OptionNode struct { Expression AST }
// Node representing an repeat quantifier. Allows zero or more occurrences of the given regular expression.
type RepeatNode struct { Expression AST }
// Node representing an repeat one or more quantifier. Allows one or more occurrences of the given regular expression.
type RepeatOneNode struct { Expression AST }

// Node representing a combination operation. Requires that one expression immediately follow the preceding expression.
type CombinationNode struct {
	A, B AST
}

// Node representing a union operation. Allows either one expression or the other to occur.
type UnionNode struct {
	A, B AST
}

// Node representing an any literal. Matches any character except newlines or the end of file.
type AnyNode struct { }
// Node representing an identifier literal.
type IdentifierNode struct { Name string }
// Node representing a string literal.
type StringNode struct { Value string }
// Node representing a class literal. Negated classes do not match the end of file but do match newlines.
type ClassNode struct {
	Chars   []rune
	Negated bool
}

func (n OptionNode) String() string { return fmt.Sprintf("(%v)?", n.Expression) }
func (n RepeatNode) String() string { return fmt.Sprintf("(%v)*", n.Expression) }
func (n RepeatOneNode) String() string { return fmt.Sprintf("(%v)+", n.Expression) }

func (n CombinationNode) String() string { return fmt.Sprintf("(%v %v)", n.A, n.B) }
func (n UnionNode) String() string { return fmt.Sprintf("(%v | %v)", n.A, n.B) }

func (n AnyNode) String() string { return "any" }
func (n IdentifierNode) String() string { return fmt.Sprintf("id:%s", n.Name) }
func (n StringNode) String() string { return fmt.Sprintf("\"%s\"", n.Value) }
func (n ClassNode) String() string {
	var builder strings.Builder
	for _, char := range n.Chars { builder.WriteRune(char) }
	if n.Negated {
		return fmt.Sprintf("[^%s]", builder.String())
	} else {
		return fmt.Sprintf("[%s]", builder.String())
	}
}
