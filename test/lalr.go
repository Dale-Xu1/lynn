package test

import (
	"fmt"
	"sort"
	"unsafe"
)

type Symbol interface { isSymbol() }
type Terminal rune
type NonTerminal uint

type Production struct {
    Left  NonTerminal
    Right []Symbol
}

type LR1Item struct {
    Production *Production
    Dot        int
    Lookahead  Terminal
}

type ItemCore struct {
    Production *Production
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
        S,
        make(map[Symbol]map[Terminal]struct{}),
    }
}

func (p *LALRParser) Parse() {
    p.findFirst()

    init := LR1Item { &p.productions[0], 0, END }
    closure := p.findClosure(map[LR1Item]struct{} { init: {} })
    fmt.Println("== CLOSURE ==")
    fmt.Println(getItemStateKey(closure))
    copy := make(map[LR1Item]struct{}, len(closure))
    for item := range closure {
        fmt.Println(item.Production.Left, item.Production.Right, item.Dot, item.Lookahead)
        copy[item] = struct{}{}
    }

    set := p.findGoto(closure, E)
    fmt.Println("== GOTO ==")
    fmt.Println(getItemStateKey(set))
    for item := range set {
        fmt.Println(item.Production.Left, item.Production.Right, item.Dot, item.Lookahead)
    }
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
        for _, production := range p.productions {
            changed = p.findRuleFirst(production.Left, production.Right) || changed
        }
    }
}

func (p *LALRParser) findRuleFirst(left NonTerminal, rule []Symbol) bool {
    // The FIRST set of a non-terminal is found by going through its productions and taking the union of the FIRST sets
    // of subsequent symbols until a non-nullable symbol is found (FIRST set does not contain epsilon)
    changed := false
    for _, s := range rule {
        if s == EPSILON { continue }
        for f := range p.first[s] {
            // Add all elements (ignoring epsilon) in FIRST set of symbol to FIRST set of production LHS non-terminal
            if f == EPSILON { continue }
            if _, ok := p.first[left][f]; !ok { p.first[left][f] = struct{}{}; changed = true }
        }
        // All elements in FIRST set have been found if symbol is non-nullable
        if _, ok := p.first[s][EPSILON]; !ok { return changed }
    }
    // If all symbols in the production are nullable, the LHS non-terminal is also nullable
    if _, ok := p.first[left][EPSILON]; !ok { p.first[left][EPSILON] = struct{}{}; changed = true }
    return changed
}

func (p *LALRParser) findSequenceFirst(sequence []Symbol) map[Terminal]struct{} {
    first := make(map[Terminal]struct{})
    for _, s := range sequence {
        if s == EPSILON { continue }
        for f := range p.first[s] {
            // Add all elements (ignoring epsilon) in FIRST set of symbol to set
            if f != EPSILON { first[f] = struct{}{} }
        }
        // All elements in FIRST set have been found if symbol is non-nullable
        if _, ok := p.first[s][EPSILON]; !ok { return first }
    }
    // If all symbols in the production are nullable, the sequence is also nullable
    first[EPSILON] = struct{}{}
    return first
}

// Computes the closure set of a set of LR(1) items.
func (p *LALRParser) findClosure(items map[LR1Item]struct{}) map[LR1Item]struct{} {
    // Transfer items in set to closure set and work list
    closure := make(map[LR1Item]struct{}, len(items))
    work := make([]LR1Item, 0, len(items))
    for item := range items { closure[item] = struct{}{}; work = append(work, item) }
    // Iterative implementation, process items in work list until no new items are added
    for index := 0; index < len(work); {
        item := work[index]; index++
        // Ensure that LR(1) item still has symbols to shift and that the current symbol is a non-terminal
        right := item.Production.Right
        if item.Dot >= len(right) { continue }
        t, ok := right[item.Dot].(NonTerminal)
        if !ok { continue }
        // Find first first set of beta sequence, if sequence is nullable, add lookahead to set
        first := p.findSequenceFirst(right[item.Dot + 1:])
        if _, ok := first[EPSILON]; ok { delete(first, EPSILON); first[item.Lookahead] = struct{}{} }
        // Look for all productions with the given non terminal as its LHS
        for i := range p.productions {
            production := &p.productions[i]
            if production.Left != t { continue }
            // Add productions to the work list, starting dot at the start and create copies for each terminal in the first set
            for lookahead := range first {
                i := LR1Item { production, 0, lookahead }
                if _, ok := closure[i]; !ok { closure[i] = struct{}{}; work = append(work, i) }
            }
        }
    }
    return closure
}

// Computes the goto set of a given item and symbol to transition on.
func (p *LALRParser) findGoto(items map[LR1Item]struct{}, symbol Symbol) map[LR1Item]struct{} {
    set := make(map[LR1Item]struct{})
    for item := range items {
        // Find LR(1) items in set where the next symbol matches the one passed to the function
        right := item.Production.Right
        if item.Dot < len(right) && right[item.Dot] == symbol {
            i := LR1Item { item.Production, item.Dot + 1, item.Lookahead }
            set[i] = struct{}{}
        }
    }
    if len(set) == 0 { return set }
    return p.findClosure(set)
}

// Creates unique identifier string given a set of LR(1) items for use in a map.
func getItemStateKey(items map[LR1Item]struct{}) string {
    // Sort states by address to ensure identical sets map to the same key
    list := make([]LR1Item, 0, len(items))
    for item := range items { list = append(list, item) }
    sort.Slice(list, func (i, j int) bool {
        a, b := list[i], list[j]
        if a.Production != b.Production { return uintptr(unsafe.Pointer(a.Production)) < uintptr(unsafe.Pointer(b.Production)) }
        if a.Dot != b.Dot { return a.Dot < b.Dot }
        return a.Lookahead < b.Lookahead
    })
    // Interpret state memory addresses as consecutive bytes, then reinterpret as string
    const ITEM_SIZE int = int(unsafe.Sizeof(LR1Item { }))
    bytes := make([]byte, 0, len(list) * ITEM_SIZE)
    for _, item := range list {
        b := (*[ITEM_SIZE]byte)(unsafe.Pointer(&item))
        bytes = append(bytes, b[:]...)
    }
    return string(bytes)
}

// TODO: buildLR1States()
func (p *LALRParser) buildLR1States() {
}

// TODO: buildLALRStates()

const (EPSILON Terminal = 'ε'; END = '$')
const (S NonTerminal = iota; E; T; F)
var nonTerminalName = map[NonTerminal]string { S: "S", E: "E", T: "T", F: "F" }

func (t Terminal) isSymbol() { }
func (t NonTerminal) isSymbol() { }

func (t Terminal) String() string { return string(t) }
func (t NonTerminal) String() string { return nonTerminalName[t] }
