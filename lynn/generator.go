package lynn

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"unsafe"
)

// Represents a range between characters.
type Range struct { Min, Max rune }

// Non-deterministic finite automata struct.
type NFA struct {
    Start  *NFAState
    Accept map[*NFAState]NFAAccept
}
// Deterministic finite automata struct.
type DFA struct {
    Start  *DFAState
    States []*DFAState
    Accept map[*DFAState]string
}

// Non-deterministic finite automata fragment struct. Holds references to the in and out states.
type NFAFragment struct { In, Out *NFAState }
// Non-deterministic finite automata accepting state metadata. Holds token identifier and priority.
type NFAAccept struct { Identifier string; Priority int }
// Non-deterministic finite automata state struct.
// Holds references to outgoing states and each transition's associated character range.
type NFAState struct {
    Transitions map[Range]*NFAState
    Epsilon     []*NFAState
}
// Deterministic finite automata state struct.
// Holds references to outgoing states and each transition's associated character range.
type DFAState struct { Transitions map[Range]*DFAState }

// Generator struct. Converts abstract syntax tree (AST) to finite automata (FA).
type Generator struct {
    fragments map[string]NFAFragment	
    ranges    map[Range]bool
    accept    map[*DFAState]string
}

// Returns a new generator struct.
func NewGenerator() *Generator { return &Generator { } }
// Converts regular expressions defined in grammar into a non-deterministic finite automata.
func (g *Generator) GenerateNFA(grammar *GrammarNode) (NFA, []Range) {
    g.fragments = make(map[string]NFAFragment)
    g.ranges = make(map[Range]bool)
    for _, fragment := range grammar.Fragments {
        // Convert fragment expressions to NFAs and add fragment to identifier map
        nfa, ok := g.expressionNFA(fragment.Expression)
        if _, exists := g.fragments[fragment.Identifier]; exists {
            fmt.Printf("Generation error: Fragment \"%s\" is already defined\n", fragment.Identifier)
        } else if ok { g.fragments[fragment.Identifier] = nfa }
    }
    start := &NFAState { make(map[Range]*NFAState, 0), make([]*NFAState, 0, len(grammar.Tokens)) }
    accept, tokens := make(map[*NFAState]NFAAccept, len(grammar.Tokens)), make(map[string]bool)
    for i, token := range grammar.Tokens {
        // Convert token expressions to NFAs and attach fragment to final NFA
        nfa, ok := g.expressionNFA(token.Expression)
        if tokens[token.Identifier] {
            fmt.Printf("Generation error: Token \"%s\" is already defined\n", token.Identifier)
            continue
        }
        tokens[token.Identifier] = true
        if ok {
            start.AddEpsilon(nfa.In)
            accept[nfa.Out] = NFAAccept { token.Identifier, i }
        }
    }
    // Disjoin all occurring ranges and apply to transitions in NFA
    ranges, expansion := disjoinRanges(g.ranges)
    start.expand(expansion, make(map[*NFAState]bool))
    return NFA { start, accept }, ranges
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
func disjoinRanges(ranges map[Range]bool) ([]Range, map[Range][]Range) {
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

// Converts non-deterministic finite automata to deterministic finite automata.
func (g *Generator) NFAtoDFA(nfa NFA, ranges []Range) DFA {
    subsets := make(map[string]*DFAState)
    g.accept = make(map[*DFAState]string, len(nfa.Accept))
    // Convert NFA to initial DFA through power-set construction
    start := g.mergeTransitions([]*NFAState { nfa.Start }, nfa, subsets)
    // Extract all DFA states from subset map
    states := make([]*DFAState, 0, len(subsets))
    for _, state := range subsets { states = append(states, state) }
    // Find minimal DFA after construction
    dfa := DFA { start, states, g.accept }
    return minimizeDFA(dfa, ranges) 
}

func (g *Generator) mergeTransitions(states []*NFAState, nfa NFA, subsets map[string]*DFAState) *DFAState {
    // Find epsilon closure of a given set of states
    closure, visited := make([]*NFAState, 0), make(map[*NFAState]bool)
    for _, state := range states { epsilonClosure(state, &closure, visited) }
    // If closure has already been processed, return reference to DFA state
    key := getClosureKey(closure)
    if state := subsets[key]; state != nil { return state }

    // Go through all transitions of all states in closure and find possible subsets reachable through each range
    l := 0; for _, state := range closure { l += len(state.Transitions) }
    merged := make(map[Range][]*NFAState, l)
    for _, state := range closure {
        for value, s := range state.Transitions {
            if merged[value] != nil {
                merged[value] = append(merged[value], s)
            } else { merged[value] = []*NFAState { s } }
        }
    }
    // Create and store DFA state
    state := &DFAState{ make(map[Range]*DFAState, len(merged)) }
    subsets[key] = state
    if identifier := nfa.resolveAccept(closure); identifier != "" { g.accept[state] = identifier }
    // Convert each subset that this state may transition into its own DFA state recursively
    for value, states := range merged {
        state.Transitions[value] = g.mergeTransitions(states, nfa, subsets)
    }
    return state
}

func epsilonClosure(state *NFAState, closure *[]*NFAState, visited map[*NFAState]bool) {
    if visited[state] { return }
    visited[state] = true // Mark state as visited
    *closure = append(*closure, state)
    for _, s := range state.Epsilon {
        // Add states reachable through epsilon transitions to closure
        epsilonClosure(s, closure, visited)
    }
}

func getClosureKey(closure []*NFAState) string {
    // Sort states by address to ensure identical subsets map to the same key
    sort.Slice(closure, func (i, j int) bool { return uintptr(unsafe.Pointer(closure[i])) < uintptr(unsafe.Pointer(closure[j])) })
    // Interpret state memory addresses as consecutive bytes, then reinterpret as string
    const UINTPTR_SIZE int = int(unsafe.Sizeof(uintptr(0)))
    bytes := make([]byte, 0, len(closure) * UINTPTR_SIZE)
    for _, state := range closure {
        b := (*[UINTPTR_SIZE]byte)(unsafe.Pointer(&state))
        bytes = append(bytes, b[:]...)
    }
    return string(bytes)
}

func (nfa NFA) resolveAccept(closure []*NFAState) string {
    // Handle accepting states, prioritizing tokens listed earlier
    accept := NFAAccept { "", -1 }
    for _, state := range closure {
        s, ok := nfa.Accept[state]
        if ok && (accept.Priority == -1 || s.Priority < accept.Priority) { accept = s }
    }
    return accept.Identifier
}

func minimizeDFA(dfa DFA, ranges []Range) DFA {
    // Create initial coarse partition and work list
    partition := getInitialPartition(dfa)
    work := make([]map[*DFAState]bool, len(partition))
    copy(work, partition)
    // DFA minimization via Hopcroft's algorithm
    for len(work) > 0 {
        target := work[0]; work = work[1:]
        for _, r := range ranges {
            for i, p := range partition {
                // Split subset in partition based on if they are able to transition to the target set on r
                p1, p2 := splitSubset(p, target, r)
                if len(p1) == 0 || len(p2) == 0 { continue }
                partition[i] = p1; partition = append(partition, p2) // Does not affect the current loop
                // If p is present in the work list, also split the subset there
                includes := false
                for i, w := range work {
                    if reflect.ValueOf(w).Pointer() != reflect.ValueOf(p).Pointer() { continue }
                    includes = true
                    work[i] = p1; work = append(work, p2)
                    break
                }
                if !includes {
                    // Otherwise add the smaller set to the work list
                    if len(p1) < len(p2) {
                        work = append(work, p1)
                    } else {
                        work = append(work, p2)
                    }
                }
            }
        }
    }
    return mergeIndistinguishable(dfa, partition)
}

func getInitialPartition(dfa DFA) []map[*DFAState]bool {
    // Separate all DFA states based on whether they are accepting or non-accepting
    // Accepting states that resolve to different token identifiers are placed in different subsets
    subsets := make(map[string]map[*DFAState]bool)
    for _, state := range dfa.States {
        identifier := dfa.Accept[state]
        subset := subsets[identifier]
        if subset == nil {
            subset = make(map[*DFAState]bool)
            subsets[identifier] = subset
        }
        subset[state] = true
    }
    partition := make([]map[*DFAState]bool, 0, len(subsets))
    for _, subset := range subsets {
        partition = append(partition, subset)
    }
    return partition
}

func splitSubset(subset map[*DFAState]bool, target map[*DFAState]bool, r Range) (map[*DFAState]bool, map[*DFAState]bool) {
    // Given a target set and a range, go through all states in the subset and test if a transition on the given range
    // exists such that we are able to move from the state to one in the target set
    // p1 is the set of states where this holds, and p2 is the complement
    p1, p2 := make(map[*DFAState]bool), make(map[*DFAState]bool)
    for state := range subset {
        // For the transition from the state on r, see if the next state is included in the target set
        if target[state.Transitions[r]] {
            p1[state] = true
        } else {
            p2[state] = true
        }
    }
    return p1, p2
}

func mergeIndistinguishable(dfa DFA, partition []map[*DFAState]bool) DFA {
    // TODO: Create replacement map
    return dfa
}

func (r Range) String() string {
    if r.Min == r.Max { return formatChar(r.Min) }
    return fmt.Sprintf("%s-%s", formatChar(r.Min), formatChar(r.Max))
}

// Adds epsilon transition to non-deterministic finite automata state.
func (s *NFAState) AddEpsilon(states ...*NFAState) { s.Epsilon = append(s.Epsilon, states...) }

func (s *NFAState) copy(copied map[*NFAState]*NFAState) *NFAState {
    // If state has already been copied, return stored copy
    if state, ok := copied[s]; ok { return state }
    // Create and store new state struct and copy transitions
    copy := &NFAState { make(map[Range]*NFAState, len(s.Transitions)), make([]*NFAState, len(s.Epsilon), cap(s.Epsilon)) }
    copied[s] = copy
    for value, state := range s.Transitions { copy.Transitions[value] = state.copy(copied) }
    for i, state := range s.Epsilon { copy.Epsilon[i] = state.copy(copied) }
    return copy
}

func (s *NFAState) expand(expansion map[Range][]Range, visited map[*NFAState]bool) {
    // If state has already been expanded, exit
    if visited[s] { return }
    visited[s] = true
    // Expand transitions based on provided expansion map
    expanded := make(map[Range]*NFAState, len(s.Transitions))
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

// FOR DEBUG PURPOSES:
// Prints all transitions formatted for use in a graph visualizer.
func (n NFA) PrintTransitions() {
    fmt.Println("digraph {")
    fmt.Println("    node [shape=\"circle\"];")
    states := make(map[*NFAState]int)
    n.Start.printTransitions(states)
    for state := range n.Accept { fmt.Printf("    %d [shape=\"doublecircle\"];\n", states[state]) }
    fmt.Println("}")
}
// FOR DEBUG PURPOSES:
// Prints all transitions formatted for use in a graph visualizer.
func (d DFA) PrintTransitions() {
    fmt.Println("digraph {")
    fmt.Println("    node [shape=\"circle\"];")
    states := make(map[*DFAState]int)
    d.Start.printTransitions(states)
    for state := range d.Accept { fmt.Printf("    %d [shape=\"doublecircle\"];\n", states[state]) }
    fmt.Println("}")
}

func (s *NFAState) printTransitions(states map[*NFAState]int) int {
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
func (s *DFAState) printTransitions(states map[*DFAState]int) int {
    if v, ok := states[s]; ok { return v }
    re := regexp.MustCompile(`\\([^"\\])`)
    i := len(states); states[s] = i
    for value, state := range s.Transitions {
        str := re.ReplaceAllString(value.String(), "\\\\$1")
        fmt.Printf("    %d -> %d [label=\"%s\"];\n", i, state.printTransitions(states), str)
    }
    return i
}
