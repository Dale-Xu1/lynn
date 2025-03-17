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
	if left == nil { return nil }
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
		case p.lexer.Match(BAR):
			right := p.parseExpression(next + 1)
			if right == nil { return nil }
			left = &UnionNode { left, right }
		case p.lexer.Match(QUESTION): left = &OptionNode { left }
		case p.lexer.Match(STAR): left = &RepeatNode { left }
		case p.lexer.Match(PLUS): left = &RepeatOneNode { left }
		
		// This relies on the current token being a valid beginning of an expression since combinations have no delimiter
		default:
			right := p.parseExpression(next + 1)
			if right == nil { return nil }
			left = &CombinationNode { left, right }
		}
	}
	return left
}

func (p *Parser) parsePrimary() AST {
	switch token := p.lexer.Token; {
	// Parentheses enclose a group, precedence is reset for inner expression
	case p.lexer.Match(L_PAREN):
		expr := p.parseExpression(UNION)
		if expr == nil || !p.lexer.Expect(R_PAREN) { return nil }
		return expr

	case p.lexer.Match(DOT): return &AnyNode { }
	case p.lexer.Match(IDENTIFIER): return &IdentifierNode { token.Value }
	case p.lexer.Match(STRING):
		value := token.Value[1:len(token.Value) - 1] // Remove quotation marks
		// Replace escape sequences with special characters
		value = strings.ReplaceAll(value, "\\t", "\t")
		value = strings.ReplaceAll(value, "\\n", "\n")
		value = strings.ReplaceAll(value, "\\r", "\r")
		value = strings.ReplaceAll(value, "\\b", "\b")
		value = strings.ReplaceAll(value, "\\f", "\f")
		value = strings.ReplaceAll(value, "\\0", "\x00")

		// Remove backslash for non-special characters
		re := regexp.MustCompile(`\\(.)`)
		value = re.ReplaceAllString(value, "$1")
		return &StringNode { value }
	case p.lexer.Match(CLASS):
		value := token.Value[1:len(token.Value) - 1] // Remove brackets
		// If caret occurs, flag class as negated and remove caret
		negated := len(value) > 0 && value[0] == '^'
		if negated { value = value[1:] }
		return &ClassNode { expandClass([]rune(value)), negated }
	default: return nil // Invalid expression
	}
}

func expandClass(chars []rune) []rune {
	expanded := make([]rune, 0, len(chars))
	for i := 0; i < len(chars); i++ {
		char := chars[i]
		switch {
		case char == '\\':
			i++
			// Replace escape sequences with special characters
			switch chars[i] {
			case 't': expanded = append(expanded, '\t')
			case 'n': expanded = append(expanded, '\n')
			case 'r': expanded = append(expanded, '\r')
			case 'b': expanded = append(expanded, '\b')
			case 'f': expanded = append(expanded, '\f')
			case '0': expanded = append(expanded, 0)
			default: expanded = append(expanded, chars[i]) // Backslash is ignored for non-special characters
			}
		case char == '-' && i > 0 && i < len(chars) - 1: // Hyphen for range cannot be first or last character in class
			if chars[i - 1] <= chars[i + 1] {
				// Expand range to all characters between endpoints
				for c := chars[i - 1] + 1; c <= chars[i + 1]; c++ {
					expanded = append(expanded, c)
				}
			} else {
				// Raise error and ignore range if endpoint order is reversed
				fmt.Printf("Syntax error: Invalid range from %c to %c\n", chars[i - 1], chars[i + 1])
				expanded = expanded[:len(expanded) - 1]
			}
			i++
		default: expanded = append(expanded, chars[i])
		}
	}
	return expanded
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
