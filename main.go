package main

import (
	"bufio"
	"lynn/lynn"
	"os"
)

func main() {
    f, _ := os.Open("lynn.ln")
    defer f.Close()

    r := bufio.NewReader(f)
    tokens := lynn.Lex(r)
    _ = tokens
    // for _, token := range tokens {
    //     fmt.Printf("%-16s %s\n", token.Type, token.Value)
    // }
}
