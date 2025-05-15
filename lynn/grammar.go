package lynn

import (
	"fmt"
	"strings"
)

// Symbol interface. Can either be a Terminal or NonTerminal struct.
type Symbol interface { fmt.Stringer }
// Terminal type. Represented by string identifier.
type Terminal string
const (EPSILON Terminal = ""; EOF_TERMINAL = "EOF")
// Non-terminal type. Represented by string identifier.
type NonTerminal string

// Grammar struct. Tracks all terminals and non-terminals, the start non-terminal, and all production rules.
type Grammar struct {
    Terminals    []Terminal
    NonTerminals []NonTerminal
    Start        NonTerminal
    Productions  []Production
}
// Production struct. Expresses a sequence of symbols that a given non-terminal may be expanded to in a grammar.
type Production struct { Left NonTerminal; Right []Symbol }

// Lexer generator struct. Converts EBNF rule definitions to context-free grammar (CFG) production rules.
type GrammarGenerator struct {
    terminals      map[string]struct{}
    strings        map[string]Terminal
    nonTerminals   []NonTerminal
    nonTerminalMap map[string]struct{}
    parents        map[NonTerminal]NonTerminal
    children       map[NonTerminal]int
    productions    []Production
}

// Returns a grammar generator struct.
func NewGrammarGenerator() *GrammarGenerator { return &GrammarGenerator { } }
// Converts EBNF rules defined in AST into CFG production rules.
func (g *GrammarGenerator) GenerateCFG(grammar *GrammarNode) *Grammar {
    // Generate set of valid terminals and create map from simple string tokens to their corresponding terminal
    g.terminals, g.strings = make(map[string]struct{}, len(grammar.Tokens)), make(map[string]Terminal, len(grammar.Tokens))
    terminals := make([]Terminal, 0, len(grammar.Tokens))
    for _, token := range grammar.Tokens {
        id := token.Identifier
        if _, ok := g.terminals[id.Name]; ok {
            fmt.Printf("Generation error: Token \"%s\" is already defined - %d:%d\n", id.Name, id.Location.Line, id.Location.Col)
            continue
        }
        g.terminals[id.Name] = struct{}{}
        t := Terminal(id.Name); terminals = append(terminals, t)
        // If a token's expression is only a string, the string itself is allowed to be used in rule expressions
        if str, ok := token.Expression.(*StringNode); ok {
            g.strings[string(str.Chars)] = t
        }
    }
    // If EOF is not defined, add to terminal list
    if _, ok := g.terminals[EOF_TERMINAL]; !ok {
        g.terminals[EOF_TERMINAL] = struct{}{}
        terminals = append(terminals, EOF_TERMINAL)
    }
    // Create list of valid non-terminals and initialize parent-children relationship data constructs
    g.nonTerminals, g.nonTerminalMap = make([]NonTerminal, 0, len(grammar.Rules)), make(map[string]struct{}, len(grammar.Rules))
    g.parents, g.children = make(map[NonTerminal]NonTerminal), make(map[NonTerminal]int, len(grammar.Rules))
    for _, rule := range grammar.Rules {
        id := rule.Identifier
        // Ensure identifier does not collide with an existing token
        if _, ok := g.terminals[id.Name]; ok {
            fmt.Printf("Generation error: Identifier \"%s\" is already taken by a token - %d:%d\n",
                id.Name, id.Location.Line, id.Location.Col)
            continue
        }
        if _, ok := g.nonTerminalMap[id.Name]; !ok {
            g.nonTerminals = append(g.nonTerminals, NonTerminal(id.Name))
            g.nonTerminalMap[id.Name] = struct{}{}
        }
    }
    // Convert AST expression to grammar
    g.productions = make([]Production, 0)
    for _, rule := range grammar.Rules {
        t := NonTerminal(rule.Identifier.Name); 
        g.expressionCFG(t, rule.Expression)
    }
    g.removeAmbiguities()
    // Collect accumulated data into grammar struct
    return &Grammar { terminals, g.nonTerminals, g.nonTerminals[0], g.productions }
}

// For a given expression node from the AST, adds to a list of productions in CFG format.
func (g *GrammarGenerator) expressionCFG(left NonTerminal, expression AST) {
    // Find all production cases for a given non-terminal by obtaining leaf nodes in a union
    var cases []AST
    if n, ok := expression.(*UnionNode); ok {
        // Recursively find all nodes reachable through only union nodes
        cases = flattenUnion(n, make([]AST, 0))
    } else {
        cases = []AST { expression }
    }
    // For each case, flatten concatenated nodes to obtain symbols of each production
    // Same procedure as finding all nodes reachable through unions, but for concatenations
    for _, c := range cases {
        var nodes []AST
        if n, ok := c.(*ConcatenationNode); ok {
            nodes = flattenConcatenation(n, make([]AST, 0))
        } else {
            nodes = []AST { c }
        }
        // Convert nodes in list to symbols
        symbols := make([]Symbol, 0, len(nodes))
        for _, n := range nodes {
            symbol := g.symbolCFG(left, n)
            if symbol != nil { symbols = append(symbols, symbol) }
        }
        g.productions = append(g.productions, Production { left, symbols })
    }
}

