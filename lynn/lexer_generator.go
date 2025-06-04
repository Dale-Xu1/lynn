package lynn

import (
	"fmt"
	"regexp"
	"slices"
	"sort"
	"unsafe"
)

// Non-deterministic finite automata struct.
type LNFA struct {
    Start  *LNFAState
    Accept map[*LNFAState]LNFAAccept
}
// Non-deterministic finite automata fragment struct. Holds references to the in and out states.
type LNFAFragment struct { In, Out *LNFAState }
// Non-deterministic finite automata accepting state metadata. Holds token identifier and priority.
type LNFAAccept struct { Identifier string; Priority int }
// Non-deterministic finite automata state struct.
// Holds references to outgoing states and each transition's associated character range.
type LNFAState struct {
    Transitions map[Range]*LNFAState
    Epsilon     []*LNFAState
}

// Deterministic finite automata struct.
type LDFA = DFA[Range]
// Deterministic finite automata state struct.
// Holds references to outgoing states and each transition's associated character range.
type LDFAState = DFAState[Range]

// Lexer generator struct. Converts token definitions in abstract syntax tree (AST) to finite automata (FA).
type LexerGenerator struct {
    fragments map[string]LNFAFragment	
    ranges    map[Range]struct{}
    accept    map[*LDFAState]string
}

// Returns a new lexer generator struct.
func NewLexerGenerator() *LexerGenerator { return &LexerGenerator { } }
// Converts regular expressions defined in grammar into a non-deterministic finite automata.
func (g *LexerGenerator) GenerateNFA(grammar *GrammarNode) (LNFA, []Range) {
    g.fragments, g.ranges = make(map[string]LNFAFragment), make(map[Range]struct{})
    tokens := make(map[string]struct{}, len(grammar.Fragments) + len(grammar.Tokens))
    for _, fragment := range grammar.Fragments {
        // Convert fragment expressions to NFAs and add fragment to identifier map
        nfa, ok := g.expressionNFA(fragment.Expression)
        id := fragment.Identifier
        if _, exists := tokens[id.Name]; exists {
            Error(fmt.Sprintf("Fragment \"%s\" is already defined - %d:%d", id.Name, id.Start.Line, id.Start.Col))
        } else if ok {
            g.fragments[id.Name] = nfa
            tokens[id.Name] = struct{}{} // Register fragment name as used so token names don't overlap
        }
    }
    // If EOF is not defined, add default token to list
    injectEOF(grammar)
    start := &LNFAState { make(map[Range]*LNFAState, 0), make([]*LNFAState, 0, len(grammar.Tokens)) }
    accept := make(map[*LNFAState]LNFAAccept, len(grammar.Tokens))
    for i, token := range grammar.Tokens {
        // Convert token expressions to NFAs and attach fragment to final NFA
        nfa, ok := g.expressionNFA(token.Expression)
        id := token.Identifier
        if _, ok := tokens[id.Name]; ok {
            Error(fmt.Sprintf("Token \"%s\" is already defined - %d:%d", id.Name, id.Start.Line, id.Start.Col))
            continue
        }
        tokens[id.Name] = struct{}{}
        if ok {
            // Invalid if accept node can be reached from the start node through only epsilon transitions
            if isAccessible(nfa.In, nfa.Out, make(map[*LNFAState]struct{})) {
                Error(fmt.Sprintf("Invalid regular expression for token \"%s\" - %d:%d", id.Name, id.Start.Line, id.Start.Col))
                continue
            }
            start.AddEpsilon(nfa.In)
            accept[nfa.Out] = LNFAAccept { id.Name, i }
        }
    }
    // Disjoin all occurring ranges and apply to transitions in NFA
    ranges, expansion := disjoinRanges(g.ranges)
    start.expand(expansion, make(map[*LNFAState]struct{}))
    return LNFA { start, accept }, ranges
}

func injectEOF(grammar *GrammarNode) {
    // Test if EOF token is already defined
    for _, token := range grammar.Tokens {
        if token.Identifier.Name == EOF_TERMINAL { return }
    }
    // Provide default EOF token if not defined
    grammar.Tokens = append(grammar.Tokens, &TokenNode {
        Identifier: &IdentifierNode { Name: EOF_TERMINAL },
        Expression: &StringNode { Chars: []rune { 0 } },
        Skip: false,
    })
}

