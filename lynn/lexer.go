package lynn

import (
	"bufio"
	"fmt"
	"strings"
)

type TokenType int
type Token struct {
    Type  TokenType
    Value string
}

type Lexer struct {
    Token  Token
    reader *bufio.Reader
    temp   string
}

// Returns new lexer struct. Initializes lexer with initial token.
func NewLexer(reader *bufio.Reader) *Lexer {
    lexer := &Lexer { Token { }, reader, "" }
    lexer.Next()
    return lexer
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

// Emits next token in stream.
func (l *Lexer) Next() Token {
    l.Token = l.next()
    return l.Token
}

// Tests if the type of the current token in the stream matches the provided type. If the types match, the next token is emitted.
func (l *Lexer) Match(token TokenType) bool {
    if l.Token.Type == token {
        l.Next()
        return true
    }
    return false
}

// Reads character stream until new token is emitted. Recursively calls if pattern is skipped or an unexpected character is encountered.
func (l *Lexer) next() Token {
    current := l.char()
    main: switch current {
    case "\x00": return Token { EOF, "" }
    case " ", "\t", "\n", "\r":
        // Repeatedly read whitespace characters until non-whitespace is found
        for {
            switch c := l.char(); c {
            case " ", "\t", "\n", "\r":
            default:
                l.temp = c // Store next character in temporary buffer for next token
                return l.next()
            }
        }
    case "/":
        // Match comment patterns and skip
        switch c := l.char(); c {
        case "/": // Line comment, read characters until termination found
            for {
                switch l.char() { case "\n", "\r", "\x00": return l.next() }
            }
        case "*": // Block comment, match for closing pattern
            for {
                switch c := l.char(); c {
                case "*":
                    if c := l.char(); c == "/" {
                        return l.next()
                    } else {
                        l.temp = c // Store in buffer and keep reading stream
                    }
                case "\x00": current = c; break main
                }
            }
        default: l.temp = c // Possible that character after / belongs to valid token
        }

    // Tokenize symbols
    case "=": return Token { EQUAL, current }
    case "+": return Token { PLUS, current }
    case "-":
        if c := l.char(); c != ">" { current = c; break }
        return Token { ARROW, "->" }
    case "*": return Token { STAR, current }
    case "?": return Token { QUESTION, current }
    case ".": return Token { DOT, current }
    case "|": return Token { BAR, current }
    case ";": return Token { SEMI, current }
    case ":": return Token { COLON, current }
    case "(": return Token { L_PAREN, current }
    case ")": return Token { R_PAREN, current }

    case "\"": // Tokenize string
        var builder strings.Builder; builder.WriteString(current)
        for {
            switch c := l.char(); c {
            case "\"": // Terminate string when input stream emits "
                builder.WriteString(c)
                return Token { STRING, builder.String() }
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
    case "[": // Tokenize class
        var builder strings.Builder; builder.WriteString(current)
        for {
            switch c := l.char(); c {
            case "]": // Terminate class when input stream emits ]
                builder.WriteString(c)
                return Token { CLASS, builder.String() }
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
            case "token":    return Token { TOKEN, id }
            case "fragment": return Token { FRAGMENT, id }
            case "skip":     return Token { SKIP, id }
            default:         return Token { IDENTIFIER, id }
            }
        }
    }	

    unexpected(current)
    return l.next()
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
    fmt.Println("Syntax error: Unexpected", char)
}

// FOR DEBUG PURPOSES:
// Consumes all tokens emitted by lexer and prints them to the standard output.
func (l *Lexer) PrintTokenStream() {
    for l.Token.Type != EOF {
        fmt.Printf("%-16s %s\n", l.Token.Type, l.Token.Value)
        l.Next()
    }
}

const (
    EOF TokenType = iota

    TOKEN
    FRAGMENT
    SKIP

    EQUAL
    PLUS
    STAR
    QUESTION
    DOT
    BAR
    SEMI
    COLON
    L_PAREN
    R_PAREN
    ARROW

    IDENTIFIER
    STRING
    CLASS
)

func (t TokenType) String() string { return typeName[t] }
var typeName = map[TokenType]string {
    EOF:        "EOF",

    TOKEN:      "TOKEN",
    FRAGMENT:   "FRAGMENT",
    SKIP:       "SKIP",

    EQUAL:      "EQUAL",
    PLUS:       "PLUS",
    STAR:       "STAR",
    QUESTION:   "QUESTION",
    DOT:        "DOT",
    BAR:        "BAR",
    SEMI:       "SEMI",
    COLON:      "COLON",
    L_PAREN:    "L_PAREN",
    R_PAREN:    "R_PAREN",
    ARROW:      "ARROW",

    IDENTIFIER: "IDENTIFIER",
    STRING:     "STRING",
    CLASS:      "CLASS",
}