// Converts a given expression node from the AST to a CFG symbol.
func (g *GrammarGenerator) symbolCFG(left NonTerminal, expression AST) Symbol {
    switch node := expression.(type) {
    case *OptionNode:
        // Create new non-terminal with possible productions including an epsilon production
        t := g.deriveNonTerminal(left)
        g.expressionCFG(t, node.Expression)
        g.productions = append(g.productions, Production { t, []Symbol { } })
        return t
    case *RepeatNode:
        // Two new non-terminals are created for repetitions:
        // E_0 -> E_0 E_1 (E_1 can be repeated 0 or more times)
        // E_0 -> epsilon
        t1, t2 := g.deriveNonTerminal(left), g.deriveNonTerminal(left)
        g.productions = append(g.productions, Production { t1, []Symbol { t1, t2 } }, Production { t1, []Symbol { } })
        g.expressionCFG(t2, node.Expression)
        return t1
    case *RepeatOneNode:
        // E_0 -> E_0 E_1
        // E_0 -> E_1 (E_1 can be repeated 1 or more times)
        t1, t2 := g.deriveNonTerminal(left), g.deriveNonTerminal(left)
        g.productions = append(g.productions, Production { t1, []Symbol { t1, t2 } }, Production { t1, []Symbol { t2 } })
        g.expressionCFG(t2, node.Expression)
        return t1
    case *UnionNode:
        // Separate inner union to new non-terminal
        t := g.deriveNonTerminal(left)
        g.expressionCFG(t, node)
        return t

    case *IdentifierNode:
        // Finds the terminal or non-terminal that the identifier is referring to
        if _, ok := g.terminals[node.Name];      ok { return Terminal(node.Name) }
        if _, ok := g.nonTerminalMap[node.Name]; ok { return NonTerminal(node.Name) }
        fmt.Printf("Generation error: Identifier \"%s\" is not defined - %d:%d\n", node.Name, node.Location.Line, node.Location.Col)
        return nil
    case *StringNode:
        // Find terminal associated with a string, resolves if explicit string definition exists in AST
        str := string(node.Chars); 
        if t, ok := g.strings[str]; ok { return t }
        fmt.Printf("Generation error: No token explicitly matches \"%s\" - %d:%d\n", str, node.Location.Line, node.Location.Col)
        return nil
    case *ClassNode:
        fmt.Printf("Generation error: Classes can not be used in rule expressions - %d:%d\n", node.Location.Line, node.Location.Col)
        return nil
    default: panic("Invalid expression passed to GrammarGenerator.symbolCFG()")
    }
}

// TODO: Allow grammar to specify associativity, further generalization (dangling-else ambiguity?)

// Removes simple precedence and associativity operator-form ambiguities in the grammar.
// Not guaranteed to remove all ambiguities, but will resolve those of infix, prefix, and postfix operations.
func (g *GrammarGenerator) removeAmbiguities() {
    type AmbiguityType uint
    type Ambiguity struct { t AmbiguityType; production Production }
    const (NONE AmbiguityType = iota; INFIX; PREFIX; POSTFIX)
    // Group productions together based on their non-terminal
    productions := make(map[NonTerminal][]Production, len(g.nonTerminals))
    modified := make([]Production, 0, len(g.productions))
    for _, p := range g.productions { productions[p.Left] = append(productions[p.Left], p) }
    for nt, group := range productions {
        // Split productions for a given non-terminal based on if there exists left or right-recursion
        a, rest := make([]Ambiguity, 0), make([]Production, 0)
        for _, p := range group {
            if len(p.Right) >= 2 {
                l, r := p.Right[0] == nt, p.Right[len(p.Right) - 1] == nt
                // Determine type of recursion and type of operator
                t := NONE
                switch {
                case l && r: t = INFIX   // E -> E ... E
                case l:      t = POSTFIX // E -> E ...
                case r:      t = PREFIX  // E -> ... E
                }
                if t != NONE { a = append(a, Ambiguity { t, p }); continue }
            }
            rest = append(rest, p)
        }
        // Divide non-terminal productions into multiple precedence levels
        left := nt
        for i, m := range a {
            // Create new auxiliary non-terminal
            // Don't generate new non-terminal when last precedence level is a prefix or postfix operator
            next := left
            if m.t == INFIX || i < len(a) - 1 { next = g.deriveNonTerminal(nt) }
            // Create modified copy of production
            right := make([]Symbol, len(m.production.Right)); copy(right, m.production.Right)
            modified = append(modified, Production { left, right })
            switch m.t {
            case INFIX:
                // For infix ambiguity, eliminate either left or right-recursion to force associativity
                // For left-associative productions,  E_k -> E_k ... E_{k + 1} (eliminate right-recursion)
                // For right-associative productions, E_k -> E_{k + 1} ... E_k
                right[0], right[len(right) - 1] = left, next
            case PREFIX:  right[len(right) - 1] = left
            case POSTFIX: right[0] = left
            }
            if next != left {
                modified = append(modified, Production { left, []Symbol { next } })
                left = next
            }
        }
        // Add all non-operation productions to highest precedence level non-terminal
        for _, p := range rest { modified = append(modified, Production { left, p.Right }) }
    }
    g.productions = modified
}