// Converts a given expression node from the AST to an NFA fragment.
func (g *LexerGenerator) expressionNFA(expression AST) (LNFAFragment, bool) {
    switch node := expression.(type) {
    // Implementation of Thompson's construction to generate an NFA from the AST
    case *OptionNode:
        nfa, ok := g.expressionNFA(node.Expression); if !ok { return nfa, ok }
        out := &LNFAState { make(map[Range]*LNFAState, 0), make([]*LNFAState, 0) }
        in := &LNFAState { make(map[Range]*LNFAState, 0), []*LNFAState { nfa.In, out } }
        nfa.Out.AddEpsilon(out)
        return LNFAFragment { in, out }, true
    case *RepeatNode:
        nfa, ok := g.expressionNFA(node.Expression); if !ok { return nfa, ok }
        out := &LNFAState { make(map[Range]*LNFAState, 0), make([]*LNFAState, 0) }
        in := &LNFAState { make(map[Range]*LNFAState, 0), []*LNFAState { nfa.In, out } }
        nfa.Out.AddEpsilon(nfa.In, out)
        return LNFAFragment { in, out }, true
    case *RepeatOneNode:
        nfa, ok := g.expressionNFA(node.Expression); if !ok { return nfa, ok }
        out := &LNFAState { make(map[Range]*LNFAState, 0), make([]*LNFAState, 0) }
        in := &LNFAState { make(map[Range]*LNFAState, 0), []*LNFAState { nfa.In } }
        nfa.Out.AddEpsilon(nfa.In, out)
        return LNFAFragment { in, out }, true

    case *ConcatNode:
        a, ok := g.expressionNFA(node.A); if !ok { return a, ok }
        b, ok := g.expressionNFA(node.B); if !ok { return b, ok }
        // Create epsilon transition between two fragments
        a.Out.AddEpsilon(b.In)
        return LNFAFragment { a.In, b.Out }, true
    case *UnionNode:
        a, ok := g.expressionNFA(node.A); if !ok { return a, ok }
        b, ok := g.expressionNFA(node.B); if !ok { return b, ok }
        // Create in state with epsilon transitions to in states of both fragments
        in, out := &LNFAState { make(map[Range]*LNFAState, 0), []*LNFAState { a.In, b.In } },
            &LNFAState { make(map[Range]*LNFAState, 0), make([]*LNFAState, 0) }
        a.Out.AddEpsilon(out) // Create epsilon transitions from out states of fragments to final out state
        b.Out.AddEpsilon(out)
        return LNFAFragment { in, out }, true

    // Generate NFAs for literals
    case *IdentifierNode:
        fa, ok := g.fragments[node.Name]
        if !ok {
            Error(fmt.Sprintf("Fragment \"%s\" is not defined - %d:%d", node.Name, node.Start.Line, node.Start.Col))
            return LNFAFragment { }, false
        }
        states := make(map[*LNFAState]*LNFAState)
        return LNFAFragment { fa.In.copy(states), states[fa.Out] }, true
    case *StringNode:
        // String cannot be empty
        if len(node.Chars) == 0 {
            Error(fmt.Sprintf("String must contain at least one character - %d:%d", node.Start.Line, node.Start.Col))
            return LNFAFragment { }, false
        }
        // Generate chain of states with transitions at each consecutive character
        out := &LNFAState { make(map[Range]*LNFAState, 0), make([]*LNFAState, 0) }
        state := out
        for i := len(node.Chars) - 1; i >= 0; i-- {
            char := node.Chars[i]
            r := Range { char, char }; g.ranges[r] = struct{}{}
            state = &LNFAState { map[Range]*LNFAState { r: state }, make([]*LNFAState, 0) }
        }
        return LNFAFragment { state, out }, true
    case *ClassNode:
        var in, out *LNFAState
        // Add transition from in to out for each character in class
        in, out = &LNFAState { make(map[Range]*LNFAState, len(node.Ranges)), make([]*LNFAState, 0) },
            &LNFAState { make(map[Range]*LNFAState, 0), make([]*LNFAState, 0) }
        for _, r := range node.Ranges {
            in.Transitions[r] = out
            g.ranges[r] = struct{}{}
        }
        return LNFAFragment { in, out }, true
    case *ErrorNode:
        Error(fmt.Sprintf("Error terminals cannot be used in token expressions - %d:%d", node.Start.Line, node.Start.Col))
        return LNFAFragment { }, false
    case *LabelNode:
        Error(fmt.Sprintf("Labels cannot be used in token expressions - %d:%d", node.Start.Line, node.Start.Col))
        return LNFAFragment { }, false
    case *AliasNode:
        Error(fmt.Sprintf("Aliases cannot be used in token expressions - %d:%d", node.Start.Line, node.Start.Col))
        return LNFAFragment { }, false
    default: panic("Invalid expression passed to LexerGenerator.expressionNFA()")
    }
}

