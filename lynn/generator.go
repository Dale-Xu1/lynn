package lynn

import (
	"fmt"
	"regexp"
	"sort"
)

// Non-deterministic finite automata struct.
type NFA struct {
    Start  *NFAState
    Accept []*NFAState
}

// Non-deterministic finite automata fragment struct. Holds references to the in and out states.
type NFAFragment struct { In, Out *NFAState }
// Non-deterministic finite automata state struct. Holds references to outgoing states and each transition's associated character.
type NFAState struct {
    Transitions map[Range]*NFAState
    Epsilon     []*NFAState
}

// Represents a range between characters.
type Range struct { Min, Max rune }

// Generator struct. Converts abstract syntax tree (AST) to finite automata (FA).
type Generator struct {
    fragments map[string]NFAFragment	
    ranges    map[Range]bool
}

// Returns a new generator struct.
func NewGenerator() *Generator { return &Generator { make(map[string]NFAFragment, 20), make(map[Range]bool, 50) } }
// Converts regular expressions defined in grammar into a non-deterministic finite automata.
func (g *Generator) GenerateNFA(grammar *GrammarNode) NFA {
    for _, fragment := range grammar.Fragments {
        // Convert fragment expressions to NFAs and add fragment to identifier map
        nfa, ok := g.expressionNFA(fragment.Expression)
        if ok { g.fragments[fragment.Identifier] = nfa }
    }
    start := &NFAState { make(map[Range]*NFAState, 0), make([]*NFAState, 0, len(grammar.Tokens)) }
    accept := make([]*NFAState, 0, len(grammar.Tokens))
    for _, token := range grammar.Tokens {
        // Convert token expressions to NFAs and attach fragment to final NFA
        if nfa, ok := g.expressionNFA(token.Expression); ok {
            start.AddEpsilon(nfa.In)
            accept = append(accept, nfa.Out)
        }
    }
    // Disjoin all occurring ranges and apply to transitions in NFA
    expansion := disjoinRanges(g.ranges)
    start.expand(expansion, make(map[*NFAState]bool, 50))
    return NFA { start, accept }
}

func (g *Generator) expressionNFA(expression AST) (NFAFragment, bool) {
    switch node := expression.(type) {
    // Implementation of Thompson's construction to generate an NFA from the AST
    case *OptionNode:
        nfa, ok := g.expressionNFA(node.Expression); if !ok { return nfa, ok }
        out := &NFAState { make(map[Range]*NFAState, 0), make([]*NFAState, 0, 1) }
        in := &NFAState { make(map[Range]*NFAState, 0), []*NFAState { nfa.In, out } }
        nfa.Out.AddEpsilon(out)
        return NFAFragment { in, out }, true
    case *RepeatNode:
        nfa, ok := g.expressionNFA(node.Expression); if !ok { return nfa, ok }
        out := &NFAState { make(map[Range]*NFAState, 0), make([]*NFAState, 0, 1) }
        in := &NFAState { make(map[Range]*NFAState, 0), []*NFAState { nfa.In, out } }
        nfa.Out.AddEpsilon(nfa.In, out)
        return NFAFragment { in, out }, true
    case *RepeatOneNode:
        nfa, ok := g.expressionNFA(node.Expression); if !ok { return nfa, ok }
        out := &NFAState { make(map[Range]*NFAState, 0), make([]*NFAState, 0, 1) }
        in := &NFAState { make(map[Range]*NFAState, 0), []*NFAState { nfa.In } }
        nfa.Out.AddEpsilon(nfa.In, out)
        return NFAFragment { in, out }, true

    case *ConcatenationNode:
        a, ok := g.expressionNFA(node.A); if !ok { return a, ok }
        b, ok := g.expressionNFA(node.B); if !ok { return b, ok }
        // Create epsilon transition between two fragments
        a.Out.AddEpsilon(b.In)
        return NFAFragment { a.In, b.Out }, true
    case *UnionNode:
        a, ok := g.expressionNFA(node.A); if !ok { return a, ok }
        b, ok := g.expressionNFA(node.B); if !ok { return b, ok }
        // Create in state with epsilon transitions to in states of both fragments
        in, out := &NFAState { make(map[Range]*NFAState, 0), []*NFAState { a.In, b.In } },
            &NFAState { make(map[Range]*NFAState, 0), make([]*NFAState, 0, 1) }
        a.Out.AddEpsilon(out) // Create epsilon transitions from out states of fragments to final out state
        b.Out.AddEpsilon(out)
        return NFAFragment { in, out }, true

    // Generate NFAs for literals
    case *IdentifierNode: return g.getFragment(node.Name)
    case *StringNode:
        // Generate chain of states with transitions at each consecutive character
        out := &NFAState { make(map[Range]*NFAState, 0), make([]*NFAState, 0, 1) }
        state := out
        for i := len(node.Chars) - 1; i >= 0; i-- {
            char := node.Chars[i]
            r := Range { char, char }; g.ranges[r] = true
            state = &NFAState { map[Range]*NFAState { r: state }, make([]*NFAState, 0, 1) }
        }
        return NFAFragment { state, out }, true
    case *ClassNode:
        var in, out *NFAState
        // Add transition from in to out for each character in class
        in, out = &NFAState { make(map[Range]*NFAState, len(node.Ranges)), make([]*NFAState, 0, 1) },
            &NFAState { make(map[Range]*NFAState, 0), make([]*NFAState, 0, 1) }
        for _, r := range node.Ranges {
            in.Transitions[r] = out
            g.ranges[r] = true
        }
        return NFAFragment { in, out }, true
    default: panic("Invalid expression passed to expressionNFA()")
    }
}

