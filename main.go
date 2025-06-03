package main

import (
	"bufio"
	"fmt"
	"lynn/lynn"
	"os"
)

func main() {
    f, e := os.Open(os.Args[1])
    name := os.Args[2]
    if e != nil { panic(e) }
    defer f.Close()
    // Parse input grammar file and generate abstract syntax tree
    fmt.Println("== Parsing grammar definition file ==")

    err := false
    lexer := lynn.NewLexer(bufio.NewReader(f), func (stream *lynn.InputStream, char rune, location lynn.Location) {
        lynn.DEFAULT_LEXER_HANDLER(stream, char, location)
        err = true
    })
    tree := lynn.NewParser(lexer, func (token lynn.Token) {
        lynn.DEFAULT_PARSER_HANDLER(token)
        err = true
    }).Parse()
    if err { return }
    fmt.Println("1/8 - Generated parse tree")

    ast := lynn.NewParseTreeVisitor().VisitGrammar(tree).(*lynn.GrammarNode)
    if len(ast.Rules) == 0 || len(ast.Tokens) == 0 {
        fmt.Println("Generation error: Grammar definition must contain at least one rule and token")
        return
    }
    fmt.Println("2/8 - Created abstract syntax tree")

    // Generate lexer data and compile to program
    fmt.Println("== Compiling lexer program file ==")

    generator := lynn.NewLexerGenerator()
    nfa, ranges := generator.GenerateNFA(ast)
    fmt.Println("3/8 - Generated non-deterministic finite automata")

    dfa := generator.NFAtoDFA(nfa, ranges)
    fmt.Println("4/8 - Generated deterministic finite automata")

    lynn.CompileLexer(name, dfa, ranges, ast)
    fmt.Println("5/8 - Compiled lexer program")

    // Generate parser data and compile to program
    fmt.Println("== Compiling parser program file ==")

    grammar, maps := lynn.NewGrammarGenerator().GenerateCFG(ast)
    fmt.Println("6/8 - Generated context-free grammar")

    table := lynn.NewLALRParserGenerator().Generate(grammar)
    fmt.Println("7/8 - Generated LALR(1) parse table")

    lynn.CompileParser(name, table, maps, ast)
    fmt.Println("8/8 - Compiled parser program")
}
