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
    Productions  []*Production
}
// Production struct. Expresses a sequence of symbols that a given non-terminal may be expanded to in a grammar.
// Auxiliary productions must have a right-hand side with a single non-terminal.
// Flatten productions must follow the form E -> E E_0.
// Removed productions must have a length of 0 (epsilon productions).
type Production struct {
    Type    ProductionType
    Left    NonTerminal; Right []Symbol
    Visitor string
}

// Lexer generator struct. Converts EBNF rule definitions to context-free grammar (CFG) production rules.
type GrammarGenerator struct {
    terminals      map[string]struct{}
    strings        map[string]Terminal
    nonTerminals   []NonTerminal
    nonTerminalMap map[string]struct{}
    parents        map[NonTerminal]NonTerminal
    children       map[NonTerminal]int
    productions    []*Production
    aliasMaps      map[*Production]map[string]int
    labels         map[*Production]*LabelNode
}

// Returns a grammar generator struct.
func NewGrammarGenerator() *GrammarGenerator { return &GrammarGenerator { } }
// Converts EBNF rules defined in AST into CFG production rules.
func (g *GrammarGenerator) GenerateCFG(grammar *GrammarNode) (*Grammar, map[*Production]map[string]int) {
    // Generate set of valid terminals and create map from simple string tokens to their corresponding terminal
    g.terminals, g.strings = make(map[string]struct{}, len(grammar.Tokens)), make(map[string]Terminal, len(grammar.Tokens))
    terminals := make([]Terminal, 0, len(grammar.Tokens))
    for _, token := range grammar.Tokens {
        if token.Skip { continue }
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
    g.productions = make([]*Production, 0)
    g.aliasMaps, g.labels = make(map[*Production]map[string]int), make(map[*Production]*LabelNode)
    for _, rule := range grammar.Rules {
        t := NonTerminal(rule.Identifier.Name)
        g.flattenProductions(t, rule.Expression)
    }
    g.removeAmbiguities()
    // Collect accumulated data into grammar struct
    return &Grammar { terminals, g.nonTerminals, g.nonTerminals[0], g.productions }, g.aliasMaps
}

// For a given expression node from the AST, adds to a list of productions in CFG format.
func (g *GrammarGenerator) flattenProductions(left NonTerminal, expression AST) {
    // Find all production cases for a given non-terminal by obtaining leaf nodes in a union
    var cases []AST
    if n, ok := expression.(*UnionNode); ok {
        // Recursively find all nodes reachable through only union nodes
        cases = flattenUnion(n, make([]AST, 0))
    } else {
        cases = []AST { expression }
    }
    for _, node := range cases {
        // If a label exists, get visitor name, otherwise default to non-terminal name
        label, ok := node.(*LabelNode)
        visitor := string(left)
        if ok { node = label.Expression; visitor = label.Identifier.Name }
        // For each case, flatten concatenated nodes and convert nodes in list to symbols
        production := g.flattenConcatCFG(left, node, visitor)
        g.productions = append(g.productions, production)
        // Associate production with label node for later use in disambiguation
        if ok { g.labels[production] = label }
    }
}

// Converts a given expression node from the AST to its corresponding CFG structure.
func (g *GrammarGenerator) expressionCFG(left NonTerminal, expression AST) {
    switch node := expression.(type) {
    case *OptionNode:
        // Create production including an epsilon production
        g.expressionCFG(left, node.Expression)
        g.productions = append(g.productions, &Production { REMOVED, left, []Symbol { }, "" })
    case *RepeatNode:
        // E -> E E' (E' can be repeated 0 or more times)
        // E -> epsilon
        if t := g.expandExpressionCFG(left, node.Expression); t != nil {
            g.productions = append(g.productions,
                &Production { FLATTEN, left, []Symbol { left, t }, "" },
                &Production { NORMAL, left, []Symbol { }, "" })
        }
    case *RepeatOneNode:
        // E -> E E'
        // E -> E' (E' can be repeated 1 or more times)
        if t := g.expandExpressionCFG(left, node.Expression); t != nil {
            g.productions = append(g.productions,
                &Production { FLATTEN, left, []Symbol { left, t }, "" },
                &Production { NORMAL, left, []Symbol { t }, "" })
        }
    // New non-terminals are auxiliary when production is for a non-derived non-terminal
    case *ConcatNode: g.productions = append(g.productions, g.flattenConcatCFG(left, node, ""))
    case *UnionNode: g.flattenUnionCFG(left, node)
    default:
        if s, ok := g.literalCFG(node); s != nil {
            g.productions = append(g.productions, &Production{ NORMAL, left, []Symbol { s }, "" })
        } else if !ok { panic("Invalid expression node passed to GrammarGenerator.symbolCFG()") }
    }
}

// Determines whether or not parts of an expression needs to be expanded to a new non-terminal.
func (g *GrammarGenerator) expandExpressionCFG(left NonTerminal, expression AST) Symbol {
    // For literals, convert and return terminal directly
    s, ok := g.literalCFG(expression); if ok { return s }
    // Otherwise, create new non-terminal and expand expression
    t := g.deriveNonTerminal(left)
    g.expressionCFG(t, expression)
    return t
}

// For a given set of nodes in a union from the AST, adds to a list of productions in CFG format.
func (g *GrammarGenerator) flattenUnionCFG(left NonTerminal, node *UnionNode) {
    // Find all production cases for a given non-terminal by obtaining leaf nodes in a union
    cases := flattenUnion(node, make([]AST, 0))
    // For each case, flatten concatenated nodes
    for _, node := range cases {
        if n, ok := node.(*ConcatNode); ok {
            // Do not generate new non-terminal if the production is a concatenation
            g.productions = append(g.productions, g.flattenConcatCFG(left, n, ""))
        } else if s := g.expandExpressionCFG(left, node); s != nil {
            // For unions of multiple productions, ensure option/repeat constructs are given separate non-terminals
            g.productions = append(g.productions, &Production { AUXILIARY, left, []Symbol { s }, "" })
        }
    }
}

// For a given concatenated string of nodes from the AST, create production after generating list of symbols.
func (g *GrammarGenerator) flattenConcatCFG(left NonTerminal, expression AST, visitor string) *Production {
    // Flatten concatenated nodes to obtain symbols of each production
    var nodes []AST
    if n, ok := expression.(*ConcatNode); ok {
        nodes = flattenConcat(n, make([]AST, 0))
    } else {
        nodes = []AST { expression }
    }
    // Convert nodes in list to symbols
    symbols := make([]Symbol, 0, len(nodes))
    aliases, identifiers := make(map[string]int), make(map[string][]int)
    for i, node := range nodes {
        switch n := node.(type) {
        case *AliasNode:
            // Intercept aliases within concatenation expressions (not allowed elsewhere)
            // Build map between aliases and indices associated with the current production
            id := n.Identifier
            if _, ok := aliases[id.Name]; !ok {
                aliases[id.Name] = i
            } else {
                fmt.Printf("Generation error: Alias \"%s\" is already defined - %d:%d\n",
                    id.Name, id.Location.Line, id.Location.Col)
            }
            node = n.Expression
        // Identifiers are accumulated as potential implicit aliases
        case *IdentifierNode: identifiers[n.Name] = append(identifiers[n.Name], i)
        // If the expression inside a quantifier is only an identifier, add as potential alias
        case *OptionNode:
            if id, ok := n.Expression.(*IdentifierNode); ok { identifiers[id.Name] = append(identifiers[id.Name], i) }
        case *RepeatNode:
            if id, ok := n.Expression.(*IdentifierNode); ok { identifiers[id.Name] = append(identifiers[id.Name], i) }
        case *RepeatOneNode:
            if id, ok := n.Expression.(*IdentifierNode); ok { identifiers[id.Name] = append(identifiers[id.Name], i) }
        }
        symbols = append(symbols, g.expandExpressionCFG(left, node))
    }
    // If an identifier has only one occurrence and no explicit alias of the same name exists, add as implicit alias
    for id, indices := range identifiers {
        if _, ok := aliases[id]; ok || len(indices) > 1 { continue }
        aliases[id] = indices[0]
    }
    // Create production struct
    // If alias map contains entries, create association between it and production
    production := &Production { NORMAL, left, symbols, visitor }
    if len(aliases) > 0 { g.aliasMaps[production] = aliases }
    return production
}

// Converts a given literal node from the AST to a CFG symbol.
func (g *GrammarGenerator) literalCFG(expression AST) (Symbol, bool) {
    switch node := expression.(type) {
    case *IdentifierNode:
        // Finds the terminal or non-terminal that the identifier is referring to
        if _, ok := g.terminals[node.Name];      ok { return Terminal(node.Name), true }
        if _, ok := g.nonTerminalMap[node.Name]; ok { return NonTerminal(node.Name), true }
        fmt.Printf("Generation error: Identifier \"%s\" is not defined - %d:%d\n", node.Name, node.Location.Line, node.Location.Col)
    case *StringNode:
        // Find terminal associated with a string, resolves if explicit string definition exists in AST
        str := string(node.Chars); 
        if t, ok := g.strings[str]; ok { return t, true }
        fmt.Printf("Generation error: No token explicitly matches \"%s\" - %d:%d\n", str, node.Location.Line, node.Location.Col)
    case *ClassNode:
        fmt.Printf("Generation error: Classes cannot be used in rule expressions - %d:%d\n", node.Location.Line, node.Location.Col)
    case *LabelNode: fmt.Printf("Generation error: Invalid use of label - %d:%d\n", node.Location.Line, node.Location.Col)
    case *AliasNode: fmt.Printf("Generation error: Invalid use of alias - %d:%d\n", node.Location.Line, node.Location.Col)
    default: return nil, false
    }
    return nil, true
}

// Removes simple precedence and associativity operator-form ambiguities in the grammar.
// Not guaranteed to remove all ambiguities, but will resolve those of infix, prefix, and postfix operations.
func (g *GrammarGenerator) removeAmbiguities() {
    type AmbiguityType uint
    const (INFIX AmbiguityType = iota; PREFIX; POSTFIX)
    type Ambiguity struct { ambiguityType AmbiguityType; production *Production }
    // Group productions together based on their non-terminal
    productions := make(map[NonTerminal][]*Production, len(g.nonTerminals))
    for _, p := range g.productions { productions[p.Left] = append(productions[p.Left], p) }
    for nt, group := range productions {
        // Split productions for a given non-terminal based on if there exists explicit left or right-recursion
        a, rest := make([]Ambiguity, 0), make([]*Production, 0)
        for _, p := range group {
            if len(p.Right) >= 2 {
                l, r := p.Right[0] == nt, p.Right[len(p.Right) - 1] == nt
                // Determine type of recursion and type of operator
                switch {
                case l && r: a = append(a, Ambiguity { INFIX, p }) // E -> E ... E
                case l: // E -> E ...
                    g.expectNoLabel(p)
                    a = append(a, Ambiguity { POSTFIX, p })
                case r: // E -> ... E
                    g.expectNoLabel(p)
                    a = append(a, Ambiguity { PREFIX, p })
                }
                if l || r { continue }
            }
            // If production is not both left and right recursive, expect no associativity label
            g.expectNoLabel(p)
            rest = append(rest, p)
        }
        // At least one non-operation production must exist to perform disambiguation
        if len(rest) == 0 { continue }
        // Divide non-terminal productions into multiple precedence levels
        left := nt
        for i, m := range a {
            // Create new auxiliary non-terminal
            // Don't generate new non-terminal when last precedence level is a prefix or postfix operator
            next := left
            if m.ambiguityType == INFIX || i < len(a) - 1 { next = g.deriveNonTerminal(nt) }
            // Modify existing productions
            p := m.production; right := p.Right
            p.Left = left
            switch m.ambiguityType {
            case INFIX:
                // For infix ambiguity, eliminate either left or right-recursion to force associativity
                // For left-associative productions,  E_k -> E_k ... E_{k + 1} (eliminate right-recursion)
                // For right-associative productions, E_k -> E_{k + 1} ... E_k
                if node, ok := g.labels[m.production]; ok && node.Associativity == RIGHT_ASSOC {
                    right[0], right[len(right) - 1] = next, left // Make right-associative if explicit label exists
                } else {
                    right[0], right[len(right) - 1] = left, next
                }
            case PREFIX:  right[len(right) - 1] = left
            case POSTFIX: right[0] = left
            }
            if next != left {
                g.productions = append(g.productions, &Production { AUXILIARY, left, []Symbol { next }, "" })
                left = next
            }
        }
        // Add all non-operation productions to highest precedence level non-terminal
        for _, p := range rest { p.Left = left }
    }
}

func (g *GrammarGenerator) expectNoLabel(production *Production) {
    // Raise error message if production has a label
    if node, ok := g.labels[production]; ok && node.Associativity != NO_ASSOC {
        fmt.Printf("Generation error: Associativity label cannot be used for current production - %d:%d\n",
            node.Location.Line, node.Location.Col)
    }
}

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
    production := &Production { NORMAL, t, []Symbol { g.Start }, "" }
    g.NonTerminals = append(g.NonTerminals, t)
    g.Productions = append(g.Productions, production)
    return production
}

