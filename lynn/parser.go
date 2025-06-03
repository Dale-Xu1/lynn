package lynn

import (
	"fmt"
	"strings"
)

// Production data struct. Expresses a sequence of symbols that a given non-terminal may be expanded to in a grammar.
type productionData struct {
    productionType int
    left, length   int
    visitor        string
    aliases        map[string]int
}
// Parse table action entry struct. Holds action type and integer parameter.
type actionEntry struct {
    actionType, value int // For 0 actions, value represents a state identifier, for 1 actions, a production identifier
}

// Parse tree child interface. May either be a Token or ParseTreeNode struct.
type ParseTreeChild interface { string(indent string) string }
// Parse tree node struct. Contains child nodes and location range.
type ParseTreeNode struct {
    Children   []ParseTreeChild
    Start, End Location
    data       *productionData
}

var productions = []productionData {
    { 2, 3, 2, "", nil },
    { 0, 3, 0, "", nil },
    { 0, 0, 1, "grammar", map[string]int { "stmt": 0 } },
    { 0, 1, 5, "ruleStmt", map[string]int { "RULE": 0, "IDENTIFIER": 1, "expr": 3 } },
    { 0, 4, 2, "", map[string]int { "SKIP": 1 } },
    { 3, 4, 0, "", nil },
    { 0, 1, 6, "tokenStmt", map[string]int { "TOKEN": 0, "IDENTIFIER": 1, "expr": 3, "s": 4 } },
    { 0, 1, 5, "fragmentStmt", map[string]int { "FRAGMENT": 0, "IDENTIFIER": 1, "expr": 3 } },
    { 0, 1, 2, "stmt", nil },
    { 0, 2, 3, "unionExpr", map[string]int { "l": 0, "r": 2 } },
    { 1, 5, 1, "", nil },
    { 1, 5, 1, "", nil },
    { 3, 5, 0, "", nil },
    { 0, 7, 4, "labelExpr", map[string]int { "IDENTIFIER": 2, "a": 3, "expr": 0 } },
    { 0, 8, 2, "concatExpr", map[string]int { "l": 0, "r": 1 } },
    { 0, 9, 3, "aliasExpr", map[string]int { "IDENTIFIER": 0, "expr": 2 } },
    { 1, 6, 1, "", nil },
    { 1, 6, 1, "", nil },
    { 1, 6, 1, "", nil },
    { 0, 10, 2, "quantifierExpr", map[string]int { "op": 1, "expr": 0 } },
    { 0, 10, 3, "groupExpr", map[string]int { "expr": 1 } },
    { 0, 10, 1, "identifierExpr", map[string]int { "IDENTIFIER": 0 } },
    { 0, 10, 1, "stringExpr", map[string]int { "STRING": 0 } },
    { 0, 10, 1, "classExpr", map[string]int { "CLASS": 0 } },
    { 0, 10, 1, "errorExpr", map[string]int { "ERROR": 0 } },
    { 0, 10, 1, "anyExpr", nil },
    { 1, 2, 1, "", nil },
    { 1, 7, 1, "", nil },
    { 1, 8, 1, "", nil },
    { 1, 9, 1, "", nil },
}
var actionTable = []map[int]actionEntry {
    { 4: { 1, 1 }, 24: { 1, 1 }, 2: { 1, 1 }, 3: { 1, 1 }, -1: { 1, 1 } },
    { 3: { 0, 6 }, 2: { 0, 7 }, 24: { 1, 2 }, 4: { 0, 3 }, -1: { 0, 4 } },
    { 24: { 2, 0 } },
    { 21: { 0, 8 } },
    { 16: { 0, 9 } },
    { 2: { 1, 0 }, 3: { 1, 0 }, 4: { 1, 0 }, 24: { 1, 0 }, -1: { 1, 0 } },
    { 21: { 0, 10 } },
    { 21: { 0, 11 } },
    { 17: { 0, 12 } },
    { 3: { 1, 8 }, 2: { 1, 8 }, 24: { 1, 8 }, 4: { 1, 8 }, -1: { 1, 8 } },
    { 17: { 0, 13 } },
    { 17: { 0, 14 } },
    { 13: { 0, 23 }, 22: { 0, 17 }, 18: { 0, 19 }, 7: { 0, 22 }, 23: { 0, 16 }, 21: { 0, 21 } },
    { 21: { 0, 21 }, 7: { 0, 22 }, 23: { 0, 16 }, 13: { 0, 23 }, 22: { 0, 17 }, 18: { 0, 19 } },
    { 22: { 0, 17 }, 18: { 0, 19 }, 21: { 0, 21 }, 7: { 0, 22 }, 23: { 0, 16 }, 13: { 0, 23 } },
    { 24: { 1, 27 }, 22: { 0, 17 }, 18: { 0, 19 }, 7: { 0, 22 }, 13: { 0, 23 }, 21: { 0, 21 }, 16: { 1, 27 }, -1: { 1, 27 }, 15: { 1, 27 }, 3: { 1, 27 }, 14: { 1, 27 }, 20: { 1, 27 }, 2: { 1, 27 }, 19: { 1, 27 }, 23: { 0, 16 }, 4: { 1, 27 } },
    { 4: { 1, 23 }, 24: { 1, 23 }, 23: { 1, 23 }, 15: { 1, 23 }, -1: { 1, 23 }, 19: { 1, 23 }, 13: { 1, 23 }, 20: { 1, 23 }, 11: { 1, 23 }, 10: { 1, 23 }, 16: { 1, 23 }, 21: { 1, 23 }, 2: { 1, 23 }, 22: { 1, 23 }, 14: { 1, 23 }, 7: { 1, 23 }, 18: { 1, 23 }, 3: { 1, 23 }, 12: { 1, 23 } },
    { 3: { 1, 22 }, 18: { 1, 22 }, 12: { 1, 22 }, 21: { 1, 22 }, -1: { 1, 22 }, 19: { 1, 22 }, 14: { 1, 22 }, 16: { 1, 22 }, 24: { 1, 22 }, 11: { 1, 22 }, 2: { 1, 22 }, 4: { 1, 22 }, 20: { 1, 22 }, 23: { 1, 22 }, 13: { 1, 22 }, 22: { 1, 22 }, 7: { 1, 22 }, 10: { 1, 22 }, 15: { 1, 22 } },
    { 13: { 1, 28 }, 18: { 1, 28 }, 19: { 1, 28 }, 14: { 1, 28 }, 23: { 1, 28 }, 16: { 1, 28 }, 24: { 1, 28 }, 3: { 1, 28 }, 20: { 1, 28 }, 22: { 1, 28 }, 21: { 1, 28 }, -1: { 1, 28 }, 4: { 1, 28 }, 7: { 1, 28 }, 15: { 1, 28 }, 2: { 1, 28 } },
    { 7: { 0, 22 }, 22: { 0, 17 }, 18: { 0, 19 }, 13: { 0, 23 }, 23: { 0, 16 }, 21: { 0, 21 } },
    { 4: { 1, 29 }, 3: { 1, 29 }, 14: { 1, 29 }, 24: { 1, 29 }, 15: { 1, 29 }, 2: { 1, 29 }, 18: { 1, 29 }, 19: { 1, 29 }, 22: { 1, 29 }, 21: { 1, 29 }, 11: { 0, 31 }, 10: { 0, 32 }, 12: { 0, 33 }, 23: { 1, 29 }, 13: { 1, 29 }, 20: { 1, 29 }, 7: { 1, 29 }, 16: { 1, 29 }, -1: { 1, 29 } },
    { 4: { 1, 21 }, 9: { 0, 34 }, 19: { 1, 21 }, -1: { 1, 21 }, 11: { 1, 21 }, 20: { 1, 21 }, 15: { 1, 21 }, 18: { 1, 21 }, 13: { 1, 21 }, 21: { 1, 21 }, 16: { 1, 21 }, 10: { 1, 21 }, 23: { 1, 21 }, 7: { 1, 21 }, 24: { 1, 21 }, 14: { 1, 21 }, 12: { 1, 21 }, 2: { 1, 21 }, 3: { 1, 21 }, 22: { 1, 21 } },
    { 20: { 1, 24 }, 3: { 1, 24 }, 13: { 1, 24 }, 16: { 1, 24 }, 7: { 1, 24 }, -1: { 1, 24 }, 4: { 1, 24 }, 22: { 1, 24 }, 21: { 1, 24 }, 2: { 1, 24 }, 14: { 1, 24 }, 18: { 1, 24 }, 11: { 1, 24 }, 19: { 1, 24 }, 24: { 1, 24 }, 15: { 1, 24 }, 10: { 1, 24 }, 12: { 1, 24 }, 23: { 1, 24 } },
    { 7: { 1, 25 }, -1: { 1, 25 }, 10: { 1, 25 }, 22: { 1, 25 }, 24: { 1, 25 }, 14: { 1, 25 }, 11: { 1, 25 }, 12: { 1, 25 }, 16: { 1, 25 }, 21: { 1, 25 }, 4: { 1, 25 }, 19: { 1, 25 }, 15: { 1, 25 }, 23: { 1, 25 }, 18: { 1, 25 }, 20: { 1, 25 }, 13: { 1, 25 }, 2: { 1, 25 }, 3: { 1, 25 } },
    { 14: { 0, 35 }, 16: { 0, 36 } },
    { -1: { 1, 26 }, 19: { 1, 26 }, 20: { 1, 26 }, 4: { 1, 26 }, 15: { 0, 37 }, 2: { 1, 26 }, 3: { 1, 26 }, 16: { 1, 26 }, 14: { 1, 26 }, 24: { 1, 26 } },
    { 14: { 0, 35 }, 20: { 0, 39 }, 16: { 1, 5 } },
    { 14: { 0, 35 }, 16: { 0, 40 } },
    { 2: { 1, 14 }, 21: { 1, 14 }, 24: { 1, 14 }, 22: { 1, 14 }, 3: { 1, 14 }, 4: { 1, 14 }, 16: { 1, 14 }, 7: { 1, 14 }, 23: { 1, 14 }, 18: { 1, 14 }, -1: { 1, 14 }, 15: { 1, 14 }, 14: { 1, 14 }, 13: { 1, 14 }, 20: { 1, 14 }, 19: { 1, 14 } },
    { 14: { 0, 35 }, 19: { 0, 41 } },
    { 7: { 1, 19 }, 12: { 1, 19 }, 4: { 1, 19 }, 15: { 1, 19 }, 20: { 1, 19 }, 16: { 1, 19 }, 22: { 1, 19 }, 13: { 1, 19 }, 23: { 1, 19 }, 10: { 1, 19 }, 24: { 1, 19 }, 18: { 1, 19 }, 2: { 1, 19 }, 3: { 1, 19 }, -1: { 1, 19 }, 19: { 1, 19 }, 21: { 1, 19 }, 14: { 1, 19 }, 11: { 1, 19 } },
    { 13: { 1, 17 }, 3: { 1, 17 }, 14: { 1, 17 }, 7: { 1, 17 }, 10: { 1, 17 }, 4: { 1, 17 }, 15: { 1, 17 }, 22: { 1, 17 }, 24: { 1, 17 }, 12: { 1, 17 }, 2: { 1, 17 }, 18: { 1, 17 }, 23: { 1, 17 }, 16: { 1, 17 }, 11: { 1, 17 }, 21: { 1, 17 }, -1: { 1, 17 }, 19: { 1, 17 }, 20: { 1, 17 } },
    { 21: { 1, 18 }, 4: { 1, 18 }, 15: { 1, 18 }, 10: { 1, 18 }, 14: { 1, 18 }, 20: { 1, 18 }, 11: { 1, 18 }, 18: { 1, 18 }, 22: { 1, 18 }, 12: { 1, 18 }, -1: { 1, 18 }, 7: { 1, 18 }, 19: { 1, 18 }, 16: { 1, 18 }, 24: { 1, 18 }, 13: { 1, 18 }, 23: { 1, 18 }, 2: { 1, 18 }, 3: { 1, 18 } },
    { 16: { 1, 16 }, -1: { 1, 16 }, 23: { 1, 16 }, 15: { 1, 16 }, 13: { 1, 16 }, 24: { 1, 16 }, 12: { 1, 16 }, 18: { 1, 16 }, 10: { 1, 16 }, 7: { 1, 16 }, 4: { 1, 16 }, 20: { 1, 16 }, 19: { 1, 16 }, 2: { 1, 16 }, 3: { 1, 16 }, 14: { 1, 16 }, 22: { 1, 16 }, 21: { 1, 16 }, 11: { 1, 16 } },
    { 18: { 0, 19 }, 21: { 0, 21 }, 22: { 0, 17 }, 7: { 0, 22 }, 23: { 0, 16 }, 13: { 0, 23 } },
    { 23: { 0, 16 }, 13: { 0, 23 }, 22: { 0, 17 }, 7: { 0, 22 }, 21: { 0, 21 }, 18: { 0, 19 } },
    { -1: { 1, 7 }, 3: { 1, 7 }, 4: { 1, 7 }, 24: { 1, 7 }, 2: { 1, 7 } },
    { 21: { 0, 44 } },
    { 16: { 0, 45 } },
    { 8: { 0, 46 } },
    { 4: { 1, 3 }, 24: { 1, 3 }, 3: { 1, 3 }, -1: { 1, 3 }, 2: { 1, 3 } },
    { 14: { 1, 20 }, 18: { 1, 20 }, 2: { 1, 20 }, 20: { 1, 20 }, 19: { 1, 20 }, 24: { 1, 20 }, 15: { 1, 20 }, 13: { 1, 20 }, 4: { 1, 20 }, 10: { 1, 20 }, 16: { 1, 20 }, -1: { 1, 20 }, 21: { 1, 20 }, 22: { 1, 20 }, 3: { 1, 20 }, 11: { 1, 20 }, 23: { 1, 20 }, 7: { 1, 20 }, 12: { 1, 20 } },
    { 23: { 1, 15 }, 15: { 1, 15 }, 4: { 1, 15 }, -1: { 1, 15 }, 24: { 1, 15 }, 16: { 1, 15 }, 3: { 1, 15 }, 14: { 1, 15 }, 18: { 1, 15 }, 21: { 1, 15 }, 13: { 1, 15 }, 7: { 1, 15 }, 22: { 1, 15 }, 2: { 1, 15 }, 19: { 1, 15 }, 20: { 1, 15 } },
    { 14: { 1, 9 }, 19: { 1, 9 }, 2: { 1, 9 }, 20: { 1, 9 }, 15: { 0, 37 }, 24: { 1, 9 }, -1: { 1, 9 }, 4: { 1, 9 }, 16: { 1, 9 }, 3: { 1, 9 } },
    { 3: { 1, 12 }, 4: { 1, 12 }, 24: { 1, 12 }, 20: { 1, 12 }, 16: { 1, 12 }, -1: { 1, 12 }, 5: { 0, 49 }, 19: { 1, 12 }, 2: { 1, 12 }, 15: { 1, 12 }, 6: { 0, 47 }, 14: { 1, 12 } },
    { 4: { 1, 6 }, 2: { 1, 6 }, 24: { 1, 6 }, 3: { 1, 6 }, -1: { 1, 6 } },
    { 16: { 1, 4 } },
    { 3: { 1, 11 }, 24: { 1, 11 }, 20: { 1, 11 }, 15: { 1, 11 }, 4: { 1, 11 }, -1: { 1, 11 }, 16: { 1, 11 }, 19: { 1, 11 }, 14: { 1, 11 }, 2: { 1, 11 } },
    { 16: { 1, 13 }, 24: { 1, 13 }, 2: { 1, 13 }, 19: { 1, 13 }, 14: { 1, 13 }, 20: { 1, 13 }, 3: { 1, 13 }, 15: { 1, 13 }, 4: { 1, 13 }, -1: { 1, 13 } },
    { 14: { 1, 10 }, 2: { 1, 10 }, 19: { 1, 10 }, 16: { 1, 10 }, 20: { 1, 10 }, -1: { 1, 10 }, 3: { 1, 10 }, 15: { 1, 10 }, 4: { 1, 10 }, 24: { 1, 10 } },
}
var gotoTable = []map[int]int {
    { 3: 1, 0: 2 },
    { 1: 5 },
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
    { 9: 18, 8: 15, 10: 20, 2: 24, 7: 25 },
    { 8: 15, 7: 25, 2: 26, 10: 20, 9: 18 },
    { 9: 18, 2: 27, 10: 20, 8: 15, 7: 25 },
    { 9: 28, 10: 20 },
    { },
    { },
    { },
    { 2: 29, 10: 20, 7: 25, 9: 18, 8: 15 },
    { 6: 30 },
    { },
    { },
    { },
    { },
    { },
    { 4: 38 },
    { },
    { },
    { },
    { },
    { },
    { },
    { },
    { 10: 20, 9: 42 },
    { 8: 15, 9: 18, 7: 43, 10: 20 },
    { },
    { },
    { },
    { },
    { },
    { },
    { },
    { },
    { 5: 48 },
    { },
    { },
    { },
    { },
    { },
}

