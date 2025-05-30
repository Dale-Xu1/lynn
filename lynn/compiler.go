package lynn

import (
	"fmt"
	"os"
	"strings"
)

// TODO: Remove Lexer.Token and Lexer.Match()

func CompileLexer(file string, grammar *GrammarNode, ranges []Range, dfa LDFA) {
    const LEXER_TEMPLATE string = "lynn/spec/lexer.template"
    // Read template information
    data, err := os.ReadFile(LEXER_TEMPLATE)
    if err != nil { panic(err) }
    template := string(data)
    // Format token type information
    tokens, typeName := make([]string, len(grammar.Tokens)), make([]string, len(grammar.Tokens))
    skip := make([]string, 0)
    for i, token := range grammar.Tokens {
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

func CompileParser(file string) {
    const PARSER_TEMPLATE string = "lynn/spec/parser.template"
    // Read template information
    data, err := os.ReadFile(PARSER_TEMPLATE)
    if err != nil { panic(err) }
    template := string(data)
    // TODO: Compile to generated parser program

    result := template
    // Write modified template to lexer program file
    if err := os.MkdirAll("out", 0755); err != nil { panic(err) }
    f, err := os.Create("out/lexer.go")
    if err != nil { panic(err) }
    defer f.Close()
    f.WriteString(result)
}

// ------------------------------------------------------------------------------------------------------------------------------

// FOR DEBUG PURPOSES:
// Consumes all tokens emitted by lexer and prints them to the standard output.
func (l *Lexer) PrintTokenStream() {
    for {
        location := fmt.Sprintf("%d:%d-%d:%d", l.Token.Start.Line, l.Token.Start.Col, l.Token.End.Line, l.Token.End.Col)
        fmt.Printf("%-16s | %-16s %-16s\n", location, l.Token.Type, l.Token.Value)
        if l.Token.Type == EOF { break }
        l.Next()
    }
}


type ShiftReduceParser struct {
    table LRParseTable
    // Productions []Production
    // Action      []map[Token]ActionEntry
    // Goto        []map[NonTerminal]int
}

type StackState struct {
    State int
    Node  ParseTreeChild
}

type ParseTreeChild interface { String(indent string) string }
type ParseTreeNode struct {
    Production *Production
    Children   []ParseTreeChild
}

func NewShiftReduceParser(table LRParseTable) *ShiftReduceParser { return &ShiftReduceParser { table } }
func (p *ShiftReduceParser) Parse() *ParseTreeNode {
    input := []Terminal { "TOKEN", "IDENTIFIER", "COLON", "STRING", "CLASS", "HASH", "IDENTIFIER", "LEFT", "BAR", "IDENTIFIER", "PLUS", "SEMI", "RULE", "IDENTIFIER", "COLON", "L_PAREN", "IDENTIFIER", "R_PAREN", "QUESTION", "SEMI", EOF_TERMINAL }
    ip := 0
    stack := []StackState { { 0, nil } }
    for {
        state, token := stack[len(stack) - 1].State, input[ip]
        action, ok := p.table.Action[state][token]
        if !ok {
            fmt.Printf("Syntax error: Unexpected token %s\n", token)
            return nil
        }
        switch action.Type {
        case SHIFT:
            // Add new state to the stack along with token
            stack = append(stack, StackState { action.Value, Token { EOF, token.String(), Location{}, Location{} } })
            ip++
        case REDUCE:
            // Find production to reduce by
            production := p.table.Grammar.Productions[action.Value]
            var node ParseTreeChild
            switch production.Type {
            case NORMAL:
                r := len(production.Right); l := len(stack) - r
                // Collect child nodes from current states on the stack and create node for reduction
                children := make([]ParseTreeChild, r)
                for i, s := range stack[l:] { children[i] = s.Node }
                node = &ParseTreeNode { production, children }
                stack = stack[:l]
            case FLATTEN:
                l := len(stack) - 2
                list := stack[l].Node.(*ParseTreeNode); element := stack[l + 1].Node
                list.Children = append(list.Children, element)
                node = list
                stack = stack[:l]
            case AUXILIARY:
                l := len(stack) - 1
                node = stack[l].Node; stack = stack[:l]
            case REMOVED: node = nil
            }
            // Find next state based on goto table
            state := stack[len(stack) - 1].State
            next := p.table.Goto[state][production.Left]
            stack = append(stack, StackState { next, node })
        case ACCEPT: return stack[1].Node.(*ParseTreeNode)
        }
    }
}

func (t Token) String(indent string) string { return fmt.Sprintf("%s<%s %s>", indent, t.Type, t.Value) }
func (n *ParseTreeNode) String(indent string) string {
    children := make([]string, len(n.Children))
    next := indent + "  "
    for i, c := range n.Children {
        str := "\n"
        if c == nil {
            str += fmt.Sprintf("%s<nil>", next)
        } else {
            str += c.String(next)
        }
        children[i] = str
    }
    return fmt.Sprintf("%s[%s]%s", indent, n.Production, strings.Join(children, ""))
}
