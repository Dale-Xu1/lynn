package main

import (
	"bufio"
	"fmt"
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
    // fmt.Println("== abstract syntax tree == ")
    // fmt.Println(ast)

    // generator := lynn.NewLexerGenerator()
    // nfa, ranges := generator.GenerateNFA(ast)
    // dfa := generator.NFAtoDFA(nfa, ranges)
    // lynn.CompileLexer("lynn", ast, ranges, dfa)

    grammar := lynn.NewGrammarGenerator().GenerateCFG(ast)
    fmt.Println("== grammar ==")
    grammar.PrintGrammar()
    table := lynn.NewLALRParserGenerator().Generate(grammar)
    fmt.Println("== parse table ==")
    table.PrintTable()

    // p := lynn.NewShiftReduceParser(table)
    // p.Parse()
}
