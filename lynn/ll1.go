package lynn

import "fmt"

type Symbol interface { isSymbol() }
type NonTerminal uint
type Terminal rune

type LL1Parser struct {
    rules map[NonTerminal][][]Symbol
    start NonTerminal
    first map[NonTerminal]map[Terminal]struct{}
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
    }
}

func (p *LL1Parser) Parse() {
    for nt := range p.rules {
        fmt.Println(nt, p.findFirst(nt), p.findFollow())
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

func (p *LL1Parser) findFollow() map[Terminal]struct{} {
    follow := make(map[Terminal]struct{})
    return follow
}

func (t NonTerminal) isSymbol() { }
func (t Terminal) isSymbol() { }

func (t Terminal) String() string { return string(t) }
