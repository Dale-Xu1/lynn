package lynn

import (
	"fmt"
	"strings"
)

// Symbol interface. Can either be a Terminal or NonTerminal struct.
type Symbol interface { String(grammar *Grammar) string }
// Terminal type. Represented by string identifier.
type Terminal string
const (EPSILON Terminal = ""; EOF_TERMINAL = "EOF")
// Non-terminal type. Represented by integer index.
type NonTerminal int

// Grammar struct. Tracks all terminals and non-terminals, the start non-terminal, and all production rules.
type Grammar struct {
    Terminals    []Terminal
    NonTerminals map[NonTerminal]string
    Start        NonTerminal
    Productions  []Production
}
// Production struct. Expresses a sequence of symbols that a given non-terminal may be expanded to in a grammar.
type Production struct { Left NonTerminal; Right []Symbol }

// TODO: Conversion from AST to grammar
// TODO: Temporary function, delete
func NewTestGrammar() *Grammar {
    return &Grammar {
        []Terminal { "+", "*", "(", ")", "id", "EOF" },
        map[NonTerminal]string { 0: "E", 1: "T", 2: "F" },
        0,
        []Production {
            { 0, []Symbol { NonTerminal(0), Terminal("+"), NonTerminal(1) } },
            { 0, []Symbol { NonTerminal(1) } },
            { 1, []Symbol { NonTerminal(1), Terminal("*"), NonTerminal(2) } },
            { 1, []Symbol { NonTerminal(2) } },
            { 2, []Symbol { Terminal("("), NonTerminal(0), Terminal(")") } },
            { 2, []Symbol { Terminal("id") } },
        },
    }
}

// TODO: Precedence and associativity ambiguity elimination
func (g *Grammar) RemoveAmbiguities() {
	productions := make(map[NonTerminal][]Production, len(g.NonTerminals))
	modified := make([]Production, 0, len(g.Productions))
	// Group productions together based on their non-terminal
	for _, p := range g.Productions { productions[p.Left] = append(productions[p.Left], p) }
	for t, group := range productions {
		a, r := make([]Production, 0), make([]Production, 0)
		for _, p := range group {
			if len(p.Right) >= 2 && p.Right[0] == t && p.Right[len(p.Right) - 1] == t {
				a = append(a, p)
			} else {
				r = append(r, p)
			}
		}
	}
	_ = modified
}

// Augment grammar with new start state S'. Returns production for augmented start state.
func (g *Grammar) Augment() *Production {
    t := NonTerminal(len(g.NonTerminals))
    g.NonTerminals[t] = "S'"
    g.Productions = append(g.Productions, Production { t, []Symbol { g.Start } })
    return &g.Productions[len(g.Productions) - 1]
}

func (t Terminal)    String(grammar *Grammar) string { return string(t) }
func (t NonTerminal) String(grammar *Grammar) string { return grammar.NonTerminals[t] }
func (p Production)  String(grammar *Grammar) string {
    var builder strings.Builder
    builder.WriteString(fmt.Sprintf("%s ->", p.Left.String(grammar)))
    for _, s := range p.Right { builder.WriteString(fmt.Sprintf(" %s", s.String(grammar))) }
    return builder.String()
}

// FOR DEBUG PURPOSES:
// Prints all production rules of the grammar.
func (g *Grammar) PrintGrammar() {
	fmt.Printf("start: %s\n", g.Productions[g.Start].Left.String(g))
	for _, production := range g.Productions { fmt.Println(production.String(g)) }
}
