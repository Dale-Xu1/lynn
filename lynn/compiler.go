package lynn

import (
	"fmt"
	"os"
	"strings"
	"unicode"
)

// Compiles relevant lexer data to lexer program.
func CompileLexer(name string, dfa LDFA, ranges []Range, grammar *GrammarNode) {
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
        id := token.Identifier.Name
        tokens[i], tokenIndices[id] = id, i
        typeName[i] = fmt.Sprintf("%d: \"%s\"", i, id)
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
        "/*{0}*/", name,
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

// Compiles relevant parser data to parser program.
func CompileParser(name string, table LRParseTable, maps map[*Production]map[string]int, grammar *GrammarNode) {
    const PARSER_TEMPLATE string = "lynn/spec/parser.template"
    // Read template information
    data, err := os.ReadFile(PARSER_TEMPLATE)
    if err != nil { panic(err) }
    template := string(data)
    // Get token indices
    tokenIndices := make(map[string]int, len(grammar.Tokens))
    tokenIndices[ERROR_TERMINAL] = -1
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
    existingVisitors, existingAliases := make(map[string]struct{}), make(map[string]struct{})
    visitors, dispatchers := make([]string, 0), make([]string, 0)
    aliases := make([]string, 0)
    for i, p := range table.Grammar.Productions[:len(productions)] {
        var out string
        if m, ok := maps[p]; ok {
            // If alias map is present, compile to string
            entries := make([]string, 0, len(m))
            for id, i := range m {
                entries = append(entries, fmt.Sprintf("\"%s\": %d", id, i))
                // Track all unique aliases
                if _, ok := existingAliases[id]; ok { continue }
                existingAliases[id] = struct{}{}
                // Generate a method for the parse tree node for each alias
                n := []rune(id); n[0] = unicode.ToUpper(n[0]) // Capitalize first character
                alias := string(n)
                aliases = append(aliases,
                    fmt.Sprintf("func (n *ParseTreeNode) %s() ParseTreeChild { return n.GetAlias(\"%s\") }", alias, id))
            }
            out = fmt.Sprintf("map[string]int { %s }", strings.Join(entries, ", "))
        } else {
            out = "nil"
        }
        productions[i] = fmt.Sprintf("    { %d, %d, %d, \"%s\", %s },",
            p.Type, nonTerminalIndices[p.Left], len(p.Right), p.Visitor, out)
        // Test if new dispatcher needs to be generated
        if len(p.Visitor) == 0 { continue }
        if _, ok := existingVisitors[p.Visitor]; ok { continue }
        existingVisitors[p.Visitor] = struct{}{}
        // Add visitor entries and dispatcher lines
        n := []rune(p.Visitor); n[0] = unicode.ToUpper(n[0]) // Capitalize first character
        visitor := string(n)
        visitors = append(visitors, fmt.Sprintf("    Visit%s(node *ParseTreeNode) T", visitor))
        dispatchers = append(dispatchers, fmt.Sprintf("        case \"%s\": return visitor.Visit%s(n)", p.Visitor, visitor))
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
            out = append(out, fmt.Sprintf("%d: { %d, %d }", tokenIndices[string(t)], entry.Type, entry.Value))
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
        // Format goto entry for each state
        out := make([]string, 0, l)
        for t, state := range entries {
            out = append(out, fmt.Sprintf("%d: %d", nonTerminalIndices[t], state))
        }
        gotoTable[i] = fmt.Sprintf("    { %s },", strings.Join(out, ", "))
    }
    // Replace sections with compiled parse table
    pairs := []string {
        "/*{0}*/", name,
        "/*{1}*/", strings.Join(productions, "\n"),
        "/*{2}*/", strings.Join(actionTable, "\n"),
        "/*{3}*/", strings.Join(gotoTable, "\n"),
        "/*{4}*/", strings.Join(visitors, "\n"),
        "/*{5}*/", strings.Join(dispatchers, "\n"),
        "/*{6}*/", strings.Join(aliases, "\n"),
    }
    result := strings.NewReplacer(pairs...).Replace(template)
    // Write modified template to lexer program file
    if err := os.MkdirAll("out", 0755); err != nil { panic(err) }
    f, err := os.Create("out/parser.go")
    if err != nil { panic(err) }
    defer f.Close()
    f.WriteString(result)
}
