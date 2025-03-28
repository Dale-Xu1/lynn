package main

import (
	"bufio"
	"lynn/lynn"
	"os"
)

func main() {
    f, _ := os.Open("lynn/spec/lynn.ln")
    defer f.Close()

    lexer := lynn.NewLexer(bufio.NewReader(f))
    parser := lynn.NewParser(lexer)
    generator := lynn.NewGenerator()

    ast := parser.Parse()
    nfa, ranges := generator.GenerateNFA(ast)
    dfa := generator.NFAtoDFA(nfa, ranges)
    dfa.PrintTransitions()
}