// Parser struct. Converts token stream to parse tree.
type Parser struct {
    lexer   BaseLexer
    handler ParserErrorHandler
}

// Base visitor struct. Describes functions necessary to implement to traverse parse tree.
type BaseVisitor[T any] interface {
    VisitGrammar(node *ParseTreeNode) T
    VisitRuleStmt(node *ParseTreeNode) T
    VisitTokenStmt(node *ParseTreeNode) T
    VisitFragmentStmt(node *ParseTreeNode) T
    VisitStmt(node *ParseTreeNode) T
    VisitUnionExpr(node *ParseTreeNode) T
    VisitLabelExpr(node *ParseTreeNode) T
    VisitConcatExpr(node *ParseTreeNode) T
    VisitAliasExpr(node *ParseTreeNode) T
    VisitQuantifierExpr(node *ParseTreeNode) T
    VisitGroupExpr(node *ParseTreeNode) T
    VisitIdentifierExpr(node *ParseTreeNode) T
    VisitStringExpr(node *ParseTreeNode) T
    VisitClassExpr(node *ParseTreeNode) T
    VisitErrorExpr(node *ParseTreeNode) T
    VisitAnyExpr(node *ParseTreeNode) T
}

// Function called when the parser encounters an error.
type ParserErrorHandler func (token Token)
var DEFAULT_PARSER_HANDLER = func (token Token) {
    fmt.Printf("Syntax error: Unexpected token %q - %d:%d\n", token.Value, token.Start.Line, token.Start.Col)
}

