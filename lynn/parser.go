package lynn

import (
	"fmt"
	"strings"
)

type Parser struct {
	lexer *Lexer
}

// Returns new parser struct.
func NewParser(lexer *Lexer) *Parser { return &Parser { lexer } }
// Reads tokens from stream and produces an abstract syntax tree representing the grammar.
func (p *Parser) Parse() *GrammarNode {
	rules, tokens, fragments := make([]*RuleNode, 0, 100), make([]*TokenNode, 0, 100), make([]*FragmentNode, 0, 100)
	for !p.lexer.Match(EOF) {
		switch rule := p.parseRule().(type) {
		case *RuleNode: rules = append(rules, rule)
		case *TokenNode: tokens = append(tokens, rule)
		case *FragmentNode: fragments = append(fragments, rule)
		case nil:
			p.unexpected()
			p.synchronize()
		}
	}
	return &GrammarNode { rules, tokens, fragments }
}

func (p *Parser) parseRule() AST {
	switch token := p.lexer.Token; {
	case p.lexer.Match(IDENTIFIER):
		// Parse following pattern: [identifier] : <expr> ;
		if !p.lexer.Match(COLON) { return nil }
		expression := p.parseExpressionDefault()
		if !p.lexer.Match(SEMI) { return nil }
		return &RuleNode { token.Value, expression }
	case p.lexer.Match(TOKEN):
		// Parse following pattern: token [identifier] : <expr> (-> skip) ;
		token := p.lexer.Token
		if !p.lexer.Match(IDENTIFIER) || !p.lexer.Match(COLON) { return nil }
		expression := p.parseExpressionDefault()
		// Test for presence of skip flag
		var skip = false
		if p.lexer.Match(ARROW) {
			if !p.lexer.Match(SKIP) { return nil }
			skip = true
		}
		if !p.lexer.Match(SEMI) { return nil }
		return &TokenNode { token.Value, expression, skip }
	case p.lexer.Match(FRAGMENT):
		// Parse following pattern: fragment [identifier] : <expr> ;
		token := p.lexer.Token
		if !p.lexer.Match(IDENTIFIER) || !p.lexer.Match(COLON) { return nil }
		expression := p.parseExpressionDefault()
		if !p.lexer.Match(SEMI) { return nil }
		return &FragmentNode { token.Value, expression }
	default: return nil
	}
}

// Represents operation precedence as an enumerated integer.
type Precedence uint
const (
	UNION Precedence = iota
	COMBINATION
	QUANTIFIER
)

func (p *Parser) parseExpressionDefault() AST { return p.parseExpression(UNION) }
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
		expr := p.parseExpressionDefault()
		if expr == nil || !p.lexer.Match(R_PAREN) { return nil }
		return expr

	case p.lexer.Match(DOT): return &AnyNode { }
	case p.lexer.Match(IDENTIFIER): return &IdentifierNode { token.Value }
	case p.lexer.Match(STRING):
		value := token.Value[1:len(token.Value) - 1] // Remove quotation marks
		return &StringNode { reduceString([]rune(value)) }
	case p.lexer.Match(CLASS):
		value := token.Value[1:len(token.Value) - 1] // Remove brackets
		// If caret occurs, flag class as negated and remove caret
		negated := len(value) > 0 && value[0] == '^'
		if negated { value = value[1:] }
		return &ClassNode { expandClass(reduceString([]rune(value)), token.Location), negated }
	default: return nil // Invalid expression
	}
}

func reduceString(chars []rune) []rune {
	reduced := make([]rune, 0, len(chars))
	for i := 0; i < len(chars); i++ {
		char := chars[i]
		switch {
		case char == '\\':
			i++
			// Replace escape sequences with special characters
			switch chars[i] {
			case 't': reduced = append(reduced, '\t')
			case 'n': reduced = append(reduced, '\n')
			case 'r': reduced = append(reduced, '\r')
			case 'b': reduced = append(reduced, '\b')
			case 'f': reduced = append(reduced, '\f')
			case '0': reduced = append(reduced, 0)
			default: reduced = append(reduced, chars[i]) // Backslash is ignored for non-special characters
			}
		default: reduced = append(reduced, chars[i])
		}
	}
	return reduced
}

