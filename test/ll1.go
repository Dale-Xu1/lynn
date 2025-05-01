package test

// import "fmt"

// type Symbol interface { isSymbol() }
// type NonTerminal uint
// type Terminal rune

// type LL1Parser struct {
//     rules  map[NonTerminal][][]Symbol
//     start  NonTerminal
//     first  map[NonTerminal]map[Terminal]struct{}
//     follow map[NonTerminal]map[Terminal]struct{}
//     table  map[NonTerminal]map[Terminal][]Symbol
//     stack  []Symbol
// }

// const (E NonTerminal = iota; R; T; Y; F)
// const EPSILON Terminal = '!'
// const END Terminal = '$'

// func NewLL1Parser() *LL1Parser {
//     rules := map[NonTerminal][][]Symbol {
//         E: {
//             { T, R },
//         },
//         R: {
//             { Terminal('|'), T, R },
//             { EPSILON },
//         },
//         T: {
//             { F, Y },
//         },
//         Y: {
//             { F, Y },
//             { EPSILON },
//         },
//         F: {
//             { Terminal('x') },
//             { Terminal('('), E, Terminal(')') },
//         },
//     }
//     return &LL1Parser {
//         rules,
//         E,
//         make(map[NonTerminal]map[Terminal]struct{}),
//         make(map[NonTerminal]map[Terminal]struct{}),
//         make(map[NonTerminal]map[Terminal][]Symbol),
//         make([]Symbol, 0),
//     }
// }

// // TODO: Precedence and associativity disambiguation
// // TODO: Left recursion elimination (indirect left recursion detection for error handling?)
// // TODO: Left factoring

// func (p *LL1Parser) Parse() {
//     for nt, rules := range p.rules {
//         p.table[nt] = make(map[Terminal][]Symbol)
//         for _, rule := range rules {
//             var set map[Terminal]struct{}
//             if len(rule) == 1 && rule[0] == EPSILON {
//                 set = p.findFollow(nt)
//             } else {
//                 set = p.productionFirst(rule)
//             }
//             for t := range set {
//                 if t == EPSILON { continue }
//                 if _, ok := p.table[nt][t]; ok { panic("Conflict in parse table: grammar is not LL(1)") }
//                 p.table[nt][t] = rule
//             }
//         }
//     }

//     input := []Terminal { 'x', '|', 'x', 'x', END }
//     index := 0
//     p.stack = []Symbol { END, p.start }
//     for len(p.stack) > 0 {
//         for i := len(p.stack) - 1; i >= 0; i-- { fmt.Printf("%v ", p.stack[i]) }
//         fmt.Println()

//         top := p.stack[len(p.stack) - 1]
//         i := input[index]

//         switch t := top.(type) {
//         case Terminal:
//             if t != i { panic("Unexpected token") }
//             p.stack = p.stack[:len(p.stack) - 1]
//             index++
//         case NonTerminal:
//             rule, ok := p.table[t][i]
//             if !ok { panic("Unexpected token") }
//             p.stack = p.stack[:len(p.stack) - 1]
//             if rule[0] == EPSILON { continue }
//             for i := len(rule) - 1; i >= 0; i-- {
//                 p.stack = append(p.stack, rule[i])
//             }
//         }
//     }
// }

// func (p *LL1Parser) productionFirst(rule []Symbol) map[Terminal]struct{} {
//     first := make(map[Terminal]struct{})
//     p.ruleFirst(first, rule)
//     return first
// }

// func (p *LL1Parser) findFirst(symbol Symbol) map[Terminal]struct{} {
//     switch t := symbol.(type) {
//     case Terminal: return map[Terminal]struct{} { t: {} }
//     case NonTerminal:
//         if first, ok := p.first[t]; ok { return first }
//         first := make(map[Terminal]struct{})
//         p.first[t] = first
//         for _, rule := range p.rules[t] { p.ruleFirst(first, rule) }
//         return first
//     default: panic("Invalid symbol passed to findFirst()")
//     }
// }

// func (p *LL1Parser) ruleFirst(first map[Terminal]struct{}, rule []Symbol) {
//     for _, s := range rule[:len(rule) - 1] {
//         f := p.findFirst(s)
//         for t := range f {
//             if t != EPSILON { first[t] = struct{}{} }
//         }
//         if _, ok := f[EPSILON]; !ok { return }
//     }
//     for t := range p.findFirst(rule[len(rule) - 1]) { first[t] = struct{}{} }
// }

// func (p *LL1Parser) findFollow(symbol NonTerminal) map[Terminal]struct{} {
//     if follow, ok := p.follow[symbol]; ok { return follow }
//     follow := make(map[Terminal]struct{})
//     p.follow[symbol] = follow
//     if symbol == p.start { follow[END] = struct{}{} }
//     for lhs, rules := range p.rules {
//         for _, rule := range rules { p.ruleFollow(symbol, follow, lhs, rule) }
//     }
//     return follow
// }

// func (p *LL1Parser) ruleFollow(symbol NonTerminal, follow map[Terminal]struct{}, lhs NonTerminal, rule []Symbol) {
//     main: for i, s := range rule {
//         if s == symbol {
//             for _, s := range rule[i + 1:] {
//                 f := p.findFirst(s)
//                 for t := range f {
//                     if t != EPSILON { follow[t] = struct{}{} }
//                 }
//                 if _, ok := f[EPSILON]; !ok { continue main }
//             }
//             for t := range p.findFollow(lhs) { follow[t] = struct{}{} }
//         }
//     }
// }

// func (t NonTerminal) isSymbol() { }
// func (t Terminal) isSymbol() { }

// var nonTerminalName = map[NonTerminal]string { E: "E", R: "R", T: "T", Y: "Y", F: "F" }
// func (t NonTerminal) String() string { return nonTerminalName[t] }
// func (t Terminal) String() string { return string(t) }
