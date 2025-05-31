package lynn

import (
	"fmt"
	"strings"
)

// Production type enum. Either NORMAL, AUXILIARY, FLATTEN, OR REMOVED.
type ProductionType uint
const (NORMAL ProductionType = iota; AUXILIARY; FLATTEN; REMOVED)
// Production data struct. Expresses a sequence of symbols that a given non-terminal may be expanded to in a grammar.
type ProductionData struct {
    Type         ProductionType
    Left, Length int
    Visitor      string
}

// Action type enum. Either SHIFT, REDUCE, or ACCEPT.
type ActionType uint
const (SHIFT ActionType = iota; REDUCE; ACCEPT)
// Parse table action entry struct. Holds action type and integer parameter.
type ActionEntry struct {
    Type  ActionType
    Value int // For SHIFT actions, value represents a state identifier, for REDUCE actions, a production identifier
}

// Parse tree child interface. May either be a Token or ParseTreeNode struct.
type ParseTreeChild interface { string(indent string) string }
type ParseTreeNode struct {
    Visitor  string
    Children []ParseTreeChild
}

var productions = []ProductionData {
    { FLATTEN, 3, 2, "" },
    { NORMAL, 3, 0, "" },
    { NORMAL, 0, 1, "grammar" },
    { NORMAL, 1, 5, "ruleStmt" },
    { NORMAL, 4, 2, "" },
    { REMOVED, 4, 0, "" },
    { NORMAL, 1, 6, "tokenStmt" },
    { NORMAL, 1, 5, "fragmentStmt" },
    { NORMAL, 2, 3, "unionExpr" },
    { AUXILIARY, 5, 1, "" },
    { AUXILIARY, 5, 1, "" },
    { REMOVED, 5, 0, "" },
    { NORMAL, 7, 4, "labelExpr" },
    { NORMAL, 8, 2, "concatExpr" },
    { NORMAL, 9, 3, "aliasExpr" },
    { AUXILIARY, 6, 1, "" },
    { AUXILIARY, 6, 1, "" },
    { AUXILIARY, 6, 1, "" },
    { NORMAL, 10, 2, "quantifierExpr" },
    { NORMAL, 10, 3, "groupExpr" },
    { NORMAL, 10, 1, "identifierExpr" },
    { NORMAL, 10, 1, "stringExpr" },
    { NORMAL, 10, 1, "classExpr" },
    { NORMAL, 10, 1, "anyExpr" },
    { AUXILIARY, 2, 1, "" },
    { AUXILIARY, 7, 1, "" },
    { AUXILIARY, 8, 1, "" },
    { AUXILIARY, 9, 1, "" },
}
var actionTable = []map[TokenType]ActionEntry {
    { 23: { REDUCE, 1 }, 3: { REDUCE, 1 }, 4: { REDUCE, 1 }, 2: { REDUCE, 1 } },
    { 23: { REDUCE, 2 }, 2: { SHIFT, 4 }, 4: { SHIFT, 5 }, 3: { SHIFT, 3 } },
    { 23: { ACCEPT, 0 } },
    { 20: { SHIFT, 7 } },
    { 20: { SHIFT, 8 } },
    { 20: { SHIFT, 9 } },
    { 4: { REDUCE, 0 }, 23: { REDUCE, 0 }, 3: { REDUCE, 0 }, 2: { REDUCE, 0 } },
    { 16: { SHIFT, 10 } },
    { 16: { SHIFT, 11 } },
    { 16: { SHIFT, 12 } },
    { 21: { SHIFT, 15 }, 17: { SHIFT, 17 }, 20: { SHIFT, 14 }, 22: { SHIFT, 21 }, 12: { SHIFT, 22 } },
    { 20: { SHIFT, 14 }, 17: { SHIFT, 17 }, 21: { SHIFT, 15 }, 12: { SHIFT, 22 }, 22: { SHIFT, 21 } },
    { 12: { SHIFT, 22 }, 17: { SHIFT, 17 }, 20: { SHIFT, 14 }, 21: { SHIFT, 15 }, 22: { SHIFT, 21 } },
    { 13: { REDUCE, 25 }, 3: { REDUCE, 25 }, 21: { SHIFT, 15 }, 22: { SHIFT, 21 }, 17: { SHIFT, 17 }, 18: { REDUCE, 25 }, 19: { REDUCE, 25 }, 4: { REDUCE, 25 }, 23: { REDUCE, 25 }, 14: { REDUCE, 25 }, 12: { SHIFT, 22 }, 20: { SHIFT, 14 }, 15: { REDUCE, 25 }, 2: { REDUCE, 25 } },
    { 23: { REDUCE, 20 }, 21: { REDUCE, 20 }, 9: { REDUCE, 20 }, 14: { REDUCE, 20 }, 17: { REDUCE, 20 }, 22: { REDUCE, 20 }, 10: { REDUCE, 20 }, 20: { REDUCE, 20 }, 8: { SHIFT, 26 }, 15: { REDUCE, 20 }, 13: { REDUCE, 20 }, 18: { REDUCE, 20 }, 11: { REDUCE, 20 }, 4: { REDUCE, 20 }, 3: { REDUCE, 20 }, 19: { REDUCE, 20 }, 12: { REDUCE, 20 }, 2: { REDUCE, 20 } },
    { 18: { REDUCE, 21 }, 13: { REDUCE, 21 }, 20: { REDUCE, 21 }, 10: { REDUCE, 21 }, 19: { REDUCE, 21 }, 22: { REDUCE, 21 }, 21: { REDUCE, 21 }, 3: { REDUCE, 21 }, 4: { REDUCE, 21 }, 11: { REDUCE, 21 }, 17: { REDUCE, 21 }, 15: { REDUCE, 21 }, 9: { REDUCE, 21 }, 2: { REDUCE, 21 }, 14: { REDUCE, 21 }, 12: { REDUCE, 21 }, 23: { REDUCE, 21 } },
    { 2: { REDUCE, 27 }, 11: { SHIFT, 29 }, 10: { SHIFT, 27 }, 13: { REDUCE, 27 }, 19: { REDUCE, 27 }, 22: { REDUCE, 27 }, 20: { REDUCE, 27 }, 21: { REDUCE, 27 }, 14: { REDUCE, 27 }, 23: { REDUCE, 27 }, 17: { REDUCE, 27 }, 3: { REDUCE, 27 }, 12: { REDUCE, 27 }, 18: { REDUCE, 27 }, 9: { SHIFT, 30 }, 4: { REDUCE, 27 }, 15: { REDUCE, 27 } },
    { 21: { SHIFT, 15 }, 20: { SHIFT, 14 }, 17: { SHIFT, 17 }, 22: { SHIFT, 21 }, 12: { SHIFT, 22 } },
    { 13: { SHIFT, 33 }, 19: { SHIFT, 34 }, 15: { REDUCE, 5 } },
    { 14: { SHIFT, 35 }, 18: { REDUCE, 24 }, 4: { REDUCE, 24 }, 23: { REDUCE, 24 }, 3: { REDUCE, 24 }, 13: { REDUCE, 24 }, 2: { REDUCE, 24 }, 19: { REDUCE, 24 }, 15: { REDUCE, 24 } },
    { 4: { REDUCE, 26 }, 21: { REDUCE, 26 }, 2: { REDUCE, 26 }, 17: { REDUCE, 26 }, 20: { REDUCE, 26 }, 19: { REDUCE, 26 }, 3: { REDUCE, 26 }, 12: { REDUCE, 26 }, 14: { REDUCE, 26 }, 23: { REDUCE, 26 }, 15: { REDUCE, 26 }, 22: { REDUCE, 26 }, 18: { REDUCE, 26 }, 13: { REDUCE, 26 } },
    { 10: { REDUCE, 22 }, 9: { REDUCE, 22 }, 15: { REDUCE, 22 }, 2: { REDUCE, 22 }, 20: { REDUCE, 22 }, 3: { REDUCE, 22 }, 21: { REDUCE, 22 }, 23: { REDUCE, 22 }, 12: { REDUCE, 22 }, 4: { REDUCE, 22 }, 13: { REDUCE, 22 }, 11: { REDUCE, 22 }, 14: { REDUCE, 22 }, 19: { REDUCE, 22 }, 18: { REDUCE, 22 }, 22: { REDUCE, 22 }, 17: { REDUCE, 22 } },
    { 14: { REDUCE, 23 }, 9: { REDUCE, 23 }, 17: { REDUCE, 23 }, 2: { REDUCE, 23 }, 11: { REDUCE, 23 }, 15: { REDUCE, 23 }, 12: { REDUCE, 23 }, 3: { REDUCE, 23 }, 13: { REDUCE, 23 }, 20: { REDUCE, 23 }, 4: { REDUCE, 23 }, 21: { REDUCE, 23 }, 18: { REDUCE, 23 }, 19: { REDUCE, 23 }, 23: { REDUCE, 23 }, 22: { REDUCE, 23 }, 10: { REDUCE, 23 } },
    { 13: { SHIFT, 33 }, 15: { SHIFT, 36 } },
    { 13: { SHIFT, 33 }, 15: { SHIFT, 37 } },
    { 3: { REDUCE, 13 }, 20: { REDUCE, 13 }, 17: { REDUCE, 13 }, 2: { REDUCE, 13 }, 18: { REDUCE, 13 }, 23: { REDUCE, 13 }, 4: { REDUCE, 13 }, 14: { REDUCE, 13 }, 22: { REDUCE, 13 }, 13: { REDUCE, 13 }, 12: { REDUCE, 13 }, 19: { REDUCE, 13 }, 21: { REDUCE, 13 }, 15: { REDUCE, 13 } },
    { 12: { SHIFT, 22 }, 21: { SHIFT, 15 }, 22: { SHIFT, 21 }, 20: { SHIFT, 14 }, 17: { SHIFT, 17 } },
    { 15: { REDUCE, 16 }, 13: { REDUCE, 16 }, 14: { REDUCE, 16 }, 17: { REDUCE, 16 }, 19: { REDUCE, 16 }, 20: { REDUCE, 16 }, 23: { REDUCE, 16 }, 2: { REDUCE, 16 }, 10: { REDUCE, 16 }, 21: { REDUCE, 16 }, 9: { REDUCE, 16 }, 4: { REDUCE, 16 }, 11: { REDUCE, 16 }, 3: { REDUCE, 16 }, 18: { REDUCE, 16 }, 12: { REDUCE, 16 }, 22: { REDUCE, 16 } },
    { 3: { REDUCE, 18 }, 13: { REDUCE, 18 }, 4: { REDUCE, 18 }, 10: { REDUCE, 18 }, 17: { REDUCE, 18 }, 18: { REDUCE, 18 }, 19: { REDUCE, 18 }, 22: { REDUCE, 18 }, 11: { REDUCE, 18 }, 2: { REDUCE, 18 }, 21: { REDUCE, 18 }, 12: { REDUCE, 18 }, 23: { REDUCE, 18 }, 9: { REDUCE, 18 }, 20: { REDUCE, 18 }, 14: { REDUCE, 18 }, 15: { REDUCE, 18 } },
    { 12: { REDUCE, 17 }, 9: { REDUCE, 17 }, 20: { REDUCE, 17 }, 14: { REDUCE, 17 }, 17: { REDUCE, 17 }, 22: { REDUCE, 17 }, 4: { REDUCE, 17 }, 2: { REDUCE, 17 }, 13: { REDUCE, 17 }, 18: { REDUCE, 17 }, 3: { REDUCE, 17 }, 21: { REDUCE, 17 }, 19: { REDUCE, 17 }, 10: { REDUCE, 17 }, 11: { REDUCE, 17 }, 23: { REDUCE, 17 }, 15: { REDUCE, 17 } },
    { 19: { REDUCE, 15 }, 23: { REDUCE, 15 }, 13: { REDUCE, 15 }, 3: { REDUCE, 15 }, 18: { REDUCE, 15 }, 20: { REDUCE, 15 }, 9: { REDUCE, 15 }, 22: { REDUCE, 15 }, 4: { REDUCE, 15 }, 17: { REDUCE, 15 }, 15: { REDUCE, 15 }, 10: { REDUCE, 15 }, 2: { REDUCE, 15 }, 12: { REDUCE, 15 }, 11: { REDUCE, 15 }, 14: { REDUCE, 15 }, 21: { REDUCE, 15 } },
    { 18: { SHIFT, 39 }, 13: { SHIFT, 33 } },
    { 15: { SHIFT, 40 } },
    { 12: { SHIFT, 22 }, 20: { SHIFT, 14 }, 21: { SHIFT, 15 }, 22: { SHIFT, 21 }, 17: { SHIFT, 17 } },
    { 7: { SHIFT, 42 } },
    { 20: { SHIFT, 43 } },
    { 4: { REDUCE, 3 }, 3: { REDUCE, 3 }, 2: { REDUCE, 3 }, 23: { REDUCE, 3 } },
    { 23: { REDUCE, 7 }, 4: { REDUCE, 7 }, 3: { REDUCE, 7 }, 2: { REDUCE, 7 } },
    { 14: { REDUCE, 14 }, 15: { REDUCE, 14 }, 19: { REDUCE, 14 }, 13: { REDUCE, 14 }, 22: { REDUCE, 14 }, 2: { REDUCE, 14 }, 17: { REDUCE, 14 }, 20: { REDUCE, 14 }, 12: { REDUCE, 14 }, 3: { REDUCE, 14 }, 21: { REDUCE, 14 }, 4: { REDUCE, 14 }, 23: { REDUCE, 14 }, 18: { REDUCE, 14 } },
    { 3: { REDUCE, 19 }, 17: { REDUCE, 19 }, 13: { REDUCE, 19 }, 9: { REDUCE, 19 }, 22: { REDUCE, 19 }, 14: { REDUCE, 19 }, 10: { REDUCE, 19 }, 18: { REDUCE, 19 }, 4: { REDUCE, 19 }, 19: { REDUCE, 19 }, 21: { REDUCE, 19 }, 2: { REDUCE, 19 }, 20: { REDUCE, 19 }, 23: { REDUCE, 19 }, 15: { REDUCE, 19 }, 12: { REDUCE, 19 }, 11: { REDUCE, 19 } },
    { 2: { REDUCE, 6 }, 23: { REDUCE, 6 }, 3: { REDUCE, 6 }, 4: { REDUCE, 6 } },
    { 14: { SHIFT, 35 }, 4: { REDUCE, 8 }, 15: { REDUCE, 8 }, 18: { REDUCE, 8 }, 13: { REDUCE, 8 }, 2: { REDUCE, 8 }, 23: { REDUCE, 8 }, 3: { REDUCE, 8 }, 19: { REDUCE, 8 } },
    { 15: { REDUCE, 4 } },
    { 6: { SHIFT, 44 }, 3: { REDUCE, 11 }, 4: { REDUCE, 11 }, 19: { REDUCE, 11 }, 14: { REDUCE, 11 }, 18: { REDUCE, 11 }, 5: { SHIFT, 45 }, 23: { REDUCE, 11 }, 15: { REDUCE, 11 }, 13: { REDUCE, 11 }, 2: { REDUCE, 11 } },
    { 4: { REDUCE, 10 }, 18: { REDUCE, 10 }, 3: { REDUCE, 10 }, 23: { REDUCE, 10 }, 19: { REDUCE, 10 }, 2: { REDUCE, 10 }, 15: { REDUCE, 10 }, 13: { REDUCE, 10 }, 14: { REDUCE, 10 } },
    { 19: { REDUCE, 9 }, 15: { REDUCE, 9 }, 18: { REDUCE, 9 }, 13: { REDUCE, 9 }, 3: { REDUCE, 9 }, 4: { REDUCE, 9 }, 14: { REDUCE, 9 }, 23: { REDUCE, 9 }, 2: { REDUCE, 9 } },
    { 2: { REDUCE, 12 }, 13: { REDUCE, 12 }, 15: { REDUCE, 12 }, 23: { REDUCE, 12 }, 14: { REDUCE, 12 }, 4: { REDUCE, 12 }, 18: { REDUCE, 12 }, 19: { REDUCE, 12 }, 3: { REDUCE, 12 } },
}
var gotoTable = []map[int]int {
    { 3: 1, 0: 2 },
    { 1: 6 },
    { },
    { },
    { },
    { },
    { },
    { },
    { },
    { },
    { 9: 20, 8: 13, 10: 16, 2: 18, 7: 19 },
    { 7: 19, 9: 20, 8: 13, 10: 16, 2: 23 },
    { 10: 16, 9: 20, 7: 19, 8: 13, 2: 24 },
    { 10: 16, 9: 25 },
    { },
    { },
    { 6: 28 },
    { 9: 20, 7: 19, 2: 31, 8: 13, 10: 16 },
    { 4: 32 },
    { },
    { },
    { },
    { },
    { },
    { },
    { },
    { 9: 38, 10: 16 },
    { },
    { },
    { },
    { },
    { },
    { },
    { 10: 16, 9: 20, 7: 41, 8: 13 },
    { },
    { },
    { },
    { },
    { },
    { },
    { },
    { },
    { },
    { 5: 46 },
    { },
    { },
    { },
}

