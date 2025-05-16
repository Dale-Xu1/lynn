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

const (WHITESPACE TokenType = iota; COMMENT; RULE; TOKEN; FRAGMENT; LEFT; RIGHT; SKIP; PLUS; STAR; QUESTION; DOT; BAR; HASH; SEMI; COLON; L_PAREN; R_PAREN; ARROW; IDENTIFIER; STRING; CLASS; EOF)
func (t TokenType) String() string { return typeName[t] }
var typeName = map[TokenType]string { WHITESPACE: "WHITESPACE", COMMENT: "COMMENT", RULE: "RULE", TOKEN: "TOKEN", FRAGMENT: "FRAGMENT", LEFT: "LEFT", RIGHT: "RIGHT", SKIP: "SKIP", PLUS: "PLUS", STAR: "STAR", QUESTION: "QUESTION", DOT: "DOT", BAR: "BAR", HASH: "HASH", SEMI: "SEMI", COLON: "COLON", L_PAREN: "L_PAREN", R_PAREN: "R_PAREN", ARROW: "ARROW", IDENTIFIER: "IDENTIFIER", STRING: "STRING", CLASS: "CLASS", EOF: "EOF" }
var skip = map[TokenType]struct{} { WHITESPACE: {}, COMMENT: {} }

var ranges = []Range { { '\x00', '\x00' }, { '\x01', '\b' }, { '\t', '\t' }, { '\n', '\n' }, { '\v', '\f' }, { '\r', '\r' }, { '\x0e', '\x1f' }, { ' ', ' ' }, { '!', '!' }, { '"', '"' }, { '#', '#' }, { '$', '\'' }, { '(', '(' }, { ')', ')' }, { '*', '*' }, { '+', '+' }, { ',', ',' }, { '-', '-' }, { '.', '.' }, { '/', '/' }, { '0', '9' }, { ':', ':' }, { ';', ';' }, { '<', '=' }, { '>', '>' }, { '?', '?' }, { '@', '@' }, { 'A', 'Z' }, { '[', '[' }, { '\\', '\\' }, { ']', ']' }, { '^', '^' }, { '_', '_' }, { '`', '`' }, { 'a', 'a' }, { 'b', 'd' }, { 'e', 'e' }, { 'f', 'f' }, { 'g', 'g' }, { 'h', 'h' }, { 'i', 'i' }, { 'j', 'j' }, { 'k', 'k' }, { 'l', 'l' }, { 'm', 'm' }, { 'n', 'n' }, { 'o', 'o' }, { 'p', 'p' }, { 'q', 'q' }, { 'r', 'r' }, { 's', 's' }, { 't', 't' }, { 'u', 'u' }, { 'v', 'z' }, { '{', '{' }, { '|', '|' }, { '}', '\U0010ffff' } }
var transitions = []map[int]int {
    { 47: 24, 10: 49, 5: 1, 15: 10, 2: 1, 43: 11, 40: 24, 38: 24, 22: 29, 48: 24, 28: 43, 17: 7, 9: 8, 32: 24, 46: 24, 52: 24, 44: 24, 27: 24, 34: 24, 37: 50, 36: 24, 19: 30, 41: 24, 21: 25, 35: 24, 55: 37, 18: 17, 3: 1, 50: 2, 12: 27, 0: 38, 53: 24, 45: 24, 13: 28, 7: 1, 25: 52, 51: 45, 39: 24, 14: 32, 49: 40, 42: 24 },
    { 7: 1, 2: 1, 3: 1, 5: 1 },
    { 37: 24, 27: 24, 35: 24, 44: 24, 36: 24, 39: 24, 40: 24, 45: 24, 32: 24, 53: 24, 42: 3, 43: 24, 51: 24, 47: 24, 49: 24, 41: 24, 34: 24, 50: 24, 48: 24, 38: 24, 46: 24, 20: 24, 52: 24 },
    { 36: 24, 27: 24, 40: 31, 37: 24, 45: 24, 53: 24, 35: 24, 47: 24, 52: 24, 48: 24, 46: 24, 34: 24, 32: 24, 44: 24, 38: 24, 50: 24, 42: 24, 51: 24, 49: 24, 41: 24, 43: 24, 39: 24, 20: 24 },
    { 44: 24, 51: 24, 41: 24, 48: 24, 49: 24, 47: 24, 42: 24, 52: 24, 53: 24, 34: 24, 35: 24, 32: 24, 37: 5, 40: 24, 50: 24, 20: 24, 45: 24, 38: 24, 46: 24, 43: 24, 27: 24, 36: 24, 39: 24 },
    { 35: 24, 46: 24, 53: 24, 41: 24, 40: 24, 52: 24, 51: 6, 32: 24, 20: 24, 38: 24, 36: 24, 48: 24, 37: 24, 47: 24, 43: 24, 50: 24, 45: 24, 42: 24, 44: 24, 49: 24, 39: 24, 27: 24, 34: 24 },
    { 42: 24, 51: 24, 49: 24, 47: 24, 48: 24, 53: 24, 27: 24, 34: 24, 50: 24, 35: 24, 36: 24, 44: 24, 37: 24, 40: 24, 38: 24, 41: 24, 43: 24, 32: 24, 52: 24, 39: 24, 20: 24, 46: 24, 45: 24 },
    { 24: 20 },
    { 32: 8, 28: 8, 56: 8, 8: 8, 30: 8, 25: 8, 18: 8, 41: 8, 31: 8, 38: 8, 20: 8, 34: 8, 12: 8, 39: 8, 48: 8, 2: 8, 50: 8, 47: 8, 42: 8, 54: 8, 11: 8, 13: 8, 6: 8, 52: 8, 27: 8, 19: 8, 9: 12, 10: 8, 53: 8, 16: 8, 7: 8, 17: 8, 22: 8, 15: 8, 45: 8, 29: 36, 40: 8, 1: 8, 37: 8, 51: 8, 46: 8, 24: 8, 36: 8, 4: 8, 55: 8, 14: 8, 49: 8, 43: 8, 44: 8, 33: 8, 23: 8, 35: 8, 21: 8, 26: 8 },
    { 41: 51, 12: 51, 56: 51, 44: 51, 43: 51, 19: 16, 14: 51, 4: 51, 30: 51, 22: 51, 42: 51, 25: 51, 8: 51, 6: 51, 40: 51, 39: 51, 34: 51, 53: 51, 18: 51, 35: 51, 45: 51, 15: 51, 50: 51, 1: 51, 29: 51, 55: 51, 9: 51, 31: 51, 27: 51, 5: 51, 21: 51, 33: 51, 51: 51, 16: 51, 23: 51, 13: 51, 32: 51, 38: 51, 2: 51, 36: 51, 52: 51, 24: 51, 48: 51, 11: 51, 20: 51, 49: 51, 54: 51, 28: 51, 46: 51, 17: 51, 47: 51, 26: 51, 37: 51, 7: 51, 10: 51, 3: 51 },
    { },
    { 37: 24, 39: 24, 27: 24, 34: 24, 46: 24, 49: 24, 41: 24, 43: 24, 53: 24, 35: 24, 47: 24, 51: 24, 50: 24, 32: 24, 45: 24, 38: 24, 20: 24, 44: 24, 36: 4, 48: 24, 52: 24, 40: 24, 42: 24 },
    { },
    { 51: 24, 20: 24, 36: 24, 44: 24, 38: 24, 41: 24, 50: 24, 39: 24, 35: 24, 52: 24, 53: 24, 43: 24, 37: 24, 27: 24, 47: 24, 32: 24, 48: 24, 42: 24, 34: 14, 46: 24, 49: 24, 40: 24, 45: 24 },
    { 48: 24, 37: 24, 53: 24, 47: 24, 51: 24, 45: 24, 44: 24, 35: 24, 49: 24, 34: 24, 42: 24, 40: 24, 39: 24, 20: 24, 50: 24, 52: 24, 46: 24, 38: 21, 27: 24, 32: 24, 36: 24, 41: 24, 43: 24 },
    { 43: 24, 41: 24, 47: 24, 52: 24, 45: 24, 36: 54, 37: 24, 34: 24, 35: 24, 42: 24, 51: 24, 53: 24, 39: 24, 38: 24, 48: 24, 44: 24, 49: 24, 20: 24, 32: 24, 46: 24, 27: 24, 40: 24, 50: 24 },
    { },
    { },
    { 50: 24, 42: 24, 52: 24, 35: 24, 38: 24, 49: 24, 45: 24, 37: 24, 34: 24, 53: 24, 27: 24, 32: 24, 36: 24, 44: 24, 48: 24, 41: 24, 47: 24, 43: 24, 20: 24, 39: 24, 40: 24, 46: 24, 51: 24 },
    { 51: 24, 37: 24, 45: 24, 42: 24, 52: 24, 47: 24, 44: 24, 39: 24, 50: 24, 46: 24, 49: 24, 41: 24, 20: 24, 48: 24, 27: 24, 40: 24, 32: 24, 36: 34, 38: 24, 35: 24, 53: 24, 34: 24, 43: 24 },
    { },
    { 37: 24, 45: 24, 43: 24, 51: 24, 46: 24, 39: 24, 42: 24, 36: 24, 44: 15, 49: 24, 27: 24, 34: 24, 35: 24, 47: 24, 38: 24, 40: 24, 20: 24, 32: 24, 52: 24, 48: 24, 53: 24, 41: 24, 50: 24 },
    { 35: 24, 32: 24, 53: 24, 47: 24, 42: 24, 52: 24, 40: 24, 45: 18, 50: 24, 46: 24, 27: 24, 38: 24, 49: 24, 51: 24, 43: 24, 37: 24, 41: 24, 44: 24, 39: 24, 36: 24, 34: 24, 20: 24, 48: 24 },
    { 20: 23, 36: 23, 51: 23, 54: 23, 49: 23, 47: 23, 55: 23, 42: 23, 31: 23, 29: 23, 50: 23, 15: 23, 33: 23, 8: 23, 17: 23, 34: 23, 37: 23, 16: 23, 46: 23, 2: 23, 18: 23, 32: 23, 19: 23, 7: 23, 4: 23, 11: 23, 53: 23, 12: 23, 41: 23, 43: 23, 44: 23, 23: 23, 38: 23, 56: 23, 5: 16, 39: 23, 22: 23, 26: 23, 10: 23, 24: 23, 6: 23, 35: 23, 13: 23, 9: 23, 14: 23, 27: 23, 1: 23, 21: 23, 25: 23, 52: 23, 0: 16, 45: 23, 48: 23, 3: 16, 28: 23, 40: 23, 30: 23 },
    { 35: 24, 48: 24, 41: 24, 20: 24, 36: 24, 38: 24, 40: 24, 49: 24, 46: 24, 37: 24, 44: 24, 53: 24, 45: 24, 52: 24, 27: 24, 50: 24, 32: 24, 51: 24, 42: 24, 47: 24, 34: 24, 43: 24, 39: 24 },
    { },
    { 43: 24, 20: 24, 36: 24, 49: 24, 37: 24, 53: 24, 35: 24, 45: 24, 40: 24, 52: 24, 27: 24, 46: 24, 51: 24, 39: 24, 50: 24, 32: 24, 34: 24, 42: 24, 48: 24, 38: 24, 44: 24, 41: 24, 47: 24 },
    { },
    { },
    { },
    { 19: 23, 14: 51 },
    { 38: 24, 20: 24, 35: 24, 37: 24, 39: 24, 41: 24, 32: 24, 40: 24, 43: 24, 49: 24, 27: 24, 45: 24, 48: 24, 46: 24, 51: 24, 47: 26, 36: 24, 42: 24, 44: 24, 34: 24, 53: 24, 50: 24, 52: 24 },
    { },
    { 53: 24, 44: 24, 40: 24, 52: 24, 35: 24, 32: 24, 34: 24, 37: 24, 41: 24, 36: 24, 20: 24, 50: 24, 38: 24, 43: 24, 47: 24, 39: 24, 45: 24, 46: 24, 48: 24, 51: 47, 42: 24, 49: 24, 27: 24 },
    { 27: 24, 48: 24, 49: 24, 34: 24, 35: 24, 39: 24, 52: 24, 45: 24, 20: 24, 40: 24, 47: 24, 43: 24, 41: 24, 46: 24, 51: 24, 37: 24, 42: 24, 36: 24, 53: 24, 32: 24, 38: 24, 44: 24, 50: 24 },
    { },
    { 44: 8, 21: 8, 52: 8, 4: 8, 35: 8, 6: 8, 31: 8, 18: 8, 46: 8, 47: 8, 26: 8, 19: 8, 56: 8, 34: 8, 13: 8, 24: 8, 23: 8, 49: 8, 11: 8, 55: 8, 30: 8, 37: 8, 33: 8, 25: 8, 17: 8, 50: 8, 39: 8, 38: 8, 14: 8, 36: 8, 22: 8, 48: 8, 28: 8, 8: 8, 41: 8, 54: 8, 42: 8, 40: 8, 29: 8, 1: 8, 45: 8, 10: 8, 43: 8, 7: 8, 32: 8, 9: 8, 53: 8, 27: 8, 51: 8, 12: 8, 15: 8, 20: 8, 16: 8, 2: 8 },
    { },
    { },
    { 50: 24, 53: 24, 42: 24, 38: 24, 36: 22, 47: 24, 48: 24, 46: 24, 41: 24, 52: 24, 39: 24, 34: 24, 27: 24, 37: 24, 32: 24, 43: 24, 49: 24, 35: 24, 44: 24, 51: 24, 20: 24, 45: 24, 40: 24 },
    { 43: 24, 52: 48, 47: 24, 42: 24, 34: 24, 45: 24, 35: 24, 32: 24, 41: 24, 20: 24, 51: 24, 44: 24, 36: 24, 46: 24, 49: 24, 40: 41, 53: 24, 39: 24, 50: 24, 48: 24, 37: 24, 38: 24, 27: 24 },
    { 43: 24, 44: 24, 35: 24, 37: 24, 39: 24, 48: 24, 51: 24, 52: 24, 41: 24, 34: 24, 47: 24, 40: 24, 38: 42, 53: 24, 27: 24, 50: 24, 45: 24, 42: 24, 49: 24, 20: 24, 32: 24, 46: 24, 36: 24 },
    { 40: 24, 38: 24, 41: 24, 36: 24, 52: 24, 20: 24, 34: 24, 35: 24, 43: 24, 48: 24, 51: 24, 47: 24, 46: 24, 53: 24, 37: 24, 39: 33, 42: 24, 44: 24, 32: 24, 27: 24, 45: 24, 50: 24, 49: 24 },
    { 4: 43, 47: 43, 15: 43, 34: 43, 44: 43, 56: 43, 42: 43, 31: 43, 36: 43, 46: 43, 20: 43, 28: 43, 6: 43, 49: 43, 41: 43, 24: 43, 21: 43, 43: 43, 12: 43, 39: 43, 45: 43, 54: 43, 29: 53, 18: 43, 53: 43, 1: 43, 17: 43, 19: 43, 23: 43, 37: 43, 30: 35, 33: 43, 13: 43, 35: 43, 8: 43, 25: 43, 7: 43, 9: 43, 55: 43, 38: 43, 14: 43, 48: 43, 51: 43, 2: 43, 16: 43, 10: 43, 32: 43, 50: 43, 27: 43, 52: 43, 11: 43, 22: 43, 26: 43, 40: 43 },
    { 51: 24, 27: 24, 52: 24, 38: 24, 35: 24, 36: 24, 34: 24, 47: 24, 44: 24, 37: 24, 53: 24, 20: 24, 42: 24, 46: 24, 48: 24, 32: 24, 50: 24, 49: 24, 39: 24, 41: 24, 40: 24, 43: 24, 45: 24 },
    { 39: 24, 48: 24, 53: 24, 40: 24, 47: 24, 35: 24, 20: 24, 49: 24, 51: 24, 45: 24, 41: 24, 52: 24, 32: 24, 36: 24, 34: 24, 50: 24, 43: 24, 38: 24, 37: 24, 27: 24, 42: 24, 44: 24, 46: 46 },
    { 39: 24, 47: 24, 50: 24, 43: 24, 52: 24, 20: 24, 38: 24, 51: 24, 35: 24, 45: 24, 46: 24, 40: 24, 44: 24, 32: 24, 42: 39, 49: 24, 34: 24, 41: 24, 37: 24, 27: 24, 48: 24, 53: 24, 36: 24 },
    { 36: 24, 51: 24, 35: 24, 47: 24, 41: 24, 50: 24, 43: 24, 46: 24, 49: 24, 38: 24, 39: 24, 40: 24, 48: 24, 52: 24, 53: 24, 20: 24, 37: 24, 44: 24, 27: 24, 34: 24, 32: 24, 42: 24, 45: 24 },
    { 45: 24, 38: 24, 39: 24, 36: 24, 35: 24, 50: 24, 49: 24, 43: 19, 51: 24, 27: 24, 47: 24, 42: 24, 52: 24, 32: 24, 44: 24, 46: 24, 48: 24, 20: 24, 41: 24, 53: 24, 37: 24, 34: 24, 40: 24 },
    { },
    { 48: 24, 40: 24, 45: 24, 32: 24, 46: 24, 35: 24, 50: 24, 52: 24, 38: 24, 49: 13, 44: 24, 37: 24, 27: 24, 51: 24, 39: 24, 43: 24, 41: 24, 36: 24, 42: 24, 53: 24, 20: 24, 34: 24, 47: 24 },
    { 43: 51, 31: 51, 6: 51, 38: 51, 21: 51, 52: 51, 5: 51, 24: 51, 51: 51, 3: 51, 35: 51, 28: 51, 2: 51, 37: 51, 13: 51, 23: 51, 42: 51, 14: 9, 16: 51, 54: 51, 33: 51, 56: 51, 4: 51, 50: 51, 41: 51, 40: 51, 30: 51, 45: 51, 17: 51, 22: 51, 49: 51, 36: 51, 20: 51, 7: 51, 29: 51, 9: 51, 27: 51, 53: 51, 8: 51, 46: 51, 48: 51, 1: 51, 55: 51, 19: 51, 32: 51, 15: 51, 10: 51, 12: 51, 18: 51, 34: 51, 26: 51, 47: 51, 39: 51, 44: 51, 11: 51, 25: 51 },
    { },
    { 45: 43, 35: 43, 46: 43, 24: 43, 50: 43, 55: 43, 22: 43, 34: 43, 8: 43, 36: 43, 47: 43, 23: 43, 19: 43, 40: 43, 12: 43, 20: 43, 15: 43, 37: 43, 54: 43, 21: 43, 43: 43, 27: 43, 25: 43, 39: 43, 42: 43, 33: 43, 11: 43, 18: 43, 38: 43, 31: 43, 28: 43, 30: 43, 13: 43, 6: 43, 10: 43, 1: 43, 9: 43, 7: 43, 51: 43, 53: 43, 32: 43, 52: 43, 2: 43, 14: 43, 48: 43, 4: 43, 16: 43, 44: 43, 26: 43, 56: 43, 41: 43, 17: 43, 49: 43, 29: 43 },
    { 32: 24, 38: 24, 49: 24, 44: 24, 46: 24, 47: 24, 52: 24, 40: 24, 43: 24, 27: 24, 20: 24, 37: 24, 34: 24, 48: 24, 35: 24, 45: 55, 41: 24, 50: 24, 51: 24, 42: 24, 53: 24, 36: 24, 39: 24 },
    { 51: 44, 53: 24, 20: 24, 32: 24, 44: 24, 41: 24, 36: 24, 52: 24, 40: 24, 43: 24, 37: 24, 50: 24, 42: 24, 38: 24, 46: 24, 34: 24, 35: 24, 45: 24, 48: 24, 47: 24, 49: 24, 39: 24, 27: 24 },
}
var accept = map[int]TokenType { 44: FRAGMENT, 49: HASH, 3: IDENTIFIER, 45: IDENTIFIER, 47: RIGHT, 54: IDENTIFIER, 2: IDENTIFIER, 5: IDENTIFIER, 6: LEFT, 11: IDENTIFIER, 12: STRING, 25: COLON, 28: R_PAREN, 16: COMMENT, 41: IDENTIFIER, 55: IDENTIFIER, 13: IDENTIFIER, 15: IDENTIFIER, 27: L_PAREN, 33: IDENTIFIER, 34: RULE, 39: IDENTIFIER, 50: IDENTIFIER, 52: QUESTION, 10: PLUS, 17: DOT, 26: SKIP, 35: CLASS, 42: IDENTIFIER, 21: IDENTIFIER, 29: SEMI, 37: BAR, 46: IDENTIFIER, 1: WHITESPACE, 14: IDENTIFIER, 19: IDENTIFIER, 20: ARROW, 22: IDENTIFIER, 32: STAR, 40: IDENTIFIER, 48: IDENTIFIER, 4: IDENTIFIER, 18: TOKEN, 24: IDENTIFIER, 31: IDENTIFIER, 38: EOF }

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