// Returns new parser struct.
func NewParser(lexer BaseLexer, handler ParserErrorHandler) *Parser { return &Parser { lexer, handler } }
// Generates parse tree based on token stream from lexer.
func (p *Parser) Parse() *ParseTreeNode {
    // Stack state struct. Holds the state identifier and the corresponding parse tree node.
    type StackState struct {
        state int
        node  ParseTreeChild
    }
    // Production and action type enums
    const (NORMAL int = iota; AUXILIARY; FLATTEN; REMOVED)
    const (SHIFT int = iota; REDUCE; ACCEPT)
    // Initialize current token and stack
    token, stack := p.lexer.Next(), []StackState { { 0, nil } }
    main: for {
        // Get the current state at the top of the stack and find the action to take
        // Next action is determined by action table given state index and the current token type
        state := stack[len(stack) - 1].state
        action, ok := actionTable[state][int(token.Type)]
        if !ok {
            // If the table does not have a valid action, cannot parse current token
            p.handler(token)
            for {
                // Pop states off the stack until a valid shift action on the error terminal is found
                if len(stack) == 0 { return nil }
                if action, ok := actionTable[state][-1]; ok && action.actionType == SHIFT {
                    // Shift token that caused error onto stack
                    // Then enter panic mode and read tokens until a valid action can be made
                    stack = append(stack, StackState { action.value, token })
                    for {
                        token = p.lexer.Next()
                        if _, ok := actionTable[action.value][int(token.Type)]; ok { continue main }
                    }
                }
                i := len(stack) - 1; stack = stack[:i]
                state = stack[i - 1].state
            }
        }
        switch action.actionType {
        case SHIFT:
            // For shift actions, add new state to the stack along with token
            stack = append(stack, StackState { action.value, token })
            token = p.lexer.Next()
        case REDUCE:
            // For reduce actions, pop states off stack and merge children into one node based on production
            production := &productions[action.value]
            i := len(stack) - production.length
            var node ParseTreeChild
            switch production.productionType {
            case NORMAL:
                // Handle normal productions
                // Collect child nodes from current states on the stack and create node for reduction
                children := make([]ParseTreeChild, production.length)
                for i, s := range stack[i:] { children[i] = s.node }
                // Find start and end locations
                start, end := findLocationRange(children)
                node = &ParseTreeNode { children, start, end, production }
            case FLATTEN:
                // Handle flatten productions
                // Of the two nodes popped, preserve the first and add the second as a child of the first
                // Results in quantified expressions in the grammar generating arrays of elements
                list, element := stack[i].node.(*ParseTreeNode), stack[i + 1].node
                list.Children = append(list.Children, element)
                switch n := element.(type) {
                case nil: continue
                case *ParseTreeNode: list.End = n.End
                case Token:          list.End = n.End
                }
                node = list
            case AUXILIARY: node = stack[i].node // For auxiliary productions, pass child through without generating new node
            case REMOVED: node = nil // Add nil value for removed productions
            }
            // Pop consumed states off stack
            // Given new state at the top of the stack, find next state based on the goto table
            stack = stack[:i]
            state := stack[i - 1].state
            next := gotoTable[state][production.left]
            // Add new state to top of the stack
            stack = append(stack, StackState { next, node })
        // Return non-terminal in auxiliary start production on accept
        case ACCEPT: return stack[1].node.(*ParseTreeNode)
        }
    }
}

