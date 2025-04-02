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
    { 40: 22, 43: 22, 38: 22, 2: 13, 5: 13, 48: 22, 46: 22, 13: 21, 18: 27, 41: 22, 25: 1, 37: 9, 12: 19, 39: 22, 45: 22, 47: 22, 42: 22, 23: 34, 7: 13, 27: 22, 35: 22, 21: 6, 51: 3, 11: 24, 0: 41, 16: 30, 49: 22, 3: 13, 9: 25, 36: 22, 54: 15, 17: 7, 34: 22, 20: 5, 32: 22, 50: 26, 14: 43, 52: 22, 44: 22, 28: 20 },
    { },
    { 46: 22, 48: 22, 49: 22, 34: 22, 45: 22, 19: 22, 44: 22, 47: 22, 27: 22, 52: 22, 37: 22, 42: 22, 43: 22, 41: 22, 40: 22, 35: 22, 38: 22, 51: 22, 39: 22, 36: 22, 32: 22, 50: 22 },
    { 42: 22, 47: 22, 49: 22, 19: 22, 37: 22, 36: 22, 32: 22, 39: 22, 52: 22, 41: 22, 27: 22, 45: 22, 44: 22, 38: 22, 34: 22, 43: 22, 50: 22, 51: 22, 35: 22, 40: 22, 46: 35, 48: 22 },
    { },
    { },
    { },
    { },
    { 26: 20, 13: 20, 28: 20, 45: 20, 53: 20, 32: 20, 49: 20, 31: 20, 20: 20, 7: 20, 1: 20, 24: 20, 18: 20, 29: 20, 40: 20, 54: 20, 34: 20, 55: 20, 41: 20, 9: 20, 16: 20, 37: 20, 39: 20, 12: 20, 27: 20, 30: 20, 43: 20, 48: 20, 33: 20, 51: 20, 47: 20, 50: 20, 23: 20, 11: 20, 46: 20, 4: 20, 14: 20, 35: 20, 44: 20, 42: 20, 17: 20, 36: 20, 19: 20, 22: 20, 15: 20, 10: 20, 21: 20, 2: 20, 38: 20, 52: 20, 6: 20, 25: 20, 8: 20 },
    { 19: 22, 47: 22, 40: 22, 49: 29, 43: 22, 38: 22, 37: 22, 50: 22, 44: 22, 45: 22, 42: 22, 41: 22, 46: 22, 34: 22, 39: 22, 36: 22, 51: 22, 27: 22, 52: 22, 48: 22, 32: 22, 35: 22 },
    { 42: 22, 44: 22, 34: 22, 41: 22, 32: 22, 37: 22, 46: 22, 19: 22, 39: 22, 49: 22, 51: 22, 52: 22, 35: 22, 36: 22, 47: 22, 40: 22, 38: 17, 43: 22, 45: 22, 27: 22, 48: 22, 50: 22 },
    { 19: 22, 35: 22, 51: 22, 48: 22, 52: 22, 37: 22, 47: 22, 45: 22, 38: 22, 34: 22, 43: 22, 36: 18, 42: 22, 39: 22, 41: 22, 40: 22, 46: 22, 44: 22, 50: 22, 49: 22, 32: 22, 27: 22 },
    { 44: 22, 47: 22, 19: 22, 27: 22, 41: 22, 42: 22, 36: 22, 48: 22, 37: 22, 34: 22, 38: 22, 32: 22, 43: 22, 35: 22, 45: 22, 52: 22, 39: 22, 40: 22, 46: 22, 49: 22, 51: 2, 50: 22 },
    { 2: 13, 3: 13, 5: 13, 7: 13 },
    { },
    { },
    { 24: 16, 47: 16, 3: 32, 45: 16, 53: 16, 36: 16, 28: 16, 40: 16, 1: 16, 51: 16, 5: 32, 31: 16, 26: 16, 34: 16, 39: 16, 49: 16, 41: 16, 29: 16, 4: 16, 54: 16, 17: 16, 14: 16, 18: 16, 30: 16, 20: 16, 6: 16, 52: 16, 8: 16, 0: 32, 9: 16, 27: 16, 35: 16, 7: 16, 46: 16, 10: 16, 32: 16, 38: 16, 55: 16, 2: 16, 25: 16, 50: 16, 11: 16, 21: 16, 43: 16, 33: 16, 22: 16, 16: 16, 37: 16, 44: 16, 19: 16, 15: 16, 12: 16, 23: 16, 48: 16, 13: 16, 42: 16 },
    { 32: 22, 35: 22, 39: 22, 19: 22, 37: 22, 52: 22, 46: 22, 45: 22, 44: 11, 36: 22, 51: 22, 49: 22, 47: 22, 27: 22, 40: 22, 42: 22, 38: 22, 48: 22, 34: 22, 41: 22, 43: 22, 50: 22 },
    { 37: 22, 51: 22, 50: 22, 39: 22, 32: 22, 38: 22, 48: 22, 43: 22, 36: 22, 42: 22, 35: 22, 49: 22, 47: 22, 34: 22, 40: 22, 52: 22, 41: 22, 46: 22, 45: 12, 27: 22, 44: 22, 19: 22 },
    { },
    { 2: 20, 38: 20, 41: 20, 11: 20, 55: 20, 50: 20, 29: 8, 31: 20, 54: 20, 24: 20, 48: 20, 52: 20, 1: 20, 30: 38, 45: 20, 42: 20, 23: 20, 15: 20, 32: 20, 22: 20, 10: 20, 35: 20, 16: 20, 34: 20, 43: 20, 40: 20, 14: 20, 46: 20, 28: 20, 12: 20, 44: 20, 39: 20, 53: 20, 4: 20, 33: 20, 51: 20, 20: 20, 8: 20, 27: 20, 18: 20, 21: 20, 7: 20, 13: 20, 26: 20, 19: 20, 49: 20, 6: 20, 17: 20, 47: 20, 9: 20, 36: 20, 37: 20, 25: 20 },
    { },
    { 46: 22, 42: 22, 49: 22, 35: 22, 36: 22, 48: 22, 39: 22, 34: 22, 50: 22, 41: 22, 47: 22, 44: 22, 37: 22, 52: 22, 43: 22, 19: 22, 45: 22, 38: 22, 27: 22, 51: 22, 40: 22, 32: 22 },
    { 28: 25, 36: 25, 4: 25, 38: 25, 30: 25, 50: 25, 39: 25, 35: 25, 22: 25, 7: 25, 21: 25, 43: 25, 10: 25, 55: 25, 19: 25, 8: 25, 11: 25, 1: 25, 34: 25, 25: 25, 33: 25, 27: 25, 2: 25, 31: 25, 15: 25, 45: 25, 12: 25, 6: 25, 48: 25, 14: 25, 26: 25, 46: 25, 54: 25, 52: 25, 40: 25, 49: 25, 29: 25, 32: 25, 16: 25, 23: 25, 17: 25, 9: 25, 37: 25, 47: 25, 20: 25, 53: 25, 44: 25, 13: 25, 41: 25, 51: 25, 18: 25, 24: 25, 42: 25 },
    { },
    { 11: 25, 42: 25, 51: 25, 16: 25, 9: 14, 47: 25, 43: 25, 54: 25, 25: 25, 46: 25, 22: 25, 53: 25, 38: 25, 48: 25, 1: 25, 15: 25, 6: 25, 36: 25, 2: 25, 41: 25, 23: 25, 24: 25, 13: 25, 50: 25, 45: 25, 52: 25, 4: 25, 27: 25, 12: 25, 39: 25, 19: 25, 35: 25, 10: 25, 28: 25, 33: 25, 37: 25, 17: 25, 32: 25, 34: 25, 49: 25, 26: 25, 30: 25, 40: 25, 55: 25, 8: 25, 18: 25, 21: 25, 14: 25, 29: 23, 7: 25, 31: 25, 44: 25, 20: 25 },
    { 42: 42, 45: 22, 35: 22, 40: 22, 41: 22, 39: 22, 27: 22, 50: 22, 32: 22, 52: 22, 46: 22, 37: 22, 49: 22, 44: 22, 51: 22, 48: 22, 47: 22, 36: 22, 19: 22, 34: 22, 43: 22, 38: 22 },
    { 18: 16, 13: 33 },
    { 48: 33, 7: 33, 12: 33, 45: 33, 25: 33, 10: 33, 33: 33, 37: 33, 2: 33, 50: 33, 36: 33, 54: 33, 26: 33, 42: 33, 22: 33, 23: 33, 34: 33, 15: 33, 38: 33, 41: 33, 39: 33, 17: 33, 55: 33, 32: 33, 18: 32, 4: 33, 44: 33, 14: 33, 6: 33, 21: 33, 1: 33, 53: 33, 19: 33, 24: 33, 43: 33, 49: 33, 20: 33, 30: 33, 9: 33, 27: 33, 31: 33, 5: 33, 35: 33, 3: 33, 11: 33, 28: 33, 13: 33, 29: 33, 40: 33, 8: 33, 16: 33, 46: 33, 51: 33, 52: 33, 47: 33 },
    { 50: 22, 46: 22, 44: 22, 32: 22, 48: 22, 42: 22, 52: 22, 27: 22, 45: 22, 51: 22, 36: 22, 19: 22, 43: 22, 35: 22, 39: 22, 49: 22, 37: 22, 40: 22, 34: 10, 38: 22, 41: 22, 47: 22 },
    { 24: 4 },
    { 35: 22, 45: 22, 37: 22, 42: 22, 27: 22, 34: 22, 38: 22, 51: 22, 19: 22, 49: 22, 39: 22, 48: 22, 43: 22, 40: 22, 36: 22, 41: 22, 46: 22, 50: 22, 47: 37, 52: 22, 32: 22, 44: 22 },
    { },
    { 13: 28, 52: 33, 26: 33, 51: 33, 41: 33, 25: 33, 35: 33, 44: 33, 32: 33, 55: 33, 50: 33, 10: 33, 6: 33, 19: 33, 46: 33, 31: 33, 12: 33, 23: 33, 37: 33, 17: 33, 14: 33, 3: 33, 1: 33, 5: 33, 28: 33, 29: 33, 36: 33, 21: 33, 15: 33, 38: 33, 20: 33, 30: 33, 54: 33, 7: 33, 24: 33, 8: 33, 27: 33, 4: 33, 42: 33, 18: 33, 2: 33, 22: 33, 43: 33, 45: 33, 48: 33, 9: 33, 40: 33, 47: 33, 16: 33, 11: 33, 33: 33, 34: 33, 53: 33, 39: 33, 49: 33 },
    { },
    { 44: 22, 46: 22, 48: 22, 36: 22, 32: 22, 39: 22, 52: 22, 49: 22, 19: 22, 50: 22, 41: 22, 42: 36, 45: 22, 40: 22, 27: 22, 35: 22, 38: 22, 37: 22, 47: 22, 43: 22, 51: 22, 34: 22 },
    { 50: 22, 39: 22, 19: 22, 34: 22, 45: 22, 35: 22, 48: 22, 42: 22, 41: 22, 49: 22, 37: 22, 51: 22, 32: 22, 52: 22, 27: 22, 43: 22, 38: 22, 40: 22, 36: 39, 47: 22, 44: 22, 46: 22 },
    { 19: 22, 36: 22, 50: 22, 44: 22, 45: 22, 52: 22, 34: 22, 48: 22, 43: 22, 40: 22, 41: 22, 46: 22, 47: 22, 27: 22, 37: 22, 38: 22, 51: 22, 32: 22, 42: 22, 49: 22, 35: 22, 39: 22 },
    { },
    { 47: 22, 34: 22, 50: 22, 35: 22, 42: 22, 45: 40, 49: 22, 52: 22, 44: 22, 40: 22, 37: 22, 27: 22, 19: 22, 43: 22, 41: 22, 46: 22, 51: 22, 32: 22, 38: 22, 39: 22, 48: 22, 36: 22 },
    { 19: 22, 34: 22, 52: 22, 44: 22, 39: 22, 48: 22, 49: 22, 40: 22, 32: 22, 46: 22, 42: 22, 47: 22, 51: 22, 38: 22, 27: 22, 50: 22, 43: 22, 35: 22, 37: 22, 36: 22, 45: 22, 41: 22 },
    { },
    { 43: 22, 40: 31, 52: 22, 34: 22, 44: 22, 19: 22, 42: 22, 36: 22, 51: 22, 46: 22, 32: 22, 47: 22, 49: 22, 45: 22, 35: 22, 39: 22, 27: 22, 41: 22, 50: 22, 38: 22, 48: 22, 37: 22 },
    { },
}
var accept = map[int]TokenType { 1: QUESTION, 5: COLON, 10: IDENTIFIER, 11: IDENTIFIER, 18: IDENTIFIER, 24: L_PAREN, 35: IDENTIFIER, 37: SKIP, 2: FRAGMENT, 4: ARROW, 32: COMMENT, 39: IDENTIFIER, 40: TOKEN, 41: EOF, 13: WHITESPACE, 15: BAR, 19: R_PAREN, 34: EQUAL, 36: IDENTIFIER, 6: SEMI, 7: DOT, 12: IDENTIFIER, 22: IDENTIFIER, 26: IDENTIFIER, 38: CLASS, 17: IDENTIFIER, 29: IDENTIFIER, 3: IDENTIFIER, 42: IDENTIFIER, 21: STAR, 31: IDENTIFIER, 43: PLUS, 9: IDENTIFIER, 14: STRING }

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
