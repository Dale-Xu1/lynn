package lynn

import (
	"cmp"
	"fmt"
	"sort"
	"strings"
)

// Interface for nodes of abstract syntax tree.
type AST fmt.Stringer
// Node representing the entire grammar as a list of rules.
type GrammarNode struct {
    Rules      []*RuleNode
    Precedence []*PrecedenceNode
    Tokens     []*TokenNode
    Fragments  []*FragmentNode
}

// Node representing a grammar rule. Specifies the rule's identifier and regular expression.
type RuleNode struct {
    Identifier *IdentifierNode
    Expression AST
    Start, End Location
}

// Node representing a precedence statement. Specifies the a precedence level and its associativity.
type PrecedenceNode struct {
    Identifier    *IdentifierNode
    Associativity AssociativityType
    Start, End    Location
}

// Node representing a token rule. Specifies the token's identifier and regular expression.
type TokenNode struct {
    Identifier *IdentifierNode
    Expression AST
    Skip       bool
    Start, End Location
}

// Node representing a rule fragment. Specifies the fragment's identifier and regular expression.
// Fragments are used to repeat regular expressions in token rules.
type FragmentNode struct {
    Identifier *IdentifierNode
    Expression AST
    Start, End Location
}

// Node representing an option quantifier. Allows zero or one occurrence of the given regular expression.
type OptionNode struct { Expression AST; Start, End Location }
// Node representing an repeat quantifier. Allows zero or more occurrences of the given regular expression.
type RepeatNode struct { Expression AST; Start, End Location }
// Node representing an repeat one or more quantifier. Allows one or more occurrences of the given regular expression.
type RepeatOneNode struct { Expression AST; Start, End Location }

// Associativity type enum. Either NO_ASSOC, LEFT_ASSOC, or RIGHT_ASSOC.
type AssociativityType uint
const (NO_ASSOC AssociativityType = iota; LEFT_ASSOC; RIGHT_ASSOC)
// Node representing a rule case label. Specifies the callback identifier and associativity for the disambiguation process.
type LabelNode struct {
    Expression AST
    Identifier *IdentifierNode
    Precedence *IdentifierNode
    Start, End Location
}

// Node representing an alias. Specifies the alias identifier and the corresponding expression.
type AliasNode struct {
    Identifier *IdentifierNode
    Expression AST
    Start, End Location
}

// Node representing a concatenation operation. Requires that one expression immediately follow the preceding expression.
type ConcatNode struct {
    A, B       AST
    Start, End Location
}

// Node representing a union operation. Allows either one expression or the other to occur.
type UnionNode struct {
    A, B       AST
    Start, End Location
}

// Node representing an identifier literal.
type IdentifierNode struct { Name string; Start, End Location }
// Node representing a string literal.
type StringNode struct { Chars []rune; Start, End Location }
// Node representing a class literal.
type ClassNode struct { Ranges []Range; Start, End Location }
// Node representing an error literal.
type ErrorNode struct { Start, End Location }

// Parse tree visitor struct. Converts parse tree to abstract syntax tree (AST).
type ParseTreeVisitor struct { }
// Returns new parse tree visitor struct.
func NewParseTreeVisitor() ParseTreeVisitor { return ParseTreeVisitor { } }

func (v ParseTreeVisitor) VisitGrammar(node *ParseTreeNode) AST {
    rules, precedence, tokens, fragments := make([]*RuleNode, 0), make([]*PrecedenceNode, 0), make([]*TokenNode, 0), make([]*FragmentNode, 0)
    for _, node := range node.Stmt().(*ParseTreeNode).Children {
        switch rule := VisitNode(v, node.(*ParseTreeNode)).(type) {
        case *RuleNode: rules = append(rules, rule)
        case *PrecedenceNode: precedence = append(precedence, rule)
        case *TokenNode: tokens = append(tokens, rule)
        case *FragmentNode: fragments = append(fragments, rule)
        }
    }
    if len(rules) == 0 || len(tokens) == 0 {
        location := node.Start
        Error(fmt.Sprintf("Grammar definition must contain at least one rule and token - %d:%d", location.Line, location.Col))
    }
    return &GrammarNode { rules, precedence, tokens, fragments }
}

