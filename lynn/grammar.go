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

// Lexer generator struct. Converts EBNF rule definitions to context-free grammar (CFG) production rules.
type GrammarGenerator struct {
}

// Returns a grammar generator struct.
func NewGrammarGenerator() *GrammarGenerator { return &GrammarGenerator { } }
// Converts EBNF rules defined in AST into CFG production rules.
func (g *GrammarGenerator) GenerateCFG(grammar *GrammarNode) *Grammar {
    // TODO: Conversion from AST to grammar
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

// TODO: When generating parse tree, remove auxiliary non-terminals

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

// ------------------------------------------------------------------------------------------------------------------------------
// TODO: Compile to generated parser program
// TODO: Print parse trees

type ShiftReduceParser struct {
	table LRParseTable
}

type StackState struct {
    State int
    Node  *ParseTreeNode
}

type ParseTreeNode struct {
    Symbol   Symbol
    Children []*ParseTreeNode
}

func NewShiftReduceParser(table LRParseTable) *ShiftReduceParser { return &ShiftReduceParser { table } }
func (p *ShiftReduceParser) Parse() *ParseTreeNode {
	input := []Terminal { "id", "+", "id", "*", "id", EOF_TERMINAL }
	ip := 0
	stack := []StackState { { 0, nil } }
	for {
		state, token := stack[len(stack) - 1].State, input[ip]
		action, ok := p.table.Action[state][token]
		if !ok {
            fmt.Printf("Syntax error: Unexpected token %s\n", token)
            return nil
        }
		switch action.Type {
		case SHIFT:
            // Create leaf node for terminal
            node := &ParseTreeNode { token, nil }
            // Add new state to the stack along with leaf node
			stack = append(stack, StackState { action.Value, node })
			ip++
		case REDUCE:
            // Find production to reduce by
			production := p.table.Grammar.Productions[action.Value]
            r := len(production.Right); l := len(stack) - r
            // Collect child nodes from current states on the stack and create node for reduction
            children := make([]*ParseTreeNode, r)
            for i, s := range stack[l:] { children[i] = s.Node }
            node := &ParseTreeNode { production.Left, children }
            // Pop stack and find next state based on goto table
			stack = stack[:l]; state := stack[l - 1].State
            next := p.table.Goto[state][production.Left]
			stack = append(stack, StackState { next, node })
		case ACCEPT:
            fmt.Println("accept")
            return stack[1].Node
		}
	}
}
