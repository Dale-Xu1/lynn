package lynn

import (
	"fmt"
	"strings"
)

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
    Identifier *IdentifierNode
    Expression AST
}

// Node representing a token rule. Specifies the token's identifier and regular expression.
type TokenNode struct {
    Identifier *IdentifierNode
    Expression AST
    Skip       bool
}

// Node representing a rule fragment. Specifies the fragment's identifier and regular expression.
// Fragments are used to repeat regular expressions in token rules.
type FragmentNode struct {
    Identifier *IdentifierNode
    Expression AST
}

// Node representing an option quantifier. Allows zero or one occurrence of the given regular expression.
type OptionNode struct { Expression AST }
// Node representing an repeat quantifier. Allows zero or more occurrences of the given regular expression.
type RepeatNode struct { Expression AST }
// Node representing an repeat one or more quantifier. Allows one or more occurrences of the given regular expression.
type RepeatOneNode struct { Expression AST }

// Associativity type enum. Either NO_ASSOC, LEFT_ASSOC, or RIGHT_ASSOC.
type AssociativityType uint
const (NO_ASSOC AssociativityType = iota; LEFT_ASSOC; RIGHT_ASSOC)
// Node representing a rule case label. Specifies the callback identifier and associativity for the disambiguation process.
type LabelNode struct {
    Expression    AST
    Identifier    *IdentifierNode
    Associativity AssociativityType
    Location      Location
}

// Node representing an alias. Specifies the alias identifier and the corresponding expression.
type AliasNode struct {
    Identifier *IdentifierNode
    Expression AST
    Location   Location
}

// Node representing a concatenation operation. Requires that one expression immediately follow the preceding expression.
type ConcatNode struct {
    A, B AST
}

// Node representing a union operation. Allows either one expression or the other to occur.
type UnionNode struct {
    A, B AST
}

// Node representing an identifier literal.
type IdentifierNode struct { Name string; Location Location }
// Node representing a string literal.
type StringNode struct { Chars []rune; Location Location }
// Node representing a class literal.
type ClassNode struct { Ranges []Range; Location Location }

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

func (n LabelNode) String() string {
    var assoc string
    if n.Associativity == LEFT_ASSOC {
        assoc = "left"
    } else {
        assoc = "right"
    }
    return fmt.Sprintf("(%v) #%s %s", n.Expression, n.Identifier, assoc)
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

// TODO: Parse tree to AST
type ParseTreeVisitor struct { }
func NewParseTreeVisitor() ParseTreeVisitor { return ParseTreeVisitor { } }

func (v ParseTreeVisitor) VisitGrammar(node *ParseTreeNode) AST {
    rules, tokens, fragments := make([]*RuleNode, 0), make([]*TokenNode, 0), make([]*FragmentNode, 0)
    for _, node := range node.Stmt().(*ParseTreeNode).Children {
        switch rule := VisitNode(v, node.(*ParseTreeNode)).(type) {
        case *RuleNode: rules = append(rules, rule)
        case *TokenNode: tokens = append(tokens, rule)
        case *FragmentNode: fragments = append(fragments, rule)
        }
    }
    return &GrammarNode { rules, tokens, fragments }
}

func (v ParseTreeVisitor) VisitRuleStmt(node *ParseTreeNode) AST {
    return &RuleNode { }
}

func (v ParseTreeVisitor) VisitTokenStmt(node *ParseTreeNode) AST {
    return &TokenNode { }
}

func (v ParseTreeVisitor) VisitFragmentStmt(node *ParseTreeNode) AST {
    return &FragmentNode { }
}

func (v ParseTreeVisitor) VisitUnionExpr(node *ParseTreeNode) AST {
    panic("TODO")
}

func (v ParseTreeVisitor) VisitLabelExpr(node *ParseTreeNode) AST {
    panic("TODO")
}

func (v ParseTreeVisitor) VisitConcatExpr(node *ParseTreeNode) AST {
    panic("TODO")
}

func (v ParseTreeVisitor) VisitAliasExpr(node *ParseTreeNode) AST {
    panic("TODO")
}

func (v ParseTreeVisitor) VisitQuantifierExpr(node *ParseTreeNode) AST {
    panic("TODO")
}

func (v ParseTreeVisitor) VisitGroupExpr(node *ParseTreeNode) AST {
    panic("TODO")
}

func (v ParseTreeVisitor) VisitIdentifierExpr(node *ParseTreeNode) AST {
    panic("TODO")
}

func (v ParseTreeVisitor) VisitStringExpr(node *ParseTreeNode) AST {
    panic("TODO")
}

func (v ParseTreeVisitor) VisitClassExpr(node *ParseTreeNode) AST {
    panic("TODO")
}

func (v ParseTreeVisitor) VisitAnyExpr(node *ParseTreeNode) AST {
    panic("TODO")
}
