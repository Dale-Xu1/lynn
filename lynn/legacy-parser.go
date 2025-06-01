package lynn

import "fmt"

// LegacyParser struct. Converts token stream to abstract syntax tree (AST).
type LegacyParser struct { lexer *Lexer }

// Tests if the type of the current token in the stream matches the provided type. If the types match, the next token is emitted.
func (l *Lexer) Match(token TokenType) bool {
    if l.Token.Type == token {
        l.Next()
        return true
    }
    return false
}

// Returns new parser struct.
func NewLegacyParser(lexer *Lexer) *LegacyParser { return &LegacyParser { lexer } }
// Reads tokens from stream and produces an abstract syntax tree representing the grammar.
func (p *LegacyParser) Parse() *GrammarNode {
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

func (p *LegacyParser) parseRule() AST {
    switch {
    case p.lexer.Match(RULE):
        // Parse following pattern: rule [identifier] : <expr> ;
        token := p.lexer.Token
        if !p.lexer.Match(IDENTIFIER) || !p.lexer.Match(COLON) { return nil }
        expression := p.parseExpressionDefault()
        if expression == nil || !p.lexer.Match(SEMI) { return nil }
        return &RuleNode { &IdentifierNode { token.Value, token.Start }, expression }
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
        return &TokenNode { &IdentifierNode { token.Value, token.Start }, expression, skip }
    case p.lexer.Match(FRAGMENT):
        // Parse following pattern: fragment [identifier] : <expr> ;
        token := p.lexer.Token
        if !p.lexer.Match(IDENTIFIER) || !p.lexer.Match(COLON) { return nil }
        expression := p.parseExpressionDefault()
        if expression == nil || !p.lexer.Match(SEMI) { return nil }
        return &FragmentNode { &IdentifierNode { token.Value, token.Start }, expression }
    default: return nil
    }
}

// Represents operation precedence as an enumerated integer.
type Precedence uint
const (UNION Precedence = iota; LABEL; CONCAT; ALIAS; QUANTIFIER)

func (p *LegacyParser) parseExpressionDefault() AST { return p.parseExpression(UNION) }
func (p *LegacyParser) parseExpression(precedence Precedence) AST {
    left := p.parsePrimary()
    if left == nil { return nil }
    main: for {
        // Determine precedence of current token in stream
        var next Precedence
        switch p.lexer.Token.Type {
        case BAR:  next = UNION
        case HASH: next = LABEL
        case L_PAREN, IDENTIFIER, STRING, CLASS, DOT: // All possible tokens for the beginning of a regular expression
            next = CONCAT
        case EQUAL: next = ALIAS
        case PLUS, STAR, QUESTION: next = QUANTIFIER
        default: break main
        }
        if next < precedence { break main } // Stop parsing if precedence is too low

        // Continue parsing based on type of expression
        t := p.lexer.Token
        switch {
        case p.lexer.Match(BAR):
            right := p.parseExpression(next + 1)
            if right == nil { return nil }
            left = &UnionNode { left, right }
        case p.lexer.Match(HASH):
            token := p.lexer.Token
            if !p.lexer.Match(IDENTIFIER) { return nil }
            assoc := NO_ASSOC
            if p.lexer.Match(LEFT) {
                assoc = LEFT_ASSOC
            } else if p.lexer.Match(RIGHT) {
                assoc = RIGHT_ASSOC
            }
            left = &LabelNode { left, &IdentifierNode { token.Value, token.Start }, assoc, t.Start }
        case p.lexer.Match(EQUAL):
            id, ok := left.(*IdentifierNode)
            if !ok { return nil }
            left = &AliasNode { id, p.parseExpression(next), t.Start }
        case p.lexer.Match(QUESTION): left = &OptionNode { left }
        case p.lexer.Match(STAR): left = &RepeatNode { left }
        case p.lexer.Match(PLUS): left = &RepeatOneNode { left }
    
        // This relies on the current token being a valid beginning of an expression since combinations have no delimiter
        default:
            right := p.parseExpression(next + 1)
            if right == nil { return nil }
            left = &ConcatNode { left, right }
        }
    }
    return left
}

func (p *LegacyParser) parsePrimary() AST {
    switch token := p.lexer.Token; {
    // Parentheses enclose a group, precedence is reset for inner expression
    case p.lexer.Match(L_PAREN):
        expr := p.parseExpressionDefault()
        if expr == nil || !p.lexer.Match(R_PAREN) { return nil }
        return expr

    case p.lexer.Match(DOT): return &ClassNode { negateRanges(expandClass([]rune { '\n', '\r' }, token.Start)), token.Start }
    case p.lexer.Match(IDENTIFIER): return &IdentifierNode { token.Value, token.Start }
    case p.lexer.Match(STRING):
        value := token.Value[1:len(token.Value) - 1] // Remove quotation marks
        return &StringNode { reduceString([]rune(value)), token.Start }
    case p.lexer.Match(CLASS):
        value := token.Value[1:len(token.Value) - 1] // Remove brackets
        // If caret occurs, flag class as negated and remove caret
        negated := len(value) > 0 && value[0] == '^'
        var expanded []Range
        if !negated {
            expanded = expandClass(reduceString([]rune(value)), token.Start)
        } else {
            expanded = negateRanges(expandClass(reduceString([]rune(value[1:])), token.Start))
        }
        return &ClassNode { expanded, token.Start }
    default: return nil // Invalid expression
    }
}

func (p *LegacyParser) unexpected() {
    // Raise error message describing the current token in the stream as unexpected
    token := p.lexer.Token
    fmt.Printf("Syntax error: Unexpected token %q - %d:%d\n", token.Value, token.Start.Line, token.Start.Col)
}

func (p *LegacyParser) synchronize() {
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