func (g *Generator) getFragment(identifier string) (NFAFragment, bool) {
    fa, ok := g.fragments[identifier]
    if !ok {
        fmt.Printf("Generation error: Fragment \"%s\" does not exist\n", identifier) // TODO: Store location in AST
        return NFAFragment { }, false
    }
    states := make(map[*NFAState]*NFAState, 10)
    return NFAFragment { fa.In.copy(states), states[fa.Out] }, true
}

// This function converts a set of ranges to a one that is mutually disjoint and has the same union.
// This operates by splitting the original ranges rather than merging them. Final output is a map from
// the original range to its corresponding set of disjoined ranges.
// For example: [1, 5], [4, 10] -> [1, 4], [4, 5], [5, 10]
func disjoinRanges(ranges map[Range]bool) map[Range][]Range {
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
    return createExpansionMap(ranges, disjoined)
}

func createExpansionMap(ranges map[Range]bool, disjoined []Range) map[Range][]Range {
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
        l := max - min + 1
        expanded := make([]Range, l); expansion[r] = expanded
        for i := range expanded {
            expanded[i] = disjoined[min + i]
        }
    }
    return expansion
}

func (s *NFAState) AddEpsilon(states ...*NFAState) { s.Epsilon = append(s.Epsilon, states...) }

func (s *NFAState) copy(states map[*NFAState]*NFAState) *NFAState {
    if state, ok := states[s]; ok { return state }
    copy := &NFAState { make(map[Range]*NFAState, len(s.Transitions)), make([]*NFAState, len(s.Epsilon), cap(s.Epsilon)) }
    states[s] = copy
    for value, state := range s.Transitions { copy.Transitions[value] = state.copy(states) }
    for i, state := range s.Epsilon { copy.Epsilon[i] = state.copy(states) }
    return copy
}

func (s *NFAState) expand(expansion map[Range][]Range, visited map[*NFAState]bool) {
    if visited[s] { return }
    visited[s] = true
    expanded := make(map[Range]*NFAState, len(s.Transitions))
    for value, state := range s.Transitions {
        ranges, ok := expansion[value]
        if ok {
            for _, r := range ranges { expanded[r] = state }
        } else { expanded[value] = state }
        state.expand(expansion, visited)
    }
    for _, state := range s.Epsilon { state.expand(expansion, visited) }
    s.Transitions = expanded
}

func (r Range) String() string {
    if r.Min == r.Max { return formatChar(r.Min) }
    return fmt.Sprintf("%s-%s", formatChar(r.Min), formatChar(r.Max))
}

// FOR DEBUG PURPOSES:
// Prints all transitions formatted for use in a graph visualizer.
func (n NFA) PrintTransitions() {
    fmt.Println("digraph {")
    fmt.Println("    node [shape=\"circle\"];")
    states := make(map[*NFAState]int, 50)
    n.Start.printTransitions(states)
    for _, state := range n.Accept { fmt.Printf("    %d [shape=\"doublecircle\"]\n", states[state]) }
    fmt.Println("}")
}

func (s *NFAState) printTransitions(states map[*NFAState]int) int {
    if v, ok := states[s]; ok { return v }
    i := len(states); states[s] = i
    for value, state := range s.Transitions {
        re := regexp.MustCompile(`\\([^"\\])`)
        str := re.ReplaceAllString(value.String(), "\\\\$1")
        fmt.Printf("    %d -> %d [label=\"%s\"];\n", i, state.printTransitions(states), str)
    }
    for _, state := range s.Epsilon {
        fmt.Printf("    %d -> %d [label=\"ε\"];\n", i, state.printTransitions(states))
    }
    return i
}
