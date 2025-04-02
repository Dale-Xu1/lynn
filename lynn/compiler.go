package lynn

import (
	"fmt"
	"os"
	"strings"
)

func CompileLexer(file string, ast *GrammarNode, ranges []Range, dfa LDFA) {
	const LEXER_TEMPLATE string = "lynn/spec/lexer.template"
	// Read template information
	data, err := os.ReadFile(LEXER_TEMPLATE)
	if err != nil { panic(err) }
	template := string(data)

	// Format token type information
	tokens, typeName := make([]string, len(ast.Tokens)), make([]string, len(ast.Tokens))
	skip := make([]string, 0)
	for i, token := range ast.Tokens {
		name := token.Identifier.Name
		tokens[i] = name
		typeName[i] = fmt.Sprintf("%s: \"%s\"", name, name)
		if token.Skip {
			skip = append(skip, fmt.Sprintf("%s: {}", name))
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
		accept = append(accept, fmt.Sprintf("%d: %s", stateIndices[state], token))
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