// Parser struct. Converts token stream to parse tree.
type Parser struct {
    lexer   *Lexer
    handler ParserErrorHandler
}

// Base visitor struct. Describes functions necessary to implement to traverse parse tree.
type BaseVisitor[T any] interface {
    VisitGrammar(node *ParseTreeNode) T
    VisitRuleStmt(node *ParseTreeNode) T
    VisitTokenStmt(node *ParseTreeNode) T
    VisitFragmentStmt(node *ParseTreeNode) T
    VisitUnionExpr(node *ParseTreeNode) T
    VisitLabelExpr(node *ParseTreeNode) T
    VisitConcatExpr(node *ParseTreeNode) T
    VisitAliasExpr(node *ParseTreeNode) T
    VisitQuantifierExpr(node *ParseTreeNode) T
    VisitGroupExpr(node *ParseTreeNode) T
    VisitIdentifierExpr(node *ParseTreeNode) T
    VisitStringExpr(node *ParseTreeNode) T
    VisitClassExpr(node *ParseTreeNode) T
    VisitAnyExpr(node *ParseTreeNode) T
}

// Function called when the parser encounters an error.
type ParserErrorHandler func (token Token)
var DEFAULT_PARSER_HANDLER = func (token Token) {
    fmt.Printf("Syntax error: Unexpected token %q - %d:%d\n", token.Value, token.Start.Line, token.Start.Col)
}