// Test if particular state is reachable from given state through only epsilon transitions.
func isAccessible(state *LNFAState, target *LNFAState, visited map[*LNFAState]struct{}) bool {
    if state == target { return true }
	if _, ok := visited[state]; ok { return false }
	visited[state] = struct{}{}
	for _, s := range state.Epsilon {
		if isAccessible(s, target, visited) { return true }
	}
    return false
}

// This function converts a set of ranges to a one that is mutually disjoint and has the same union.
// This operates by splitting the original ranges rather than merging them. Final output is a map from
// the original range to its corresponding set of disjoined ranges.
// For example: [1, 5], [4, 10] -> [1, 4], [4, 5], [5, 10]
func disjoinRanges(ranges map[Range]struct{}) ([]Range, map[Range][]Range) {
    type Endpoint struct { value rune; start bool }
    // Sort endpoints of all ranges, marked as start or end
    endpoints := make([]Endpoint, 0, len(ranges) * 2)
    for r := range ranges { endpoints = append(endpoints, Endpoint { r.Min, true }, Endpoint { r.Max, false }) }
    sort.Slice(endpoints, func (i, j int) bool {
        r, s := endpoints[i], endpoints[j]
        if r.value == s.value { return r.start }
        return r.value < s.value
    })
    // Sweep endpoints to convert to disjoint integer intervals
    disjoined, n := make([]Range, 0, len(endpoints)), 0
    var current rune
    for _, e := range endpoints {
        if e.start {
            if n > 0 && e.value > current { disjoined = append(disjoined, Range { current, e.value - 1 }) }
            current = e.value
            n++
        } else {
            if e.value >= current { disjoined = append(disjoined, Range { current, e.value }) }
            current = e.value + 1
            n--
        }
    }
    return disjoined, createExpansionMap(ranges, disjoined)
}

// Given a set of ranges and a sorted list of disjoined ranges (output from disjoinRanges()),
// create a map from the original range to the set of disjoined ranges that have the same union.
// Ex. [1, 5] -> [1, 2], [3, 5]
func createExpansionMap(ranges map[Range]struct{}, disjoined []Range) map[Range][]Range {
    // Assume disjoined ranges are sorted and can be searched with binary search
    search := func (target Range, value func (Range) rune) int {
        ref := value(target)
        low, high := 0, len(disjoined) - 1
        for low <= high {
            mid := (low + high) / 2
            v := value(disjoined[mid]) // Allow caller to specify which endpoint of range to use for search
            if v == ref { return mid }
            if v < ref {
                low = mid + 1
            } else {
                high = mid - 1
            }
        }
        return -1
    }
    expansion := make(map[Range][]Range, len(ranges))
    for r := range ranges {
        // Search for disjoined range with matching endpoints
        min, max := search(r, func (r Range) rune { return r.Min }), search(r, func (r Range) rune { return r.Max })
        if min == -1 || max == -1 { panic("Invalid range expansion") }
        if min == max { continue } // Expansion is unnecessary if range is unaffected
        // Map all disjoined ranges between min and max indices to original range
        expanded := make([]Range, max - min + 1); expansion[r] = expanded
        for i := range expanded {
            expanded[i] = disjoined[min + i]
        }
    }
    return expansion
}

