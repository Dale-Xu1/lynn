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

// Returns new lexer struct
func NewLexer(reader *bufio.Reader) *Lexer {
    return &Lexer { reader, "" }
}

// Returns the next character in the input stream as a string. The lexer also contains a temporary buffer to store one character.
// If the temporary buffer contains a character, it is outputted instead of having the next character read from the input stream.
func (l *Lexer) char() string {
    if l.temp != "" {
        c := l.temp; l.temp = ""
        return c
    }

    c, _, err := l.reader.ReadRune()
    if err != nil { return "\x00" } // Return a null character if stream does not have any more characters to emit
    return string(c)
}

// Reads character stream until new token is emitted. Recursively calls if pattern is skipped or an unexpected character is encountered.
func (l *Lexer) Next() Token {
    current := l.char()
    main: switch current {
    case "\x00": return Token { "EOF", "" }
    case " ", "\t", "\n", "\r":
        // Repeatedly read whitespace characters until non-whitespace is found
        for {
            switch c := l.char(); c {
            case " ", "\t", "\n", "\r":
            default:
                l.temp = c // Store next character in temporary buffer for next token
                return l.Next()
            }
        }
    case "/":
        // Match comment patterns and skip
        switch c := l.char(); c {
        case "/": // Line comment, read characters until termination found
            for {
                switch l.char() { case "\n", "\r", "\x00": return l.Next() }
            }
        case "*": // Block comment, match for closing pattern
            for {
                switch c := l.char(); c {
                case "*":
                    if c := l.char(); c == "/" {
                        return l.Next()
                    } else {
                        l.temp = c // Store in buffer and keep reading stream
                    }
                case "\x00": current = c; break main
                }
            }
        default: l.temp = c // Possible that character after / belongs to valid token
        }

    // Tokenize symbols
    case "=": return Token { "EQUAL", current }
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

    case "\"": // Tokenize string
        var builder strings.Builder; builder.WriteString(current)
        stringLiteral: for {
            switch c := l.char(); c {
            case "\"": // Terminate string when input stream emits "
                builder.WriteString(c)
                break stringLiteral
            case "\\":
                // Read any character including " after escape character
                switch e := l.char(); e {
                case "\n", "\r", "\x00": current = e; break main
                default: builder.WriteString(c + e)
                }
            case "\n", "\r", "\x00": current = c; break main
            default: builder.WriteString(c)
            }
        }
        return Token { "STRING", builder.String() }
    case "[": // Tokenize class
        var builder strings.Builder; builder.WriteString(current)
        class: for {
            switch c := l.char(); c {
            case "]": // Terminate class when input stream emits ]
                builder.WriteString(c)
                break class
            case "\\":
                // Read any character including ] after escape character
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
        // Valid identifier or keyword if letter or _ is read
        if (current >= "a" && current <= "z") || (current >= "A" && current <= "Z") || current == "_" {
            var builder strings.Builder
            builder.WriteString(current)

            // Read letters and digits until non-match is found
            c := l.char()
            for (c >= "a" && c <= "z") || (c >= "A" && c <= "Z") || c == "_" || (c >= "0" && c <= "9") {
                builder.WriteString(c)
                c = l.char()
            }
            l.temp = c // Store non-match in buffer as it could be part of a valid token

            // Test if identifier is a keyword
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

// Prints a formatted error message given an unexpected character.
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

// FOR DEBUG PURPOSES:
// Consumes all tokens emitted by lexer and prints them to the standard output.
func (l *Lexer) PrintTokenStream() {
    for {
        token := l.Next()
        if token.Type == "EOF" { break }

        fmt.Printf("%-16s %s\n", token.Type, token.Value)
    }
}
