package test

import "fmt"

type Symbol interface { isSymbol() }
type Terminal rune
type NonTerminal uint

type Production struct {
    Left  NonTerminal
    Right []Symbol
}

type LR1Item struct {
    Production Production
    Dot        int
    Lookahead  Terminal
}

type ItemCore struct {
    Production Production
    Dot        int
}

type LALRParser struct {
    terminals    []Terminal
    nonTerminals []NonTerminal
    productions  []Production
    start        NonTerminal

    first        map[Symbol]map[Terminal]struct{}
}

func NewLALRParser() *LALRParser {
    terminals := []Terminal { '+', '*', '(', ')', 'x' }
    nonTerminals := []NonTerminal { S, E, T, F }
    productions := []Production {
        { S, []Symbol { E } },
        { E, []Symbol { E, Terminal('+'), T } },
        { E, []Symbol { T } },
        { T, []Symbol { T, Terminal('*'), F } },
        { T, []Symbol { F } },
        { F, []Symbol { Terminal('('), E, Terminal(')') } },
        { F, []Symbol { Terminal('x') } },
    }
    return &LALRParser {
        terminals, nonTerminals,
        productions,
        E,
        make(map[Symbol]map[Terminal]struct{}),
    }
}

func (p *LALRParser) Parse() {
    p.findFirst()
    fmt.Println(p.first)
}

// Computes the FIRST sets of all symbols in the grammar provided by the parser.
func (p *LALRParser) findFirst() {
    // Initialize FIRST sets for all terminals and non-terminals
    // FIRST set of a terminal contains only itself, and FIRST sets of non-terminals are initialized empty
    for _, t := range p.terminals { p.first[t] = map[Terminal]struct{} { t: {} } }
    for _, t := range p.nonTerminals { p.first[t] = make(map[Terminal]struct{}) }
    // Iterative implementation, repeat procedure until no changes are made
    for changed := true; changed; {
        changed = false
        main: for _, production := range p.productions {
            left, right := production.Left, production.Right
            // The FIRST set of a non-terminal is found by going through its productions and taking the union of the FIRST sets
            // of subsequent symbols until a non-nullable symbol is found (FIRST set does not contain epsilon)
            for _, s := range right {
                if s == EPSILON { continue }
                for f := range p.first[s] {
                    // Add all elements (ignoring epsilon) in FIRST set of symbol to FIRST set of production LHS non-terminal
                    if f == EPSILON { continue }
                    if _, ok := p.first[left][f]; !ok {
                        p.first[left][f] = struct{}{}
                        changed = true
                    }
                }
                // All elements in FIRST set have been found if symbol is non-nullable
                if _, ok := p.first[s][EPSILON]; !ok { continue main }
            }
            // If all symbols in the production are nullable, the LHS non-terminal is also nullable
            if _, ok := p.first[left][EPSILON]; !ok {
                p.first[left][EPSILON] = struct{}{}
                changed = true
            }
        }
    }
}

// Computes the FOLLOW sets of all symbols in the grammar provided by the parser.
func (p *LALRParser) findFollow() {

}

const EPSILON Terminal = 'ε'
const END Terminal = '$'
const (S NonTerminal = iota; E; T; F)
var nonTerminalName = map[NonTerminal]string { S: "S", E: "E", T: "T", F: "F" }

func (t Terminal) isSymbol() { }
func (t NonTerminal) isSymbol() { }

func (t Terminal) String() string { return string(t) }
func (t NonTerminal) String() string { return nonTerminalName[t] }
