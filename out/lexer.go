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
    { 25: 4, 50: 32, 17: 14, 46: 1, 45: 1, 48: 1, 2: 27, 3: 27, 0: 21, 20: 34, 37: 35, 36: 1, 52: 1, 42: 1, 13: 8, 23: 31, 41: 1, 5: 27, 43: 1, 38: 1, 35: 1, 11: 23, 39: 1, 47: 1, 27: 1, 7: 27, 44: 1, 34: 1, 14: 24, 54: 9, 40: 1, 16: 5, 12: 17, 49: 1, 9: 6, 28: 41, 32: 1, 51: 19, 18: 30, 21: 20 },
    { 41: 1, 44: 1, 37: 1, 50: 1, 45: 1, 32: 1, 19: 1, 42: 1, 27: 1, 51: 1, 48: 1, 43: 1, 36: 1, 38: 1, 39: 1, 47: 1, 49: 1, 52: 1, 40: 1, 46: 1, 34: 1, 35: 1 },
    { 35: 1, 37: 1, 39: 1, 44: 1, 40: 1, 45: 1, 27: 1, 42: 1, 47: 1, 51: 1, 32: 1, 43: 1, 41: 1, 36: 1, 19: 1, 50: 1, 49: 1, 38: 1, 46: 1, 52: 1, 34: 1, 48: 1 },
    { 31: 3, 7: 3, 21: 3, 14: 3, 22: 3, 8: 3, 52: 3, 51: 3, 15: 3, 44: 3, 27: 3, 36: 3, 47: 3, 41: 3, 34: 3, 48: 3, 39: 3, 30: 3, 5: 38, 0: 38, 43: 3, 42: 3, 2: 3, 24: 3, 55: 3, 16: 3, 53: 3, 17: 3, 28: 3, 33: 3, 35: 3, 26: 3, 4: 3, 23: 3, 12: 3, 29: 3, 25: 3, 50: 3, 9: 3, 11: 3, 20: 3, 10: 3, 38: 3, 46: 3, 1: 3, 3: 38, 40: 3, 54: 3, 13: 3, 19: 3, 45: 3, 37: 3, 6: 3, 18: 3, 49: 3, 32: 3 },
    { },
    { 24: 16 },
    { 38: 6, 49: 6, 24: 6, 52: 6, 43: 6, 50: 6, 45: 6, 27: 6, 47: 6, 46: 6, 13: 6, 10: 6, 14: 6, 9: 7, 53: 6, 1: 6, 48: 6, 6: 6, 22: 6, 42: 6, 18: 6, 34: 6, 11: 6, 4: 6, 26: 6, 33: 6, 32: 6, 29: 12, 7: 6, 28: 6, 15: 6, 16: 6, 30: 6, 36: 6, 2: 6, 35: 6, 54: 6, 31: 6, 39: 6, 8: 6, 37: 6, 55: 6, 51: 6, 41: 6, 23: 6, 19: 6, 21: 6, 44: 6, 17: 6, 25: 6, 20: 6, 12: 6, 40: 6 },
    { },
    { },
    { },
    { 48: 1, 38: 1, 49: 1, 34: 1, 47: 33, 42: 1, 50: 1, 41: 1, 32: 1, 44: 1, 43: 1, 37: 1, 19: 1, 36: 1, 39: 1, 40: 1, 51: 1, 27: 1, 52: 1, 35: 1, 46: 1, 45: 1 },
    { 49: 1, 34: 1, 36: 15, 44: 1, 41: 1, 37: 1, 38: 1, 42: 1, 32: 1, 35: 1, 51: 1, 50: 1, 39: 1, 27: 1, 43: 1, 45: 1, 47: 1, 52: 1, 19: 1, 40: 1, 46: 1, 48: 1 },
    { 18: 6, 7: 6, 10: 6, 21: 6, 34: 6, 11: 6, 38: 6, 2: 6, 17: 6, 47: 6, 51: 6, 44: 6, 14: 6, 29: 6, 37: 6, 25: 6, 55: 6, 52: 6, 22: 6, 50: 6, 32: 6, 28: 6, 31: 6, 16: 6, 4: 6, 26: 6, 49: 6, 43: 6, 35: 6, 6: 6, 23: 6, 15: 6, 13: 6, 40: 6, 42: 6, 12: 6, 9: 6, 39: 6, 36: 6, 53: 6, 41: 6, 48: 6, 46: 6, 1: 6, 27: 6, 8: 6, 54: 6, 24: 6, 45: 6, 20: 6, 30: 6, 19: 6, 33: 6 },
    { 42: 43, 2: 43, 14: 43, 44: 43, 49: 43, 11: 43, 8: 43, 50: 43, 23: 43, 25: 43, 5: 43, 24: 43, 20: 43, 29: 43, 31: 43, 4: 43, 53: 43, 30: 43, 36: 43, 41: 43, 28: 43, 13: 43, 40: 43, 3: 43, 55: 43, 26: 43, 12: 43, 43: 43, 22: 43, 21: 43, 34: 43, 45: 43, 54: 43, 1: 43, 32: 43, 18: 38, 46: 43, 33: 43, 27: 43, 15: 43, 17: 43, 6: 43, 51: 43, 47: 43, 38: 43, 7: 43, 35: 43, 39: 43, 16: 43, 9: 43, 37: 43, 10: 43, 48: 43, 52: 43, 19: 43 },
    { },
    { 38: 1, 50: 1, 39: 1, 49: 1, 51: 1, 45: 22, 48: 1, 42: 1, 43: 1, 35: 1, 32: 1, 27: 1, 52: 1, 34: 1, 19: 1, 41: 1, 36: 1, 37: 1, 46: 1, 47: 1, 44: 1, 40: 1 },
    { },
    { },
    { 49: 41, 11: 41, 37: 41, 46: 41, 30: 41, 28: 41, 19: 41, 53: 41, 47: 41, 21: 41, 41: 41, 27: 41, 9: 41, 23: 41, 48: 41, 17: 41, 54: 41, 22: 41, 39: 41, 14: 41, 15: 41, 40: 41, 52: 41, 35: 41, 20: 41, 2: 41, 18: 41, 38: 41, 42: 41, 33: 41, 7: 41, 16: 41, 29: 41, 4: 41, 50: 41, 6: 41, 36: 41, 31: 41, 51: 41, 25: 41, 10: 41, 12: 41, 26: 41, 1: 41, 43: 41, 34: 41, 55: 41, 24: 41, 45: 41, 13: 41, 44: 41, 32: 41, 8: 41 },
    { 39: 1, 49: 1, 51: 1, 41: 1, 32: 1, 44: 1, 37: 1, 40: 1, 35: 1, 27: 1, 52: 1, 48: 1, 38: 1, 46: 42, 36: 1, 42: 1, 50: 1, 34: 1, 47: 1, 19: 1, 45: 1, 43: 1 },
    { },
    { },
    { 41: 1, 32: 1, 47: 1, 43: 1, 42: 1, 38: 1, 40: 1, 34: 1, 49: 1, 51: 28, 48: 1, 45: 1, 35: 1, 52: 1, 27: 1, 44: 1, 36: 1, 46: 1, 50: 1, 39: 1, 37: 1, 19: 1 },
    { },
    { },
    { 47: 1, 37: 1, 43: 1, 39: 1, 45: 1, 41: 1, 48: 1, 35: 1, 50: 1, 51: 1, 49: 1, 40: 1, 46: 1, 27: 1, 42: 1, 36: 29, 19: 1, 38: 1, 34: 1, 32: 1, 52: 1, 44: 1 },
    { 45: 1, 44: 1, 38: 1, 50: 1, 40: 10, 35: 1, 47: 1, 49: 1, 41: 1, 27: 1, 36: 1, 52: 1, 37: 1, 39: 1, 46: 1, 32: 1, 51: 1, 42: 1, 43: 1, 34: 1, 19: 1, 48: 1 },
    { 7: 27, 2: 27, 3: 27, 5: 27 },
    { 36: 1, 46: 1, 32: 1, 38: 1, 35: 1, 51: 1, 52: 1, 41: 1, 27: 1, 37: 1, 40: 1, 50: 1, 42: 1, 45: 1, 43: 1, 47: 1, 19: 1, 34: 1, 48: 1, 49: 1, 44: 1, 39: 1 },
    { 19: 1, 43: 1, 39: 1, 47: 1, 42: 1, 48: 1, 36: 1, 32: 1, 49: 1, 38: 1, 46: 1, 34: 1, 37: 1, 52: 1, 41: 1, 51: 1, 35: 1, 44: 1, 40: 1, 27: 1, 45: 2, 50: 1 },
    { 13: 43, 18: 3 },
    { },
    { 51: 1, 43: 1, 47: 1, 48: 1, 44: 1, 52: 1, 46: 1, 39: 1, 41: 1, 36: 1, 40: 1, 35: 1, 27: 1, 37: 1, 38: 1, 49: 1, 45: 1, 50: 1, 34: 1, 32: 1, 19: 1, 42: 26 },
    { 32: 1, 27: 1, 34: 1, 47: 1, 36: 1, 43: 1, 50: 1, 39: 1, 45: 1, 37: 1, 44: 1, 49: 1, 42: 1, 19: 1, 41: 1, 48: 1, 38: 1, 46: 1, 52: 1, 40: 1, 51: 1, 35: 1 },
    { },
    { 35: 1, 19: 1, 34: 1, 42: 1, 48: 1, 27: 1, 41: 1, 37: 1, 39: 1, 46: 1, 50: 1, 40: 1, 43: 1, 32: 1, 38: 1, 36: 1, 49: 39, 52: 1, 45: 1, 47: 1, 44: 1, 51: 1 },
    { 34: 1, 42: 1, 47: 1, 44: 11, 37: 1, 36: 1, 27: 1, 19: 1, 40: 1, 51: 1, 41: 1, 38: 1, 43: 1, 49: 1, 52: 1, 32: 1, 46: 1, 45: 1, 48: 1, 35: 1, 50: 1, 39: 1 },
    { },
    { },
    { 48: 1, 32: 1, 44: 1, 49: 1, 34: 40, 37: 1, 46: 1, 50: 1, 41: 1, 43: 1, 35: 1, 47: 1, 45: 1, 19: 1, 51: 1, 36: 1, 38: 1, 52: 1, 42: 1, 27: 1, 40: 1, 39: 1 },
    { 50: 1, 32: 1, 34: 1, 42: 1, 40: 1, 19: 1, 43: 1, 44: 1, 52: 1, 36: 1, 46: 1, 45: 1, 35: 1, 49: 1, 51: 1, 37: 1, 38: 36, 39: 1, 48: 1, 41: 1, 27: 1, 47: 1 },
    { 16: 41, 25: 41, 52: 41, 23: 41, 20: 41, 37: 41, 39: 41, 30: 37, 29: 18, 10: 41, 44: 41, 55: 41, 24: 41, 28: 41, 12: 41, 35: 41, 17: 41, 7: 41, 22: 41, 19: 41, 1: 41, 4: 41, 27: 41, 6: 41, 40: 41, 8: 41, 26: 41, 49: 41, 54: 41, 45: 41, 38: 41, 42: 41, 36: 41, 14: 41, 13: 41, 41: 41, 46: 41, 21: 41, 11: 41, 51: 41, 50: 41, 18: 41, 2: 41, 48: 41, 9: 41, 32: 41, 34: 41, 43: 41, 33: 41, 53: 41, 15: 41, 47: 41, 31: 41 },
    { 45: 1, 36: 1, 38: 1, 35: 1, 34: 1, 39: 1, 49: 1, 37: 1, 43: 1, 41: 1, 44: 1, 40: 1, 50: 1, 51: 1, 47: 1, 19: 1, 32: 1, 52: 1, 42: 25, 27: 1, 48: 1, 46: 1 },
    { 27: 43, 33: 43, 24: 43, 26: 43, 32: 43, 12: 43, 31: 43, 13: 13, 37: 43, 34: 43, 5: 43, 40: 43, 29: 43, 1: 43, 48: 43, 39: 43, 11: 43, 36: 43, 46: 43, 23: 43, 55: 43, 14: 43, 45: 43, 4: 43, 19: 43, 16: 43, 6: 43, 53: 43, 51: 43, 8: 43, 15: 43, 43: 43, 42: 43, 2: 43, 3: 43, 52: 43, 30: 43, 35: 43, 9: 43, 38: 43, 28: 43, 18: 43, 49: 43, 17: 43, 44: 43, 10: 43, 25: 43, 21: 43, 54: 43, 41: 43, 47: 43, 7: 43, 22: 43, 50: 43, 20: 43 },
}
var accept = map[int]TokenType { 26: IDENTIFIER, 39: IDENTIFIER, 16: ARROW, 19: IDENTIFIER, 20: SEMI, 21: EOF, 34: COLON, 37: CLASS, 1: IDENTIFIER, 4: QUESTION, 8: STAR, 10: IDENTIFIER, 23: L_PAREN, 25: IDENTIFIER, 28: FRAGMENT, 35: IDENTIFIER, 11: IDENTIFIER, 40: IDENTIFIER, 9: BAR, 27: WHITESPACE, 32: IDENTIFIER, 36: IDENTIFIER, 38: COMMENT, 2: TOKEN, 31: EQUAL, 14: DOT, 33: SKIP, 15: IDENTIFIER, 29: IDENTIFIER, 42: IDENTIFIER, 7: STRING, 17: R_PAREN, 22: IDENTIFIER, 24: PLUS }

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
