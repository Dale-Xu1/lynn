package lynn

import "fmt"

// Non-deterministic finite automata struct.
type NFA struct {
	Start  *NFAState
	Accept []*NFAState
}

const (
	EPSILON rune = -1
	ANY rune = -2
)

// Non-deterministic finite automata fragment struct. Holds references to the in and out states.
type NFAFragment struct { In, Out *NFAState }
// Non-deterministic finite automata state struct. Holds references to outgoing states and each transition's associated character.
type NFAState struct {
	Transitions map[rune][]*NFAState
}

// Generator struct. Converts abstract syntax tree (AST) to finite automata (FA).
type Generator struct {
	fragments map[string]NFAFragment	
}

// Returns a new generator struct.
func NewGenerator() *Generator { return &Generator { make(map[string]NFAFragment, 20) } }
// Converts regular expressions defined in grammar into a non-deterministic finite automata.
func (g *Generator) GenerateNFA(grammar *GrammarNode) NFA {
	for _, fragment := range grammar.Fragments {
		// Convert fragment expressions to NFAs and add fragment to identifier map
		nfa, ok := g.expressionNFA(fragment.Expression)
		if ok { g.fragments[fragment.Identifier] = nfa }
	}
	start := &NFAState { make(map[rune][]*NFAState, len(grammar.Tokens)) }
	accept := make([]*NFAState, 0, len(grammar.Tokens))
	for _, token := range grammar.Tokens {
		// Convert token expressions to NFAs and attach fragment to final NFA
		if nfa, ok := g.expressionNFA(token.Expression); ok {
			start.AddTransition(EPSILON, nfa.In)
			accept = append(accept, nfa.Out)
		}
	}
	return NFA { start, accept }
}

func (g *Generator) expressionNFA(expression AST) (NFAFragment, bool) {
	switch node := expression.(type) {
	// Implementation of Thompson's construction to generate an NFA from the AST
	case *OptionNode:
		nfa, ok := g.expressionNFA(node.Expression); if !ok { return nfa, ok }
		out := &NFAState { make(map[rune][]*NFAState, 1) }
		in := &NFAState { map[rune][]*NFAState { EPSILON: { nfa.In, out } } }
		nfa.Out.AddTransition(EPSILON, out)
		return NFAFragment { in, out }, true
	case *RepeatNode:
		nfa, ok := g.expressionNFA(node.Expression); if !ok { return nfa, ok }
		out := &NFAState { make(map[rune][]*NFAState, 1) }
		in := &NFAState { map[rune][]*NFAState { EPSILON: { nfa.In, out } } }
		nfa.Out.AddTransition(EPSILON, nfa.In, out)
		return NFAFragment { in, out }, true
	case *RepeatOneNode:
		nfa, ok := g.expressionNFA(node.Expression); if !ok { return nfa, ok }
		out := &NFAState { make(map[rune][]*NFAState, 1) }
		in := &NFAState { map[rune][]*NFAState { EPSILON: { nfa.In } } }
		nfa.Out.AddTransition(EPSILON, nfa.In, out)
		return NFAFragment { in, out }, true

	case *ConcatenationNode:
		a, ok := g.expressionNFA(node.A); if !ok { return a, ok }
		b, ok := g.expressionNFA(node.B); if !ok { return b, ok }
		// Create epsilon transition between two fragments
		a.Out.AddTransition(EPSILON, b.In)
		return NFAFragment { a.In, b.Out }, true
	case *UnionNode:
		a, ok := g.expressionNFA(node.A); if !ok { return a, ok }
		b, ok := g.expressionNFA(node.B); if !ok { return b, ok }
		// Create in state with epsilon transitions to in states of both fragments
		in, out := &NFAState { map[rune][]*NFAState { EPSILON: { a.In, b.In } } }, &NFAState { make(map[rune][]*NFAState, 1) }
		a.Out.AddTransition(EPSILON, out) // Create epsilon transitions from out states of fragments to final out state
		b.Out.AddTransition(EPSILON, out)
		return NFAFragment { in, out }, true

	// Generate NFAs for literals
	case *AnyNode:
		out := &NFAState { make(map[rune][]*NFAState, 1) }
		in := &NFAState { map[rune][]*NFAState { ANY: { out }, '\n': nil, '\r': nil } } // . does not match new lines
		return NFAFragment { in, out }, true
	case *IdentifierNode: return g.copyFragment(node.Name)
	case *StringNode:
		// Generate chain of states with transitions at each consecutive character
		out := &NFAState { make(map[rune][]*NFAState, 1) }
		state := out
		for i := len(node.Chars) - 1; i >= 0; i-- {
			char := node.Chars[i]
			state = &NFAState { map[rune][]*NFAState { char: { state } } }
		}
		return NFAFragment { state, out }, true
	case *ClassNode:
		var in *NFAState; var out *NFAState
		if !node.Negated {
			// Add transition from in to out for each character in class
			in, out = &NFAState { make(map[rune][]*NFAState, len(node.Chars)) }, &NFAState { make(map[rune][]*NFAState, 1) }
			for _, c := range node.Chars { in.Transitions[c] = []*NFAState { out } }
		} else {
			// Add nil transition for each character in class and create any transition from in to out
			in, out = &NFAState { make(map[rune][]*NFAState, len(node.Chars) + 1) }, &NFAState { make(map[rune][]*NFAState, 1) }
			in.Transitions[ANY] = []*NFAState { out }
			for _, c := range node.Chars { in.Transitions[c] = nil }
		}
		return NFAFragment { in, out }, true
	default: panic("Invalid expression passed to expressionNFA()")
	}
}

func (g *Generator) copyFragment(identifier string) (NFAFragment, bool) {
	fa, ok := g.fragments[identifier]
	if !ok {
		fmt.Printf("Generation error: Fragment \"%s\" does not exist\n", identifier) // TODO: Store location in AST
		return NFAFragment { }, false
	}
	states := make(map[*NFAState]*NFAState, 10)
	return NFAFragment { fa.In.copy(states), states[fa.Out] }, true
}

func (s *NFAState) copy(states map[*NFAState]*NFAState) *NFAState {
	if state, ok := states[s]; ok { return state }
	copy := &NFAState { make(map[rune][]*NFAState, len(s.Transitions)) }
	states[s] = copy
	for value, out := range s.Transitions {
		copy.Transitions[value] = make([]*NFAState, len(out), cap(out))
		for i, state := range out { copy.Transitions[value][i] = state.copy(states) }
	}
	return copy
}

func (s *NFAState) AddTransition(value rune, states ...*NFAState) {
	if out, ok := s.Transitions[value]; ok {
		for _, state := range states { s.Transitions[value] = append(out, state) }
	} else {
		s.Transitions[value] = states
	}
}
