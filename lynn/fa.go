package lynn

type FA struct {
	In, Out *State
}

type State struct {
	Transitions map[rune]*State
}

func GenerateNFA(grammar *GrammarNode) FA {
	return FA { nil, nil }
}

// func expressionFNA(expression AST) FA {
// 	switch node := expression.(type) {

// 	}
// }
