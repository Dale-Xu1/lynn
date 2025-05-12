package lynn

import (
	"bufio"
	"fmt"
	"slices"
)

// Represents type of token as an enumerated integer.
type TokenType uint
// Location struct. Holds line and column of token.
type Location struct { Line, Col int }
// Token struct. Holds type, value, and location of token.
type Token struct {
    Type     TokenType
    Value    string
    Location Location
}

// Represents a range between characters.
type Range struct { Min, Max rune }

const (WHITESPACE TokenType = iota; COMMENT; RULE; TOKEN; FRAGMENT; SKIP; EQUAL; PLUS; STAR; QUESTION; DOT; BAR; SEMI; COLON; L_PAREN; R_PAREN; ARROW; IDENTIFIER; STRING; CLASS; EOF)
func (t TokenType) String() string { return typeName[t] }
var typeName = map[TokenType]string { WHITESPACE: "WHITESPACE", COMMENT: "COMMENT", RULE: "RULE", TOKEN: "TOKEN", FRAGMENT: "FRAGMENT", SKIP: "SKIP", EQUAL: "EQUAL", PLUS: "PLUS", STAR: "STAR", QUESTION: "QUESTION", DOT: "DOT", BAR: "BAR", SEMI: "SEMI", COLON: "COLON", L_PAREN: "L_PAREN", R_PAREN: "R_PAREN", ARROW: "ARROW", IDENTIFIER: "IDENTIFIER", STRING: "STRING", CLASS: "CLASS", EOF: "EOF" }
var skip = map[TokenType]struct{} { WHITESPACE: {}, COMMENT: {} }

