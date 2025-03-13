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
    lynn.Lex(r)
}
