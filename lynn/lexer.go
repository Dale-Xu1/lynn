package lynn

import (
	"bufio"
	"fmt"
	"strings"
)

type Token struct {
    Type, Value string
}

type Lexer struct {
    reader *bufio.Reader
    temp   string
}

func NewLexer(reader *bufio.Reader) *Lexer {
    return &Lexer { reader, "" }
}

func (l *Lexer) char() string {
    if l.temp != "" {
        c := l.temp; l.temp = ""
        return c
    }

    c, _, err := l.reader.ReadRune()
    if err != nil { return "\x00" }
    return string(c)
}

func (l *Lexer) Next() Token {
    current := l.char()
    main: switch current {
    case "\x00": return Token { "EOF", "" }
    case " ", "\t", "\n", "\r":
        for {
            switch c := l.char(); c {
            case " ", "\t", "\n", "\r":
            default:
                l.temp = c
                return l.Next()
            }
        }
    case "/":
        switch c := l.char(); c {
        case "/":
            lineComment: for {
                switch l.char() { case "\n", "\r", "\x00": break lineComment }
            }
        case "*":
            blockComment: for {
                switch c := l.char(); c {
                case "*":
                    if c := l.char(); c == "/" {
                        break blockComment
                    } else {
                        l.temp = c
                    }
                case "\x00": current = c; break main
                }
            }
        }
        return l.Next()

    case "+": return Token { "PLUS", current }
    case "-":
        if c := l.char(); c != ">" { current = c; break }
        return Token { "ARROW", "->" }
    case "*": return Token { "STAR", current }
    case "?": return Token { "QUESTION", current }
    case ".": return Token { "DOT", current }
    case "|": return Token { "BAR", current }
    case ";": return Token { "SEMI", current }
    case ":": return Token { "COLON", current }
    case "(": return Token { "L_PAREN", current }
    case ")": return Token { "R_PAREN", current }

    case "\"":
        var builder strings.Builder; builder.WriteString(current)
        stringLiteral: for {
            switch c := l.char(); c {
            case "\"":
                builder.WriteString(c)
                break stringLiteral
            case "\\":
                switch e := l.char(); e {
                case "\n", "\r", "\x00": current = e; break main
                default: builder.WriteString(c + e)
                }
            case "\n", "\r", "\x00": current = c; break main
            default: builder.WriteString(c)
            }
        }
        return Token { "STRING", builder.String() }
    case "[":
        var builder strings.Builder; builder.WriteString(current)
        class: for {
            switch c := l.char(); c {
            case "]":
                builder.WriteString(c)
                break class
            case "\\":
                switch e := l.char(); e {
                case "\n", "\r", "\x00": current = e; break main
                default: builder.WriteString(c + e)
                }
            case "\n", "\r", "\x00": current = c; break main
            default: builder.WriteString(c)
            }
        }
        return Token { "CLASS", builder.String() }
    default:
        if (current >= "a" && current <= "z") || (current >= "A" && current <= "Z") || current == "_" {
            var builder strings.Builder
            builder.WriteString(current)

            c := l.char()
            for (c >= "a" && c <= "z") || (c >= "A" && c <= "Z") || c == "_" || (c >= "0" && c <= "9") {
                builder.WriteString(c)
                c = l.char()
            }
            l.temp = c

            switch id := builder.String(); id {
            case "token":    return Token { "TOKEN", id }
            case "fragment": return Token { "FRAGMENT", id }
            case "skip":     return Token { "SKIP", id }
            default:         return Token { "IDENTIFIER", id }
            }
        }
    }	

    unexpected(current)
    return l.Next()
}

func unexpected(char string) {
    switch char {
    case " ": char = "space"
    case "\n": char = "new line"
    case "\t": char = "tab"
    case "\x00": char = "end of file"
    default: char = fmt.Sprintf("character \"%s\"", char)
    }
    fmt.Println("Unexpected", char)
}