// Given a list of children, find the location range that they occupy
func findLocationRange(children []ParseTreeChild) (Location, Location) {
    var start, end Location
    for _, c := range children {
        // The start location is determined by the start of the first non-nil child
        switch n := c.(type) {
        case nil: continue
        case *ParseTreeNode: start = n.Start
        case Token:          start = n.Start
        }
        break
    }
    for i := len(children) - 1; i >= 0; i-- {
        c := children[i]
        // The end location is determined by the end of the last non-nil child
        switch n := c.(type) {
        case nil: continue
        case *ParseTreeNode: end = n.End
        case Token:          end = n.End
        }
        break
    }
    return start, end
}

// Given a parse tree node, dispatches the corresponding function in the visitor.
func VisitNode[T any](visitor BaseVisitor[T], node ParseTreeChild) T {
    if n, ok := node.(*ParseTreeNode); ok {
        switch n.data.visitor {
        case "grammar": return visitor.VisitGrammar(n)
        case "ruleStmt": return visitor.VisitRuleStmt(n)
        case "tokenStmt": return visitor.VisitTokenStmt(n)
        case "fragmentStmt": return visitor.VisitFragmentStmt(n)
        case "stmt": return visitor.VisitStmt(n)
        case "unionExpr": return visitor.VisitUnionExpr(n)
        case "labelExpr": return visitor.VisitLabelExpr(n)
        case "concatExpr": return visitor.VisitConcatExpr(n)
        case "aliasExpr": return visitor.VisitAliasExpr(n)
        case "quantifierExpr": return visitor.VisitQuantifierExpr(n)
        case "groupExpr": return visitor.VisitGroupExpr(n)
        case "identifierExpr": return visitor.VisitIdentifierExpr(n)
        case "stringExpr": return visitor.VisitStringExpr(n)
        case "classExpr": return visitor.VisitClassExpr(n)
        case "errorExpr": return visitor.VisitErrorExpr(n)
        case "anyExpr": return visitor.VisitAnyExpr(n)
        }
    }
    panic("Invalid parse tree child passed to VisitNode()")
}

