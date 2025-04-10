package lynn

import (
	"cmp"
	"fmt"
	"sort"
)

// Parser struct. Converts token stream to abstract syntax tree (AST).
type Parser struct { lexer *Lexer }

// Returns new parser struct.
func NewParser(lexer *Lexer) *Parser { return &Parser { lexer } }
// Reads tokens from stream and produces an abstract syntax tree representing the grammar.
func (p *Parser) Parse() *GrammarNode {
    rules, tokens, fragments := make([]*RuleNode, 0), make([]*TokenNode, 0), make([]*FragmentNode, 0)
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
        if expression == nil || !p.lexer.Match(SEMI) { return nil }
        return &RuleNode { &IdentifierNode { token.Value, token.Location }, expression }
    case p.lexer.Match(TOKEN):
        // Parse following pattern: token [identifier] : <expr> (-> skip) ;
        token := p.lexer.Token
        if !p.lexer.Match(IDENTIFIER) || !p.lexer.Match(COLON) { return nil }
        expression := p.parseExpressionDefault(); if expression == nil { return nil }
        // Test for presence of skip flag
        var skip = false
        if p.lexer.Match(ARROW) {
            if !p.lexer.Match(SKIP) { return nil }
            skip = true
        }
        if !p.lexer.Match(SEMI) { return nil }
        return &TokenNode { &IdentifierNode { token.Value, token.Location }, expression, skip }
    case p.lexer.Match(FRAGMENT):
        // Parse following pattern: fragment [identifier] : <expr> ;
        token := p.lexer.Token
        if !p.lexer.Match(IDENTIFIER) || !p.lexer.Match(COLON) { return nil }
        expression := p.parseExpressionDefault()
        if expression == nil || !p.lexer.Match(SEMI) { return nil }
        return &FragmentNode { &IdentifierNode { token.Value, token.Location }, expression }
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
            left = &ConcatenationNode { left, right }
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

    case p.lexer.Match(DOT): return &ClassNode { negateRanges(expandClass([]rune { '\n', '\r' }, token.Location)), token.Location }
    case p.lexer.Match(IDENTIFIER): return &IdentifierNode { token.Value, token.Location }
    case p.lexer.Match(STRING):
        value := token.Value[1:len(token.Value) - 1] // Remove quotation marks
        return &StringNode { reduceString([]rune(value)), token.Location }
    case p.lexer.Match(CLASS):
        value := token.Value[1:len(token.Value) - 1] // Remove brackets
        // If caret occurs, flag class as negated and remove caret
        negated := len(value) > 0 && value[0] == '^'
        var expanded []Range
        if !negated {
            expanded = expandClass(reduceString([]rune(value)), token.Location)
        } else {
            expanded = negateRanges(expandClass(reduceString([]rune(value[1:])), token.Location))
        }
        return &ClassNode { expanded, token.Location }
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
            case '0': reduced = append(reduced, 0)
            default: reduced = append(reduced, chars[i]) // Backslash is ignored for non-special characters
            }
        default: reduced = append(reduced, chars[i])
        }
    }
    return reduced
}

func expandClass(chars []rune, location Location) []Range {
    // Convert characters and hyphen notation to range structs
    expanded := make([]Range, 0, len(chars))
    for i := 0; i < len(chars); i++ {
        char := chars[i]
        switch {
        case char == '-' && i > 0 && i < len(chars) - 1: // Hyphen for range cannot be first or last character in class
            expanded = expanded[:len(expanded) - 1]
            if chars[i - 1] <= chars[i + 1] {
                expanded = append(expanded, Range { chars[i - 1], chars[i + 1] })
            } else {
                // Raise error and ignore range if endpoint order is reversed
                fmt.Printf("Syntax error: Invalid range from \"%s\" to \"%s\" - %d:%d\n",
                    formatChar(chars[i - 1]), formatChar(chars[i + 1]), location.Line, location.Col)
            }
            i++
        default: expanded = append(expanded, Range { char, char })
        }
    }
    if len(expanded) <= 1 { return expanded }
    return mergeRanges(expanded)
}

func mergeRanges(ranges []Range) []Range {
    // Sort ranges based on minimum
    sort.Slice(ranges, func (i, j int) bool { return ranges[i].Min < ranges[j].Min })
    // Scan ranges and merge if overlap is found
    merged := make([]Range, 1, len(ranges))
    merged[0] = ranges[0]
    for _, r := range ranges[1:] {
        last := &merged[len(merged) - 1]
        if r.Min <= last.Max + 1 {
            last.Max = max(last.Max, r.Max)
        } else {
            merged = append(merged, r)
        }
    }
    return merged
}

func negateRanges(ranges []Range) []Range {
    // Assumes ranges are already sorted and merged
    negated := make([]Range, 0, len(ranges) + 1)
    const MAX rune = 0x10ffff // Maximum unicode character
    var start rune = 1
    for _, r := range ranges {
        if r.Min > start { negated = append(negated, Range { start, r.Min - 1 }) }
        start = r.Max + 1
    }
    if start <= MAX { negated = append(negated, Range { start, MAX }) }
    return negated
}

func max[T cmp.Ordered](a, b T) T {
    if a > b { return a }
    return b
}

func (p *Parser) unexpected() {
    // Raise error message describing the current token in the stream as unexpected
    token := p.lexer.Token
    fmt.Printf("Syntax error: Unexpected token %q - %d:%d\n", token.Value, token.Location.Line, token.Location.Col)
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

func formatChar(char rune) string {
    str := fmt.Sprintf("%q", string(char))
    return str[1:len(str) - 1]
}
