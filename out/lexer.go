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

const (WHITESPACE TokenType = iota; COMMENT; RULE; TOKEN; FRAGMENT; LEFT; RIGHT; SKIP; EQUAL; PLUS; STAR; QUESTION; DOT; BAR; HASH; SEMI; COLON; L_PAREN; R_PAREN; ARROW; IDENTIFIER; STRING; CLASS; EOF)
func (t TokenType) String() string { return typeName[t] }
var typeName = map[TokenType]string { WHITESPACE: "WHITESPACE", COMMENT: "COMMENT", RULE: "RULE", TOKEN: "TOKEN", FRAGMENT: "FRAGMENT", LEFT: "LEFT", RIGHT: "RIGHT", SKIP: "SKIP", EQUAL: "EQUAL", PLUS: "PLUS", STAR: "STAR", QUESTION: "QUESTION", DOT: "DOT", BAR: "BAR", HASH: "HASH", SEMI: "SEMI", COLON: "COLON", L_PAREN: "L_PAREN", R_PAREN: "R_PAREN", ARROW: "ARROW", IDENTIFIER: "IDENTIFIER", STRING: "STRING", CLASS: "CLASS", EOF: "EOF" }
var skip = map[TokenType]struct{} { WHITESPACE: {}, COMMENT: {} }