// TODO: Grammar simplification
// func (g *GrammarGenerator) simplifyGrammar() {
//     // Group productions together based on their non-terminal
//     productions := make(map[NonTerminal][]Production, len(g.nonTerminals))
//     simplified := make([]Production, 0)
//     for _, p := range g.productions { productions[p.Left] = append(productions[p.Left], p) }
//     // Generate replacement map for equivalent non-terminals
//     // A non-terminal is only equivalent if there exists a non-terminal whose only production is N_1 -> N_2
//     replacement := make(map[NonTerminal]NonTerminal)
//     for nt, group := range productions {
//         if _, ok := replacement[nt]; ok || len(group) > 1 { continue }
//         p := group[0]; if len(p.Right) != 1 { continue }
//         if t, ok := p.Right[0].(NonTerminal); ok { replacement[nt] = t }
//     }
//     // Replace occurrences according to map
//     for _, p := range g.productions {
//         if _, ok := replacement[p.Left]; ok { continue }
//         // Create modified copy of production
//         right := make([]Symbol, len(p.Right))
//         for i, s := range p.Right {
//             if t, ok := s.(NonTerminal); ok {
//                 if r, ok := replacement[t]; ok { right[i] = r; continue }
//             }
//             right[i] = s
//         }
//         simplified = append(simplified, Production { p.Left, right })
//     }
//     g.productions = simplified
// }

// Creates a new non-terminal derived from a parent non-terminal.
func (g *GrammarGenerator) deriveNonTerminal(nt NonTerminal) NonTerminal {
    // Derive new non-terminal from parent
    parent, ok := g.parents[nt]; if !ok { parent = nt }
    // Generate new name from parent name
    var id string
    for {
        id = fmt.Sprintf("%s_%d", parent, g.children[parent])
        g.children[parent]++
        // Handle naming collisions
        if _, ok := g.nonTerminalMap[id]; !ok { break }
    }
    // Register non-terminal and reference to its parent
    t := NonTerminal(id)
    g.nonTerminals = append(g.nonTerminals, t); g.nonTerminalMap[id] = struct{}{}
    g.parents[t] = parent
    return t
}

// Augment grammar with new start state. Returns production for augmented start state.
func (g *Grammar) Augment() *Production {
    t := NonTerminal("S'")
    g.NonTerminals = append(g.NonTerminals, t)
    g.Productions = append(g.Productions, Production { t, []Symbol { g.Start } })
    return &g.Productions[len(g.Productions) - 1]
}

// ------------------------------------------------------------------------------------------------------------------------------

func flattenConcatenation(node *ConcatenationNode, nodes []AST) []AST {
    if a, ok := node.A.(*ConcatenationNode); ok { nodes = flattenConcatenation(a, nodes) } else { nodes = append(nodes, node.A) }
    if b, ok := node.B.(*ConcatenationNode); ok { nodes = flattenConcatenation(b, nodes) } else { nodes = append(nodes, node.B) }
    return nodes
}

func flattenUnion(node *UnionNode, nodes []AST) []AST {
    if a, ok := node.A.(*UnionNode); ok { nodes = flattenUnion(a, nodes) } else { nodes = append(nodes, node.A) }
    if b, ok := node.B.(*UnionNode); ok { nodes = flattenUnion(b, nodes) } else { nodes = append(nodes, node.B) }
    return nodes
}

func (t Terminal)    String() string { return string(t) }
func (t NonTerminal) String() string { return string(t) }
func (p Production) String() string {
    var builder strings.Builder
    builder.WriteString(fmt.Sprintf("%s ->", p.Left))
    if len(p.Right) > 0 {
        for _, s := range p.Right { builder.WriteString(fmt.Sprintf(" %s", s)) }
    } else {
        builder.WriteString(" ε")
    }
    return builder.String()
}

// FOR DEBUG PURPOSES:
// Prints all production rules of the grammar.
func (g *Grammar) PrintGrammar() {
    fmt.Printf("start: %s\n", g.Start)
    for _, production := range g.Productions { fmt.Println(production) }
}

// ------------------------------------------------------------------------------------------------------------------------------
// TODO: When generating parse tree, remove auxiliary non-terminals
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
    input := []Terminal { "TOKEN", "IDENTIFIER", "COLON", "STRING", "CLASS", "BAR", "IDENTIFIER", "PLUS", "SEMI", "RULE", "IDENTIFIER", "COLON", "L_PAREN", "IDENTIFIER", "R_PAREN", "QUESTION", "SEMI", EOF_TERMINAL }
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
