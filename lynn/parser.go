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
func NewParser(lexer *Lexer) *Parser { return &Parser { lexer } }
// Reads tokens from stream and produces an abstract syntax tree representing the grammar.
func (p *Parser) Parse() *GrammarNode {
	rules, tokens, fragments := make([]*RuleNode, 0, 100), make([]*TokenNode, 0, 100), make([]*FragmentNode, 0, 100)
	for !p.lexer.Match(EOF) {
		switch rule := p.parseRule(); rule.(type) {
		case *RuleNode: rules = append(rules, rule.(*RuleNode))
		case *TokenNode: tokens = append(tokens, rule.(*TokenNode))
		case *FragmentNode: fragments = append(fragments, rule.(*FragmentNode))
		case nil: p.synchronize()
		}
	}
	return &GrammarNode { rules, tokens, fragments }
}

func (p *Parser) parseRule() AST {
	switch token := p.lexer.Token; {
	case p.lexer.Match(IDENTIFIER):
		// Parse following pattern: [identifier] : <expr> ;
		if !p.lexer.Expect(COLON) { return nil }
		expression := p.parseExpressionDefault()
		if !p.lexer.Expect(SEMI) { return nil }
		return &RuleNode { token.Value, expression }
	case p.lexer.Match(TOKEN):
		// Parse following pattern: token [identifier] : <expr> (-> skip) ;
		token := p.lexer.Token
		if !p.lexer.Expect(IDENTIFIER) || !p.lexer.Expect(COLON) { return nil }
		expression := p.parseExpressionDefault()
		// Test for presence of skip flag
		var skip = false
		if p.lexer.Match(ARROW) {
			if !p.lexer.Expect(SKIP) { return nil }
			skip = true
		}
		if !p.lexer.Expect(SEMI) { return nil }
		return &TokenNode { token.Value, expression, skip }
	case p.lexer.Match(FRAGMENT):
		// Parse following pattern: fragment [identifier] : <expr> ;
		token := p.lexer.Token
		if !p.lexer.Expect(IDENTIFIER) || !p.lexer.Expect(COLON) { return nil }
		expression := p.parseExpressionDefault()
		if !p.lexer.Expect(SEMI) { return nil }
		return &FragmentNode { token.Value, expression }
	default: return nil
	}
}

// Represents operation precedence as an enumerated integer.
type Precedence int
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
type StringNode struct { Value string }
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
	value := strings.ReplaceAll(n.Value, "\t", "\\t")
	value = strings.ReplaceAll(value, "\n", "\\n")
	value = strings.ReplaceAll(value, "\r", "\\r")
	value = strings.ReplaceAll(value, "\b", "\\b")
	value = strings.ReplaceAll(value, "\f", "\\f")
	value = strings.ReplaceAll(value, "\x00", "\\0")
	return fmt.Sprintf("\"%s\"", value)
}

func (n ClassNode) String() string {
	var builder strings.Builder
	for _, char := range n.Chars {
		switch char {
		case '\t': builder.WriteString("\\t") 
		case '\n': builder.WriteString("\\n") 
		case '\r': builder.WriteString("\\r") 
		case '\b': builder.WriteString("\\b") 
		case '\f': builder.WriteString("\\f") 
		case 0: builder.WriteString("\\0") 
		default: builder.WriteRune(char)
		}
	}
	if n.Negated {
		return fmt.Sprintf("[^%s]", builder.String())
	} else {
		return fmt.Sprintf("[%s]", builder.String())
	}
}