// ------------------------------------------------------------------------------------------------------------------------------

// Converts non-deterministic finite automata to deterministic finite automata.
func (g *LexerGenerator) NFAtoDFA(nfa LNFA, ranges []Range) LDFA {
    subsets := make(map[string]*LDFAState)
    g.accept = make(map[*LDFAState]string, len(nfa.Accept))
    // Convert NFA to initial DFA through power-set construction
    start := g.mergeTransitions([]*LNFAState { nfa.Start }, nfa, subsets)
    // Extract all DFA states from subset map
    states := make([]*LDFAState, 0, len(subsets))
    for _, state := range subsets { states = append(states, state) }
    // Find minimal DFA after construction
    dfa := LDFA { start, states, g.accept }
    return minimizeDFA(dfa, ranges) 
}

// Given a set of possible NFA states we may currently be in, first expand set to include states reachable through epsilon
// transitions. Merge transitions on the same range to generate sets of NFA states reachable on that range and recursively
// merge transitions for each of these sets.
func (g *LexerGenerator) mergeTransitions(states []*LNFAState, nfa LNFA, subsets map[string]*LDFAState) *LDFAState {
    // Find epsilon closure of a given set of states
    closure := make(map[*LNFAState]struct{})
    for _, state := range states { epsilonClosure(state, closure) }
    // If closure has already been processed, return reference to DFA state
    key := getClosureKey(closure)
    if state := subsets[key]; state != nil { return state }

    // Go through all transitions of all states in closure and find possible subsets reachable through each range
    l := 0; for state := range closure { l += len(state.Transitions) }
    merged := make(map[Range][]*LNFAState, l)
    for state := range closure {
        for value, s := range state.Transitions {
            if merged[value] != nil {
                merged[value] = append(merged[value], s)
            } else { merged[value] = []*LNFAState { s } }
        }
    }
    // Create and store DFA state
    state := &LDFAState{ make(map[Range]*LDFAState, len(merged)) }
    subsets[key] = state
    if id, ok := nfa.resolveAccept(closure); ok { g.accept[state] = id }
    // Convert each subset that this state may transition into its own DFA state recursively
    for value, states := range merged {
        state.Transitions[value] = g.mergeTransitions(states, nfa, subsets)
    }
    return state
}

// Finds set of states reachable from given state through only epsilon transitions.
func epsilonClosure(state *LNFAState, closure map[*LNFAState]struct{}) {
    if _, ok := closure[state]; ok { return }
    closure[state] = struct{}{} // Mark state as part of the closure
    for _, s := range state.Epsilon {
        // Add states reachable through epsilon transitions to closure
        epsilonClosure(s, closure)
    }
}

// For a set of NFA states, find accept status with the minimum priority value in the broader NFA.
func (n LNFA) resolveAccept(closure map[*LNFAState]struct{}) (string, bool) {
    // Handle accepting states, prioritizing tokens listed earlier
    accept := LNFAAccept { "", -1 }
    for state := range closure {
        id, ok := n.Accept[state]
        if ok && (accept.Priority == -1 || id.Priority < accept.Priority) { accept = id }
    }
    return accept.Identifier, accept.Priority != -1
}

// ------------------------------------------------------------------------------------------------------------------------------

// Creates unique identifier string given a closure of NFA states for use in a map.
func getClosureKey(closure map[*LNFAState]struct{}) string {
    // Sort states by address to ensure identical subsets map to the same key
    pointers := make([]uintptr, 0, len(closure))
    for state := range closure { pointers = append(pointers, uintptr(unsafe.Pointer(state))) }
    slices.Sort(pointers)
    // Interpret state memory addresses as consecutive bytes, then reinterpret as string
    const UINTPTR_SIZE int = int(unsafe.Sizeof(uintptr(0)))
    bytes := make([]byte, 0, len(pointers) * UINTPTR_SIZE)
    for _, p := range pointers {
        b := (*[UINTPTR_SIZE]byte)(unsafe.Pointer(&p))
        bytes = append(bytes, b[:]...)
    }
    return string(bytes)
}

