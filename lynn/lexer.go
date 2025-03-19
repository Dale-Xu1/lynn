package lynn

import (
	"bufio"
	"fmt"
	"strings"
)

// Represents type of token as an enumerated integer.
type TokenType uint
// Location struct. Holds line and column of token.
type Location struct { Line, Col uint }
// Token struct. Holds type, value, and location of token.
type Token struct {
    Type     TokenType
    Value    string
    Location Location
}

// Lexer struct. Produces token stream.
type Lexer struct {
    Token     Token
    reader    *bufio.Reader
    temp      rune
    line, col uint
}

// Returns new lexer struct. Initializes lexer with initial token.
func NewLexer(reader *bufio.Reader) *Lexer {
    lexer := &Lexer { Token { }, reader, -1, 1, 1 }
    lexer.Next()
    return lexer
}

// Returns the next character in the input stream as a string while maintaining location.
// If the temporary buffer contains a character, it is outputted instead of having the next character read from the input stream.
func (l *Lexer) char() rune {
    var c rune
    if l.temp != -1 {
        c = l.temp; l.temp = -1
    } else {
        var err error
        c, _, err = l.reader.ReadRune()
        if err != nil { return 0 } // Return a null character if stream does not have any more characters to emit
    }
    switch c {
    case '\n': l.line++; l.col = 1
    default: l.col++
    }
    return c
}

// Reads the next character in the input stream if the temporary buffer is empty, but does not maintain location.
// Result is stored in the lexer's temporary buffer, which is emitted on the next call to char().
func (l *Lexer) lookahead() rune {
    if l.temp != -1 { return l.temp }
    var err error
    l.temp, _, err = l.reader.ReadRune()
    if err != nil { return 0 } // Return a null character if stream does not have any more characters to emit
    return l.temp
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

func (l *Lexer) location() Location { return Location { l.line, l.col } }
func (l *Lexer) next() Token {
    // Read character stream until new token is emitted
    // Recursively calls if pattern is skipped or an unexpected character is encountered
    location, current := l.location(), l.char()
    main: switch current {
    case 0: return Token { EOF, "", location }
    case ' ', '\t', '\n', '\r':
        // Repeatedly read whitespace characters until non-whitespace is found
        for {
            switch c := l.lookahead(); c {
            case ' ', '\t', '\n', '\r': l.char()
            default: return l.next()
            }
        }
    case '/':
        // Match comment patterns and skip
        switch c := l.lookahead(); c {
        case '/': // Line comment, read characters until termination found
            l.char()
            for {
                switch l.char() { case '\n', '\r', 0: return l.next() }
            }
        case '*': // Block comment, match for closing pattern
            l.char()
            for {
                switch c := l.char(); c {
                case '*':
                    if c := l.lookahead(); c == '/' {
                        l.char()
                        return l.next()
                    }
                case 0:
                    unexpected(c, l.location())
                    break main
                }
            }
        }

    // Tokenize symbols
    case '=': return Token { EQUAL,    string(current), location }
    case '+': return Token { PLUS,     string(current), location }
    case '*': return Token { STAR,     string(current), location }
    case '?': return Token { QUESTION, string(current), location }
    case '.': return Token { DOT,      string(current), location }
    case '|': return Token { BAR,      string(current), location }
    case ';': return Token { SEMI,     string(current), location }
    case ':': return Token { COLON,    string(current), location }
    case '(': return Token { L_PAREN,  string(current), location }
    case ')': return Token { R_PAREN,  string(current), location }
    case '-':
        n := l.location()
        if c := l.char(); c != '>' { unexpected(c, n); break }
        return Token { ARROW, "->", location }

    case '"': // Tokenize string
        var builder strings.Builder; builder.WriteRune(current)
        for {
            n := l.location()
            switch c := l.char(); c {
            case '"': // Terminate string when input stream emits "
                builder.WriteRune(c)
                return Token { STRING, builder.String(), location }
            case '\\':
                // Read any character including " after escape character
                n := l.location()
                switch e := l.char(); e {
                case '\n', '\r', 0: unexpected(e, n); break main
                default: builder.WriteRune(c); builder.WriteRune(e)
                }
            case '\n', '\r', 0: unexpected(c, n); break main
            default: builder.WriteRune(c)
            }
        }
    case '[': // Tokenize class
        var builder strings.Builder; builder.WriteRune(current)
        for {
            n := l.location()
            switch c := l.char(); c {
            case ']': // Terminate class when input stream emits ]
                builder.WriteRune(c)
                return Token { CLASS, builder.String(), location }
            case '\\':
                // Read any character including ] after escape character
                n := l.location()
                switch e := l.char(); e {
                case '\n', '\r', 0: unexpected(e, n); break main
                default: builder.WriteRune(c); builder.WriteRune(e)
                }
            case '\n', '\r', 0: unexpected(c, n); break main
            default: builder.WriteRune(c)
            }
        }
    default:
        // Valid identifier or keyword if letter or _ is read
        if (current >= 'a' && current <= 'z') || (current >= 'A' && current <= 'Z') || current == '_' {
            var builder strings.Builder
            builder.WriteRune(current)
            // Read letters and digits until non-match is found
            c := l.lookahead()
            for (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_' || (c >= '0' && c <= '9') {
                builder.WriteRune(c)
                l.char(); c = l.lookahead()
            }

            // Test if identifier is a keyword
            switch id := builder.String(); id {
            case "token":    return Token { TOKEN,      id, location }
            case "fragment": return Token { FRAGMENT,   id, location }
            case "skip":     return Token { SKIP,       id, location }
            default:         return Token { IDENTIFIER, id, location }
            }
        } else {
            unexpected(current, location)
        }
    }	

    return l.next()
}

func unexpected(char rune, location Location) {
    // Format special characters
    var str string
    switch char {
    case ' ':        str = "space"
    case '\t':       str = "tab"
    case '\n', '\r': str = "new line"
    case 0:          str = "end of file"
    default:         str = fmt.Sprintf("character \"%c\"", char)
    }
    // Print formatted error message given an unexpected character
    fmt.Printf("Syntax error: Unexpected %s - %d:%d\n", str, location.Line, location.Col)
}

// FOR DEBUG PURPOSES:
// Consumes all tokens emitted by lexer and prints them to the standard output.
func (l *Lexer) PrintTokenStream() {
    for l.Token.Type != EOF {
        fmt.Printf("%-7s | %-16s %-16s\n", fmt.Sprintf("%d:%d", l.Token.Location.Line, l.Token.Location.Col), l.Token.Type, l.Token.Value)
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
