package lynn

import (
	"fmt"
	"os"
	"strings"
)

// Production data struct. Expresses a sequence of symbols that a given non-terminal may be expanded to in a grammar.
type productionData struct {
    productionType int
    left, length   int
    visitor        string
    aliases        map[string]int
}

// Parse table entry struct. Holds action entries and goto table for a specific state.
type tableEntry struct {
    actions map[int]actionEntry
    gotos   map[int]int
}
// Parse table action entry struct. Holds action type and integer parameter.
type actionEntry struct {
    actionType, value int // For shift actions, value represents a state identifier, for reduce actions, a production identifier
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
    { 0, 1, 5, "ruleStmt", map[string]int { "IDENTIFIER": 1, "expr": 3, "RULE": 0 } },
    { 1, 5, 1, "", nil },
    { 1, 5, 1, "", nil },
    { 0, 4, 2, "", map[string]int { "a": 1 } },
    { 3, 4, 0, "", nil },
    { 0, 1, 3, "precedenceStmt", map[string]int { "v": 2, "PRECEDENCE": 0, "IDENTIFIER": 1 } },
    { 0, 6, 2, "", map[string]int { "SKIP": 1 } },
    { 3, 6, 0, "", nil },
    { 0, 1, 6, "tokenStmt", map[string]int { "IDENTIFIER": 1, "expr": 3, "s": 4, "TOKEN": 0 } },
    { 0, 1, 5, "fragmentStmt", map[string]int { "FRAGMENT": 0, "IDENTIFIER": 1, "expr": 3 } },
    { 0, 1, 2, "stmt", nil },
    { 0, 2, 3, "unionExpr", map[string]int { "r": 2, "l": 0 } },
    { 0, 7, 2, "", map[string]int { "IDENTIFIER": 1 } },
    { 3, 7, 0, "", nil },
    { 0, 9, 4, "labelExpr", map[string]int { "expr": 0, "IDENTIFIER": 2, "p": 3 } },
    { 0, 10, 2, "concatExpr", map[string]int { "l": 0, "r": 1 } },
    { 0, 11, 3, "aliasExpr", map[string]int { "IDENTIFIER": 0, "expr": 2 } },
    { 1, 8, 1, "", nil },
    { 1, 8, 1, "", nil },
    { 1, 8, 1, "", nil },
    { 0, 12, 2, "quantifierExpr", map[string]int { "op": 1, "expr": 0 } },
    { 0, 12, 3, "groupExpr", map[string]int { "expr": 1 } },
    { 0, 12, 1, "identifierExpr", map[string]int { "IDENTIFIER": 0 } },
    { 0, 12, 1, "stringExpr", map[string]int { "STRING": 0 } },
    { 0, 12, 1, "classExpr", map[string]int { "CLASS": 0 } },
    { 0, 12, 1, "errorExpr", map[string]int { "ERROR": 0 } },
    { 0, 12, 1, "anyExpr", nil },
    { 1, 2, 1, "", nil },
    { 1, 9, 1, "", nil },
    { 1, 10, 1, "", nil },
    { 1, 11, 1, "", nil },
}
var parseTable = []tableEntry {
    { map[int]actionEntry { 4: { 1, 1 }, 5: { 1, 1 }, -1: { 1, 1 }, 26: { 1, 1 }, 3: { 1, 1 }, 2: { 1, 1 } }, map[int]int { 3: 2, 0: 1 } },
    { map[int]actionEntry { 26: { 2, 0 } }, map[int]int { } },
    { map[int]actionEntry { 4: { 0, 6 }, 3: { 0, 7 }, -1: { 0, 8 }, 2: { 0, 3 }, 26: { 1, 2 }, 5: { 0, 5 } }, map[int]int { 1: 4 } },
    { map[int]actionEntry { 23: { 0, 9 } }, map[int]int { } },
    { map[int]actionEntry { 2: { 1, 0 }, 3: { 1, 0 }, 5: { 1, 0 }, -1: { 1, 0 }, 4: { 1, 0 }, 26: { 1, 0 } }, map[int]int { } },
    { map[int]actionEntry { 23: { 0, 10 } }, map[int]int { } },
    { map[int]actionEntry { 23: { 0, 11 } }, map[int]int { } },
    { map[int]actionEntry { 23: { 0, 12 } }, map[int]int { } },
    { map[int]actionEntry { 18: { 0, 13 } }, map[int]int { } },
    { map[int]actionEntry { 19: { 0, 14 } }, map[int]int { } },
    { map[int]actionEntry { 19: { 0, 15 } }, map[int]int { } },
    { map[int]actionEntry { 19: { 0, 16 } }, map[int]int { } },
    { map[int]actionEntry { 26: { 1, 7 }, 2: { 1, 7 }, -1: { 1, 7 }, 19: { 0, 18 }, 3: { 1, 7 }, 5: { 1, 7 }, 4: { 1, 7 } }, map[int]int { 4: 17 } },
    { map[int]actionEntry { 5: { 1, 13 }, 3: { 1, 13 }, 4: { 1, 13 }, 26: { 1, 13 }, 2: { 1, 13 }, -1: { 1, 13 } }, map[int]int { } },
    { map[int]actionEntry { 23: { 0, 25 }, 14: { 0, 29 }, 8: { 0, 21 }, 24: { 0, 22 }, 25: { 0, 28 }, 20: { 0, 24 } }, map[int]int { 2: 27, 11: 23, 12: 19, 9: 20, 10: 26 } },
    { map[int]actionEntry { 14: { 0, 29 }, 8: { 0, 21 }, 23: { 0, 25 }, 25: { 0, 28 }, 24: { 0, 22 }, 20: { 0, 24 } }, map[int]int { 11: 23, 10: 26, 12: 19, 2: 30, 9: 20 } },
    { map[int]actionEntry { 25: { 0, 28 }, 14: { 0, 29 }, 20: { 0, 24 }, 24: { 0, 22 }, 8: { 0, 21 }, 23: { 0, 25 } }, map[int]int { 2: 31, 11: 23, 10: 26, 12: 19, 9: 20 } },
    { map[int]actionEntry { 3: { 1, 8 }, 4: { 1, 8 }, 2: { 1, 8 }, 5: { 1, 8 }, -1: { 1, 8 }, 26: { 1, 8 } }, map[int]int { } },
    { map[int]actionEntry { 6: { 0, 34 }, 7: { 0, 33 } }, map[int]int { 5: 32 } },
    { map[int]actionEntry { 8: { 1, 33 }, 4: { 1, 33 }, -1: { 1, 33 }, 14: { 1, 33 }, 2: { 1, 33 }, 22: { 1, 33 }, 26: { 1, 33 }, 25: { 1, 33 }, 11: { 0, 38 }, 13: { 0, 35 }, 24: { 1, 33 }, 21: { 1, 33 }, 15: { 1, 33 }, 5: { 1, 33 }, 16: { 1, 33 }, 3: { 1, 33 }, 23: { 1, 33 }, 18: { 1, 33 }, 20: { 1, 33 }, 12: { 0, 36 } }, map[int]int { 8: 37 } },
    { map[int]actionEntry { 16: { 0, 39 }, -1: { 1, 30 }, 22: { 1, 30 }, 21: { 1, 30 }, 2: { 1, 30 }, 18: { 1, 30 }, 26: { 1, 30 }, 15: { 1, 30 }, 3: { 1, 30 }, 4: { 1, 30 }, 5: { 1, 30 } }, map[int]int { } },
    { map[int]actionEntry { 26: { 1, 28 }, 3: { 1, 28 }, -1: { 1, 28 }, 11: { 1, 28 }, 14: { 1, 28 }, 22: { 1, 28 }, 4: { 1, 28 }, 5: { 1, 28 }, 25: { 1, 28 }, 20: { 1, 28 }, 18: { 1, 28 }, 23: { 1, 28 }, 8: { 1, 28 }, 2: { 1, 28 }, 15: { 1, 28 }, 21: { 1, 28 }, 16: { 1, 28 }, 13: { 1, 28 }, 24: { 1, 28 }, 12: { 1, 28 } }, map[int]int { } },
    { map[int]actionEntry { 21: { 1, 26 }, 8: { 1, 26 }, 26: { 1, 26 }, 18: { 1, 26 }, 23: { 1, 26 }, 20: { 1, 26 }, 25: { 1, 26 }, 5: { 1, 26 }, 3: { 1, 26 }, -1: { 1, 26 }, 4: { 1, 26 }, 24: { 1, 26 }, 11: { 1, 26 }, 22: { 1, 26 }, 13: { 1, 26 }, 12: { 1, 26 }, 15: { 1, 26 }, 14: { 1, 26 }, 2: { 1, 26 }, 16: { 1, 26 } }, map[int]int { } },
    { map[int]actionEntry { 2: { 1, 32 }, 14: { 1, 32 }, 25: { 1, 32 }, 18: { 1, 32 }, -1: { 1, 32 }, 26: { 1, 32 }, 24: { 1, 32 }, 22: { 1, 32 }, 21: { 1, 32 }, 23: { 1, 32 }, 20: { 1, 32 }, 15: { 1, 32 }, 8: { 1, 32 }, 4: { 1, 32 }, 5: { 1, 32 }, 16: { 1, 32 }, 3: { 1, 32 } }, map[int]int { } },
    { map[int]actionEntry { 23: { 0, 25 }, 24: { 0, 22 }, 25: { 0, 28 }, 8: { 0, 21 }, 20: { 0, 24 }, 14: { 0, 29 } }, map[int]int { 9: 20, 12: 19, 2: 40, 10: 26, 11: 23 } },
    { map[int]actionEntry { 13: { 1, 25 }, 16: { 1, 25 }, 2: { 1, 25 }, 26: { 1, 25 }, 21: { 1, 25 }, 14: { 1, 25 }, 25: { 1, 25 }, 5: { 1, 25 }, 4: { 1, 25 }, 15: { 1, 25 }, -1: { 1, 25 }, 10: { 0, 41 }, 18: { 1, 25 }, 3: { 1, 25 }, 12: { 1, 25 }, 8: { 1, 25 }, 22: { 1, 25 }, 23: { 1, 25 }, 24: { 1, 25 }, 20: { 1, 25 }, 11: { 1, 25 } }, map[int]int { } },
    { map[int]actionEntry { 25: { 0, 28 }, 8: { 0, 21 }, 23: { 0, 25 }, 5: { 1, 31 }, 21: { 1, 31 }, 16: { 1, 31 }, 14: { 0, 29 }, 3: { 1, 31 }, 26: { 1, 31 }, 22: { 1, 31 }, 2: { 1, 31 }, 24: { 0, 22 }, -1: { 1, 31 }, 4: { 1, 31 }, 20: { 0, 24 }, 18: { 1, 31 }, 15: { 1, 31 } }, map[int]int { 11: 42, 12: 19 } },
    { map[int]actionEntry { 15: { 0, 43 }, 18: { 0, 44 } }, map[int]int { } },
    { map[int]actionEntry { 3: { 1, 27 }, 12: { 1, 27 }, 24: { 1, 27 }, 22: { 1, 27 }, -1: { 1, 27 }, 8: { 1, 27 }, 20: { 1, 27 }, 18: { 1, 27 }, 14: { 1, 27 }, 21: { 1, 27 }, 5: { 1, 27 }, 13: { 1, 27 }, 4: { 1, 27 }, 2: { 1, 27 }, 15: { 1, 27 }, 16: { 1, 27 }, 25: { 1, 27 }, 11: { 1, 27 }, 26: { 1, 27 }, 23: { 1, 27 } }, map[int]int { } },
    { map[int]actionEntry { 26: { 1, 29 }, 18: { 1, 29 }, 13: { 1, 29 }, 16: { 1, 29 }, 22: { 1, 29 }, 15: { 1, 29 }, 3: { 1, 29 }, 2: { 1, 29 }, 12: { 1, 29 }, 20: { 1, 29 }, 21: { 1, 29 }, 25: { 1, 29 }, 23: { 1, 29 }, 4: { 1, 29 }, 11: { 1, 29 }, 24: { 1, 29 }, 8: { 1, 29 }, 14: { 1, 29 }, 5: { 1, 29 }, -1: { 1, 29 } }, map[int]int { } },
    { map[int]actionEntry { 15: { 0, 43 }, 18: { 0, 45 } }, map[int]int { } },
    { map[int]actionEntry { 15: { 0, 43 }, 22: { 0, 46 }, 18: { 1, 10 } }, map[int]int { 6: 47 } },
    { map[int]actionEntry { 26: { 1, 6 }, 3: { 1, 6 }, 2: { 1, 6 }, -1: { 1, 6 }, 5: { 1, 6 }, 4: { 1, 6 } }, map[int]int { } },
    { map[int]actionEntry { 2: { 1, 5 }, 26: { 1, 5 }, 3: { 1, 5 }, 5: { 1, 5 }, 4: { 1, 5 }, -1: { 1, 5 } }, map[int]int { } },
    { map[int]actionEntry { 2: { 1, 4 }, 26: { 1, 4 }, 3: { 1, 4 }, 5: { 1, 4 }, 4: { 1, 4 }, -1: { 1, 4 } }, map[int]int { } },
    { map[int]actionEntry { 12: { 1, 20 }, 25: { 1, 20 }, 15: { 1, 20 }, -1: { 1, 20 }, 23: { 1, 20 }, 22: { 1, 20 }, 16: { 1, 20 }, 14: { 1, 20 }, 2: { 1, 20 }, 24: { 1, 20 }, 3: { 1, 20 }, 26: { 1, 20 }, 18: { 1, 20 }, 11: { 1, 20 }, 5: { 1, 20 }, 21: { 1, 20 }, 20: { 1, 20 }, 8: { 1, 20 }, 4: { 1, 20 }, 13: { 1, 20 } }, map[int]int { } },
    { map[int]actionEntry { 18: { 1, 21 }, 23: { 1, 21 }, 26: { 1, 21 }, 15: { 1, 21 }, 3: { 1, 21 }, 12: { 1, 21 }, 20: { 1, 21 }, 4: { 1, 21 }, 13: { 1, 21 }, 8: { 1, 21 }, 25: { 1, 21 }, -1: { 1, 21 }, 22: { 1, 21 }, 2: { 1, 21 }, 14: { 1, 21 }, 11: { 1, 21 }, 5: { 1, 21 }, 24: { 1, 21 }, 16: { 1, 21 }, 21: { 1, 21 } }, map[int]int { } },
    { map[int]actionEntry { 23: { 1, 23 }, 21: { 1, 23 }, 5: { 1, 23 }, 22: { 1, 23 }, 4: { 1, 23 }, 3: { 1, 23 }, 24: { 1, 23 }, 12: { 1, 23 }, 20: { 1, 23 }, 8: { 1, 23 }, 16: { 1, 23 }, 2: { 1, 23 }, 15: { 1, 23 }, 14: { 1, 23 }, 26: { 1, 23 }, 11: { 1, 23 }, -1: { 1, 23 }, 13: { 1, 23 }, 25: { 1, 23 }, 18: { 1, 23 } }, map[int]int { } },
    { map[int]actionEntry { 3: { 1, 22 }, 23: { 1, 22 }, 21: { 1, 22 }, -1: { 1, 22 }, 14: { 1, 22 }, 2: { 1, 22 }, 11: { 1, 22 }, 15: { 1, 22 }, 24: { 1, 22 }, 8: { 1, 22 }, 13: { 1, 22 }, 16: { 1, 22 }, 4: { 1, 22 }, 26: { 1, 22 }, 22: { 1, 22 }, 25: { 1, 22 }, 20: { 1, 22 }, 18: { 1, 22 }, 12: { 1, 22 }, 5: { 1, 22 } }, map[int]int { } },
    { map[int]actionEntry { 23: { 0, 48 } }, map[int]int { } },
    { map[int]actionEntry { 15: { 0, 43 }, 21: { 0, 49 } }, map[int]int { } },
    { map[int]actionEntry { 24: { 0, 22 }, 8: { 0, 21 }, 14: { 0, 29 }, 25: { 0, 28 }, 20: { 0, 24 }, 23: { 0, 25 } }, map[int]int { 12: 19, 11: 50 } },
    { map[int]actionEntry { 20: { 1, 18 }, 5: { 1, 18 }, 8: { 1, 18 }, 25: { 1, 18 }, 2: { 1, 18 }, 23: { 1, 18 }, 26: { 1, 18 }, -1: { 1, 18 }, 24: { 1, 18 }, 16: { 1, 18 }, 15: { 1, 18 }, 18: { 1, 18 }, 14: { 1, 18 }, 4: { 1, 18 }, 22: { 1, 18 }, 21: { 1, 18 }, 3: { 1, 18 } }, map[int]int { } },
    { map[int]actionEntry { 20: { 0, 24 }, 8: { 0, 21 }, 23: { 0, 25 }, 24: { 0, 22 }, 25: { 0, 28 }, 14: { 0, 29 } }, map[int]int { 11: 23, 12: 19, 10: 26, 9: 51 } },
    { map[int]actionEntry { -1: { 1, 3 }, 3: { 1, 3 }, 2: { 1, 3 }, 5: { 1, 3 }, 4: { 1, 3 }, 26: { 1, 3 } }, map[int]int { } },
    { map[int]actionEntry { -1: { 1, 12 }, 5: { 1, 12 }, 3: { 1, 12 }, 2: { 1, 12 }, 26: { 1, 12 }, 4: { 1, 12 } }, map[int]int { } },
    { map[int]actionEntry { 9: { 0, 52 } }, map[int]int { } },
    { map[int]actionEntry { 18: { 0, 53 } }, map[int]int { } },
    { map[int]actionEntry { 15: { 1, 16 }, 5: { 1, 16 }, 3: { 1, 16 }, 22: { 1, 16 }, 26: { 1, 16 }, 21: { 1, 16 }, 18: { 1, 16 }, 4: { 1, 16 }, 17: { 0, 54 }, 16: { 1, 16 }, -1: { 1, 16 }, 2: { 1, 16 } }, map[int]int { 7: 55 } },
    { map[int]actionEntry { 11: { 1, 24 }, 18: { 1, 24 }, 24: { 1, 24 }, 23: { 1, 24 }, 4: { 1, 24 }, 5: { 1, 24 }, 14: { 1, 24 }, 16: { 1, 24 }, 3: { 1, 24 }, 13: { 1, 24 }, 8: { 1, 24 }, 22: { 1, 24 }, 26: { 1, 24 }, 21: { 1, 24 }, 20: { 1, 24 }, 15: { 1, 24 }, -1: { 1, 24 }, 2: { 1, 24 }, 12: { 1, 24 }, 25: { 1, 24 } }, map[int]int { } },
    { map[int]actionEntry { 5: { 1, 19 }, 14: { 1, 19 }, 15: { 1, 19 }, 8: { 1, 19 }, 22: { 1, 19 }, 2: { 1, 19 }, 26: { 1, 19 }, 24: { 1, 19 }, 21: { 1, 19 }, 23: { 1, 19 }, 16: { 1, 19 }, 20: { 1, 19 }, 3: { 1, 19 }, -1: { 1, 19 }, 18: { 1, 19 }, 25: { 1, 19 }, 4: { 1, 19 } }, map[int]int { } },
    { map[int]actionEntry { 16: { 0, 39 }, 22: { 1, 14 }, 5: { 1, 14 }, 3: { 1, 14 }, 18: { 1, 14 }, 26: { 1, 14 }, -1: { 1, 14 }, 2: { 1, 14 }, 21: { 1, 14 }, 4: { 1, 14 }, 15: { 1, 14 } }, map[int]int { } },
    { map[int]actionEntry { 18: { 1, 9 } }, map[int]int { } },
    { map[int]actionEntry { 26: { 1, 11 }, 2: { 1, 11 }, -1: { 1, 11 }, 4: { 1, 11 }, 3: { 1, 11 }, 5: { 1, 11 } }, map[int]int { } },
    { map[int]actionEntry { 23: { 0, 56 } }, map[int]int { } },
    { map[int]actionEntry { 21: { 1, 17 }, 18: { 1, 17 }, 15: { 1, 17 }, 5: { 1, 17 }, -1: { 1, 17 }, 22: { 1, 17 }, 4: { 1, 17 }, 3: { 1, 17 }, 2: { 1, 17 }, 16: { 1, 17 }, 26: { 1, 17 } }, map[int]int { } },
    { map[int]actionEntry { 5: { 1, 15 }, 21: { 1, 15 }, 16: { 1, 15 }, 18: { 1, 15 }, 2: { 1, 15 }, 3: { 1, 15 }, 26: { 1, 15 }, 15: { 1, 15 }, 4: { 1, 15 }, 22: { 1, 15 }, -1: { 1, 15 } }, map[int]int { } },
}