var ranges = []Range { { '\x00', '\x00' }, { '\x01', '\b' }, { '\t', '\t' }, { '\n', '\n' }, { '\v', '\f' }, { '\r', '\r' }, { '\x0e', '\x1f' }, { ' ', ' ' }, { '!', '!' }, { '"', '"' }, { '#', '#' }, { '$', '\'' }, { '(', '(' }, { ')', ')' }, { '*', '*' }, { '+', '+' }, { ',', ',' }, { '-', '-' }, { '.', '.' }, { '/', '/' }, { '0', '9' }, { ':', ':' }, { ';', ';' }, { '<', '<' }, { '=', '=' }, { '>', '>' }, { '?', '?' }, { '@', '@' }, { 'A', 'Z' }, { '[', '[' }, { '\\', '\\' }, { ']', ']' }, { '^', '^' }, { '_', '_' }, { '`', '`' }, { 'a', 'a' }, { 'b', 'd' }, { 'e', 'e' }, { 'f', 'f' }, { 'g', 'g' }, { 'h', 'h' }, { 'i', 'i' }, { 'j', 'j' }, { 'k', 'k' }, { 'l', 'l' }, { 'm', 'm' }, { 'n', 'n' }, { 'o', 'o' }, { 'p', 'p' }, { 'q', 'q' }, { 'r', 'r' }, { 's', 's' }, { 't', 't' }, { 'u', 'u' }, { 'v', 'z' }, { '{', '{' }, { '|', '|' }, { '}', '\U0010ffff' } }
var transitions = []map[int]int {
    { 39: 34, 33: 34, 52: 11, 17: 33, 43: 34, 37: 34, 14: 31, 29: 4, 47: 34, 49: 34, 9: 35, 28: 34, 44: 55, 2: 1, 50: 26, 3: 1, 18: 15, 5: 1, 41: 34, 54: 34, 13: 14, 21: 36, 26: 16, 40: 34, 12: 7, 53: 34, 51: 47, 24: 8, 45: 34, 0: 20, 19: 53, 15: 38, 7: 1, 35: 34, 46: 34, 56: 42, 48: 34, 10: 39, 36: 34, 38: 40, 42: 34, 22: 43 },
    { 2: 1, 3: 1, 5: 1, 7: 1 },
    { 48: 34, 38: 34, 44: 34, 35: 34, 43: 34, 50: 34, 53: 34, 39: 34, 37: 34, 33: 34, 52: 27, 28: 34, 51: 34, 47: 34, 45: 34, 54: 34, 20: 34, 42: 34, 46: 34, 36: 34, 41: 34, 49: 34, 40: 34 },
    { 43: 18, 40: 34, 35: 34, 39: 34, 37: 34, 41: 34, 48: 34, 47: 34, 46: 34, 33: 34, 38: 34, 36: 34, 44: 34, 52: 34, 50: 34, 53: 34, 45: 34, 49: 34, 54: 34, 28: 34, 42: 34, 20: 34, 51: 34 },
    { 11: 4, 57: 4, 47: 4, 16: 4, 4: 4, 46: 4, 56: 4, 32: 4, 18: 4, 34: 4, 55: 4, 41: 4, 51: 4, 43: 4, 30: 23, 14: 4, 53: 4, 8: 4, 10: 4, 23: 4, 2: 4, 17: 4, 7: 4, 25: 4, 39: 4, 48: 4, 35: 4, 1: 4, 37: 4, 44: 4, 9: 4, 31: 32, 27: 4, 42: 4, 52: 4, 54: 4, 40: 4, 6: 4, 12: 4, 45: 4, 49: 4, 26: 4, 15: 4, 22: 4, 38: 4, 19: 4, 13: 4, 33: 4, 21: 4, 24: 4, 28: 4, 36: 4, 29: 4, 50: 4, 20: 4 },
    { 24: 35, 15: 35, 11: 35, 34: 35, 45: 35, 31: 35, 41: 35, 55: 35, 21: 35, 38: 35, 25: 35, 4: 35, 19: 35, 2: 35, 10: 35, 40: 35, 27: 35, 1: 35, 28: 35, 6: 35, 14: 35, 9: 35, 16: 35, 39: 35, 42: 35, 23: 35, 57: 35, 17: 35, 7: 35, 8: 35, 33: 35, 18: 35, 54: 35, 29: 35, 53: 35, 46: 35, 35: 35, 30: 35, 49: 35, 22: 35, 43: 35, 50: 35, 48: 35, 56: 35, 32: 35, 26: 35, 12: 35, 36: 35, 20: 35, 47: 35, 44: 35, 52: 35, 37: 35, 51: 35, 13: 35 },
    { 46: 34, 48: 34, 41: 34, 54: 34, 37: 34, 40: 34, 33: 34, 51: 34, 35: 34, 44: 34, 36: 34, 45: 34, 52: 34, 49: 34, 28: 34, 43: 34, 50: 34, 53: 34, 39: 34, 42: 34, 47: 34, 38: 34, 20: 34 },
    { },
    { },
    { },
    { 28: 34, 38: 34, 43: 34, 53: 34, 46: 34, 36: 34, 42: 34, 45: 34, 48: 34, 54: 34, 52: 34, 37: 34, 35: 34, 39: 34, 20: 34, 33: 34, 51: 34, 44: 34, 49: 34, 50: 34, 40: 34, 41: 34, 47: 34 },
    { 36: 34, 42: 34, 43: 34, 46: 34, 48: 34, 41: 34, 52: 34, 20: 34, 47: 3, 54: 34, 51: 34, 49: 34, 53: 34, 39: 34, 44: 34, 50: 34, 40: 34, 45: 34, 28: 34, 33: 34, 37: 34, 35: 34, 38: 34 },
    { 49: 34, 47: 34, 38: 25, 43: 34, 48: 34, 41: 34, 50: 34, 20: 34, 28: 34, 37: 34, 53: 34, 44: 34, 36: 34, 39: 34, 51: 34, 40: 34, 54: 34, 45: 34, 46: 34, 42: 34, 52: 34, 35: 34, 33: 34 },
    { 33: 34, 38: 34, 42: 34, 46: 34, 40: 34, 41: 34, 52: 34, 35: 34, 51: 34, 28: 34, 20: 34, 43: 34, 44: 34, 49: 34, 39: 51, 47: 34, 50: 34, 48: 34, 53: 34, 37: 34, 36: 34, 45: 34, 54: 34 },
    { },
    { },
    { },
    { 41: 34, 39: 34, 45: 30, 49: 34, 51: 34, 47: 34, 52: 34, 50: 34, 44: 34, 53: 34, 43: 34, 40: 34, 33: 34, 42: 34, 35: 34, 48: 34, 20: 34, 28: 34, 36: 34, 37: 34, 46: 34, 54: 34, 38: 34 },
    { 52: 34, 39: 34, 20: 34, 36: 34, 54: 34, 46: 34, 28: 34, 40: 34, 41: 34, 44: 34, 33: 34, 43: 34, 38: 34, 45: 34, 49: 34, 53: 34, 37: 50, 51: 34, 50: 34, 48: 34, 35: 34, 47: 34, 42: 34 },
    { 45: 34, 39: 34, 50: 34, 47: 34, 37: 34, 35: 34, 28: 34, 40: 34, 53: 34, 42: 34, 49: 34, 36: 34, 44: 34, 46: 34, 48: 34, 54: 34, 43: 34, 41: 34, 51: 34, 52: 34, 20: 34, 33: 34, 38: 34 },
    { },
    { 54: 34, 45: 34, 50: 34, 40: 34, 37: 34, 44: 34, 35: 34, 38: 34, 47: 34, 53: 34, 46: 54, 42: 34, 43: 34, 49: 34, 20: 34, 51: 34, 28: 34, 41: 34, 52: 34, 48: 34, 39: 34, 33: 34, 36: 34 },
    { 28: 34, 40: 34, 43: 34, 54: 34, 51: 34, 41: 34, 36: 34, 50: 34, 20: 34, 35: 34, 46: 34, 37: 34, 42: 34, 45: 34, 38: 34, 39: 34, 44: 34, 48: 52, 49: 34, 47: 34, 33: 34, 52: 34, 53: 34 },
    { 21: 4, 18: 4, 15: 4, 20: 4, 42: 4, 8: 4, 50: 4, 13: 4, 38: 4, 16: 4, 28: 4, 55: 4, 37: 4, 19: 4, 34: 4, 2: 4, 6: 4, 12: 4, 36: 4, 4: 4, 11: 4, 10: 4, 26: 4, 23: 4, 27: 4, 41: 4, 45: 4, 17: 4, 39: 4, 40: 4, 56: 4, 52: 4, 1: 4, 35: 4, 22: 4, 14: 4, 32: 4, 49: 4, 29: 4, 31: 4, 46: 4, 51: 4, 57: 4, 44: 4, 47: 4, 54: 4, 48: 4, 53: 4, 33: 4, 24: 4, 30: 4, 7: 4, 9: 4, 43: 4, 25: 4 },
    { },
    { 44: 34, 46: 34, 20: 34, 41: 34, 42: 34, 38: 34, 50: 34, 33: 34, 49: 34, 40: 34, 47: 34, 51: 34, 43: 34, 37: 34, 39: 34, 48: 34, 53: 34, 54: 34, 45: 34, 52: 6, 35: 34, 36: 34, 28: 34 },
    { 52: 34, 40: 34, 20: 34, 51: 34, 50: 34, 48: 34, 54: 34, 28: 34, 38: 34, 35: 34, 41: 13, 44: 34, 45: 34, 43: 34, 46: 34, 37: 34, 36: 34, 39: 34, 49: 34, 53: 44, 42: 34, 47: 34, 33: 34 },
    { 43: 34, 35: 34, 20: 34, 52: 34, 46: 34, 28: 34, 37: 34, 38: 34, 51: 34, 54: 34, 40: 34, 49: 34, 39: 34, 42: 34, 50: 34, 48: 34, 45: 34, 41: 34, 53: 34, 36: 34, 44: 34, 47: 34, 33: 34 },
    { 46: 34, 47: 34, 44: 34, 52: 34, 35: 34, 49: 34, 20: 34, 54: 34, 43: 34, 38: 34, 50: 34, 33: 34, 48: 34, 37: 34, 42: 34, 45: 34, 53: 34, 36: 34, 40: 34, 28: 34, 39: 34, 51: 34, 41: 34 },
    { 52: 34, 54: 34, 40: 34, 42: 34, 36: 34, 43: 34, 47: 34, 33: 34, 50: 34, 38: 34, 41: 22, 44: 34, 20: 34, 35: 34, 37: 34, 39: 34, 48: 34, 49: 34, 46: 34, 45: 34, 51: 34, 53: 34, 28: 34 },
    { 20: 34, 36: 34, 42: 34, 53: 34, 28: 34, 38: 34, 49: 34, 50: 34, 46: 34, 33: 34, 39: 34, 54: 34, 37: 21, 45: 34, 41: 34, 43: 34, 48: 34, 47: 34, 44: 34, 51: 34, 40: 34, 52: 34, 35: 34 },
    { },
    { },
    { 25: 9 },
    { 39: 34, 49: 34, 48: 34, 42: 34, 20: 34, 33: 34, 52: 34, 36: 34, 51: 34, 38: 34, 50: 34, 46: 34, 40: 34, 37: 34, 44: 34, 54: 34, 41: 34, 35: 34, 28: 34, 45: 34, 43: 34, 47: 34, 53: 34 },
    { 43: 35, 29: 35, 36: 35, 53: 35, 21: 35, 28: 35, 55: 35, 6: 35, 22: 35, 37: 35, 26: 35, 45: 35, 41: 35, 48: 35, 54: 35, 51: 35, 11: 35, 12: 35, 32: 35, 50: 35, 57: 35, 8: 35, 52: 35, 31: 35, 42: 35, 38: 35, 9: 24, 34: 35, 44: 35, 24: 35, 18: 35, 56: 35, 23: 35, 14: 35, 15: 35, 30: 5, 7: 35, 13: 35, 33: 35, 2: 35, 46: 35, 35: 35, 16: 35, 4: 35, 19: 35, 1: 35, 25: 35, 17: 35, 20: 35, 27: 35, 49: 35, 39: 35, 10: 35, 47: 35, 40: 35 },
    { },
    { 43: 45, 2: 45, 27: 45, 39: 45, 44: 45, 3: 45, 24: 45, 8: 45, 38: 45, 50: 45, 52: 45, 33: 45, 4: 45, 17: 45, 37: 45, 6: 45, 10: 45, 5: 45, 51: 45, 32: 45, 48: 45, 21: 45, 31: 45, 28: 45, 13: 45, 26: 45, 34: 45, 20: 45, 9: 45, 30: 45, 12: 45, 1: 45, 22: 45, 46: 45, 7: 45, 55: 45, 49: 45, 45: 45, 19: 48, 47: 45, 36: 45, 15: 45, 25: 45, 18: 45, 40: 45, 23: 45, 42: 45, 53: 45, 14: 45, 35: 45, 57: 45, 11: 45, 29: 45, 54: 45, 16: 45, 56: 45, 41: 45 },
    { },
    { },
    { 36: 34, 20: 34, 47: 34, 33: 34, 54: 34, 41: 34, 53: 34, 43: 34, 42: 34, 39: 34, 48: 34, 37: 34, 45: 34, 51: 34, 49: 34, 28: 34, 44: 34, 38: 34, 52: 34, 46: 34, 40: 34, 35: 34, 50: 46 },
    { 8: 41, 35: 41, 19: 41, 38: 41, 41: 41, 33: 41, 2: 41, 39: 41, 15: 41, 16: 41, 23: 41, 11: 41, 53: 41, 50: 41, 26: 41, 54: 41, 10: 41, 57: 41, 36: 41, 49: 41, 6: 41, 46: 41, 7: 41, 51: 41, 29: 41, 32: 41, 0: 48, 1: 41, 55: 41, 28: 41, 24: 41, 12: 41, 3: 48, 31: 41, 4: 41, 43: 41, 56: 41, 21: 41, 42: 41, 45: 41, 20: 41, 48: 41, 9: 41, 25: 41, 13: 41, 44: 41, 17: 41, 27: 41, 14: 41, 22: 41, 37: 41, 40: 41, 34: 41, 30: 41, 5: 48, 52: 41, 47: 41, 18: 41 },
    { },
    { },
    { 43: 34, 40: 34, 38: 34, 46: 34, 54: 34, 47: 34, 42: 34, 49: 34, 36: 34, 53: 34, 44: 56, 45: 34, 50: 34, 41: 34, 52: 34, 39: 34, 48: 34, 20: 34, 33: 34, 51: 34, 28: 34, 37: 34, 35: 34 },
    { 51: 45, 54: 45, 21: 45, 26: 45, 10: 45, 2: 45, 9: 45, 12: 45, 18: 45, 49: 45, 20: 45, 19: 45, 50: 45, 39: 45, 7: 45, 32: 45, 33: 45, 3: 45, 24: 45, 38: 45, 8: 45, 41: 45, 48: 45, 17: 45, 31: 45, 36: 45, 25: 45, 44: 45, 5: 45, 35: 45, 45: 45, 53: 45, 42: 45, 23: 45, 52: 45, 11: 45, 56: 45, 16: 45, 57: 45, 29: 45, 14: 37, 47: 45, 4: 45, 55: 45, 27: 45, 13: 45, 22: 45, 30: 45, 40: 45, 6: 45, 1: 45, 34: 45, 28: 45, 15: 45, 46: 45, 37: 45, 43: 45 },
    { 44: 34, 36: 34, 47: 34, 52: 34, 20: 34, 43: 34, 48: 34, 51: 34, 54: 34, 28: 34, 33: 34, 49: 34, 42: 34, 35: 49, 50: 34, 39: 34, 37: 34, 53: 34, 46: 34, 40: 34, 45: 34, 38: 34, 41: 34 },
    { 49: 34, 48: 34, 54: 34, 42: 34, 35: 34, 28: 34, 37: 34, 45: 34, 52: 34, 20: 34, 36: 34, 33: 34, 53: 34, 39: 34, 46: 34, 51: 34, 47: 34, 43: 29, 44: 34, 50: 34, 40: 34, 38: 34, 41: 34 },
    { },
    { 40: 34, 54: 34, 51: 34, 46: 34, 20: 34, 43: 34, 39: 17, 41: 34, 50: 34, 37: 34, 35: 34, 36: 34, 49: 34, 28: 34, 44: 34, 45: 34, 42: 34, 48: 34, 33: 34, 38: 34, 53: 34, 47: 34, 52: 34 },
    { 28: 34, 36: 34, 35: 34, 40: 34, 33: 34, 37: 34, 48: 34, 53: 34, 44: 34, 45: 34, 20: 34, 46: 19, 47: 34, 38: 34, 39: 34, 51: 34, 49: 34, 50: 34, 54: 34, 41: 34, 42: 34, 52: 34, 43: 34 },
    { 28: 34, 46: 34, 50: 34, 36: 34, 53: 34, 48: 34, 39: 34, 41: 34, 38: 34, 44: 34, 35: 34, 54: 34, 47: 34, 52: 34, 49: 34, 40: 2, 42: 34, 51: 34, 33: 34, 45: 34, 37: 34, 20: 34, 43: 34 },
    { 40: 34, 41: 34, 48: 34, 33: 34, 44: 34, 51: 34, 52: 34, 43: 34, 50: 34, 46: 34, 28: 34, 37: 34, 35: 34, 42: 34, 36: 34, 47: 34, 53: 34, 39: 34, 20: 34, 38: 34, 49: 34, 45: 34, 54: 34 },
    { 19: 41, 14: 45 },
    { 54: 34, 52: 10, 20: 34, 36: 34, 37: 34, 40: 34, 41: 34, 48: 34, 51: 34, 43: 34, 39: 34, 47: 34, 33: 34, 42: 34, 44: 34, 46: 34, 28: 34, 35: 34, 38: 34, 49: 34, 45: 34, 50: 34, 53: 34 },
    { 50: 34, 47: 34, 51: 34, 28: 34, 46: 34, 36: 34, 40: 34, 35: 34, 54: 34, 37: 12, 33: 34, 45: 34, 53: 34, 44: 34, 43: 34, 48: 34, 49: 34, 42: 34, 52: 34, 39: 34, 41: 34, 38: 34, 20: 34 },
    { 43: 34, 49: 34, 41: 34, 46: 34, 33: 34, 44: 34, 53: 34, 51: 34, 20: 34, 39: 34, 47: 34, 38: 34, 42: 34, 35: 34, 50: 34, 54: 34, 45: 34, 52: 34, 40: 34, 28: 34, 37: 28, 48: 34, 36: 34 },
}
var accept = map[int]TokenType { 7: L_PAREN, 11: IDENTIFIER, 17: IDENTIFIER, 30: IDENTIFIER, 31: STAR, 32: CLASS, 39: HASH, 2: IDENTIFIER, 10: FRAGMENT, 15: DOT, 20: EOF, 22: IDENTIFIER, 27: RIGHT, 44: IDENTIFIER, 52: SKIP, 12: IDENTIFIER, 14: R_PAREN, 29: IDENTIFIER, 38: PLUS, 42: BAR, 48: COMMENT, 56: IDENTIFIER, 8: EQUAL, 16: QUESTION, 25: IDENTIFIER, 36: COLON, 51: IDENTIFIER, 54: IDENTIFIER, 6: LEFT, 19: TOKEN, 40: IDENTIFIER, 43: SEMI, 13: IDENTIFIER, 47: IDENTIFIER, 49: IDENTIFIER, 55: IDENTIFIER, 3: IDENTIFIER, 18: IDENTIFIER, 34: IDENTIFIER, 1: WHITESPACE, 9: ARROW, 21: IDENTIFIER, 24: STRING, 26: IDENTIFIER, 28: RULE, 46: IDENTIFIER, 50: IDENTIFIER }

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
