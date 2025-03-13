package lynn

import (
	"bufio"
	"fmt"
	"strings"
)

func Lex(reader *bufio.Reader) {
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

    next()
    main: for {
        switch current {
        case "\x00":
            fmt.Println("EOF")
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
                    case "\x00": panic("Unexpected EOF")
                    default: next()
                    }
                }
            }

        case "+": fmt.Println("PLUS     +")
        case "-":
            if next() == ">" {
                fmt.Println("ARROW    ->")
            } else {
                fmt.Println("DASH     -")
                continue
            }
        case "*": fmt.Println("STAR     *")
        case "?": fmt.Println("QUESTION ?")
        case ".": fmt.Println("DOT      .")
        case "|": fmt.Println("BAR      |")
        case ";": fmt.Println("SEMI     ;")
        case ":": fmt.Println("COLON    :")
        case "(": fmt.Println("L_PAREN  (")
        case ")": fmt.Println("R_PAREN  )")

        case "\"":
            var builder strings.Builder
            builder.WriteString(current)
            builder.WriteString(next())
            stringLiteral: for {
                switch current {
                case "\"": break stringLiteral
                case "\\":
                    switch c := next(); c {
                    case "\n", "\r", "\x00": panic("Unexpected character in escape sequence")
                    default: builder.WriteString(c + next())
                    }
                case "\n", "\r", "\x00": panic("Unexpected character in string literal")
                default: builder.WriteString(next())
                }
            }
            fmt.Printf("STRING   %s\n", builder.String())
        case "[":
            var builder strings.Builder
            builder.WriteString(current)
            builder.WriteString(next())
            class: for {
                switch current {
                case "]": break class
                case "\\":
                    switch c := next(); c {
                    case "\n", "\r", "\x00": panic("Unexpected character in escape sequence")
                    default: builder.WriteString(c + next())
                    }
                case "\n", "\r", "\x00": panic("Unexpected character in class")
                default: builder.WriteString(next())
                }
            }
            fmt.Printf("CLASS    %s\n", builder.String())
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
                case "token": fmt.Printf("TOKEN    %s\n", id)
                case "fragment": fmt.Printf("FRAGMENT %s\n", id)
                case "skip": fmt.Printf("SKIP     %s\n", id)
                default: fmt.Printf("IDENTIFIER %s\n", id)
                }
                continue main
            } else {
                panic("Unexpected character")
            }	
        }
        next()
    }
}