// Parser struct. Converts token stream to parse tree.
type Parser struct {
    lexer   BaseLexer
    handler ParserErrorHandler
}

// Base visitor interface. Describes functions necessary to implement to traverse parse tree.
type BaseVisitor[T any] interface {
    VisitGrammar(node *ParseTreeNode) T
    VisitRuleStmt(node *ParseTreeNode) T
    VisitPrecedenceStmt(node *ParseTreeNode) T
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
    fmt.Fprintf(os.Stderr, "Syntax error: Unexpected token %q - %d:%d\n", token.Value, token.Start.Line, token.Start.Col)
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
        action, ok := parseTable[state].actions[int(token.Type)]
        if !ok {
            // If the table does not have a valid action, cannot parse current token
            p.handler(token)
            for {
                // Pop states off the stack until a valid shift action on the error terminal is found
                if len(stack) == 0 { return nil }
                if action, ok := parseTable[state].actions[-1]; ok && action.actionType == SHIFT {
                    // Shift token that caused error onto stack
                    // Then enter panic mode and read tokens until a valid action can be made
                    stack = append(stack, StackState { action.value, token })
                    for {
                        token = p.lexer.Next()
                        if _, ok := parseTable[action.value].actions[int(token.Type)]; ok { continue main }
                        if token.Type == EOF { return nil }
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
                case *ParseTreeNode: list.End = n.End
                case Token:          list.End = n.End
                }
                node = list
            case AUXILIARY: node = stack[i].node // For auxiliary productions, pass child through without generating new node
            case REMOVED:   node = nil // Add nil value for removed productions
            }
            // Pop consumed states off stack
            // Given new state at the top of the stack, find next state based on the goto table
            stack = stack[:i]
            state := stack[i - 1].state
            next := parseTable[state].gotos[production.left]
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
        case "precedenceStmt": return visitor.VisitPrecedenceStmt(n)
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
func (n *ParseTreeNode) IDENTIFIER() ParseTreeChild { return n.GetAlias("IDENTIFIER") }
func (n *ParseTreeNode) Expr() ParseTreeChild { return n.GetAlias("expr") }
func (n *ParseTreeNode) RULE() ParseTreeChild { return n.GetAlias("RULE") }
func (n *ParseTreeNode) A() ParseTreeChild { return n.GetAlias("a") }
func (n *ParseTreeNode) V() ParseTreeChild { return n.GetAlias("v") }
func (n *ParseTreeNode) PRECEDENCE() ParseTreeChild { return n.GetAlias("PRECEDENCE") }
func (n *ParseTreeNode) SKIP() ParseTreeChild { return n.GetAlias("SKIP") }
func (n *ParseTreeNode) S() ParseTreeChild { return n.GetAlias("s") }
func (n *ParseTreeNode) TOKEN() ParseTreeChild { return n.GetAlias("TOKEN") }
func (n *ParseTreeNode) FRAGMENT() ParseTreeChild { return n.GetAlias("FRAGMENT") }
func (n *ParseTreeNode) R() ParseTreeChild { return n.GetAlias("r") }
func (n *ParseTreeNode) L() ParseTreeChild { return n.GetAlias("l") }
func (n *ParseTreeNode) P() ParseTreeChild { return n.GetAlias("p") }
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
