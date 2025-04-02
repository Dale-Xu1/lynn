package lynn

// Non-deterministic finite automata struct.
type PNFA struct {
    Start  *PNFAState
    Accept map[*PNFAState]PNFAAccept
}
// Non-deterministic finite automata fragment struct. Holds references to the in and out states.
type PNFAFragment struct { In, Out *PNFAState }
// Non-deterministic finite automata accepting state metadata. Holds rule identifier and priority.
type PNFAAccept struct { Identifier string; Priority int }
// Non-deterministic finite automata state struct.
// Holds references to outgoing states and each transition's associated token type.
type PNFAState struct {
    Transitions map[Range]*PNFAState
    Epsilon     []*PNFAState
}

// Deterministic finite automata struct.
type PDFA = DFA[TokenType]
// Deterministic finite automata state struct.
// Holds references to outgoing states and each transition's associated token type.
type PDFAState = DFAState[TokenType]

// Parser generator struct. Converts grammar definition in abstract syntax tree (AST) to finite automata (FA).
type ParserGenerator struct {
}

// Returns a new parser generator struct.
func NewParserGenerator() *ParserGenerator { return &ParserGenerator { } }