var ranges = []Range { { '\x00', '\x00' }, { '\x01', '\b' }, { '\t', '\t' }, { '\n', '\n' }, { '\v', '\f' }, { '\r', '\r' }, { '\x0e', '\x1f' }, { ' ', ' ' }, { '!', '!' }, { '"', '"' }, { '#', '\'' }, { '(', '(' }, { ')', ')' }, { '*', '*' }, { '+', '+' }, { ',', ',' }, { '-', '-' }, { '.', '.' }, { '/', '/' }, { '0', '9' }, { ':', ':' }, { ';', ';' }, { '<', '<' }, { '=', '=' }, { '>', '>' }, { '?', '?' }, { '@', '@' }, { 'A', 'Z' }, { '[', '[' }, { '\\', '\\' }, { ']', ']' }, { '^', '^' }, { '_', '_' }, { '`', '`' }, { 'a', 'a' }, { 'b', 'd' }, { 'e', 'e' }, { 'f', 'f' }, { 'g', 'g' }, { 'h', 'h' }, { 'i', 'i' }, { 'j', 'j' }, { 'k', 'k' }, { 'l', 'l' }, { 'm', 'm' }, { 'n', 'n' }, { 'o', 'o' }, { 'p', 'p' }, { 'q', 'q' }, { 'r', 'r' }, { 's', 's' }, { 't', 't' }, { 'u', 'u' }, { 'v', 'z' }, { '{', '{' }, { '|', '|' }, { '}', '\U0010ffff' } }
var transitions = []map[int]int {
    { 17: 9, 46: 31, 13: 46, 51: 12, 53: 31, 55: 41, 9: 32, 44: 31, 7: 10, 41: 31, 28: 45, 37: 6, 16: 16, 0: 7, 18: 17, 48: 31, 35: 31, 43: 31, 12: 3, 3: 10, 50: 4, 20: 20, 27: 31, 52: 31, 39: 31, 36: 31, 11: 36, 21: 23, 47: 31, 42: 31, 2: 10, 23: 24, 25: 47, 32: 31, 14: 37, 38: 31, 40: 31, 49: 11, 34: 31, 45: 31, 5: 10 },
    { 52: 31, 49: 31, 51: 31, 41: 31, 37: 31, 40: 31, 27: 31, 47: 31, 39: 31, 19: 31, 38: 31, 50: 31, 45: 31, 32: 31, 53: 31, 43: 31, 48: 31, 36: 22, 42: 31, 44: 31, 34: 31, 35: 31, 46: 31 },
    { 39: 31, 52: 31, 53: 31, 46: 31, 40: 31, 43: 31, 44: 31, 51: 29, 32: 31, 19: 31, 38: 31, 49: 31, 27: 31, 48: 31, 35: 31, 45: 31, 37: 31, 34: 31, 50: 31, 47: 31, 36: 31, 41: 31, 42: 31 },
    { },
    { 45: 31, 39: 31, 32: 31, 35: 31, 42: 18, 43: 31, 46: 31, 37: 31, 41: 31, 44: 31, 34: 31, 27: 31, 19: 31, 49: 31, 38: 31, 40: 31, 50: 31, 53: 31, 47: 31, 48: 31, 36: 31, 52: 31, 51: 31 },
    { },
    { 42: 31, 53: 31, 49: 13, 51: 31, 35: 31, 19: 31, 40: 31, 44: 31, 36: 31, 50: 31, 43: 31, 32: 31, 47: 31, 39: 31, 41: 31, 27: 31, 37: 31, 52: 31, 48: 31, 34: 31, 45: 31, 46: 31, 38: 31 },
    { },
    { },
    { },
    { 7: 10, 2: 10, 3: 10, 5: 10 },
    { 19: 31, 38: 31, 50: 31, 40: 31, 35: 31, 45: 31, 34: 31, 32: 31, 49: 31, 37: 31, 52: 38, 53: 31, 47: 31, 51: 31, 42: 31, 46: 31, 41: 31, 48: 31, 27: 31, 36: 31, 44: 31, 43: 31, 39: 31 },
    { 37: 31, 27: 31, 42: 31, 41: 31, 34: 31, 51: 31, 32: 31, 35: 31, 43: 31, 39: 31, 40: 31, 38: 31, 48: 31, 49: 31, 19: 31, 52: 31, 45: 31, 53: 31, 36: 31, 50: 31, 44: 31, 47: 31, 46: 39 },
    { 27: 31, 32: 31, 38: 31, 50: 31, 39: 31, 43: 31, 40: 31, 53: 31, 45: 31, 42: 31, 51: 31, 46: 31, 49: 31, 19: 31, 37: 31, 52: 31, 41: 31, 34: 27, 47: 31, 36: 31, 44: 31, 48: 31, 35: 31 },
    { 34: 31, 47: 31, 50: 31, 39: 31, 27: 31, 19: 31, 44: 31, 36: 31, 37: 31, 45: 2, 41: 31, 52: 31, 32: 31, 49: 31, 51: 31, 42: 31, 35: 31, 38: 31, 53: 31, 46: 31, 40: 31, 43: 31, 48: 31 },
    { },
    { 24: 15 },
    { 18: 34, 13: 30 },
    { 52: 31, 39: 31, 35: 31, 53: 31, 38: 31, 19: 31, 27: 31, 40: 35, 41: 31, 43: 31, 37: 31, 49: 31, 51: 31, 36: 31, 34: 31, 46: 31, 48: 31, 42: 31, 45: 31, 50: 31, 44: 31, 47: 31, 32: 31 },
    { 34: 31, 45: 31, 48: 31, 39: 31, 19: 31, 38: 31, 50: 31, 42: 31, 32: 31, 47: 31, 44: 31, 43: 31, 53: 31, 49: 31, 51: 31, 36: 31, 37: 31, 40: 31, 27: 31, 41: 31, 35: 31, 46: 31, 52: 31 },
    { },
    { 12: 45, 47: 45, 18: 45, 21: 45, 37: 45, 16: 45, 25: 45, 4: 45, 55: 45, 34: 45, 20: 45, 41: 45, 31: 45, 51: 45, 33: 45, 52: 45, 54: 45, 35: 45, 7: 45, 1: 45, 29: 45, 43: 45, 27: 45, 17: 45, 14: 45, 56: 45, 39: 45, 40: 45, 23: 45, 30: 45, 8: 45, 32: 45, 28: 45, 9: 45, 36: 45, 49: 45, 15: 45, 2: 45, 42: 45, 50: 45, 26: 45, 13: 45, 38: 45, 46: 45, 22: 45, 53: 45, 19: 45, 45: 45, 24: 45, 44: 45, 48: 45, 6: 45, 11: 45, 10: 45 },
    { 38: 31, 34: 31, 39: 31, 40: 31, 41: 31, 46: 31, 44: 31, 50: 31, 45: 31, 52: 31, 32: 31, 27: 31, 35: 31, 48: 31, 49: 31, 51: 31, 36: 31, 42: 31, 43: 31, 53: 31, 19: 31, 47: 31, 37: 31 },
    { },
    { },
    { 48: 31, 49: 31, 41: 31, 36: 26, 35: 31, 37: 31, 45: 31, 44: 31, 52: 31, 43: 31, 53: 31, 47: 31, 50: 31, 39: 31, 27: 31, 19: 31, 51: 31, 42: 31, 46: 31, 40: 31, 32: 31, 34: 31, 38: 31 },
    { 35: 31, 44: 31, 37: 31, 34: 31, 36: 31, 39: 31, 52: 31, 53: 31, 40: 31, 41: 31, 38: 31, 47: 31, 51: 31, 50: 31, 19: 31, 49: 31, 48: 31, 46: 31, 27: 31, 45: 40, 32: 31, 43: 31, 42: 31 },
    { 45: 31, 38: 33, 39: 31, 27: 31, 44: 31, 34: 31, 19: 31, 47: 31, 46: 31, 49: 31, 53: 31, 42: 31, 52: 31, 51: 31, 41: 31, 43: 31, 48: 31, 37: 31, 50: 31, 40: 31, 32: 31, 36: 31, 35: 31 },
    { 14: 30, 27: 30, 40: 30, 39: 30, 20: 30, 31: 30, 29: 30, 35: 30, 48: 30, 30: 30, 43: 30, 25: 30, 28: 30, 52: 30, 37: 30, 36: 30, 13: 30, 4: 30, 7: 30, 46: 30, 53: 30, 1: 30, 9: 30, 34: 30, 47: 30, 6: 30, 18: 8, 16: 30, 12: 30, 38: 30, 32: 30, 17: 30, 21: 30, 3: 30, 50: 30, 8: 30, 23: 30, 22: 30, 56: 30, 5: 30, 24: 30, 2: 30, 51: 30, 19: 30, 10: 30, 45: 30, 33: 30, 42: 30, 55: 30, 26: 30, 41: 30, 11: 30, 44: 30, 49: 30, 54: 30, 15: 30 },
    { 43: 31, 53: 31, 38: 31, 46: 31, 40: 31, 44: 31, 48: 31, 50: 31, 36: 31, 41: 31, 39: 31, 47: 31, 34: 31, 49: 31, 35: 31, 51: 31, 52: 31, 42: 31, 32: 31, 45: 31, 27: 31, 37: 31, 19: 31 },
    { 41: 30, 44: 30, 23: 30, 4: 30, 28: 30, 1: 30, 35: 30, 10: 30, 15: 30, 5: 30, 14: 30, 32: 30, 6: 30, 3: 30, 29: 30, 39: 30, 46: 30, 48: 30, 2: 30, 38: 30, 33: 30, 56: 30, 13: 28, 21: 30, 7: 30, 17: 30, 45: 30, 34: 30, 49: 30, 26: 30, 12: 30, 36: 30, 52: 30, 18: 30, 54: 30, 37: 30, 42: 30, 16: 30, 24: 30, 47: 30, 27: 30, 11: 30, 40: 30, 22: 30, 55: 30, 19: 30, 30: 30, 20: 30, 9: 30, 25: 30, 43: 30, 53: 30, 31: 30, 8: 30, 50: 30, 51: 30 },
    { 41: 31, 35: 31, 43: 31, 44: 31, 38: 31, 52: 31, 19: 31, 39: 31, 53: 31, 37: 31, 47: 31, 40: 31, 27: 31, 46: 31, 36: 31, 42: 31, 50: 31, 34: 31, 45: 31, 48: 31, 49: 31, 32: 31, 51: 31 },
    { 37: 32, 4: 32, 13: 32, 43: 32, 15: 32, 6: 32, 30: 32, 45: 32, 27: 32, 24: 32, 12: 32, 28: 32, 31: 32, 33: 32, 20: 32, 39: 32, 25: 32, 17: 32, 32: 32, 51: 32, 7: 32, 23: 32, 50: 32, 44: 32, 26: 32, 11: 32, 18: 32, 48: 32, 10: 32, 2: 32, 35: 32, 34: 32, 22: 32, 56: 32, 36: 32, 46: 32, 40: 32, 21: 32, 29: 44, 52: 32, 53: 32, 16: 32, 54: 32, 14: 32, 41: 32, 55: 32, 9: 42, 49: 32, 38: 32, 47: 32, 8: 32, 42: 32, 1: 32, 19: 32 },
    { 27: 31, 53: 31, 34: 31, 19: 31, 49: 31, 37: 31, 38: 31, 43: 31, 44: 43, 47: 31, 39: 31, 51: 31, 45: 31, 52: 31, 48: 31, 36: 31, 32: 31, 42: 31, 41: 31, 46: 31, 40: 31, 35: 31, 50: 31 },
    { 5: 8, 56: 34, 35: 34, 43: 34, 23: 34, 49: 34, 22: 34, 45: 34, 1: 34, 17: 34, 31: 34, 52: 34, 48: 34, 36: 34, 19: 34, 32: 34, 38: 34, 40: 34, 13: 34, 8: 34, 55: 34, 30: 34, 7: 34, 51: 34, 25: 34, 44: 34, 0: 8, 47: 34, 28: 34, 50: 34, 46: 34, 39: 34, 29: 34, 4: 34, 12: 34, 34: 34, 21: 34, 18: 34, 24: 34, 41: 34, 14: 34, 15: 34, 16: 34, 9: 34, 54: 34, 2: 34, 27: 34, 20: 34, 37: 34, 11: 34, 10: 34, 33: 34, 42: 34, 26: 34, 6: 34, 3: 8, 53: 34 },
    { 36: 31, 42: 31, 51: 31, 19: 31, 49: 31, 50: 31, 27: 31, 46: 31, 39: 31, 34: 31, 35: 31, 45: 31, 48: 31, 38: 31, 32: 31, 47: 19, 53: 31, 44: 31, 43: 31, 37: 31, 40: 31, 41: 31, 52: 31 },
    { },
    { },
    { 53: 31, 46: 31, 27: 31, 35: 31, 44: 31, 19: 31, 49: 31, 32: 31, 37: 31, 43: 1, 36: 31, 51: 31, 45: 31, 42: 31, 39: 31, 47: 31, 41: 31, 38: 31, 50: 31, 34: 31, 48: 31, 40: 31, 52: 31 },
    { 53: 31, 41: 31, 43: 31, 49: 31, 39: 31, 40: 31, 45: 31, 32: 31, 38: 31, 48: 31, 19: 31, 34: 31, 44: 31, 46: 31, 51: 31, 42: 25, 50: 31, 52: 31, 35: 31, 36: 31, 37: 31, 27: 31, 47: 31 },
    { 50: 31, 40: 31, 34: 31, 53: 31, 35: 31, 39: 31, 52: 31, 44: 31, 48: 31, 49: 31, 51: 31, 42: 31, 19: 31, 46: 31, 36: 31, 37: 31, 41: 31, 27: 31, 43: 31, 45: 31, 38: 31, 47: 31, 32: 31 },
    { },
    { },
    { 43: 31, 53: 31, 34: 31, 48: 31, 39: 31, 40: 31, 32: 31, 47: 31, 41: 31, 27: 31, 51: 31, 45: 31, 50: 31, 52: 31, 49: 31, 44: 31, 37: 31, 38: 31, 36: 14, 35: 31, 42: 31, 19: 31, 46: 31 },
    { 6: 32, 28: 32, 15: 32, 34: 32, 46: 32, 9: 32, 16: 32, 19: 32, 41: 32, 25: 32, 48: 32, 4: 32, 36: 32, 56: 32, 11: 32, 37: 32, 31: 32, 17: 32, 30: 32, 50: 32, 29: 32, 14: 32, 21: 32, 20: 32, 39: 32, 18: 32, 32: 32, 23: 32, 35: 32, 45: 32, 12: 32, 33: 32, 40: 32, 52: 32, 26: 32, 55: 32, 10: 32, 7: 32, 13: 32, 49: 32, 51: 32, 8: 32, 24: 32, 44: 32, 47: 32, 2: 32, 22: 32, 27: 32, 43: 32, 53: 32, 42: 32, 1: 32, 38: 32, 54: 32 },
    { 43: 45, 47: 45, 12: 45, 29: 21, 42: 45, 38: 45, 16: 45, 24: 45, 45: 45, 31: 45, 19: 45, 2: 45, 21: 45, 20: 45, 34: 45, 39: 45, 22: 45, 54: 45, 50: 45, 4: 45, 56: 45, 18: 45, 53: 45, 44: 45, 14: 45, 8: 45, 35: 45, 40: 45, 17: 45, 25: 45, 52: 45, 28: 45, 27: 45, 33: 45, 51: 45, 15: 45, 23: 45, 9: 45, 30: 5, 26: 45, 1: 45, 55: 45, 6: 45, 32: 45, 11: 45, 41: 45, 10: 45, 46: 45, 48: 45, 7: 45, 13: 45, 49: 45, 37: 45, 36: 45 },
    { },
    { },
}
var accept = map[int]TokenType { 22: RULE, 24: EQUAL, 40: TOKEN, 42: STRING, 1: IDENTIFIER, 6: IDENTIFIER, 15: ARROW, 23: SEMI, 29: FRAGMENT, 43: IDENTIFIER, 47: QUESTION, 2: IDENTIFIER, 31: IDENTIFIER, 33: IDENTIFIER, 38: IDENTIFIER, 13: IDENTIFIER, 18: IDENTIFIER, 39: IDENTIFIER, 12: IDENTIFIER, 26: IDENTIFIER, 41: BAR, 9: DOT, 20: COLON, 25: IDENTIFIER, 36: L_PAREN, 5: CLASS, 8: COMMENT, 11: IDENTIFIER, 27: IDENTIFIER, 35: IDENTIFIER, 46: STAR, 14: IDENTIFIER, 19: SKIP, 37: PLUS, 3: R_PAREN, 4: IDENTIFIER, 7: EOF, 10: WHITESPACE }

