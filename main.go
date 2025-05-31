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

    // lexer := lynn.NewLexer(bufio.NewReader(f), lynn.DEFAULT_LEXER_HANDLER)
    // tree := lynn.NewParser(lexer, lynn.DEFAULT_PARSER_HANDLER).Parse()
    // tree.Print()
    // // ast := lynn.NewParseTreeVisitor().VisitGrammar(tree)
    // // fmt.Println(ast)

    lexer := lynn.NewLexer(bufio.NewReader(f), lynn.DEFAULT_LEXER_HANDLER)
    parser := lynn.NewLegacyParser(lexer)
    ast := parser.Parse()

    // generator := lynn.NewLexerGenerator()
    // nfa, ranges := generator.GenerateNFA(ast)
    // dfa := generator.NFAtoDFA(nfa, ranges)
    // lynn.CompileLexer("lynn", ast, ranges, dfa)

    grammar := lynn.NewGrammarGenerator().GenerateCFG(ast)
    grammar.PrintGrammar()
    // table := lynn.NewLALRParserGenerator().Generate(grammar)
    // lynn.CompileParser("lynn", ast, table)
}
