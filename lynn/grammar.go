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
func NewTestGrammar() *Grammar {
    return &Grammar {
        []Terminal { "+", "*", "(", ")", "id", "EOF" },
        map[NonTerminal]string { 0: "E" },
        0,
        []Production {
            { 0, []Symbol { NonTerminal(0), Terminal("+"), NonTerminal(0) } },
            { 0, []Symbol { NonTerminal(0), Terminal("*"), NonTerminal(0) } },
            { 0, []Symbol { Terminal("("), NonTerminal(0), Terminal(")") } },
            { 0, []Symbol { Terminal("id") } },
        },
    }
}

// Removes simple precedence and associativity ambiguities in the grammar.
// Not guaranteed to remove all ambiguities, but will resolve those of infix operations.
func (g *Grammar) RemoveAmbiguities() {
    productions := make(map[NonTerminal][]Production, len(g.NonTerminals))
    modified := make([]Production, 0, len(g.Productions))
    // Group productions together based on their non-terminal
    for _, p := range g.Productions { productions[p.Left] = append(productions[p.Left], p) }
    for t, group := range productions {
        // Split productions for a given non-terminal based on if there exists self-recursion at both ends of the production
        // Productions classified as ambiguous follow the form E -> E ... E
        a, r := make([]Production, 0), make([]Production, 0)
        for _, p := range group {
            if len(p.Right) >= 2 && p.Right[0] == t && p.Right[len(p.Right) - 1] == t {
                a = append(a, p)
            } else {
                r = append(r, p)
            }
        }
        // For each ambiguous production, replace with left or right-recursive form
        left := t
        for _, p := range a {
            // Create new auxiliary non-terminal
            next := NonTerminal(len(g.NonTerminals))
            g.NonTerminals[next] = fmt.Sprintf("%s'", g.NonTerminals[left]) // Use name of previous non-terminal with a '
            // Modify non-terminals at the ends
            // For left-associative productions,  E_k -> E_k ... E_{k + 1}
            // For right-associative productions, E_k -> E_{k + 1} ... E_k
            symbols := make([]Symbol, len(p.Right))
            copy(symbols[1:], p.Right[1:len(p.Right) - 1])
            symbols[0], symbols[len(symbols) - 1] = left, next
            // TODO: Allow grammar to specify associativity
            // symbols[0], symbols[len(symbols) - 1] = next, left
            // Add modified productions to new list
            modified = append(modified, Production { left, symbols })
            modified = append(modified, Production { left, []Symbol { next } })
            left = next
        }
        // For remaining productions, replace left non-terminal with new non-terminal of highest precedence
        for _, p := range r { modified = append(modified, Production { left, p.Right }) }
    }
    g.Productions = modified
}

// Augment grammar with new start state. Returns production for augmented start state.
func (g *Grammar) Augment() *Production {
    t := NonTerminal(len(g.NonTerminals))
    g.NonTerminals[t] = "_S"
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