// Lexer struct. Produces token stream.
type Lexer struct {
    Token   Token
    stream  *InputStream
    handler ErrorHandler
}

// Input stream struct. Produces character stream.
type InputStream struct {
    reader        *bufio.Reader
    location      Location
    buffer, stack []streamData
}
type streamData struct { char rune; location Location }

// Function called when the lexer encounters an error. Expected to bring input stream to synchronization point.
type ErrorHandler func (stream *InputStream, char rune, location Location)
var DEFAULT_HANDLER = func (stream *InputStream, char rune, location Location) {
    // Format special characters
    var str string
    switch char {
    case ' ':        str = "space"
    case '\t':       str = "tab"
    case '\n', '\r': str = "new line"
    case 0:          str = "end of file"
    default:         str = fmt.Sprintf("character %q", string(char))
    }
    // Print formatted error message given an unexpected character
    fmt.Printf("Syntax error: Unexpected %s - %d:%d\n", str, location.Line, location.Col)
    // Find synchronization point
    var whitespace = []rune { 0, ' ', '\t', '\n', '\r' }
    for {
        if char := stream.next(); slices.Contains(whitespace, char) { break }
    }
}

// Returns new lexer struct. Initializes lexer with initial token.
func NewLexer(reader *bufio.Reader, handler ErrorHandler) *Lexer {
    stream := &InputStream { reader, Location { 1, 1 }, make([]streamData, 0), make([]streamData, 0) }
    lexer := &Lexer { Token { }, stream, handler }
    lexer.Next()
    return lexer
}

