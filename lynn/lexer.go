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

const (EOF TokenType = iota; WHITESPACE; COMMENT; TOKEN; FRAGMENT; SKIP; EQUAL; PLUS; STAR; QUESTION; DOT; BAR; SEMI; COLON; L_PAREN; R_PAREN; ARROW; IDENTIFIER; STRING; CLASS)
func (t TokenType) String() string { return typeName[t] }
var typeName = map[TokenType]string { EOF: "EOF", WHITESPACE: "WHITESPACE", COMMENT: "COMMENT", TOKEN: "TOKEN", FRAGMENT: "FRAGMENT", SKIP: "SKIP", EQUAL: "EQUAL", PLUS: "PLUS", STAR: "STAR", QUESTION: "QUESTION", DOT: "DOT", BAR: "BAR", SEMI: "SEMI", COLON: "COLON", L_PAREN: "L_PAREN", R_PAREN: "R_PAREN", ARROW: "ARROW", IDENTIFIER: "IDENTIFIER", STRING: "STRING", CLASS: "CLASS" }
var skip = map[TokenType]struct{} { WHITESPACE: {}, COMMENT: {} }

var ranges = []Range { { '\x00', '\x00' }, { '\x01', '\b' }, { '\t', '\t' }, { '\n', '\n' }, { '\v', '\f' }, { '\r', '\r' }, { '\x0e', '\x1f' }, { ' ', ' ' }, { '!', '!' }, { '"', '"' }, { '#', '\'' }, { '(', '(' }, { ')', ')' }, { '*', '*' }, { '+', '+' }, { ',', ',' }, { '-', '-' }, { '.', '.' }, { '/', '/' }, { '0', '9' }, { ':', ':' }, { ';', ';' }, { '<', '<' }, { '=', '=' }, { '>', '>' }, { '?', '?' }, { '@', '@' }, { 'A', 'Z' }, { '[', '[' }, { '\\', '\\' }, { ']', ']' }, { '^', '^' }, { '_', '_' }, { '`', '`' }, { 'a', 'a' }, { 'b', 'd' }, { 'e', 'e' }, { 'f', 'f' }, { 'g', 'g' }, { 'h', 'h' }, { 'i', 'i' }, { 'j', 'j' }, { 'k', 'k' }, { 'l', 'l' }, { 'm', 'm' }, { 'n', 'n' }, { 'o', 'o' }, { 'p', 'p' }, { 'q', 'q' }, { 'r', 'r' }, { 's', 's' }, { 't', 't' }, { 'u', 'z' }, { '{', '{' }, { '|', '|' }, { '}', '\U0010ffff' } }
var transitions = []map[int]int {
    { 50: 7, 13: 40, 3: 9, 9: 34, 52: 24, 43: 24, 35: 24, 18: 30, 5: 9, 12: 18, 47: 24, 16: 31, 37: 41, 2: 9, 46: 24, 27: 24, 14: 14, 20: 33, 51: 26, 28: 10, 0: 2, 38: 24, 45: 24, 32: 24, 25: 28, 42: 24, 44: 24, 54: 39, 23: 15, 48: 24, 49: 24, 34: 24, 7: 9, 41: 24, 17: 6, 21: 20, 40: 24, 36: 24, 11: 3, 39: 24 },
    { },
    { },
    { },
    { 34: 24, 36: 24, 45: 24, 49: 24, 19: 24, 47: 8, 42: 24, 41: 24, 44: 24, 38: 24, 37: 24, 43: 24, 50: 24, 40: 24, 52: 24, 39: 24, 48: 24, 27: 24, 32: 24, 51: 24, 35: 24, 46: 24 },
    { 37: 24, 48: 24, 36: 24, 35: 24, 44: 24, 50: 24, 32: 24, 51: 24, 40: 24, 41: 24, 38: 24, 34: 24, 39: 24, 45: 24, 27: 24, 47: 24, 52: 24, 42: 24, 43: 24, 49: 24, 19: 24, 46: 24 },
    { },
    { 37: 24, 46: 24, 49: 24, 36: 24, 41: 24, 35: 24, 40: 24, 38: 24, 51: 24, 47: 24, 45: 24, 43: 24, 27: 24, 42: 29, 50: 24, 39: 24, 32: 24, 44: 24, 34: 24, 48: 24, 19: 24, 52: 24 },
    { 44: 24, 42: 24, 43: 24, 41: 24, 46: 24, 39: 24, 40: 24, 48: 24, 34: 24, 32: 24, 38: 24, 35: 24, 52: 24, 45: 24, 49: 24, 37: 24, 51: 24, 47: 24, 50: 24, 27: 24, 19: 24, 36: 24 },
    { 3: 9, 5: 9, 7: 9, 2: 9 },
    { 48: 10, 12: 10, 54: 10, 47: 10, 4: 10, 8: 10, 35: 10, 31: 10, 19: 10, 55: 10, 18: 10, 27: 10, 37: 10, 43: 10, 51: 10, 21: 10, 45: 10, 22: 10, 33: 10, 25: 10, 49: 10, 10: 10, 32: 10, 26: 10, 53: 10, 44: 10, 20: 10, 41: 10, 6: 10, 11: 10, 29: 11, 46: 10, 2: 10, 15: 10, 16: 10, 13: 10, 34: 10, 1: 10, 30: 1, 23: 10, 24: 10, 36: 10, 28: 10, 38: 10, 9: 10, 42: 10, 52: 10, 40: 10, 39: 10, 50: 10, 7: 10, 17: 10, 14: 10 },
    { 23: 10, 19: 10, 18: 10, 48: 10, 25: 10, 2: 10, 46: 10, 8: 10, 38: 10, 10: 10, 33: 10, 7: 10, 52: 10, 17: 10, 21: 10, 51: 10, 16: 10, 36: 10, 45: 10, 49: 10, 30: 10, 41: 10, 55: 10, 54: 10, 29: 10, 22: 10, 28: 10, 24: 10, 15: 10, 14: 10, 34: 10, 6: 10, 31: 10, 11: 10, 40: 10, 43: 10, 37: 10, 20: 10, 53: 10, 47: 10, 32: 10, 26: 10, 35: 10, 39: 10, 4: 10, 9: 10, 50: 10, 1: 10, 42: 10, 13: 10, 12: 10, 27: 10, 44: 10 },
    { },
    { },
    { },
    { },
    { },
    { 12: 22, 47: 22, 23: 22, 49: 22, 24: 22, 54: 22, 36: 22, 31: 22, 42: 22, 46: 22, 40: 22, 50: 22, 51: 22, 16: 22, 41: 22, 2: 22, 8: 22, 13: 22, 4: 22, 28: 22, 25: 22, 11: 22, 37: 22, 39: 22, 43: 22, 15: 22, 38: 22, 45: 22, 26: 22, 53: 22, 33: 22, 35: 22, 7: 22, 17: 22, 22: 22, 5: 22, 52: 22, 18: 12, 19: 22, 10: 22, 6: 22, 20: 22, 3: 22, 44: 22, 55: 22, 14: 22, 21: 22, 34: 22, 48: 22, 1: 22, 27: 22, 9: 22, 32: 22, 30: 22, 29: 22 },
    { },
    { 45: 24, 46: 24, 52: 24, 41: 24, 42: 24, 44: 32, 47: 24, 43: 24, 51: 24, 19: 24, 36: 24, 34: 24, 40: 24, 37: 24, 38: 24, 39: 24, 49: 24, 32: 24, 50: 24, 48: 24, 27: 24, 35: 24 },
    { },
    { 9: 34, 16: 34, 28: 34, 8: 34, 1: 34, 18: 34, 23: 34, 4: 34, 25: 34, 45: 34, 42: 34, 37: 34, 52: 34, 10: 34, 26: 34, 34: 34, 29: 34, 32: 34, 43: 34, 50: 34, 14: 34, 30: 34, 36: 34, 12: 34, 11: 34, 7: 34, 24: 34, 6: 34, 19: 34, 47: 34, 17: 34, 31: 34, 38: 34, 39: 34, 44: 34, 15: 34, 53: 34, 51: 34, 33: 34, 48: 34, 35: 34, 46: 34, 20: 34, 55: 34, 21: 34, 41: 34, 49: 34, 40: 34, 27: 34, 22: 34, 13: 34, 2: 34, 54: 34 },
    { 7: 22, 10: 22, 51: 22, 9: 22, 3: 22, 54: 22, 11: 22, 41: 22, 26: 22, 48: 22, 15: 22, 38: 22, 20: 22, 21: 22, 50: 22, 42: 22, 45: 22, 23: 22, 14: 22, 12: 22, 52: 22, 39: 22, 16: 22, 8: 22, 27: 22, 43: 22, 46: 22, 1: 22, 13: 17, 55: 22, 4: 22, 36: 22, 40: 22, 34: 22, 35: 22, 33: 22, 2: 22, 22: 22, 29: 22, 31: 22, 24: 22, 53: 22, 18: 22, 37: 22, 47: 22, 28: 22, 17: 22, 6: 22, 25: 22, 49: 22, 32: 22, 19: 22, 5: 22, 30: 22, 44: 22 },
    { 18: 23, 24: 23, 10: 23, 0: 12, 37: 23, 42: 23, 49: 23, 45: 23, 51: 23, 14: 23, 41: 23, 7: 23, 29: 23, 13: 23, 47: 23, 32: 23, 43: 23, 16: 23, 34: 23, 38: 23, 15: 23, 35: 23, 44: 23, 26: 23, 5: 12, 33: 23, 36: 23, 20: 23, 9: 23, 30: 23, 17: 23, 25: 23, 39: 23, 50: 23, 54: 23, 22: 23, 23: 23, 11: 23, 55: 23, 28: 23, 12: 23, 4: 23, 1: 23, 21: 23, 2: 23, 27: 23, 52: 23, 53: 23, 48: 23, 6: 23, 31: 23, 46: 23, 8: 23, 3: 12, 40: 23, 19: 23 },
    { 49: 24, 41: 24, 38: 24, 19: 24, 50: 24, 37: 24, 43: 24, 47: 24, 42: 24, 52: 24, 45: 24, 51: 24, 36: 24, 27: 24, 34: 24, 39: 24, 44: 24, 35: 24, 48: 24, 40: 24, 32: 24, 46: 24 },
    { 35: 24, 49: 24, 52: 24, 42: 24, 34: 24, 48: 24, 43: 24, 32: 24, 46: 24, 44: 24, 47: 24, 51: 24, 40: 24, 41: 24, 39: 24, 38: 19, 36: 24, 50: 24, 27: 24, 37: 24, 45: 24, 19: 24 },
    { 32: 24, 37: 24, 44: 24, 41: 24, 40: 24, 50: 24, 49: 24, 36: 24, 43: 24, 46: 43, 47: 24, 52: 24, 38: 24, 19: 24, 42: 24, 34: 24, 39: 24, 45: 24, 27: 24, 51: 24, 48: 24, 35: 24 },
    { 46: 24, 48: 24, 50: 24, 38: 24, 44: 24, 47: 24, 45: 24, 39: 24, 34: 24, 43: 24, 40: 24, 52: 24, 49: 24, 19: 24, 37: 24, 32: 24, 35: 24, 51: 24, 36: 37, 41: 24, 42: 24, 27: 24 },
    { },
    { 42: 24, 48: 24, 35: 24, 39: 24, 41: 24, 37: 24, 52: 24, 40: 4, 44: 24, 38: 24, 49: 24, 50: 24, 19: 24, 36: 24, 51: 24, 47: 24, 46: 24, 32: 24, 34: 24, 45: 24, 43: 24, 27: 24 },
    { 13: 22, 18: 23 },
    { 24: 13 },
    { 34: 24, 45: 24, 41: 24, 27: 24, 51: 24, 36: 36, 52: 24, 48: 24, 47: 24, 50: 24, 39: 24, 46: 24, 38: 24, 44: 24, 42: 24, 49: 24, 19: 24, 40: 24, 32: 24, 37: 24, 43: 24, 35: 24 },
    { },
    { 47: 34, 41: 34, 45: 34, 20: 34, 40: 34, 51: 34, 39: 34, 11: 34, 27: 34, 7: 34, 4: 34, 44: 34, 21: 34, 42: 34, 2: 34, 50: 34, 34: 34, 14: 34, 19: 34, 26: 34, 36: 34, 8: 34, 55: 34, 33: 34, 35: 34, 37: 34, 10: 34, 17: 34, 9: 16, 22: 34, 16: 34, 12: 34, 24: 34, 31: 34, 49: 34, 28: 34, 32: 34, 53: 34, 43: 34, 18: 34, 30: 34, 48: 34, 25: 34, 29: 21, 46: 34, 1: 34, 13: 34, 54: 34, 23: 34, 6: 34, 52: 34, 38: 34, 15: 34 },
    { 43: 24, 51: 24, 47: 24, 42: 24, 49: 24, 40: 24, 45: 24, 38: 24, 50: 24, 37: 24, 44: 24, 46: 24, 39: 24, 32: 24, 27: 24, 36: 24, 19: 24, 35: 24, 41: 24, 52: 24, 34: 25, 48: 24 },
    { 44: 24, 40: 24, 39: 24, 42: 24, 36: 24, 41: 24, 52: 24, 32: 24, 45: 42, 27: 24, 47: 24, 34: 24, 49: 24, 37: 24, 43: 24, 46: 24, 51: 24, 35: 24, 50: 24, 48: 24, 38: 24, 19: 24 },
    { 36: 24, 39: 24, 35: 24, 19: 24, 32: 24, 50: 24, 43: 24, 34: 24, 45: 38, 49: 24, 40: 24, 48: 24, 41: 24, 44: 24, 47: 24, 46: 24, 51: 24, 52: 24, 27: 24, 38: 24, 42: 24, 37: 24 },
    { 36: 24, 44: 24, 34: 24, 39: 24, 49: 24, 51: 24, 41: 24, 46: 24, 47: 24, 27: 24, 52: 24, 50: 24, 48: 24, 45: 24, 38: 24, 40: 24, 42: 24, 32: 24, 43: 24, 35: 24, 37: 24, 19: 24 },
    { },
    { },
    { 46: 24, 48: 24, 37: 24, 19: 24, 32: 24, 27: 24, 51: 24, 45: 24, 36: 24, 44: 24, 34: 24, 42: 24, 52: 24, 38: 24, 35: 24, 40: 24, 41: 24, 47: 24, 49: 35, 50: 24, 43: 24, 39: 24 },
    { 27: 24, 40: 24, 41: 24, 42: 24, 38: 24, 39: 24, 43: 24, 36: 24, 47: 24, 50: 24, 52: 24, 19: 24, 48: 24, 51: 5, 37: 24, 46: 24, 45: 24, 32: 24, 44: 24, 34: 24, 49: 24, 35: 24 },
    { 40: 24, 47: 24, 42: 27, 45: 24, 19: 24, 38: 24, 44: 24, 43: 24, 51: 24, 27: 24, 34: 24, 39: 24, 48: 24, 35: 24, 46: 24, 50: 24, 37: 24, 36: 24, 52: 24, 49: 24, 41: 24, 32: 24 },
}
var accept = map[int]TokenType { 4: IDENTIFIER, 5: FRAGMENT, 9: WHITESPACE, 13: ARROW, 19: IDENTIFIER, 25: IDENTIFIER, 35: IDENTIFIER, 2: EOF, 15: EQUAL, 33: COLON, 43: IDENTIFIER, 6: DOT, 12: COMMENT, 18: R_PAREN, 28: QUESTION, 1: CLASS, 14: PLUS, 16: STRING, 3: L_PAREN, 7: IDENTIFIER, 27: IDENTIFIER, 42: IDENTIFIER, 26: IDENTIFIER, 32: IDENTIFIER, 36: IDENTIFIER, 38: TOKEN, 39: BAR, 20: SEMI, 24: IDENTIFIER, 29: IDENTIFIER, 37: IDENTIFIER, 40: STAR, 8: SKIP, 41: IDENTIFIER }

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
type ErrorHandler func (stream *InputStream)
var DEFAULT_HANDLER = func (stream *InputStream) {
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
            unexpected(input[i], l.stream.location)
            l.stream.synchronize(l.handler)
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
func (i *InputStream) synchronize(handler ErrorHandler) {
    i.reset()
    handler(i)
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

func unexpected(char rune, location Location) {
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
}

// FOR DEBUG PURPOSES:
// Consumes all tokens emitted by lexer and prints them to the standard output.
func (l *Lexer) PrintTokenStream() {
    for l.Token.Type != EOF {
        fmt.Printf("%-7s | %-16s %-16s\n", fmt.Sprintf("%d:%d", l.Token.Location.Line, l.Token.Location.Col), l.Token.Type, l.Token.Value)
        l.Next()
    }
}
