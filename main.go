package main

import (
	"bufio"
	"fmt"
	"lynn/parser"
	"strings"
)

func main() {
    // f, _ := os.Open("lynn/spec/lynn.ln")
    // defer f.Close()

    // lexer := lynn.NewLexer(bufio.NewReader(f))
    // parser := lynn.NewParser(lexer)
    // generator := lynn.NewGenerator()

    // ast := parser.Parse()
    // nfa, ranges := generator.GenerateNFA(ast)
    // dfa := generator.NFAtoDFA(nfa, ranges)
    // dfa.PrintTransitions()

    lexer := parser.NewLexer(bufio.NewReader(strings.NewReader("abcda")))
    fmt.Println(lexer.Token)
    lexer.Next()
    fmt.Println(lexer.Token)
    lexer.Next()
    fmt.Println(lexer.Token)
}
