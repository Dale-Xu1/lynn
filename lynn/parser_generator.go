package lynn

import (
	"fmt"
	"sort"
	"strings"
	"unsafe"
)

// LR(0) item struct. Specifies how much of a production's sequence of symbols have been read.
type LR0Item struct { Production *Production; Dot int }
// LR(1) item struct. Specifies how much of a production's sequence of symbols have been read and a lookahead terminal.
type LR1Item struct { Production *Production; Dot int; Lookahead Terminal }
// LR(1) state, determined by set of LR(1) items. Transitions to other states on symbols.
type LRState struct {
    Items       map[LR1Item]struct{}
    Transitions map[Symbol]*LRState
}

// LR(1) parse table. Represents action table and goto table.
type LRParseTable struct {
    Grammar *Grammar
    Action  []map[Terminal]ActionEntry
    Goto    []map[NonTerminal]int
}

// Action type enum. Either SHIFT, REDUCE, or ACCEPT.
type ActionType uint
const (SHIFT ActionType = iota; REDUCE; ACCEPT)
// Parse table action entry struct. Holds action type and integer parameter.
type ActionEntry struct {
    Type  ActionType
    Value int // For SHIFT actions, value represents a state identifier, for REDUCE actions, a production identifier
}

// LALR parser generator struct. Converts a given grammar to an LR(1) parse table.
type LALRParserGenerator struct {
    grammar   *Grammar
    augmented *Production
    first     map[Symbol]map[Terminal]struct{}
}

// Returns a new LALR parser generator struct.
func NewLALRParserGenerator() *LALRParserGenerator { return &LALRParserGenerator { } }
// Converts a grammar definition to an LR(1) parse table.
func (g *LALRParserGenerator) Generate(grammar *Grammar) LRParseTable {
    // Initialize generator with augmented grammar
    g.grammar = grammar
    g.augmented = grammar.Augment()
    // Find first set and construct LALR(1) states from LR(1) states
    g.first = make(map[Symbol]map[Terminal]struct{})
    g.findFirst()
    states := buildLALRStates(g.buildLR1States())
    // Generate parse table and pass to shift-reduce parser
    table := g.buildParseTable(states)
    return table
}

// Computes the FIRST sets of all symbols in the grammar provided by the parser.
func (g *LALRParserGenerator) findFirst() {
    // Initialize FIRST sets for all terminals and non-terminals
    // FIRST set of a terminal contains only itself, and FIRST sets of non-terminals are initialized empty
    for _, t := range g.grammar.Terminals {
        if t != EOF_TERMINAL { g.first[t] = map[Terminal]struct{} { t: {} } }
    }
    for _, t := range g.grammar.NonTerminals { g.first[t] = make(map[Terminal]struct{}) }
    // Iterative implementation, repeat procedure until no changes are made
    for changed := true; changed; {
        changed = false
        for _, production := range g.grammar.Productions {
            changed = g.findRuleFirst(production.Left, production.Right) || changed
        }
    }
}

func (g *LALRParserGenerator) findRuleFirst(left NonTerminal, rule []Symbol) bool {
    // The FIRST set of a non-terminal is found by going through its productions and taking the union of the FIRST sets
    // of subsequent symbols until a non-nullable symbol is found (FIRST set does not contain epsilon)
    changed := false
    for _, s := range rule {
        for f := range g.first[s] {
            // Add all elements in FIRST set of symbol to FIRST set of production LHS non-terminal
            if _, ok := g.first[left][f]; !ok { g.first[left][f] = struct{}{}; changed = true }
        }
        // All elements in FIRST set have been found if symbol is non-nullable
        if _, ok := g.first[s][EPSILON]; !ok { return changed }
    }
    // If all symbols in the production are nullable, the LHS non-terminal is also nullable
    if _, ok := g.first[left][EPSILON]; !ok { g.first[left][EPSILON] = struct{}{}; changed = true }
    return changed
}