// Returns new parser struct.
func NewParser(lexer *Lexer, handler ParserErrorHandler) *Parser { return &Parser { lexer, handler } }
// Generates parse tree based on token stream from lexer.
func (p *Parser) Parse() *ParseTreeNode {
    // Stack state struct. Holds the state identifier and the corresponding parse tree node.
    type StackState struct {
        state int
        node  ParseTreeChild
    }
    // Initialize current token and stack
    token, stack := p.lexer.Token, []StackState { { 0, nil } }
    for {
        // Get the current state at the top of the stack and find the action to take
        // Next action is determined by action table given state index and the current token type
        state := stack[len(stack) - 1].state
        action, ok := actionTable[state][token.Type]
        if !ok {
            // If the table does not have a valid action, cannot parse current token
            fmt.Printf("Syntax error: Unexpected token %q - %d:%d\n", token.Value, token.Start.Line, token.Start.Col)
            return nil
        }
        switch action.Type {
        case SHIFT:
            // For shift actions, add new state to the stack along with token
            stack = append(stack, StackState { action.Value, token })
            token = p.lexer.Next()
        case REDUCE:
            // For reduce actions, pop states off stack and merge children into one node based on production
            production := &productions[action.Value]
            i := len(stack) - production.Length
            var node ParseTreeChild
            switch production.Type {
            case NORMAL:
                // Collect child nodes from current states on the stack and create node for reduction
                children := make([]ParseTreeChild, production.Length)
                for i, s := range stack[i:] { children[i] = s.node }
                node = &ParseTreeNode { production.Visitor, children }
            case FLATTEN:
                // Of the two nodes popped, preserve the first and add the second as a child of the first
                // Results in quantified expressions in the grammar generating arrays of elements
                list, element := stack[i].node.(*ParseTreeNode), stack[i + 1].node
                list.Children = append(list.Children, element)
                node = list
            case AUXILIARY: node = stack[i].node // Passes child through without generating new node
            case REMOVED:   node = nil // Adds a nil value
            }
            // Pop consumed states off stack
            // Given new state at the top of the stack, find next state based on the goto table
            stack = stack[:i]
            state := stack[i - 1].state
            next := gotoTable[state][production.Left]
            // Add new state to top of the stack
            stack = append(stack, StackState { next, node })
        case ACCEPT: return stack[1].node.(*ParseTreeNode)
        }
    }
}

