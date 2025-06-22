package main

import (
	"flag"
	"fmt"
	"lynn/lynn"
	"lynn/lynn/parser"
	"os"
	"path/filepath"
)

func main() {
    // Configure CLI flags
    cmd := filepath.Base(os.Args[0])
    var name, lang string; var log bool
    flag.StringVar(&name, "o", "parser", "Output Go package name")
    flag.StringVar(&lang, "l", "go", "Output program language (\"go\" or \"ts\")")
    flag.BoolVar(&log, "a", false, "Log syntax tree and augmented grammar")
    flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] <path>\n", cmd)
		fmt.Fprintln(os.Stderr, "Arguments:")
		fmt.Fprintln(os.Stderr, "  <path>\n    \tThe path to the input file")
		flag.PrintDefaults()
    }
    flag.Parse()
    args := flag.Args()
    if len(args) != 1 { flag.Usage(); return }
    path := args[0]
    f, e := os.Open(path)
    if e != nil { panic(e) }
    defer f.Close()

    // Parse input grammar file and generate abstract syntax tree
    fmt.Println("== Parsing grammar definition file... ==")
    err := false
    lexer := parser.NewLexer(f, func (stream *parser.InputStream, char rune, location parser.Location) {
        parser.DEFAULT_LEXER_HANDLER(stream, char, location)
        err = true
    })
    tree := parser.NewParser(lexer, func (token parser.Token) {
        parser.DEFAULT_PARSER_HANDLER(token)
        err = true
    }).Parse()
    if err { Fail(); return }
    fmt.Println("1/8 - Generated parse tree")

    ast := lynn.NewParseTreeVisitor().VisitGrammar(tree).(*lynn.GrammarNode)
    if lynn.Panic() { Fail(); return }
    fmt.Println("2/8 - Created abstract syntax tree")
    if log { fmt.Println(ast) }

    // Generate lexer data and compile to program
    fmt.Println("== Generating lexer data... ==")
    generator := lynn.NewLexerGenerator()
    nfa, ranges := generator.GenerateNFA(ast)
    if lynn.Panic() { Fail(); return }
    fmt.Println("3/8 - Generated non-deterministic finite automata")

    dfa := generator.NFAtoDFA(nfa, ranges)
    fmt.Println("4/8 - Generated deterministic finite automata")

    // Generate parser data and compile to program
    fmt.Println("== Generating parser data... ==")
    grammar, maps := lynn.NewGrammarGenerator().GenerateCFG(ast)
    if lynn.Panic() { Fail(); return }
    fmt.Println("5/8 - Generated context-free grammar")
    if log { grammar.PrintGrammar() }

    table := lynn.NewLALRParserGenerator().Generate(grammar)
    if lynn.Panic() { Fail(); return }
    fmt.Println("6/8 - Generated LALR(1) parse table")

    fmt.Println("== Compiling generated programs... ==")
    switch lang {
    case "go":
        lynn.CompileLexerGo(name, dfa, ranges, ast)
        fmt.Println("7/8 - Compiled lexer program")
        lynn.CompileParserGo(name, table, maps, ast)
        fmt.Println("8/8 - Compiled parser program")
    case "ts":
        lynn.CompileLexerTS(dfa, ranges, ast)
        fmt.Println("7/8 - Compiled lexer program")
        lynn.CompileParserTS(table, maps, ast)
        fmt.Println("8/8 - Compiled parser program")
    default: Fail()
    }
}

func Fail() { fmt.Fprintln(os.Stderr, "Error occurred: Failed to generate programs") }
