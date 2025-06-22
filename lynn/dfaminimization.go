package lynn

import (
	"fmt"
	"reflect"
)

// Deterministic finite automata transition value. Must be comparable and able to be converted to a string.
type DFAValue interface {
    comparable
    fmt.Stringer
}
// Deterministic finite automata struct.
type DFA[T DFAValue] struct {
    Start  *DFAState[T]
    States []*DFAState[T]
    Accept map[*DFAState[T]]string
}
// Deterministic finite automata state struct.
// Holds references to outgoing states and each transition's associated value.
type DFAState[T DFAValue] struct { Transitions map[T]*DFAState[T] }

// Implementation of Hopcroft's algorithm. Computes the minimal DFA equivalent to the original.
// This function modifies the original DFA states and thus is destructive.
func minimizeDFA[T DFAValue](dfa DFA[T], values []T) DFA[T] {
    // Create initial coarse partition and work list
    partition := getInitialPartition(dfa)
    work := make([]map[*DFAState[T]]struct{}, len(partition))
    copy(work, partition)
    // DFA minimization via Hopcroft's algorithm
    for len(work) > 0 {
        target := work[0]; work = work[1:]
        for _, value := range values {
            main: for i, p := range partition {
                // Split subset in partition based on if they are able to transition to the target set on r
                p1, p2 := splitSubset(p, target, value)
                if len(p1) == 0 || len(p2) == 0 { continue }
                partition[i] = p1; partition = append(partition, p2) // Does not affect the current loop
                // If p is present in the work list, also split the subset there
                for i, w := range work {
                    if reflect.ValueOf(w).Pointer() != reflect.ValueOf(p).Pointer() { continue }
                    work[i] = p1; work = append(work, p2)
                    continue main
                }
                // Otherwise add the smaller set to the work list
                if len(p1) < len(p2) {
                    work = append(work, p1)
                } else {
                    work = append(work, p2)
                }
            }
        }
    }
    return mergeIndistinguishable(dfa, partition)
}

// Separate all DFA states based on whether they are accepting or non-accepting.
// Accepting states that resolve to different token identifiers are placed in different subsets.
func getInitialPartition[T DFAValue](dfa DFA[T]) []map[*DFAState[T]]struct{} {
    // Group states based on their accept status
    // Non-accepting states are given the empty string as an identifier
    subsets := make(map[string]map[*DFAState[T]]struct{})
    for _, state := range dfa.States {
        id := dfa.Accept[state]
        subset := subsets[id]
        if subset == nil {
            subset = make(map[*DFAState[T]]struct{})
            subsets[id] = subset
        }
        subset[state] = struct{}{}
    }
    // Transfer groups to slice
    partition := make([]map[*DFAState[T]]struct{}, 0, len(subsets))
    for _, subset := range subsets {
        partition = append(partition, subset)
    }
    return partition
}

// Given a target set and a range, go through all states in the subset and test if a transition on the given range
// exists such that we are able to move from the state to one in the target set. Splits the subset accordingly.
func splitSubset[T DFAValue](subset map[*DFAState[T]]struct{}, target map[*DFAState[T]]struct{},
    value T) (map[*DFAState[T]]struct{}, map[*DFAState[T]]struct{}) {
    // p1 is the set of states where this holds, and p2 is the complement
    p1, p2 := make(map[*DFAState[T]]struct{}), make(map[*DFAState[T]]struct{})
    for state := range subset {
        // For the transition from the state on r, see if the next state is included in the target set
        if _, ok := target[state.Transitions[value]]; ok {
            p1[state] = struct{}{}
        } else {
            p2[state] = struct{}{}
        }
    }
    return p1, p2
}

// Given a partition of the DFA states, choose a representative among each subset and replace all non-representatives
// with the representative of their subset.
func mergeIndistinguishable[T DFAValue](dfa DFA[T], partition []map[*DFAState[T]]struct{}) DFA[T] {
    // Create replacement map
    merge := make(map[*DFAState[T]]*DFAState[T])
    for _, subset := range partition {
        if len(subset) == 1 { continue }
        states := make([]*DFAState[T], 0, len(subset))
        for state := range subset { states = append(states, state) }
        // For each subset, choose a representative state and map all other states to the representative
        representative := states[0]
        for _, state := range states[1:] { merge[state] = representative }
    }
    // Replace start node with representative from subset
    start := merge[dfa.Start]
    if start == nil { start = dfa.Start }
    // Recursively replace all states with their representatives
    visited := make(map[*DFAState[T]]struct{})
    start.merge(merge, visited)
    states, accept := make([]*DFAState[T], 0, len(visited)), make(map[*DFAState[T]]string)
    for state := range visited {
        states = append(states, state)
        id, ok := dfa.Accept[state]
        if ok { accept[state] = id }
    }
    return DFA[T] { start, states, accept }
}

// Recursively replace states with their representatives as described in mergeIndistinguishable().
func (s *DFAState[T]) merge(merge map[*DFAState[T]]*DFAState[T], visited map[*DFAState[T]]struct{}) {
    // If state has already been visited, exit
    if _, ok := visited[s]; ok { return }
    visited[s] = struct{}{}
    // For all states reachable through this state, replace with subset representative
    for value, state := range s.Transitions {
        if representative := merge[state]; representative != nil {
            state = representative
            s.Transitions[value] = state
        }
        state.merge(merge, visited)
    }
}