func (g *LALRParserGenerator) findSequenceFirst(sequence []Symbol) map[Terminal]struct{} {
    first := make(map[Terminal]struct{})
    for _, s := range sequence {
        // Add all elements in FIRST set of symbol to set
        for f := range g.first[s] { first[f] = struct{}{} }
        // All elements in FIRST set have been found if symbol is non-nullable
        if _, ok := g.first[s][EPSILON]; !ok { return first }
    }
    // If all symbols in the production are nullable, the sequence is also nullable
    first[EPSILON] = struct{}{}
    return first
}

// Computes the closure set of a set of LR(1) items.
func (g *LALRParserGenerator) findClosure(items map[LR1Item]struct{}) map[LR1Item]struct{} {
    // Transfer items in set to closure set and work list
    closure := make(map[LR1Item]struct{}, len(items))
    work := make([]LR1Item, 0, len(items))
    for item := range items { closure[item] = struct{}{}; work = append(work, item) }
    // Iterative implementation, process items in work list until no new items are added
    for len(work) > 0 {
        item := work[0]; work = work[1:]
        // Ensure that LR(1) item still has symbols to shift and that the current symbol is a non-terminal
        right := item.Production.Right
        if item.Dot >= len(right) { continue }
        t, ok := right[item.Dot].(NonTerminal); if !ok { continue }
        // Find first first set of beta sequence, if sequence is nullable, add lookahead to set
        first := g.findSequenceFirst(right[item.Dot + 1:])
        if _, ok := first[EPSILON]; ok { delete(first, EPSILON); first[item.Lookahead] = struct{}{} }
        // Look for all productions with the given non terminal as its LHS
        for i := range g.grammar.Productions {
            production := g.grammar.Productions[i]
            if production.Left != t { continue }
            // Add productions to the work list, starting dot at the start and create copies for each terminal in the first set
            for lookahead := range first {
                i := LR1Item { production, 0, lookahead }
                if _, ok := closure[i]; !ok { closure[i] = struct{}{}; work = append(work, i) }
            }
        }
    }
    return closure
}

// Computes the GOTO set of a given item and symbol to transition on.
func (g *LALRParserGenerator) findGoto(items map[LR1Item]struct{}, symbol Symbol) map[LR1Item]struct{} {
    set := make(map[LR1Item]struct{})
    for item := range items {
        // Find LR(1) items in set where the next symbol matches the one passed to the function
        right := item.Production.Right
        if item.Dot < len(right) && right[item.Dot] == symbol {
            i := LR1Item { item.Production, item.Dot + 1, item.Lookahead }
            set[i] = struct{}{}
        }
    }
    if len(set) == 0 { return set }
    return g.findClosure(set)
}

// Constructs the canonical collection of LR(1) item sets.
// The first state in the list is the start state (LR(1) item set that contains augmented start production).
func (g *LALRParserGenerator) buildLR1States() []*LRState {
    // Find closure of initial item
    closure := g.findClosure(map[LR1Item]struct{} { { g.augmented, 0, EOF_TERMINAL }: {} })
    // Initialize work list with state associated with set
    start := &LRState { closure, make(map[Symbol]*LRState) }
    states := map[string]*LRState { getLR1ItemStateKey(start.Items): start } // Register state key to state object
    // Iterative implementation, process items in list list until no new items are added
    list := []*LRState { start }
    for index := 0; index < len(list); {
        state := list[index]; index++
        // Find all outgoing transition symbols from current state
        // Determined by symbols after the dot in the LR(1) item's production
        transitions := make(map[Symbol]struct{})
        for item := range state.Items {
            right := item.Production.Right
            if item.Dot < len(right) {
                symbol := right[item.Dot]
                transitions[symbol] = struct{}{}
            }
        }
        // For each outgoing transition, compute the next state determined by the GOTO set
        for symbol := range transitions {
            set := g.findGoto(state.Items, symbol); if len(set) == 0 { continue }
            // Determine if the state has already been visited
            key := getLR1ItemStateKey(set); next := states[key]
            if next == nil {
                // Create a new state and add to the work list if it has not been seen before
                next = &LRState{ set, make(map[Symbol]*LRState) }
                states[key] = next
                list = append(list, next)
            }
            // Add transition to the new state to the current state
            state.Transitions[symbol] = next
        }
    }
    return list
}

