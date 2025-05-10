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
    terminals    map[string]struct{}
    strings      map[string]Terminal
    nonTerminals map[string]NonTerminal
    parentData   map[NonTerminal]*nonTerminalData
    productions  []Production
}
type nonTerminalData struct {
    identifier  string
    derivatives int
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
    // Create list of unique non-terminals and assign index identifiers to each
    g.nonTerminals = make(map[string]NonTerminal,           len(grammar.Rules))
    g.parentData   = make(map[NonTerminal]*nonTerminalData, len(grammar.Rules))
    for _, rule := range grammar.Rules {
        id := rule.Identifier
        // Ensure identifier does not collide with an existing token
        if _, ok := g.terminals[id.Name]; ok {
            fmt.Printf("Generation error: Identifier \"%s\" is already taken by a token - %d:%d\n",
                id.Name, id.Location.Line, id.Location.Col)
            continue
        }
        t, ok := g.nonTerminals[id.Name]
        if !ok {
            t = NonTerminal(len(g.nonTerminals))
            g.nonTerminals[id.Name] = t
            g.parentData[t] = &nonTerminalData { id.Name, 0 }
        }
    }
    // Convert AST expression to grammar
    g.productions = make([]Production, 0)
    for _, rule := range grammar.Rules {
        t := g.nonTerminals[rule.Identifier.Name]
        g.expressionCFG(t, t, rule.Expression)
    }
    nonTerminals := make(map[NonTerminal]string, len(g.nonTerminals))
    for id, t := range g.nonTerminals { nonTerminals[t] = id }
    return &Grammar { terminals, nonTerminals, 0, g.productions }
}

// For a given expression node from the AST, adds to a list of productions in CFG format.
func (g *GrammarGenerator) expressionCFG(parent NonTerminal, left NonTerminal, expression AST) {
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
        symbols := make([]Symbol, 0)
        for _, n := range nodes {
            symbol := g.symbolCFG(parent, n)
            if symbol != nil { symbols = append(symbols, symbol) }
        }
        g.productions = append(g.productions, Production { left, symbols })
    }
}

// Converts a given expression node from the AST to a CFG symbol.
func (g *GrammarGenerator) symbolCFG(parent NonTerminal, expression AST) Symbol {
    switch node := expression.(type) {
    case *OptionNode:
        // Create new non-terminal with possible productions including an epsilon production
        t := g.deriveNonTerminal(parent)
        g.expressionCFG(parent, t, node.Expression)
        g.productions = append(g.productions, Production { t, []Symbol { } })
        return t
    case *RepeatNode:
        // Two new non-terminals are created for repetitions:
        // E_0 -> E_0 E_1 (E_1 can be repeated 0 or more times)
        // E_0 -> epsilon
        t1, t2 := g.deriveNonTerminal(parent), g.deriveNonTerminal(parent)
        g.productions = append(g.productions, Production { t1, []Symbol { t1, t2 } }, Production { t1, []Symbol { } })
        g.expressionCFG(parent, t2, node.Expression)
        return t1
    case *RepeatOneNode:
        // E_0 -> E_0 E_1
        // E_0 -> E_1 (E_1 can be repeated 1 or more times)
        t1, t2 := g.deriveNonTerminal(parent), g.deriveNonTerminal(parent)
        g.productions = append(g.productions, Production { t1, []Symbol { t1, t2 } }, Production { t1, []Symbol { t2 } })
        g.expressionCFG(parent, t2, node.Expression)
        return t1
    case *UnionNode:
        // Separate inner union to new non-terminal
        t := g.deriveNonTerminal(parent)
        g.expressionCFG(parent, t, node)
        return t

    case *IdentifierNode:
        // Finds the terminal or non-terminal that the identifier is referring to
        if _, ok := g.terminals[node.Name]; ok { return Terminal(node.Name) }
        if t, ok := g.nonTerminals[node.Name]; ok { return t }
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

// Creates a new non-terminal derived from a parent non-terminal.
func (g *GrammarGenerator) deriveNonTerminal(parent NonTerminal) NonTerminal {
    // Derive new non-terminal from parent
    data := g.parentData[parent]
    id := fmt.Sprintf("%s_%d", data.identifier, data.derivatives)
    data.derivatives++
    // Register non-terminal
    t := NonTerminal(len(g.nonTerminals)); g.nonTerminals[id] = t
    return t
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
            // TODO: Look more into automatic disambiguation (or manual markers)
            // symbols[0], symbols[len(symbols) - 1] = next, left
            // Add modified productions to new list
            modified = append(modified, Production { left, symbols })
            modified = append(modified, Production { left, []Symbol { next } })
            left = next
        }
        // For remaining productions, replace left non-terminal with new non-terminal of highest precedence
        for _, p := range r {
            if len(p.Right) > 0 {
                // Replace left and right recursion with high precedence non-terminal
                if p.Right[0]                == t { p.Right[0]                = left }
                if p.Right[len(p.Right) - 1] == t { p.Right[len(p.Right) - 1] = left }
            }
            modified = append(modified, Production { left, p.Right })
        }
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

func (t Terminal)    String(grammar *Grammar) string { return string(t) }
func (t NonTerminal) String(grammar *Grammar) string { return grammar.NonTerminals[t] }
func (p Production)  String(grammar *Grammar) string {
    var builder strings.Builder
    builder.WriteString(fmt.Sprintf("%s ->", p.Left.String(grammar)))
    if len(p.Right) > 0 {
        for _, s := range p.Right { builder.WriteString(fmt.Sprintf(" %s", s.String(grammar))) }
    } else {
        builder.WriteString(" ε")
    }
    return builder.String()
}

// FOR DEBUG PURPOSES:
// Prints all production rules of the grammar.
func (g *Grammar) PrintGrammar() {
    fmt.Printf("start: %s\n", g.Start.String(g))
    for _, production := range g.Productions { fmt.Println(production.String(g)) }
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
    input := []Terminal { "TOKEN", "IDENTIFIER", "COLON", "STRING", "CLASS", "BAR", "IDENTIFIER", "PLUS", "SEMI", EOF_TERMINAL }
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
