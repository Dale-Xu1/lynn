package lynn

import (
	"fmt"
	"os"
	"strings"
)

func CompileLexer(file string, grammar *GrammarNode, ranges []Range, dfa LDFA) {
    const LEXER_TEMPLATE string = "lynn/spec/lexer.template"
    // Read template information
    data, err := os.ReadFile(LEXER_TEMPLATE)
    if err != nil { panic(err) }
    template := string(data)
    // Format token type information
    tokens, typeName := make([]string, len(grammar.Tokens)), make([]string, len(grammar.Tokens))
    tokenIndices := make(map[string]int, len(grammar.Tokens))
    skip := make([]string, 0)
    for i, token := range grammar.Tokens {
        name := token.Identifier.Name
        tokens[i], tokenIndices[name] = name, i
        typeName[i] = fmt.Sprintf("%d: \"%s\"", i, name)
        if token.Skip {
            skip = append(skip, fmt.Sprintf("%d: {}", i))
        }
    }
    tokens[0] += " TokenType = iota"
    // Format range information
    rangeIndices := make(map[Range]int, len(ranges))
    rangeStrings := make([]string, len(ranges))
    for i, r := range ranges { rangeIndices[r] = i }
    for i, r := range ranges {
        rangeStrings[i] = fmt.Sprintf("{ %q, %q }", r.Min, r.Max)
    }
    // Format state information
    stateIndices := map[*LDFAState]int { dfa.Start: 0 }
    transitions := make([]string, len(dfa.States))
    for _, state := range dfa.States {
        if state != dfa.Start { stateIndices[state] = len(stateIndices) }
    }
    for _, state := range dfa.States {
        i, l := stateIndices[state], len(state.Transitions)
        if l == 0 {
            transitions[i] = "    { },"
            continue
        }
        // Format outgoing transitions for each state
        out := make([]string, 0, l)
        for r, state := range state.Transitions {
            out = append(out, fmt.Sprintf("%d: %d", rangeIndices[r], stateIndices[state]))
        }
        transitions[i] = fmt.Sprintf("    { %s },", strings.Join(out, ", "))
    }
    // Format accepting states
    accept := make([]string, 0, len(dfa.Accept))
    for state, token := range dfa.Accept {
        accept = append(accept, fmt.Sprintf("%d: %d", stateIndices[state], tokenIndices[token]))
    }
    // Replace sections with compiled DFA
    pairs := []string {
        "/*{0}*/", file,
        "/*{1}*/", strings.Join(tokens, "; "),
        "/*{2}*/", strings.Join(typeName, ", "),
        "/*{3}*/", strings.Join(skip, ", "),
        "/*{4}*/", strings.Join(rangeStrings, ", "),
        "/*{5}*/", strings.Join(transitions, "\n"),
        "/*{6}*/", strings.Join(accept, ", "),
    }
    result := strings.NewReplacer(pairs...).Replace(template)
    // Write modified template to lexer program file
    if err := os.MkdirAll("out", 0755); err != nil { panic(err) }
    f, err := os.Create("out/lexer.go")
    if err != nil { panic(err) }
    defer f.Close()
    f.WriteString(result)
}

func CompileParser(file string, grammar *GrammarNode, table LRParseTable) {
    const PARSER_TEMPLATE string = "lynn/spec/parser.template"
    // Read template information
    data, err := os.ReadFile(PARSER_TEMPLATE)
    if err != nil { panic(err) }
    template := string(data)
    // Get token indices
    tokenIndices := make(map[string]int, len(grammar.Tokens))
    for i, token := range grammar.Tokens {
        tokenIndices[token.Identifier.Name] = i
    }
    // Get non-terminal indices (skip augmented start non-terminal)
    l := len(table.Grammar.NonTerminals) - 1
    nonTerminalIndices := make(map[NonTerminal]int, l)
    for i, t := range table.Grammar.NonTerminals[:l] {
        nonTerminalIndices[t] = i
    }
    // Format production data
    // Remove last production, which is the augmented start production
    productions := make([]string, len(table.Grammar.Productions) - 1)
    for i, p := range table.Grammar.Productions[:len(productions)] {
        var s string
        switch p.Type {
        case NORMAL:    s = "NORMAL"
        case AUXILIARY: s = "AUXILIARY"
        case FLATTEN:   s = "FLATTEN"
        case REMOVED:   s = "REMOVED"
        }
        productions[i] = fmt.Sprintf("    { %s, %d, %d, \"%s\" },",
            s, nonTerminalIndices[p.Left], len(p.Right), p.Visitor)
    }
    // Format action table
    actionTable := make([]string, len(table.Action))
    for i, entries := range table.Action {
        l := len(entries)
        if l == 0 {
            actionTable[i] = "    { },"
            continue
        }
        // Format action entries for each state
        out := make([]string, 0, l)
        for t, entry := range entries {
            var s string
            switch entry.Type {
            case SHIFT:  s = "SHIFT"
            case REDUCE: s = "REDUCE"
            case ACCEPT: s = "ACCEPT"
            }
            out = append(out, fmt.Sprintf("%d: { %s, %d }", tokenIndices[string(t)], s, entry.Value))
        }
        actionTable[i] = fmt.Sprintf("    { %s },", strings.Join(out, ", "))
    }
    // Format goto table
    gotoTable := make([]string, len(table.Goto))
    for i, entries := range table.Goto {
        l := len(entries)
        if l == 0 {
            gotoTable[i] = "    { },"
            continue
        }
        // Format action entries for each state
        out := make([]string, 0, l)
        for t, state := range entries {
            out = append(out, fmt.Sprintf("%d: %d", nonTerminalIndices[t], state))
        }
        gotoTable[i] = fmt.Sprintf("    { %s },", strings.Join(out, ", "))
    }
    // Replace sections with compiled parse table
    pairs := []string {
        "/*{0}*/", file,
        "/*{1}*/", strings.Join(productions, "\n"),
        "/*{2}*/", strings.Join(actionTable, "\n"),
        "/*{3}*/", strings.Join(gotoTable, "\n"),
    }
    result := strings.NewReplacer(pairs...).Replace(template)
    // Write modified template to lexer program file
    if err := os.MkdirAll("out", 0755); err != nil { panic(err) }
    f, err := os.Create("out/parser.go")
    if err != nil { panic(err) }
    defer f.Close()
    f.WriteString(result)
}



// TODO: Remove Lexer.Token, Lexer.Match(), and initial call to Lexer.Next()
// TODO: Remove this after bootstrapping

// Production type enum. Either NORMAL, AUXILIARY, FLATTEN, OR REMOVED.
type ProductionType uint
const (NORMAL ProductionType = iota; AUXILIARY; FLATTEN; REMOVED)
// Production data struct. Expresses a sequence of symbols that a given non-terminal may be expanded to in a grammar.
type ProductionData struct {
    Type         ProductionType
    Left, Length int
    Visitor      string
}

// Action type enum. Either SHIFT, REDUCE, or ACCEPT.
type ActionType uint
const (SHIFT ActionType = iota; REDUCE; ACCEPT)
// Parse table action entry struct. Holds action type and integer parameter.
type ActionEntry struct {
    Type  ActionType
    Value int // For SHIFT actions, value represents a state identifier, for REDUCE actions, a production identifier
}