// Merges LR(1) states with identical LR(0) cores to create LALR(1) states.
func buildLALRStates(states []*LRState) []*LRState {
    // Create replacement map according to partition LR(1) of states based on their LR(0) cores
    representatives, merge := make(map[string]*LRState), make(map[*LRState]*LRState)
    for _, state := range states {
        // Compute LR(0) core
        core := make(map[LR0Item]struct{})
        for item := range state.Items { core[LR0Item { item.Production, item.Dot }] = struct{}{} }
        // If the core has not been seen before, assign the current state as the subset's representative
        // Otherwise map current state to its representative
        key := getLR0ItemStateKey(core)
        if representative := representatives[key]; representative != nil {
            merge[state] = representative
        } else {
            representatives[key] = state
        }
    }
    // Merge states according to replacement map
    merged := make([]*LRState, 0)
    for _, state := range states {
        if representative := merge[state]; representative != nil { continue }
        // Representatives are added straight to merged state list
        merged = append(merged, state)
        for symbol, next := range state.Transitions {
            // For all outgoing transitions of a state, replace the next state with its representative
            if r := merge[next]; r != nil { state.Transitions[symbol] = r }
        }
    }
    for _, state := range states {
        representative := merge[state]; if representative == nil { continue }
        // If state is not a representative, its data is merged into its corresponding representative
        for item := range state.Items { representative.Items[item] = struct{}{} }
        for symbol, next := range state.Transitions {
            // States cannot be successfully merged if multiple transitions on the same symbol exist that go to different states
            if r := merge[next]; r != nil { next = r }
            representative.Transitions[symbol] = next
        }
    }
    return merged
}

// Construct LALR(1) parse table.
func (g *LALRParserGenerator) buildParseTable(states []*LRState) LRParseTable {
    // Create map from state and production structs to their respective integer identifiers
    stateId := make(map[*LRState]int)
    productionId := make(map[*Production]int)
    for i, state := range states { stateId[state] = i }
    for i, p := range g.grammar.Productions { productionId[p] = i }
    // Initialize action and goto tables in parse table
    table := LRParseTable {
        g.grammar,
        make([]map[Terminal]ActionEntry, len(states)),
        make([]map[NonTerminal]int, len(states)),
    }
    for i, state := range states {
        action, jump := make(map[Terminal]ActionEntry), make(map[NonTerminal]int)
        table.Action[i], table.Goto[i] = action, jump
        // For all transitions of the current state, generate the corresponding shift and goto actions
        for symbol, next := range state.Transitions {
            id := stateId[next]
            // Shift if the transition is on a terminal, goto if it is on a non-terminal
            switch t := symbol.(type) {
            case Terminal:    action[t] = ActionEntry { SHIFT, stateId[next] }
            case NonTerminal: jump[t] = id
            }
        }
        for item := range state.Items {
            // Identify all LR(1) items of the state where all symbols have been consumed
            if item.Dot < len(item.Production.Right) { continue }
            if item.Production == g.augmented {
                // Register an accept action if the production being reduced is the augmented start non-terminal
                action[EOF_TERMINAL] = ActionEntry{ Type: ACCEPT }
            } else {
                id := productionId[item.Production]
                if existing, ok := action[item.Lookahead]; ok {
                    switch existing.Type {
                    case SHIFT:
                        // Reduce action is ignored, preferring shift action if it already exists
                        fmt.Printf("Generation error: Shift/reduce conflict on token %s for production %v\n",
                            item.Lookahead, item.Production)
                        continue
                    case REDUCE:
                        // Resolve reduce/reduce conflict by choosing reduce action with lower production identifier
                        fmt.Printf("Generation error: Reduce/reduce conflict on token %s between productions %v and %v\n",
                            item.Lookahead, item.Production, g.grammar.Productions[existing.Value])
                        if id > existing.Value { continue }
                    }
                }
                action[item.Lookahead] = ActionEntry { REDUCE, id }
            }
        }
    }
    return table
}

// ------------------------------------------------------------------------------------------------------------------------------

