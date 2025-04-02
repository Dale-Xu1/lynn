package lynn

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