// Emits next token in stream.
func (l *Lexer) Next() Token {
    location := l.stream.location
    input, stack := make([]rune, 0), make([]int, 0)
    i, state := 0, 0
    for {
        // Read current character in stream and add to input
        char := l.stream.Read(); input = append(input, char)
        next, ok := transitions[state][searchRange(char)]
        if !ok { break } // Exit loop if we cannot transition from this state on the character
        // Store the visited states since the last occurring accepting state
        if _, ok := accept[state]; ok { stack = stack[:0] }
        stack = append(stack, state)
        state = next
        i++
    }
    // Backtrack to last accepting state
    var token TokenType
    for {
        // Unread current character
        l.stream.Unread()
        if t, ok := accept[state]; ok { token = t; break }
        if len(stack) == 0 {
            // If no accepting state was encountered, raise error and synchronize
            l.stream.synchronize(l.handler, input[i], l.stream.location)
            return l.Next() // Attempt to read token again
        }
        // Restore previously visited states
        state, stack = stack[len(stack) - 1], stack[:len(stack) - 1]
        i--
    }
    l.stream.reset()
    if _, ok := skip[token]; ok { return l.Next() } // Skip token
    // Create token and store as current token
    l.Token = Token { token, string(input[:i]), location }
    return l.Token
}