// ------------------------------------------------------------------------------------------------------------------------------

func flattenConcat(node *ConcatNode, nodes []AST) []AST {
    if a, ok := node.A.(*ConcatNode); ok { nodes = flattenConcat(a, nodes) } else { nodes = append(nodes, node.A) }
    if b, ok := node.B.(*ConcatNode); ok { nodes = flattenConcat(b, nodes) } else { nodes = append(nodes, node.B) }
    return nodes
}

func flattenUnion(node *UnionNode, nodes []AST) []AST {
    if a, ok := node.A.(*UnionNode); ok { nodes = flattenUnion(a, nodes) } else { nodes = append(nodes, node.A) }
    if b, ok := node.B.(*UnionNode); ok { nodes = flattenUnion(b, nodes) } else { nodes = append(nodes, node.B) }
    return nodes
}

func (t Terminal)    String() string { return string(t) }
func (t NonTerminal) String() string { return string(t) }
func (p Production)  String() string {
    var builder strings.Builder
    builder.WriteString(fmt.Sprintf("%s ->", p.Left))
    if len(p.Right) > 0 {
        for _, s := range p.Right { builder.WriteString(fmt.Sprintf(" %s", s)) }
    } else {
        builder.WriteString(" ε")
    }
    if p.Visitor != "" { builder.WriteString(fmt.Sprintf("  #%s", p.Visitor)) }
    return builder.String()
}

// FOR DEBUG PURPOSES:
// Prints all production rules of the grammar.
func (g *Grammar) PrintGrammar() {
    fmt.Printf("[start: %s]\n", g.Start)
    for _, production := range g.Productions { fmt.Println(production) }
}
