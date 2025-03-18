package main

import (
	"bufio"
	"fmt"
	"lynn/lynn"
	"os"
)

func main() {
    f, _ := os.Open("lynn/spec/lynn.ln")
    defer f.Close()

    lexer := lynn.NewLexer(bufio.NewReader(f))
    parser := lynn.NewParser(lexer)

    ast := parser.Parse()
    fmt.Printf("%v", ast)
}
