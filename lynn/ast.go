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

// Node representing a concatenation operation. Requires that one expression immediately follow the preceding expression.
type ConcatenationNode struct {
    A, B AST
}

// Node representing a union operation. Allows either one expression or the other to occur.
type UnionNode struct {
    A, B AST
}

// Node representing an identifier literal.
type IdentifierNode struct { Name string; Location Location }
// Node representing a string literal.
type StringNode struct { Chars []rune }
// Node representing a class literal.
type ClassNode struct { Ranges []Range }

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

func (n ConcatenationNode) String() string { return fmt.Sprintf("(%v %v)", n.A, n.B) }
func (n UnionNode) String() string { return fmt.Sprintf("(%v | %v)", n.A, n.B) }

func (n IdentifierNode) String() string { return fmt.Sprintf("id:%s", n.Name) }
func (n StringNode) String() string { return fmt.Sprintf("%q", string(n.Chars)) }
func (n ClassNode) String() string {
    ranges := make([]string, len(n.Ranges))
    for i, r := range n.Ranges { ranges[i] = r.String() }
    return fmt.Sprintf("[%s]", strings.Join(ranges, ","))
}