func (v ParseTreeVisitor) VisitStmt(node *ParseTreeNode) AST { panic("Invalid statement") }
func (v ParseTreeVisitor) VisitRuleStmt(node *ParseTreeNode) AST {
    id := node.IDENTIFIER().(Token)
    identifier := &IdentifierNode { id.Value, id.Start, id.End }
    return &RuleNode { identifier, VisitNode(v, node.Expr()), node.Start, node.End }
}

func (v ParseTreeVisitor) VisitPrecedenceStmt(node *ParseTreeNode) AST {
    id := node.IDENTIFIER().(Token)
    identifier := &IdentifierNode { id.Value, id.Start, id.End }
    var assoc AssociativityType
    if v, ok := node.V().(*ParseTreeNode); ok {
        switch v.A().(Token).Type {
        case LEFT:  assoc = LEFT_ASSOC
        case RIGHT: assoc = RIGHT_ASSOC
        default: panic("Invalid associativity type")
        }
    } else { assoc = NO_ASSOC }
    return &PrecedenceNode { identifier, assoc, node.Start, node.End }
}

func (v ParseTreeVisitor) VisitTokenStmt(node *ParseTreeNode) AST {
    id := node.IDENTIFIER().(Token)
    identifier := &IdentifierNode { id.Value, id.Start, id.End }
    skip := node.S() != nil
    return &TokenNode { identifier, VisitNode(v, node.Expr()), skip, node.Start, node.End }
}

func (v ParseTreeVisitor) VisitFragmentStmt(node *ParseTreeNode) AST {
    id := node.IDENTIFIER().(Token)
    identifier := &IdentifierNode { id.Value, id.Start, id.End }
    return &FragmentNode { identifier, VisitNode(v, node.Expr()), node.Start, node.End }
}

func (v ParseTreeVisitor) VisitUnionExpr(node *ParseTreeNode) AST {
    left, right := VisitNode(v, node.L()), VisitNode(v, node.R())
    return &UnionNode { left, right, node.Start, node.End }
}

func (v ParseTreeVisitor) VisitLabelExpr(node *ParseTreeNode) AST {
    id := node.IDENTIFIER().(Token)
    identifier := &IdentifierNode { id.Value, id.Start, id.End }
    var precedence *IdentifierNode
    if t, ok := node.P().(*ParseTreeNode); ok {
        n := t.IDENTIFIER().(Token)
        precedence = &IdentifierNode { n.Value, n.Start, n.End }
    }
    return &LabelNode { VisitNode(v, node.Expr()), identifier, precedence, node.Start, node.End }
}

func (v ParseTreeVisitor) VisitConcatExpr(node *ParseTreeNode) AST {
    left, right := VisitNode(v, node.L()), VisitNode(v, node.R())
    return &ConcatNode { left, right, node.Start, node.End }
}

func (v ParseTreeVisitor) VisitAliasExpr(node *ParseTreeNode) AST {
    id := node.IDENTIFIER().(Token)
    identifier := &IdentifierNode { id.Value, id.Start, id.End }
    return &AliasNode { identifier, VisitNode(v, node.Expr()), node.Start, node.End }
}

func (v ParseTreeVisitor) VisitQuantifierExpr(node *ParseTreeNode) AST {
    switch node.Op().(Token).Type {
    case QUESTION: return &OptionNode    { VisitNode(v, node.Expr()), node.Start, node.End }
    case STAR:     return &RepeatNode    { VisitNode(v, node.Expr()), node.Start, node.End }
    case PLUS:     return &RepeatOneNode { VisitNode(v, node.Expr()), node.Start, node.End }
    default: panic("Invalid quantifier operation")
    }
}