// Returns the next character in the input stream while maintaining location.
func (i *InputStream) Read() rune {
    // Store previous location in stack and read next character
    l := i.location; char := i.next()
    i.stack = append(i.stack, streamData { char, l })
    return char
}

func (i *InputStream) next() rune {
    // If buffered data exists, consume it before requesting new data from the reader
    if len(i.buffer) > 0 {
        data := i.buffer[len(i.buffer) - 1]; i.buffer = i.buffer[:len(i.buffer) - 1]
        i.location = data.location
        return data.char
    }
    char, _, err := i.reader.ReadRune()
    if err != nil { return 0 } // Return a null character if stream does not have any more characters to emit
    // Update current location based on character read
    l := &i.location
    switch char {
    case '\n': l.Line++; l.Col = 1
    default: l.Col++
    }
    return char
}

// Unreads the current character in the input stream while maintaining location.
func (i *InputStream) Unread() {
    if len(i.stack) == 0 { return }
    data := i.stack[len(i.stack) - 1]; i.stack = i.stack[:len(i.stack) - 1]
    l := i.location; i.location = data.location
    i.buffer = append(i.buffer, streamData{ data.char, l })
}

// Releases previously read characters.
func (i *InputStream) reset() { i.stack = i.stack[:0] }
func (i *InputStream) synchronize(handler ErrorHandler, char rune, location Location) {
    i.reset()
    handler(i, char, location)
    i.reset()
}

// Run binary search on character to find index associated with the range that contains the character.
func searchRange(char rune) int {
    low, high := 0, len(ranges) - 1
    for low <= high {
        mid := (low + high) / 2
        r := ranges[mid]
        if char >= r.Min && char <= r.Max { return mid }
        if char > r.Max {
            low = mid + 1
        } else {
            high = mid - 1
        }
    }
    return -1
}

// Tests if the type of the current token in the stream matches the provided type. If the types match, the next token is emitted.
func (l *Lexer) Match(token TokenType) bool {
    if l.Token.Type == token {
        l.Next()
        return true
    }
    return false
}

// FOR DEBUG PURPOSES:
// Consumes all tokens emitted by lexer and prints them to the standard output.
func (l *Lexer) PrintTokenStream() {
    for l.Token.Type != EOF {
        fmt.Printf("%-7s | %-16s %-16s\n", fmt.Sprintf("%d:%d", l.Token.Location.Line, l.Token.Location.Col), l.Token.Type, l.Token.Value)
        l.Next()
    }
}
