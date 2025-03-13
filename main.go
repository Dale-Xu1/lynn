package main

import (
	"bufio"
	"fmt"
	"lynn/lynn"
	"os"
)

func main() {
    f, _ := os.Open("lynn.ln")
    defer f.Close()

    r := bufio.NewReader(f)
    lexer := lynn.NewLexer(r)
    for {
        token := lexer.Next()
        if token.Type == "EOF" { break }

        fmt.Printf("%-16s %s\n", token.Type, token.Value)
    }
}
