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

    generator.GenerateNFA(parser.Parse())
}
