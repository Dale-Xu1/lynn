package lynn

// import "fmt"

// // Identifier of a token. Used as transition values for parser finite automata.
// type TokenIdentifier string
// func (t TokenIdentifier) String() string { return string(t) }

// // Non-deterministic finite automata struct.
// type PNFA struct {
//     Start  *PNFAState
//     Accept map[*PNFAState]PNFAAccept
// }
// // Non-deterministic finite automata fragment struct. Holds references to the in and out states.
// type PNFAFragment struct { In, Out *PNFAState }
// // Non-deterministic finite automata accepting state metadata. Holds rule identifier and priority.
// type PNFAAccept struct { Identifier string; Priority int }
// // Non-deterministic finite automata state struct.
// // Holds references to outgoing states and each transition's associated token identifier.
// type PNFAState struct {
//     Transitions map[TokenIdentifier]*PNFAState
//     Epsilon     []*PNFAState
// }

// // Deterministic finite automata struct.
// type PDFA = DFA[TokenIdentifier]
// // Deterministic finite automata state struct.
// // Holds references to outgoing states and each transition's associated token identifier.
// type PDFAState = DFAState[TokenIdentifier]

// // Parser generator struct. Converts grammar definition in abstract syntax tree (AST) to finite automata (FA).
// type ParserGenerator struct {
//     tokens map[TokenIdentifier]struct{}
//     strings map[string]TokenIdentifier
// }

// // Returns a new parser generator struct.
// func NewParserGenerator() *ParserGenerator { return &ParserGenerator { } }

// // Converts grammar rules into a non-deterministic finite automata.
// func (g *ParserGenerator) GenerateNFA(grammar *GrammarNode) {
//     // Generate set of valid tokens and create map from simple string tokens to their identifier
//     g.tokens = make(map[TokenIdentifier]struct{})
//     g.strings = make(map[string]TokenIdentifier)
//     for _, token := range grammar.Tokens {
//         id := TokenIdentifier(token.Identifier.Name)
//         g.tokens[id] = struct{}{}
//         // If a token's expression is only a string, the string itself is allowed to be used in rule expressions
//         if str, ok := token.Expression.(*StringNode); ok {
//             g.strings[string(str.Chars)] = id
//         }
//     }
// }

// // Converts a given expression node from the AST to an NFA fragment.
// func (g *ParserGenerator) expressionNFA(expression AST) (PNFAFragment, bool) {
//     switch node := expression.(type) {
//     // Implementation of Thompson's construction to generate an NFA from the AST
//     case *OptionNode:
//         nfa, ok := g.expressionNFA(node.Expression); if !ok { return nfa, ok }
//         out := &PNFAState { make(map[TokenIdentifier]*PNFAState, 0), make([]*PNFAState, 0) }
//         in := &PNFAState { make(map[TokenIdentifier]*PNFAState, 0), []*PNFAState { nfa.In, out } }
//         nfa.Out.AddEpsilon(out)
//         return PNFAFragment { in, out }, true
//     case *RepeatNode:
//         nfa, ok := g.expressionNFA(node.Expression); if !ok { return nfa, ok }
//         out := &PNFAState { make(map[TokenIdentifier]*PNFAState, 0), make([]*PNFAState, 0) }
//         in := &PNFAState { make(map[TokenIdentifier]*PNFAState, 0), []*PNFAState { nfa.In, out } }
//         nfa.Out.AddEpsilon(nfa.In, out)
//         return PNFAFragment { in, out }, true
//     case *RepeatOneNode:
//         nfa, ok := g.expressionNFA(node.Expression); if !ok { return nfa, ok }
//         out := &PNFAState { make(map[TokenIdentifier]*PNFAState, 0), make([]*PNFAState, 0) }
//         in := &PNFAState { make(map[TokenIdentifier]*PNFAState, 0), []*PNFAState { nfa.In } }
//         nfa.Out.AddEpsilon(nfa.In, out)
//         return PNFAFragment { in, out }, true

//     case *ConcatenationNode:
//         a, ok := g.expressionNFA(node.A); if !ok { return a, ok }
//         b, ok := g.expressionNFA(node.B); if !ok { return b, ok }
//         // Create epsilon transition between two fragments
//         a.Out.AddEpsilon(b.In)
//         return PNFAFragment { a.In, b.Out }, true
//     case *UnionNode:
//         a, ok := g.expressionNFA(node.A); if !ok { return a, ok }
//         b, ok := g.expressionNFA(node.B); if !ok { return b, ok }
//         // Create in state with epsilon transitions to in states of both fragments
//         in, out := &PNFAState { make(map[TokenIdentifier]*PNFAState, 0), []*PNFAState { a.In, b.In } },
//             &PNFAState { make(map[TokenIdentifier]*PNFAState, 0), make([]*PNFAState, 0) }
//         a.Out.AddEpsilon(out) // Create epsilon transitions from out states of fragments to final out state
//         b.Out.AddEpsilon(out)
//         return PNFAFragment { in, out }, true

//     // Generate NFAs for literals
//     case *IdentifierNode:
//         // Follows NFA construction for rule references based on LL(*)
//         panic("") // TODO: Token and rule references
//     case *StringNode:
//         // Find associated token identifier and create transition on that identifier
//         str := string(node.Chars)
//         id, ok := g.strings[str]
//         if ok {
//             fmt.Printf("Generation error: No token explicitly matches \"%s\" - %d:%d\n", str, node.Location.Line, node.Location.Col)
//             return PNFAFragment { }, false
//         }
//         out := &PNFAState { make(map[TokenIdentifier]*PNFAState, 0), make([]*PNFAState, 0) }
//         in := &PNFAState { map[TokenIdentifier]*PNFAState { id: out }, make([]*PNFAState, 0) }
//         return PNFAFragment { in, out }, true
//     case *ClassNode:
//         fmt.Printf("Generation error: Classes can not be used in rule expressions - %d:%d\n", node.Location.Line, node.Location.Col)
//         return PNFAFragment { }, false
//     default: panic("Invalid expression passed to ParserGenerator.expressionNFA()")
//     }
// }

// // Adds epsilon transition to non-deterministic finite automata state.
// func (s *PNFAState) AddEpsilon(states ...*PNFAState) { s.Epsilon = append(s.Epsilon, states...) }
