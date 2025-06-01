package lynn

import (
	"fmt"
	"strings"
)

// Production data struct. Expresses a sequence of symbols that a given non-terminal may be expanded to in a grammar.
type productionData struct {
    Type, Left, Length int
    Visitor            string
}
// Parse table action entry struct. Holds action type and integer parameter.
type actionEntry struct {
    Type, Value int // For 0 actions, value represents a state identifier, for 1 actions, a production identifier
}

// Parse tree child interface. May either be a Token or ParseTreeNode struct.
type ParseTreeChild interface { string(indent string) string }
type ParseTreeNode struct {
    data     *productionData
    Children []ParseTreeChild
}

var productions = []productionData {
    { 2, 3, 2, "" },
    { 0, 3, 0, "" },
    { 0, 0, 1, "grammar" },
    { 0, 1, 5, "ruleStmt" },
    { 0, 4, 2, "" },
    { 3, 4, 0, "" },
    { 0, 1, 6, "tokenStmt" },
    { 0, 1, 5, "fragmentStmt" },
    { 0, 2, 3, "unionExpr" },
    { 1, 5, 1, "" },
    { 1, 5, 1, "" },
    { 3, 5, 0, "" },
    { 0, 7, 4, "labelExpr" },
    { 0, 8, 2, "concatExpr" },
    { 0, 9, 3, "aliasExpr" },
    { 1, 6, 1, "" },
    { 1, 6, 1, "" },
    { 1, 6, 1, "" },
    { 0, 10, 2, "quantifierExpr" },
    { 0, 10, 3, "groupExpr" },
    { 0, 10, 1, "identifierExpr" },
    { 0, 10, 1, "stringExpr" },
    { 0, 10, 1, "classExpr" },
    { 0, 10, 1, "anyExpr" },
    { 1, 2, 1, "" },
    { 1, 7, 1, "" },
    { 1, 8, 1, "" },
    { 1, 9, 1, "" },
}
var actionTable = []map[TokenType]actionEntry {
    { 23: { 1, 1 }, 2: { 1, 1 }, 3: { 1, 1 }, 4: { 1, 1 } },
    { 23: { 2, 0 } },
    { 3: { 0, 3 }, 4: { 0, 5 }, 2: { 0, 6 }, 23: { 1, 2 } },
    { 20: { 0, 7 } },
    { 3: { 1, 0 }, 2: { 1, 0 }, 4: { 1, 0 }, 23: { 1, 0 } },
    { 20: { 0, 8 } },
    { 20: { 0, 9 } },
    { 16: { 0, 10 } },
    { 16: { 0, 11 } },
    { 16: { 0, 12 } },
    { 22: { 0, 21 }, 20: { 0, 13 }, 12: { 0, 16 }, 21: { 0, 19 }, 17: { 0, 17 } },
    { 20: { 0, 13 }, 17: { 0, 17 }, 12: { 0, 16 }, 21: { 0, 19 }, 22: { 0, 21 } },
    { 22: { 0, 21 }, 21: { 0, 19 }, 17: { 0, 17 }, 12: { 0, 16 }, 20: { 0, 13 } },
    { 23: { 1, 20 }, 20: { 1, 20 }, 2: { 1, 20 }, 3: { 1, 20 }, 17: { 1, 20 }, 15: { 1, 20 }, 11: { 1, 20 }, 13: { 1, 20 }, 14: { 1, 20 }, 10: { 1, 20 }, 8: { 0, 25 }, 4: { 1, 20 }, 18: { 1, 20 }, 9: { 1, 20 }, 21: { 1, 20 }, 12: { 1, 20 }, 19: { 1, 20 }, 22: { 1, 20 } },
    { 22: { 0, 21 }, 20: { 0, 13 }, 17: { 0, 17 }, 21: { 0, 19 }, 13: { 1, 25 }, 23: { 1, 25 }, 18: { 1, 25 }, 2: { 1, 25 }, 12: { 0, 16 }, 3: { 1, 25 }, 15: { 1, 25 }, 4: { 1, 25 }, 19: { 1, 25 }, 14: { 1, 25 } },
    { 13: { 0, 27 }, 19: { 0, 29 }, 15: { 1, 5 } },
    { 11: { 1, 23 }, 22: { 1, 23 }, 23: { 1, 23 }, 2: { 1, 23 }, 20: { 1, 23 }, 18: { 1, 23 }, 3: { 1, 23 }, 19: { 1, 23 }, 14: { 1, 23 }, 21: { 1, 23 }, 4: { 1, 23 }, 12: { 1, 23 }, 9: { 1, 23 }, 13: { 1, 23 }, 10: { 1, 23 }, 17: { 1, 23 }, 15: { 1, 23 } },
    { 20: { 0, 13 }, 17: { 0, 17 }, 12: { 0, 16 }, 22: { 0, 21 }, 21: { 0, 19 } },
    { 3: { 1, 27 }, 2: { 1, 27 }, 20: { 1, 27 }, 11: { 0, 32 }, 17: { 1, 27 }, 13: { 1, 27 }, 21: { 1, 27 }, 22: { 1, 27 }, 15: { 1, 27 }, 14: { 1, 27 }, 9: { 0, 31 }, 12: { 1, 27 }, 23: { 1, 27 }, 10: { 0, 33 }, 18: { 1, 27 }, 4: { 1, 27 }, 19: { 1, 27 } },
    { 3: { 1, 21 }, 12: { 1, 21 }, 9: { 1, 21 }, 23: { 1, 21 }, 17: { 1, 21 }, 18: { 1, 21 }, 21: { 1, 21 }, 2: { 1, 21 }, 11: { 1, 21 }, 22: { 1, 21 }, 4: { 1, 21 }, 13: { 1, 21 }, 15: { 1, 21 }, 20: { 1, 21 }, 14: { 1, 21 }, 10: { 1, 21 }, 19: { 1, 21 } },
    { 21: { 1, 26 }, 12: { 1, 26 }, 3: { 1, 26 }, 17: { 1, 26 }, 4: { 1, 26 }, 14: { 1, 26 }, 18: { 1, 26 }, 22: { 1, 26 }, 23: { 1, 26 }, 13: { 1, 26 }, 15: { 1, 26 }, 20: { 1, 26 }, 2: { 1, 26 }, 19: { 1, 26 } },
    { 12: { 1, 22 }, 19: { 1, 22 }, 11: { 1, 22 }, 22: { 1, 22 }, 10: { 1, 22 }, 23: { 1, 22 }, 4: { 1, 22 }, 20: { 1, 22 }, 21: { 1, 22 }, 2: { 1, 22 }, 18: { 1, 22 }, 9: { 1, 22 }, 17: { 1, 22 }, 14: { 1, 22 }, 3: { 1, 22 }, 13: { 1, 22 }, 15: { 1, 22 } },
    { 14: { 0, 35 }, 2: { 1, 24 }, 4: { 1, 24 }, 23: { 1, 24 }, 19: { 1, 24 }, 3: { 1, 24 }, 15: { 1, 24 }, 13: { 1, 24 }, 18: { 1, 24 } },
    { 13: { 0, 27 }, 15: { 0, 36 } },
    { 15: { 0, 37 }, 13: { 0, 27 } },
    { 20: { 0, 13 }, 21: { 0, 19 }, 12: { 0, 16 }, 22: { 0, 21 }, 17: { 0, 17 } },
    { 18: { 1, 13 }, 2: { 1, 13 }, 21: { 1, 13 }, 19: { 1, 13 }, 22: { 1, 13 }, 12: { 1, 13 }, 17: { 1, 13 }, 3: { 1, 13 }, 14: { 1, 13 }, 20: { 1, 13 }, 23: { 1, 13 }, 13: { 1, 13 }, 15: { 1, 13 }, 4: { 1, 13 } },
    { 17: { 0, 17 }, 21: { 0, 19 }, 12: { 0, 16 }, 22: { 0, 21 }, 20: { 0, 13 } },
    { 15: { 0, 40 } },
    { 7: { 0, 41 } },
    { 13: { 0, 27 }, 18: { 0, 42 } },
    { 2: { 1, 15 }, 9: { 1, 15 }, 15: { 1, 15 }, 14: { 1, 15 }, 3: { 1, 15 }, 10: { 1, 15 }, 22: { 1, 15 }, 4: { 1, 15 }, 23: { 1, 15 }, 19: { 1, 15 }, 21: { 1, 15 }, 18: { 1, 15 }, 20: { 1, 15 }, 13: { 1, 15 }, 12: { 1, 15 }, 11: { 1, 15 }, 17: { 1, 15 } },
    { 21: { 1, 17 }, 2: { 1, 17 }, 10: { 1, 17 }, 9: { 1, 17 }, 4: { 1, 17 }, 18: { 1, 17 }, 13: { 1, 17 }, 14: { 1, 17 }, 3: { 1, 17 }, 17: { 1, 17 }, 20: { 1, 17 }, 23: { 1, 17 }, 15: { 1, 17 }, 12: { 1, 17 }, 22: { 1, 17 }, 11: { 1, 17 }, 19: { 1, 17 } },
    { 19: { 1, 16 }, 20: { 1, 16 }, 22: { 1, 16 }, 12: { 1, 16 }, 11: { 1, 16 }, 3: { 1, 16 }, 23: { 1, 16 }, 21: { 1, 16 }, 17: { 1, 16 }, 9: { 1, 16 }, 2: { 1, 16 }, 10: { 1, 16 }, 14: { 1, 16 }, 18: { 1, 16 }, 15: { 1, 16 }, 4: { 1, 16 }, 13: { 1, 16 } },
    { 13: { 1, 18 }, 19: { 1, 18 }, 15: { 1, 18 }, 22: { 1, 18 }, 23: { 1, 18 }, 21: { 1, 18 }, 18: { 1, 18 }, 9: { 1, 18 }, 20: { 1, 18 }, 12: { 1, 18 }, 14: { 1, 18 }, 4: { 1, 18 }, 11: { 1, 18 }, 2: { 1, 18 }, 3: { 1, 18 }, 10: { 1, 18 }, 17: { 1, 18 } },
    { 20: { 0, 43 } },
    { 23: { 1, 7 }, 3: { 1, 7 }, 4: { 1, 7 }, 2: { 1, 7 } },
    { 23: { 1, 3 }, 4: { 1, 3 }, 2: { 1, 3 }, 3: { 1, 3 } },
    { 23: { 1, 14 }, 18: { 1, 14 }, 4: { 1, 14 }, 3: { 1, 14 }, 14: { 1, 14 }, 13: { 1, 14 }, 2: { 1, 14 }, 20: { 1, 14 }, 12: { 1, 14 }, 17: { 1, 14 }, 19: { 1, 14 }, 15: { 1, 14 }, 22: { 1, 14 }, 21: { 1, 14 } },
    { 13: { 1, 8 }, 18: { 1, 8 }, 3: { 1, 8 }, 2: { 1, 8 }, 15: { 1, 8 }, 23: { 1, 8 }, 4: { 1, 8 }, 19: { 1, 8 }, 14: { 0, 35 } },
    { 23: { 1, 6 }, 3: { 1, 6 }, 2: { 1, 6 }, 4: { 1, 6 } },
    { 15: { 1, 4 } },
    { 18: { 1, 19 }, 22: { 1, 19 }, 13: { 1, 19 }, 12: { 1, 19 }, 4: { 1, 19 }, 14: { 1, 19 }, 21: { 1, 19 }, 9: { 1, 19 }, 15: { 1, 19 }, 11: { 1, 19 }, 3: { 1, 19 }, 17: { 1, 19 }, 10: { 1, 19 }, 19: { 1, 19 }, 2: { 1, 19 }, 20: { 1, 19 }, 23: { 1, 19 } },
    { 2: { 1, 11 }, 5: { 0, 45 }, 4: { 1, 11 }, 23: { 1, 11 }, 14: { 1, 11 }, 15: { 1, 11 }, 18: { 1, 11 }, 6: { 0, 46 }, 13: { 1, 11 }, 3: { 1, 11 }, 19: { 1, 11 } },
    { 3: { 1, 12 }, 23: { 1, 12 }, 15: { 1, 12 }, 14: { 1, 12 }, 19: { 1, 12 }, 2: { 1, 12 }, 13: { 1, 12 }, 4: { 1, 12 }, 18: { 1, 12 } },
    { 2: { 1, 9 }, 14: { 1, 9 }, 3: { 1, 9 }, 23: { 1, 9 }, 18: { 1, 9 }, 15: { 1, 9 }, 19: { 1, 9 }, 4: { 1, 9 }, 13: { 1, 9 } },
    { 4: { 1, 10 }, 18: { 1, 10 }, 14: { 1, 10 }, 2: { 1, 10 }, 19: { 1, 10 }, 13: { 1, 10 }, 3: { 1, 10 }, 23: { 1, 10 }, 15: { 1, 10 } },
}
var gotoTable = []map[int]int {
    { 0: 1, 3: 2 },
    { },
    { 1: 4 },
    { },
    { },
    { },
    { },
    { },
    { },
    { },
    { 2: 15, 9: 20, 10: 18, 7: 22, 8: 14 },
    { 7: 22, 8: 14, 9: 20, 10: 18, 2: 23 },
    { 7: 22, 2: 24, 9: 20, 8: 14, 10: 18 },
    { },
    { 10: 18, 9: 26 },
    { 4: 28 },
    { },
    { 2: 30, 7: 22, 10: 18, 9: 20, 8: 14 },
    { 6: 34 },
    { },
    { },
    { },
    { },
    { },
    { },
    { 10: 18, 9: 38 },
    { },
    { 7: 39, 8: 14, 9: 20, 10: 18 },
    { },
    { },
    { },
    { },
    { },
    { },
    { },
    { },
    { },
    { },
    { },
    { },
    { },
    { },
    { },
    { 5: 44 },
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
        case 0:
            // For shift actions, add new state to the stack along with token
            stack = append(stack, StackState { action.Value, token })
            token = p.lexer.Next()
        case 1:
            // For reduce actions, pop states off stack and merge children into one node based on production
            production := &productions[action.Value]
            i := len(stack) - production.Length
            var node ParseTreeChild
            switch production.Type {
            case 0:
                // Handle normal productions
                // Collect child nodes from current states on the stack and create node for reduction
                children := make([]ParseTreeChild, production.Length)
                for i, s := range stack[i:] { children[i] = s.node }
                node = &ParseTreeNode { production, children }
            case 2:
                // Handle flatten productions
                // Of the two nodes popped, preserve the first and add the second as a child of the first
                // Results in quantified expressions in the grammar generating arrays of elements
                list, element := stack[i].node.(*ParseTreeNode), stack[i + 1].node
                list.Children = append(list.Children, element)
                node = list
            case 1: node = stack[i].node // For auxiliary productions, pass child through without generating new node
            case 3: node = nil // Add nil value for removed productions
            }
            // Pop consumed states off stack
            // Given new state at the top of the stack, find next state based on the goto table
            stack = stack[:i]
            state := stack[i - 1].state
            next := gotoTable[state][production.Left]
            // Add new state to top of the stack
            stack = append(stack, StackState { next, node })
        // Return non-terminal in auxiliary start production on accept
        case 2: return stack[1].node.(*ParseTreeNode)
        }
    }
}

// Given a parse tree node, dispatches the corresponding function in the visitor.
func VisitNode[T any](visitor BaseVisitor[T], node *ParseTreeNode) T {
    switch node.data.Visitor {
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
    return fmt.Sprintf("%s[%s]%s", indent, n.data.Visitor, strings.Join(children, ""))
}
