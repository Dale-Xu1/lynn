package lynn

import (
	"fmt"
	"slices"
)

type Symbol interface { isSymbol() }
type NonTerminal uint
type Terminal rune

type LL1Parser struct {
    rules  map[NonTerminal][][]Symbol
    start  NonTerminal
    first  map[NonTerminal]map[Terminal]struct{}
    follow map[NonTerminal]map[Terminal]struct{}
}

const EPSILON Terminal = '!'
const END Terminal = '$'

func NewLL1Parser() *LL1Parser {
    const (E NonTerminal = iota; R; T; Y; F)
    rules := map[NonTerminal][][]Symbol {
        E: {
            { T, R },
        },
        R: {
            { Terminal('+'), T, R },
            { EPSILON },
        },
        T: {
            { F, Y },
        },
        Y: {
            { Terminal('*'), F, Y },
            { EPSILON },
        },
        F: {
            { Terminal('x') },
            { Terminal('('), E, Terminal(')') },
        },
    }
    return &LL1Parser {
        rules,
        E,
        make(map[NonTerminal]map[Terminal]struct{}),
        make(map[NonTerminal]map[Terminal]struct{}),
    }
}

// TODO: Precedence and associativity disambiguation
// TODO: Left recursion elimination (indirect left recursion detection for error handling?)
// TODO: Left factoring

func (p *LL1Parser) Parse() {
    for t := range p.rules { p.findFirst(t) }
    for t := range p.rules { p.findFollow(t) }

    // TODO: Construct parse table

    nts := make([]NonTerminal, 0, len(p.rules))
    for nt := range p.rules { nts = append(nts, nt) }
    slices.Sort(nts)
    for _, nt := range nts {
        fmt.Println(nt, p.first[nt], p.follow[nt])
    }
}

func (p *LL1Parser) findFirst(symbol Symbol) map[Terminal]struct{} {
    switch t := symbol.(type) {
    case Terminal: return map[Terminal]struct{} { t: {} }
    case NonTerminal:
        if first, ok := p.first[t]; ok { return first }
        first := make(map[Terminal]struct{})
        p.first[t] = first
        for _, rule := range p.rules[t] { p.ruleFirst(first, rule) }
        return first
    default: panic("Invalid symbol passed to findFirst()")
    }
}

func (p *LL1Parser) ruleFirst(first map[Terminal]struct{}, rule []Symbol) {
    for _, s := range rule[:len(rule) - 1] {
        f := p.findFirst(s)
        for t := range f {
            if t != EPSILON { first[t] = struct{}{} }
        }
        if _, ok := f[EPSILON]; !ok { return }
    }
    for t := range p.findFirst(rule[len(rule) - 1]) {
        first[t] = struct{}{}
    }
}

func (p *LL1Parser) findFollow(symbol NonTerminal) map[Terminal]struct{} {
    if follow, ok := p.follow[symbol]; ok { return follow }
    follow := make(map[Terminal]struct{})
    p.follow[symbol] = follow
    if symbol == p.start { follow[END] = struct{}{} }
    for lhs, rules := range p.rules {
        for _, rule := range rules { p.ruleFollow(symbol, follow, lhs, rule) }
    }
    return follow
}

func (p *LL1Parser) ruleFollow(symbol NonTerminal, follow map[Terminal]struct{}, lhs NonTerminal, rule []Symbol) {
    main: for i, s := range rule {
        if s == symbol {
            for _, s := range rule[i + 1:] {
                f := p.findFirst(s)
                for t := range f {
                    if t != EPSILON { follow[t] = struct{}{} }
                }
                if _, ok := f[EPSILON]; !ok { continue main }
            }
            for t := range p.findFollow(lhs) { follow[t] = struct{}{} }
        }
    }
}

func (t NonTerminal) isSymbol() { }
func (t Terminal) isSymbol() { }

func (t Terminal) String() string { return string(t) }