func (v ParseTreeVisitor) VisitGroupExpr(node *ParseTreeNode) AST { return VisitNode(v, node.Expr()) }
func (v ParseTreeVisitor) VisitIdentifierExpr(node *ParseTreeNode) AST {
    return &IdentifierNode { node.IDENTIFIER().(Token).Value, node.Start, node.End }
}

func (v ParseTreeVisitor) VisitStringExpr(node *ParseTreeNode) AST {
    str := node.STRING().(Token)
    value := str.Value[1:len(str.Value) - 1] // Remove quotation marks
    return &StringNode { reduceString([]rune(value)), node.Start, node.End }
}

func (v ParseTreeVisitor) VisitClassExpr(node *ParseTreeNode) AST {
    class := node.CLASS().(Token)
    value := class.Value[1:len(class.Value) - 1] // Remove brackets
    // If caret occurs, flag class as negated and remove caret
    negated, location := len(value) > 0 && value[0] == '^', node.Start
    var expanded []Range
    if !negated {
        expanded = expandClass(reduceString([]rune(value)), location)
    } else {
        expanded = negateRanges(expandClass(reduceString([]rune(value[1:])), location))
    }
    return &ClassNode { expanded, location, node.End }
}

func (v ParseTreeVisitor) VisitErrorExpr(node *ParseTreeNode) AST { return &ErrorNode { node.Start, node.End } }
func (v ParseTreeVisitor) VisitAnyExpr(node *ParseTreeNode) AST {
    location := node.Start
    return &ClassNode { negateRanges(expandClass([]rune { '\n', '\r' }, location)), location, node.End }
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
                Error(fmt.Sprintf("Invalid range from \"%s\" to \"%s\" - %d:%d",
                    formatChar(chars[i - 1]), formatChar(chars[i + 1]), location.Line, location.Col))
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

// ------------------------------------------------------------------------------------------------------------------------------

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
func (n PrecedenceNode) String() string {
    var assoc string
    if n.Associativity == LEFT_ASSOC {
        assoc = "left"
    } else {
        assoc = "right"
    }
    return fmt.Sprintf("prec %s : %s", n.Identifier, assoc)
}
func (n TokenNode) String() string {
    if n.Skip {
        return fmt.Sprintf("token skip %s : %v", n.Identifier, n.Expression)
    }
    return fmt.Sprintf("token %s : %v", n.Identifier, n.Expression)
}
func (n FragmentNode) String() string {
    return fmt.Sprintf("frag %s : %v", n.Identifier, n.Expression)
}

func (n OptionNode) String() string { return fmt.Sprintf("(%v)?", n.Expression) }
func (n RepeatNode) String() string { return fmt.Sprintf("(%v)*", n.Expression) }
func (n RepeatOneNode) String() string { return fmt.Sprintf("(%v)+", n.Expression) }

func (n LabelNode) String() string {
    var precedence string
    if n.Precedence != nil { precedence = fmt.Sprintf(" %%%s", n.Precedence) }
    return fmt.Sprintf("(%v) #%s%s", n.Expression, n.Identifier, precedence)
}
func (n AliasNode) String() string { return fmt.Sprintf("(%s = %v)", n.Identifier, n.Expression) }

func (n ConcatNode) String() string { return fmt.Sprintf("(%v %v)", n.A, n.B) }
func (n UnionNode) String() string { return fmt.Sprintf("(%v | %v)", n.A, n.B) }

func (n IdentifierNode) String() string { return fmt.Sprintf("id:%s", n.Name) }
func (n StringNode) String() string { return fmt.Sprintf("%q", string(n.Chars)) }
func (n ClassNode) String() string {
    ranges := make([]string, len(n.Ranges))
    for i, r := range n.Ranges { ranges[i] = r.String() }
    return fmt.Sprintf("[%s]", strings.Join(ranges, ","))
}
func (n ErrorNode) String() string { return "error" }
