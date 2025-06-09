package main

import (
	"bufio"
	"flag"
	"fmt"
	"lynn/lynn"
	"os"
	"path/filepath"
	"strings"
)

func main() {
    // Configure CLI flags
    cmd := filepath.Base(os.Args[0])
    name := *flag.String("o", "", "Output package name (defaults to name of input file)")
    flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s <path> [flags]\n", cmd)
		fmt.Fprintln(os.Stderr, "Arguments:")
		fmt.Fprintln(os.Stderr, "  [path]\n\tThe path to the input file")
		flag.PrintDefaults()
    }
    flag.Parse()
    args := flag.Args()
    if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s <path> [flags]\n", cmd)
        return
    }
    path := args[0]
    if name == "" { name = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)) }
    f, e := os.Open(path)
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
    if err { Fail(); return }
    fmt.Println("1/8 - Generated parse tree")

    ast := lynn.NewParseTreeVisitor().VisitGrammar(tree).(*lynn.GrammarNode)
    if lynn.Panic() { Fail(); return }
    fmt.Println("2/8 - Created abstract syntax tree")
    fmt.Println(ast)

    // Generate lexer data and compile to program
    generator := lynn.NewLexerGenerator()
    nfa, ranges := generator.GenerateNFA(ast)
    if lynn.Panic() { Fail(); return }
    fmt.Println("3/8 - Generated non-deterministic finite automata")

    dfa := generator.NFAtoDFA(nfa, ranges)
    fmt.Println("4/8 - Generated deterministic finite automata")

    // Generate parser data and compile to program
    grammar, maps := lynn.NewGrammarGenerator().GenerateCFG(ast)
    if lynn.Panic() { Fail(); return }
    fmt.Println("5/8 - Generated context-free grammar")
    grammar.PrintGrammar()

    table := lynn.NewLALRParserGenerator().Generate(grammar)
    if lynn.Panic() { Fail(); return }
    fmt.Println("6/8 - Generated LALR(1) parse table")

    lynn.CompileLexerGo(name, dfa, ranges, ast)
    fmt.Println("7/8 - Compiled lexer program")
    lynn.CompileParserGo(name, table, maps, ast)
    fmt.Println("8/8 - Compiled parser program")
}

func Fail() { fmt.Fprintln(os.Stderr, "Error occurred: Failed to generate programs") }