// Creates unique identifier string given a set of LR(0) items for use in a map.
func getLR0ItemStateKey(items map[LR0Item]struct{}) string {
    // Sort states by address to ensure identical sets map to the same key
    list := make([]LR0Item, 0, len(items))
    for item := range items { list = append(list, item) }
    sort.Slice(list, func (i, j int) bool {
        a, b := list[i], list[j]
        if a.Production != b.Production { return uintptr(unsafe.Pointer(a.Production)) < uintptr(unsafe.Pointer(b.Production)) }
        return a.Dot < b.Dot
    })
    // Interpret state memory addresses as consecutive bytes, then reinterpret as string
    const ITEM_SIZE int = int(unsafe.Sizeof(LR0Item { }))
    bytes := make([]byte, 0, len(list) * ITEM_SIZE)
    for _, item := range list {
        b := (*[ITEM_SIZE]byte)(unsafe.Pointer(&item))
        bytes = append(bytes, b[:]...)
    }
    return string(bytes)
}

// Creates unique identifier string given a set of LR(1) items for use in a map.
func getLR1ItemStateKey(items map[LR1Item]struct{}) string {
    // Sort states by address to ensure identical sets map to the same key
    list := make([]LR1Item, 0, len(items))
    for item := range items { list = append(list, item) }
    sort.Slice(list, func (i, j int) bool {
        a, b := list[i], list[j]
        if a.Production != b.Production { return uintptr(unsafe.Pointer(a.Production)) < uintptr(unsafe.Pointer(b.Production)) }
        if a.Dot != b.Dot { return a.Dot < b.Dot }
        return a.Lookahead < b.Lookahead
    })
    // Interpret state memory addresses as consecutive bytes, then reinterpret as string
    const ITEM_SIZE int = int(unsafe.Sizeof(LR1Item { }))
    bytes := make([]byte, 0, len(list) * ITEM_SIZE)
    for _, item := range list {
        b := (*[ITEM_SIZE]byte)(unsafe.Pointer(&item))
        bytes = append(bytes, b[:]...)
    }
    return string(bytes)
}

// FOR DEBUG PURPOSES:
// Prints all LR(1) states in the graph and transitions between states.
func PrintStates(states []*LRState) {
    ids := make(map[*LRState]int)
    for i, state := range states { ids[state] = i }
    for i, state := range states {
        fmt.Printf("%d ", i)
        for item := range state.Items {
            right := item.Production.Right
            str := make([]string, len(right)); for i, s := range right { str[i] = s.String() }
            fmt.Printf("[%s -> %s.%s, %s] ", item.Production.Left,
                strings.Join(str[:item.Dot], " "), strings.Join(str[item.Dot:], " "), item.Lookahead)
        }
        fmt.Println()
        for s, next := range state.Transitions {
            fmt.Printf("    %s -> %d\n", s, ids[next])
        }
    }
}

// FOR DEBUG PURPOSES:
// Prints formatted parse table.
func (t LRParseTable) PrintTable() {
    fmt.Print("state  |")
    for _, t := range t.Grammar.Terminals { fmt.Printf(" %-6.6s |", t) }; fmt.Print(" |")
    for _, t := range t.Grammar.NonTerminals { fmt.Printf(" %-6.6s |", t) }; fmt.Println()
    l := 8 + len(t.Grammar.Terminals) * 9 + 2 + len(t.Grammar.NonTerminals) * 9
    fmt.Println(strings.Repeat("-", l))
    for i := range t.Action {
        fmt.Printf("%-6d |", i)
        action, jump := t.Action[i], t.Goto[i]
        for _, t := range t.Grammar.Terminals {
            a, ok := action[t]; if !ok { fmt.Print("        |"); continue }
            var str string
            switch a.Type {
            case SHIFT:  str = fmt.Sprintf("s%d", a.Value)
            case REDUCE: str = fmt.Sprintf("r%d", a.Value)
            case ACCEPT: str = "acc"
            }
            fmt.Printf(" %-6s |", str)
        }
        fmt.Print(" |")
        for _, t := range t.Grammar.NonTerminals {
            g, ok := jump[t]
            if ok {
                fmt.Printf(" %-6d |", g)
            } else {
                fmt.Print("        |")
            }
        }
        fmt.Println()
    }
}