// Given a parse tree node, dispatches the corresponding function in the visitor.
func VisitNode[T any](visitor BaseVisitor[T], node *ParseTreeNode) T {
    switch node.Visitor {
    case "grammar": return visitor.VisitGrammar(node)
    case "ruleStmt": return visitor.VisitRuleStmt(node)
    case "tokenStmt": return visitor.VisitTokenStmt(node)
    case "fragmentStmt": return visitor.VisitFragmentStmt(node)
    case "unionExpr": return visitor.VisitUnionExpr(node)
    case "labelExpr": return visitor.VisitLabelExpr(node)
    case "concatExpr": return visitor.VisitConcatExpr(node)
    case "aliasExpr": return visitor.VisitAliasExpr(node)
    case "quantifierExpr": return visitor.VisitQuantifierExpr(node)
    case "groupExpr": return visitor.VisitGroupExpr(node)
    case "identifierExpr": return visitor.VisitIdentifierExpr(node)
    case "stringExpr": return visitor.VisitStringExpr(node)
    case "classExpr": return visitor.VisitClassExpr(node)
    case "anyExpr": return visitor.VisitAnyExpr(node)
    default: panic("Invalid parse tree node passed to VisitNode()")
    }
}

// FOR DEBUG PURPOSES:
// Prints the parse tree to the standard output.
func (n *ParseTreeNode) Print() { fmt.Println(n.string("")) }

func (t Token) string(indent string) string { return fmt.Sprintf("%s<%s %s>", indent, t.Type, t.Value) }
func (n *ParseTreeNode) string(indent string) string {
    children := make([]string, len(n.Children))
    next := indent + "  "
    for i, c := range n.Children {
        str := "\n"
        if c == nil {
            str += fmt.Sprintf("%s<nil>", next)
        } else {
            str += c.string(next)
        }
        children[i] = str
    }
    return fmt.Sprintf("%s[%s]%s", indent, n.Visitor, strings.Join(children, ""))
}