func (n *ParseTreeNode) Stmt() ParseTreeChild { return n.GetAlias("stmt") }
func (n *ParseTreeNode) RULE() ParseTreeChild { return n.GetAlias("RULE") }
func (n *ParseTreeNode) IDENTIFIER() ParseTreeChild { return n.GetAlias("IDENTIFIER") }
func (n *ParseTreeNode) Expr() ParseTreeChild { return n.GetAlias("expr") }
func (n *ParseTreeNode) SKIP() ParseTreeChild { return n.GetAlias("SKIP") }
func (n *ParseTreeNode) TOKEN() ParseTreeChild { return n.GetAlias("TOKEN") }
func (n *ParseTreeNode) S() ParseTreeChild { return n.GetAlias("s") }
func (n *ParseTreeNode) FRAGMENT() ParseTreeChild { return n.GetAlias("FRAGMENT") }
func (n *ParseTreeNode) L() ParseTreeChild { return n.GetAlias("l") }
func (n *ParseTreeNode) R() ParseTreeChild { return n.GetAlias("r") }
func (n *ParseTreeNode) A() ParseTreeChild { return n.GetAlias("a") }
func (n *ParseTreeNode) Op() ParseTreeChild { return n.GetAlias("op") }
func (n *ParseTreeNode) STRING() ParseTreeChild { return n.GetAlias("STRING") }
func (n *ParseTreeNode) CLASS() ParseTreeChild { return n.GetAlias("CLASS") }
func (n *ParseTreeNode) ERROR() ParseTreeChild { return n.GetAlias("ERROR") }

// Given an alias, return the corresponding parse tree node child based on the production data.
func (n *ParseTreeNode) GetAlias(alias string) ParseTreeChild {
    if n.data.aliases == nil { return nil }
    if i, ok := n.data.aliases[alias]; ok { return n.Children[i] }
    return nil
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
    return fmt.Sprintf("%s[%s]%s", indent, n.data.visitor, strings.Join(children, ""))
}
