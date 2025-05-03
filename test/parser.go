package test

import "fmt"

type ShiftReduceParser struct {
    productions []Production
	table       LRParseTable
	stack       []int
}

func NewShiftReduceParser(productions []Production, table LRParseTable) *ShiftReduceParser {
	return &ShiftReduceParser { productions, table, nil }
}

func (p *ShiftReduceParser) Parse() {
	input := []Terminal { 'x', '+', 'x', '*', 'x', END }
	p.stack = []int { 0 }
	ip := 0
	main: for {
		state := p.stack[len(p.stack) - 1]
		action, ok := p.table.Action[state][input[ip]]
		if !ok { panic("Unexpected symbol") }
		switch action.Type {
		case SHIFT:
			p.stack = append(p.stack, action.Value)
			ip++
			fmt.Printf("s%d\n", action.Value)
		case REDUCE:
			production := p.productions[action.Value]
			l := len(p.stack) - len(production.Right)
			p.stack = p.stack[:l]
			p.stack = append(p.stack, p.table.Goto[p.stack[l - 1]][production.Left])
			fmt.Printf("r%d\n", action.Value)
		case ACCEPT:
			fmt.Println("acc")
			break main
		}
	}
}
