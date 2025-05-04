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

const (WHITESPACE TokenType = iota; COMMENT; TOKEN; FRAGMENT; SKIP; EQUAL; PLUS; STAR; QUESTION; DOT; BAR; SEMI; COLON; L_PAREN; R_PAREN; ARROW; IDENTIFIER; STRING; CLASS; EOF)
func (t TokenType) String() string { return typeName[t] }
var typeName = map[TokenType]string { WHITESPACE: "WHITESPACE", COMMENT: "COMMENT", TOKEN: "TOKEN", FRAGMENT: "FRAGMENT", SKIP: "SKIP", EQUAL: "EQUAL", PLUS: "PLUS", STAR: "STAR", QUESTION: "QUESTION", DOT: "DOT", BAR: "BAR", SEMI: "SEMI", COLON: "COLON", L_PAREN: "L_PAREN", R_PAREN: "R_PAREN", ARROW: "ARROW", IDENTIFIER: "IDENTIFIER", STRING: "STRING", CLASS: "CLASS", EOF: "EOF" }
var skip = map[TokenType]struct{} { WHITESPACE: {}, COMMENT: {} }

var ranges = []Range { { '\x00', '\x00' }, { '\x01', '\b' }, { '\t', '\t' }, { '\n', '\n' }, { '\v', '\f' }, { '\r', '\r' }, { '\x0e', '\x1f' }, { ' ', ' ' }, { '!', '!' }, { '"', '"' }, { '#', '\'' }, { '(', '(' }, { ')', ')' }, { '*', '*' }, { '+', '+' }, { ',', ',' }, { '-', '-' }, { '.', '.' }, { '/', '/' }, { '0', '9' }, { ':', ':' }, { ';', ';' }, { '<', '<' }, { '=', '=' }, { '>', '>' }, { '?', '?' }, { '@', '@' }, { 'A', 'Z' }, { '[', '[' }, { '\\', '\\' }, { ']', ']' }, { '^', '^' }, { '_', '_' }, { '`', '`' }, { 'a', 'a' }, { 'b', 'd' }, { 'e', 'e' }, { 'f', 'f' }, { 'g', 'g' }, { 'h', 'h' }, { 'i', 'i' }, { 'j', 'j' }, { 'k', 'k' }, { 'l', 'l' }, { 'm', 'm' }, { 'n', 'n' }, { 'o', 'o' }, { 'p', 'p' }, { 'q', 'q' }, { 'r', 'r' }, { 's', 's' }, { 't', 't' }, { 'u', 'z' }, { '{', '{' }, { '|', '|' }, { '}', '\U0010ffff' } }
var transitions = []map[int]int {
    { 12: 20, 3: 10, 25: 34, 16: 28, 5: 10, 40: 25, 46: 25, 35: 25, 20: 6, 49: 25, 38: 25, 17: 14, 45: 25, 39: 25, 11: 7, 36: 25, 48: 25, 23: 43, 0: 35, 43: 25, 52: 25, 21: 15, 50: 38, 42: 25, 2: 10, 27: 25, 51: 1, 28: 39, 47: 25, 41: 25, 7: 10, 9: 22, 32: 25, 34: 25, 13: 23, 54: 40, 14: 31, 18: 24, 44: 25, 37: 41 },
    { 50: 25, 32: 25, 46: 21, 37: 25, 39: 25, 44: 25, 52: 25, 19: 25, 36: 25, 49: 25, 34: 25, 45: 25, 41: 25, 35: 25, 48: 25, 40: 25, 27: 25, 51: 25, 43: 25, 38: 25, 42: 25, 47: 25 },
    { 38: 39, 29: 39, 44: 39, 13: 39, 14: 39, 47: 39, 43: 39, 36: 39, 34: 39, 48: 39, 27: 39, 11: 39, 42: 39, 50: 39, 8: 39, 16: 39, 32: 39, 30: 39, 4: 39, 17: 39, 45: 39, 35: 39, 12: 39, 51: 39, 6: 39, 2: 39, 15: 39, 1: 39, 18: 39, 41: 39, 26: 39, 19: 39, 52: 39, 39: 39, 10: 39, 28: 39, 31: 39, 46: 39, 24: 39, 22: 39, 55: 39, 9: 39, 23: 39, 49: 39, 40: 39, 37: 39, 7: 39, 53: 39, 54: 39, 21: 39, 33: 39, 20: 39, 25: 39 },
    { 31: 3, 52: 3, 23: 3, 47: 3, 10: 3, 43: 3, 11: 3, 1: 3, 0: 19, 5: 19, 12: 3, 45: 3, 33: 3, 27: 3, 9: 3, 17: 3, 53: 3, 3: 19, 51: 3, 42: 3, 54: 3, 15: 3, 7: 3, 26: 3, 16: 3, 50: 3, 6: 3, 24: 3, 39: 3, 13: 3, 21: 3, 25: 3, 48: 3, 37: 3, 49: 3, 2: 3, 41: 3, 34: 3, 19: 3, 46: 3, 55: 3, 44: 3, 35: 3, 18: 3, 29: 3, 4: 3, 38: 3, 32: 3, 36: 3, 40: 3, 30: 3, 8: 3, 28: 3, 14: 3, 20: 3, 22: 3 },
    { 40: 25, 50: 25, 46: 25, 52: 25, 38: 9, 44: 25, 41: 25, 19: 25, 34: 25, 42: 25, 51: 25, 47: 25, 35: 25, 27: 25, 36: 25, 49: 25, 32: 25, 45: 25, 39: 25, 48: 25, 43: 25, 37: 25 },
    { 48: 25, 39: 25, 41: 25, 42: 25, 47: 25, 50: 25, 40: 25, 19: 25, 38: 25, 46: 25, 34: 25, 37: 25, 43: 25, 44: 25, 45: 27, 32: 25, 36: 25, 35: 25, 27: 25, 49: 25, 52: 25, 51: 25 },
    { },
    { },
    { 33: 8, 22: 8, 16: 8, 41: 8, 13: 32, 32: 8, 8: 8, 39: 8, 17: 8, 31: 8, 21: 8, 2: 8, 52: 8, 47: 8, 55: 8, 4: 8, 20: 8, 53: 8, 38: 8, 44: 8, 40: 8, 5: 8, 23: 8, 27: 8, 24: 8, 29: 8, 12: 8, 34: 8, 30: 8, 50: 8, 36: 8, 45: 8, 7: 8, 26: 8, 43: 8, 51: 8, 48: 8, 35: 8, 10: 8, 19: 8, 49: 8, 15: 8, 37: 8, 1: 8, 25: 8, 46: 8, 9: 8, 6: 8, 18: 8, 28: 8, 3: 8, 14: 8, 42: 8, 54: 8, 11: 8 },
    { 38: 25, 34: 25, 45: 25, 51: 25, 49: 25, 36: 25, 46: 25, 48: 25, 19: 25, 52: 25, 32: 25, 37: 25, 40: 25, 43: 25, 47: 25, 27: 25, 41: 25, 35: 25, 50: 25, 39: 25, 42: 25, 44: 26 },
    { 3: 10, 5: 10, 7: 10, 2: 10 },
    { 43: 25, 49: 25, 42: 25, 32: 25, 41: 25, 44: 25, 45: 25, 47: 25, 27: 25, 40: 25, 48: 25, 38: 25, 52: 25, 50: 25, 19: 25, 51: 25, 39: 25, 36: 25, 46: 25, 34: 25, 35: 25, 37: 25 },
    { },
    { 35: 25, 27: 25, 40: 25, 19: 25, 47: 25, 36: 25, 43: 25, 38: 25, 41: 25, 45: 25, 51: 25, 32: 25, 50: 25, 42: 25, 44: 25, 46: 25, 49: 25, 34: 25, 39: 25, 37: 25, 48: 25, 52: 25 },
    { },
    { },
    { 38: 25, 39: 25, 52: 25, 32: 25, 37: 25, 47: 29, 41: 25, 27: 25, 34: 25, 45: 25, 40: 25, 36: 25, 35: 25, 43: 25, 19: 25, 42: 25, 50: 25, 48: 25, 46: 25, 51: 25, 49: 25, 44: 25 },
    { 7: 22, 26: 22, 24: 22, 1: 22, 18: 22, 35: 22, 22: 22, 48: 22, 20: 22, 25: 22, 11: 22, 14: 22, 28: 22, 42: 22, 53: 22, 55: 22, 16: 22, 46: 22, 45: 22, 13: 22, 51: 22, 8: 22, 10: 22, 27: 22, 12: 22, 6: 22, 15: 22, 19: 22, 30: 22, 41: 22, 44: 22, 37: 22, 9: 22, 31: 22, 2: 22, 39: 22, 49: 22, 50: 22, 38: 22, 29: 22, 54: 22, 47: 22, 21: 22, 17: 22, 40: 22, 23: 22, 32: 22, 36: 22, 43: 22, 34: 22, 33: 22, 52: 22, 4: 22 },
    { },
    { },
    { },
    { 52: 25, 36: 25, 27: 25, 43: 25, 19: 25, 50: 25, 38: 25, 41: 25, 42: 37, 32: 25, 46: 25, 45: 25, 49: 25, 35: 25, 44: 25, 47: 25, 39: 25, 51: 25, 34: 25, 37: 25, 40: 25, 48: 25 },
    { 23: 22, 55: 22, 19: 22, 14: 22, 10: 22, 30: 22, 28: 22, 51: 22, 24: 22, 4: 22, 33: 22, 17: 22, 22: 22, 38: 22, 15: 22, 54: 22, 7: 22, 45: 22, 20: 22, 11: 22, 1: 22, 42: 22, 36: 22, 26: 22, 18: 22, 47: 22, 27: 22, 49: 22, 40: 22, 48: 22, 9: 18, 44: 22, 50: 22, 35: 22, 39: 22, 37: 22, 43: 22, 13: 22, 46: 22, 12: 22, 8: 22, 52: 22, 2: 22, 25: 22, 29: 17, 41: 22, 16: 22, 31: 22, 53: 22, 6: 22, 21: 22, 34: 22, 32: 22 },
    { },
    { 18: 3, 13: 8 },
    { 32: 25, 44: 25, 48: 25, 46: 25, 41: 25, 42: 25, 38: 25, 45: 25, 49: 25, 52: 25, 19: 25, 27: 25, 43: 25, 34: 25, 50: 25, 40: 25, 37: 25, 47: 25, 36: 25, 51: 25, 39: 25, 35: 25 },
    { 36: 5, 41: 25, 40: 25, 49: 25, 46: 25, 37: 25, 27: 25, 38: 25, 52: 25, 34: 25, 50: 25, 48: 25, 35: 25, 44: 25, 47: 25, 19: 25, 39: 25, 51: 25, 45: 25, 32: 25, 43: 25, 42: 25 },
    { 32: 25, 19: 25, 42: 25, 44: 25, 43: 25, 48: 25, 47: 25, 52: 25, 35: 25, 37: 25, 40: 25, 45: 25, 49: 25, 27: 25, 39: 25, 50: 25, 36: 25, 38: 25, 41: 25, 46: 25, 51: 13, 34: 25 },
    { 24: 42 },
    { 46: 25, 50: 25, 32: 25, 39: 25, 44: 25, 42: 25, 40: 25, 43: 25, 38: 25, 45: 25, 36: 25, 37: 25, 41: 25, 34: 25, 51: 25, 52: 25, 19: 25, 47: 25, 35: 25, 49: 25, 27: 25, 48: 25 },
    { 50: 25, 32: 25, 45: 11, 44: 25, 35: 25, 48: 25, 43: 25, 38: 25, 49: 25, 47: 25, 36: 25, 19: 25, 39: 25, 41: 25, 27: 25, 42: 25, 40: 25, 51: 25, 52: 25, 37: 25, 46: 25, 34: 25 },
    { },
    { 29: 8, 48: 8, 10: 8, 14: 8, 54: 8, 40: 8, 42: 8, 31: 8, 51: 8, 28: 8, 39: 8, 9: 8, 37: 8, 25: 8, 19: 8, 47: 8, 16: 8, 23: 8, 44: 8, 3: 8, 34: 8, 2: 8, 49: 8, 46: 8, 53: 8, 36: 8, 30: 8, 26: 8, 17: 8, 24: 8, 22: 8, 33: 8, 6: 8, 15: 8, 4: 8, 1: 8, 35: 8, 55: 8, 27: 8, 20: 8, 13: 8, 50: 8, 43: 8, 21: 8, 45: 8, 38: 8, 18: 19, 5: 8, 12: 8, 52: 8, 32: 8, 11: 8, 7: 8, 41: 8, 8: 8 },
    { 50: 25, 49: 25, 47: 25, 52: 25, 34: 4, 46: 25, 19: 25, 51: 25, 37: 25, 41: 25, 45: 25, 27: 25, 44: 25, 36: 25, 35: 25, 38: 25, 42: 25, 32: 25, 40: 25, 43: 25, 39: 25, 48: 25 },
    { },
    { },
    { 51: 25, 19: 25, 38: 25, 43: 25, 35: 25, 42: 25, 46: 25, 32: 25, 40: 16, 27: 25, 50: 25, 34: 25, 37: 25, 41: 25, 44: 25, 47: 25, 48: 25, 52: 25, 39: 25, 45: 25, 49: 25, 36: 25 },
    { 32: 25, 50: 25, 40: 25, 36: 30, 51: 25, 45: 25, 37: 25, 38: 25, 41: 25, 44: 25, 39: 25, 19: 25, 46: 25, 43: 25, 48: 25, 47: 25, 42: 25, 35: 25, 27: 25, 34: 25, 49: 25, 52: 25 },
    { 32: 25, 43: 25, 27: 25, 19: 25, 40: 25, 45: 25, 41: 25, 44: 25, 37: 25, 34: 25, 47: 25, 48: 25, 38: 25, 52: 25, 42: 36, 46: 25, 36: 25, 39: 25, 35: 25, 51: 25, 49: 25, 50: 25 },
    { 44: 39, 20: 39, 43: 39, 11: 39, 35: 39, 29: 2, 52: 39, 10: 39, 23: 39, 25: 39, 32: 39, 48: 39, 14: 39, 46: 39, 12: 39, 9: 39, 39: 39, 19: 39, 6: 39, 47: 39, 26: 39, 42: 39, 24: 39, 37: 39, 28: 39, 55: 39, 36: 39, 45: 39, 54: 39, 38: 39, 40: 39, 15: 39, 27: 39, 41: 39, 50: 39, 51: 39, 1: 39, 16: 39, 18: 39, 7: 39, 33: 39, 17: 39, 22: 39, 2: 39, 4: 39, 8: 39, 13: 39, 49: 39, 21: 39, 31: 39, 30: 12, 34: 39, 53: 39 },
    { },
    { 27: 25, 32: 25, 19: 25, 52: 25, 44: 25, 48: 25, 34: 25, 50: 25, 37: 25, 43: 25, 41: 25, 42: 25, 35: 25, 39: 25, 51: 25, 40: 25, 46: 25, 49: 33, 47: 25, 36: 25, 38: 25, 45: 25 },
    { },
    { },
}
var accept = map[int]TokenType { 30: IDENTIFIER, 15: SEMI, 19: COMMENT, 29: SKIP, 36: IDENTIFIER, 6: COLON, 9: IDENTIFIER, 12: CLASS, 21: IDENTIFIER, 34: QUESTION, 13: FRAGMENT, 41: IDENTIFIER, 11: TOKEN, 26: IDENTIFIER, 27: IDENTIFIER, 33: IDENTIFIER, 40: BAR, 43: EQUAL, 7: L_PAREN, 16: IDENTIFIER, 20: R_PAREN, 37: IDENTIFIER, 1: IDENTIFIER, 4: IDENTIFIER, 18: STRING, 23: STAR, 25: IDENTIFIER, 31: PLUS, 35: EOF, 38: IDENTIFIER, 5: IDENTIFIER, 10: WHITESPACE, 14: DOT, 42: ARROW }

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
