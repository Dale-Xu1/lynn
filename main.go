package main

import (
	"bufio"
	"lynn/lynn"
	"os"
)

func main() {
    f, err := os.Open("lynn/spec/lynn.ln")
    if err != nil { panic(err) }
    defer f.Close()

    lexer := lynn.NewLexer(bufio.NewReader(f), lynn.DEFAULT_HANDLER)
    parser := lynn.NewParser(lexer)
    ast := parser.Parse()
    _ = ast

    // generator := lynn.NewGenerator()
    // nfa, ranges := generator.GenerateNFA(ast)
    // dfa := generator.NFAtoDFA(nfa, ranges)
    // lynn.CompileLexer("lynn", ast, ranges, dfa)
}
