package lynn

import (
	"bufio"
	"fmt"
	"strings"
)

// Represents type of token as an enumerated integer.
type TokenType int
// Token struct. Holds type and value of token.
type Token struct {
    Type  TokenType
    Value string
}

// Lexer struct. Produces token stream.
type Lexer struct {
    Token  Token
    reader *bufio.Reader
    temp   rune
}

// Returns new lexer struct. Initializes lexer with initial token.
func NewLexer(reader *bufio.Reader) *Lexer {
    lexer := &Lexer { Token { }, reader, -1 }
    lexer.Next()
    return lexer
}

// Returns the next character in the input stream as a string. The lexer also contains a temporary buffer to store one character.
// If the temporary buffer contains a character, it is outputted instead of having the next character read from the input stream.
func (l *Lexer) char() rune {
    if l.temp != -1 {
        c := l.temp; l.temp = -1
        return c
    }

    c, _, err := l.reader.ReadRune()
    if err != nil { return 0 } // Return a null character if stream does not have any more characters to emit
    return c
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

// Tests if the type of the current token in the stream matches the provided type. Prints error message if types do not match.
func (l *Lexer) Expect(token TokenType) bool {
    r := l.Match(token)
    if !r { fmt.Printf("Syntax error: Unexpected token \"%s\"\n", l.Token.Value) }
    return r
}

func (l *Lexer) next() Token {
    // Read character stream until new token is emitted
    // Recursively calls if pattern is skipped or an unexpected character is encountered
    current := l.char()
    main: switch current {
    case 0: return Token { EOF, "" }
    case ' ', '\t', '\n', '\r':
        // Repeatedly read whitespace characters until non-whitespace is found
        for {
            switch c := l.char(); c {
            case ' ', '\t', '\n', '\r':
            default:
                l.temp = c // Store next character in temporary buffer for next token
                return l.next()
            }
        }
    case '/':
        // Match comment patterns and skip
        switch c := l.char(); c {
        case '/': // Line comment, read characters until termination found
            for {
                switch l.char() { case '\n', '\r', 0: return l.next() }
            }
        case '*': // Block comment, match for closing pattern
            for {
                switch c := l.char(); c {
                case '*':
                    if c := l.char(); c == '/' {
                        return l.next()
                    } else {
                        l.temp = c // Store in buffer and keep reading stream
                    }
                case 0: current = c; break main
                }
            }
        default: l.temp = c // Possible that character after / belongs to valid token
        }

    // Tokenize symbols
    case '=': return Token { EQUAL,    string(current) }
    case '+': return Token { PLUS,     string(current) }
    case '*': return Token { STAR,     string(current) }
    case '?': return Token { QUESTION, string(current) }
    case '.': return Token { DOT,      string(current) }
    case '|': return Token { BAR,      string(current) }
    case ';': return Token { SEMI,     string(current) }
    case ':': return Token { COLON,    string(current) }
    case '(': return Token { L_PAREN,  string(current) }
    case ')': return Token { R_PAREN,  string(current) }
    case '-':
        if c := l.char(); c != '>' { current = c; break }
        return Token { ARROW, "->" }

    case '"': // Tokenize string
        var builder strings.Builder; builder.WriteRune(current)
        for {
            switch c := l.char(); c {
            case '"': // Terminate string when input stream emits "
                builder.WriteRune(c)
                return Token { STRING, builder.String() }
            case '\\':
                // Read any character including " after escape character
                switch e := l.char(); e {
                case '\n', '\r', 0: current = e; break main
                default: builder.WriteRune(c); builder.WriteRune(e)
                }
            case '\n', '\r', 0: current = c; break main
            default: builder.WriteRune(c)
            }
        }
    case '[': // Tokenize class
        var builder strings.Builder; builder.WriteRune(current)
        for {
            switch c := l.char(); c {
            case ']': // Terminate class when input stream emits ]
                builder.WriteRune(c)
                return Token { CLASS, builder.String() }
            case '\\':
                // Read any character including ] after escape character
                switch e := l.char(); e {
                case '\n', '\r', 0: current = e; break main
                default: builder.WriteRune(c); builder.WriteRune(e)
                }
            case '\n', '\r', 0: current = c; break main
            default: builder.WriteRune(c)
            }
        }
    default:
        // Valid identifier or keyword if letter or _ is read
        if (current >= 'a' && current <= 'z') || (current >= 'A' && current <= 'Z') || current == '_' {
            var builder strings.Builder
            builder.WriteRune(current)

            // Read letters and digits until non-match is found
            c := l.char()
            for (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_' || (c >= '0' && c <= '9') {
                builder.WriteRune(c)
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

func unexpected(char rune) {
    // Print formatted error message given an unexpected character
    var str string
    switch char {
    case ' ': str = "space"
    case '\n': str = "new line"
    case '\t': str = "tab"
    case 0: str = "end of file"
    default: str = fmt.Sprintf("character \"%c\"", char)
    }
    fmt.Println("Syntax error: Unexpected", str)
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