// Adds epsilon transition to non-deterministic finite automata state.
func (s *LNFAState) AddEpsilon(states ...*LNFAState) { s.Epsilon = append(s.Epsilon, states...) }

// Duplicates an NFA state and all other states reachable through its transitions.
func (s *LNFAState) copy(copied map[*LNFAState]*LNFAState) *LNFAState {
    // If state has already been copied, return stored copy
    if state, ok := copied[s]; ok { return state }
    // Create and store new state struct and copy transitions
    copy := &LNFAState { make(map[Range]*LNFAState, len(s.Transitions)), make([]*LNFAState, len(s.Epsilon), cap(s.Epsilon)) }
    copied[s] = copy
    for value, state := range s.Transitions { copy.Transitions[value] = state.copy(copied) }
    for i, state := range s.Epsilon { copy.Epsilon[i] = state.copy(copied) }
    return copy
}

// Duplicate transitions based on expansion map created by createExpansionMap().
func (s *LNFAState) expand(expansion map[Range][]Range, visited map[*LNFAState]struct{}) {
    // If state has already been expanded, exit
    if _, ok := visited[s]; ok { return }
    visited[s] = struct{}{}
    // Expand transitions based on provided expansion map
    expanded := make(map[Range]*LNFAState, len(s.Transitions))
    for value, state := range s.Transitions {
        if ranges := expansion[value]; ranges != nil {
            for _, r := range ranges { expanded[r] = state }
        } else { expanded[value] = state } // If range has no expansion, keep as is
        state.expand(expansion, visited)
    }
    // Epsilon transitions are not expanded, but may reach states with transitions that need expansion
    for _, state := range s.Epsilon { state.expand(expansion, visited) }
    s.Transitions = expanded
}

func (r Range) String() string {
    if r.Min == r.Max { return formatChar(r.Min) }
    return fmt.Sprintf("%s-%s", formatChar(r.Min), formatChar(r.Max))
}

func formatChar(char rune) string {
    str := fmt.Sprintf("%q", string(char))
    return str[1:len(str) - 1]
}

// FOR DEBUG PURPOSES:
// Prints all transitions formatted for use in a graph visualizer.
func (n LNFA) PrintTransitions() {
    fmt.Println("digraph {")
    fmt.Println("    node [shape=\"circle\"];")
    states := make(map[*LNFAState]int)
    n.Start.printTransitions(states)
    for state := range n.Accept { fmt.Printf("    %d [shape=\"doublecircle\"];\n", states[state]) }
    fmt.Println("}")
}
// FOR DEBUG PURPOSES:
// Prints all transitions formatted for use in a graph visualizer.
func (d DFA[T]) PrintTransitions() {
    fmt.Println("digraph {")
    fmt.Println("    node [shape=\"circle\"];")
    states := make(map[*DFAState[T]]int)
    d.Start.printTransitions(states)
    for state := range d.Accept { fmt.Printf("    %d [shape=\"doublecircle\"];\n", states[state]) }
    fmt.Println("}")
}

func (s *LNFAState) printTransitions(states map[*LNFAState]int) int {
    if v, ok := states[s]; ok { return v }
    re := regexp.MustCompile(`\\([^"\\])`)
    i := len(states); states[s] = i
    for value, state := range s.Transitions {
        str := re.ReplaceAllString(value.String(), "\\\\$1")
        fmt.Printf("    %d -> %d [label=\"%s\"];\n", i, state.printTransitions(states), str)
    }
    for _, state := range s.Epsilon {
        fmt.Printf("    %d -> %d [label=\"ε\"];\n", i, state.printTransitions(states))
    }
    return i
}
func (s *DFAState[T]) printTransitions(states map[*DFAState[T]]int) int {
    if v, ok := states[s]; ok { return v }
    re := regexp.MustCompile(`\\([^"\\])`)
    i := len(states); states[s] = i
    for value, state := range s.Transitions {
        str := re.ReplaceAllString(value.String(), "\\\\$1")
        fmt.Printf("    %d -> %d [label=\"%s\"];\n", i, state.printTransitions(states), str)
    }
    return i
}
