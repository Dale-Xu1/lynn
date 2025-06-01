package main

import (
	"bufio"
	"lynn/lynn"
	"os"
)

func main() {
    f, err := os.Open("lynn/spec/lynn.ln")
    if err != nil { panic(err) }
    defer f.Close()

    lexer := lynn.NewLexer(bufio.NewReader(f), lynn.DEFAULT_LEXER_HANDLER)
    tree := lynn.NewParser(lexer, lynn.DEFAULT_PARSER_HANDLER).Parse()
    ast := lynn.NewParseTreeVisitor().VisitGrammar(tree).(*lynn.GrammarNode)

    generator := lynn.NewLexerGenerator()
    nfa, ranges := generator.GenerateNFA(ast)
    dfa := generator.NFAtoDFA(nfa, ranges)
    lynn.CompileLexer("lynn", dfa, ranges, ast)

    grammar, maps := lynn.NewGrammarGenerator().GenerateCFG(ast)
    table := lynn.NewLALRParserGenerator().Generate(grammar)
    lynn.CompileParser("lynn", table, maps, ast)
}
