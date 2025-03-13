package lynn

type Parser struct {
	lexer *Lexer
}

// Returns new parser struct
func NewParser(lexer *Lexer) *Parser {
	return &Parser { lexer }
}

func (p *Parser) Parse() {

}