func expandClass(chars []rune, location Location) []rune {
	expanded := make([]rune, 0, len(chars))
	for i := 0; i < len(chars); i++ {
		char := chars[i]
		switch {
		case char == '-' && i > 0 && i < len(chars) - 1: // Hyphen for range cannot be first or last character in class
			if chars[i - 1] <= chars[i + 1] {
				// Expand range to all characters between endpoints
				for c := chars[i - 1] + 1; c <= chars[i + 1]; c++ {
					expanded = append(expanded, c)
				}
			} else {
				// Raise error and ignore range if endpoint order is reversed
				fmt.Printf("Syntax error: Invalid range from \"%s\" to \"%s\" - %d:%d\n",
					formatRune(chars[i - 1]), formatRune(chars[i + 1]), location.Line, location.Col)
				expanded = expanded[:len(expanded) - 1]
			}
			i++
		default: expanded = append(expanded, chars[i])
		}
	}
	return expanded
}

func (p *Parser) unexpected() {
	// Raise error message describing the current token in the stream as unexpected
	token := p.lexer.Token
    fmt.Printf("Syntax error: Unexpected token \"%s\" - %d:%d\n", token.Value, token.Location.Line, token.Location.Col)
}

func (p *Parser) synchronize() {
	// Skip tokens until a semicolon or a keyword denoting the beginning of a rule is found
	main: for {
		switch p.lexer.Token.Type {
		case SEMI:
			p.lexer.Next()
			fallthrough
		case TOKEN, FRAGMENT:
			break main
		}
		p.lexer.Next()
	}
}

// Interface for nodes of abstract syntax tree.
type AST fmt.Stringer
// Node representing the entire grammar as a list of rules.
type GrammarNode struct {
	Rules     []*RuleNode
	Tokens    []*TokenNode
	Fragments []*FragmentNode
}

// Node representing a grammar rule. Specifies the rule's identifier and regular expression.
type RuleNode struct {
	Identifier string
	Expression AST
}

// Node representing a token rule. Specifies the token's identifier and regular expression.
type TokenNode struct {
	Identifier string
	Expression AST
	Skip       bool
}

// Node representing a rule fragment. Specifies the fragment's identifier and regular expression.
// Fragments are used to repeat regular expressions in token rules.
type FragmentNode struct {
	Identifier string
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
type StringNode struct { Chars []rune }
// Node representing a class literal. Negated classes do not match the end of file but do match newlines.
type ClassNode struct {
	Chars   []rune
	Negated bool
}

func (n GrammarNode) String() string {
	lines := make([]string, 0, len(n.Rules) + len(n.Tokens) + len(n.Fragments))
	for _, rule := range n.Rules { lines = append(lines, rule.String()) }
	for _, token := range n.Tokens { lines = append(lines, token.String()) }
	for _, fragment := range n.Fragments { lines = append(lines, fragment.String()) }
	return strings.Join(lines, "\n")
}

func (n RuleNode) String() string {
	return fmt.Sprintf("rule %s : %v", n.Identifier, n.Expression)
}

func (n TokenNode) String() string {
	if n.Skip {
		return fmt.Sprintf("token skip %s : %v", n.Identifier, n.Expression)
	}
	return fmt.Sprintf("token %s : %v", n.Identifier, n.Expression)
}

func (n FragmentNode) String() string {
	return fmt.Sprintf("fragment %s : %v", n.Identifier, n.Expression)
}

func (n OptionNode) String() string { return fmt.Sprintf("(%v)?", n.Expression) }
func (n RepeatNode) String() string { return fmt.Sprintf("(%v)*", n.Expression) }
func (n RepeatOneNode) String() string { return fmt.Sprintf("(%v)+", n.Expression) }

func (n CombinationNode) String() string { return fmt.Sprintf("(%v %v)", n.A, n.B) }
func (n UnionNode) String() string { return fmt.Sprintf("(%v | %v)", n.A, n.B) }

func (n AnyNode) String() string { return "any" }
func (n IdentifierNode) String() string { return fmt.Sprintf("id:%s", n.Name) }
func (n StringNode) String() string {
	return fmt.Sprintf("\"%s\"", formatChars(n.Chars))
}

func (n ClassNode) String() string {
	if n.Negated {
		return fmt.Sprintf("[^%s]", formatChars(n.Chars))
	} else {
		return fmt.Sprintf("[%s]", formatChars(n.Chars))
	}
}

func formatChars(chars []rune) string {
	var builder strings.Builder
	for _, char := range chars {
		builder.WriteString(formatRune(char))
	}
	return builder.String()
}

func formatRune(char rune) string {
    switch char {
    case '\t': return "\\t"
    case '\n': return "\\n"
    case '\r': return "\\r"
    case '\b': return "\\b"
    case '\f': return "\\f"
    case 0:    return "\\0"
	default:   return string(char)
    }
}
