package lynn

import (
	"bufio"
	"fmt"
	"strings"
)

type Token struct {
    Type, Value string
}

func unexpected(char string) {
    switch char {
    case " ": char = "space"
    case "\n": char = "new line"
    case "\t": char = "tab"
    case "\x00": char = "EOF"
    default: char = fmt.Sprintf("character \"%s\"", char)
    }
    fmt.Println("Unexpected", char)
}

func Lex(reader *bufio.Reader) []Token {
    var current string
    next := func () string {
        c, _, err := reader.ReadRune()
        if err != nil {
            current = "\x00"
        } else {
            current = string(c)
        }

        return current
    }

    tokens := make([]Token, 0, 100)
    next()
    main: for {
        current: switch current {
        case "\x00":
            tokens = append(tokens, Token { "EOF", "" })
            break main
        case " ", "\t", "\n", "\r":
        case "/":
            switch next() {
            case "/":
                next()
                lineComment: for {
                    switch current {
                    case "\n", "\r", "\x00": break lineComment
                    default:
                        next()
                    }
                }
            case "*":
                next()
                blockComment: for {
                    switch current {
                    case "*":
                        next()
                        if current == "/" { break blockComment }
                    case "\x00":
                        unexpected(current)
                        break blockComment
                    default: next()
                    }
                }
            }

        case "+": tokens = append(tokens, Token { "PLUS", current })
        case "-":
            if c := next(); c != ">" {
                unexpected(c)
                break
            }
            tokens = append(tokens, Token { "ARROW", "->" })
        case "*": tokens = append(tokens, Token { "STAR", current })
        case "?": tokens = append(tokens, Token { "QUESTION", current })
        case ".": tokens = append(tokens, Token { "DOT", current })
        case "|": tokens = append(tokens, Token { "BAR", current })
        case ";": tokens = append(tokens, Token { "SEMI", current })
        case ":": tokens = append(tokens, Token { "COLON", current })
        case "(": tokens = append(tokens, Token { "L_PAREN", current })
        case ")": tokens = append(tokens, Token { "R_PAREN", current })

        case "\"":
            var builder strings.Builder
            builder.WriteString(current)
            builder.WriteString(next())
            stringLiteral: for {
                switch current {
                case "\"": break stringLiteral
                case "\\":
                    switch c := next(); c {
                    case "\n", "\r", "\x00":
                        unexpected(c)
                        break current
                    default: builder.WriteString(c + next())
                    }
                case "\n", "\r", "\x00":
                    unexpected(current)
                    break current
                default: builder.WriteString(next())
                }
            }
            tokens = append(tokens, Token { "STRING", builder.String() }) 
        case "[":
            var builder strings.Builder
            builder.WriteString(current)
            builder.WriteString(next())
            class: for {
                switch current {
                case "]": break class
                case "\\":
                    switch c := next(); c {
                    case "\n", "\r", "\x00":
                        unexpected(c)
                        break current
                    default: builder.WriteString(c + next())
                    }
                case "\n", "\r", "\x00":
                    unexpected(current)
                    break current
                default: builder.WriteString(next())
                }
            }
            tokens = append(tokens, Token { "CLASS", builder.String() })
        default:
            if (current >= "a" && current <= "z") || (current >= "A" && current <= "Z") || current == "_" {
                var builder strings.Builder
                builder.WriteString(current)
                next()
                for (current >= "a" && current <= "z") || (current >= "A" && current <= "Z") || current == "_" || (current >= "0" && current <= "9") {
                    builder.WriteString(current)
                    next()
                }

                switch id := builder.String(); id {
                case "token": tokens = append(tokens, Token { "TOKEN", id })
                case "fragment": tokens = append(tokens, Token { "FRAGMENT", id })
                case "skip": tokens = append(tokens, Token { "SKIP", id })
                default: tokens = append(tokens, Token { "IDENTIFIER", id })
                }
                continue main
            } else {
                unexpected(current)
            }	
        }
        next()
    }

    return tokens
}
