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
    { 2: 32, 56: 41, 17: 7, 3: 32, 46: 11, 49: 11, 52: 53, 28: 11, 44: 25, 38: 30, 50: 21, 39: 11, 42: 11, 43: 11, 22: 3, 37: 11, 24: 4, 7: 32, 41: 11, 26: 13, 33: 11, 48: 11, 5: 32, 10: 56, 35: 11, 14: 19, 53: 11, 51: 14, 0: 15, 15: 28, 47: 11, 40: 11, 29: 10, 36: 11, 18: 45, 12: 34, 21: 35, 54: 11, 19: 48, 45: 11, 13: 17, 9: 52 },
    { 28: 11, 47: 11, 51: 11, 46: 11, 52: 11, 20: 11, 38: 11, 40: 11, 50: 11, 49: 11, 45: 11, 39: 11, 37: 50, 44: 11, 33: 11, 53: 11, 36: 11, 54: 11, 35: 11, 48: 11, 41: 11, 43: 11, 42: 11 },
    { 38: 11, 46: 11, 51: 11, 40: 11, 42: 11, 28: 11, 49: 11, 53: 11, 33: 11, 39: 11, 54: 11, 35: 11, 20: 11, 44: 11, 47: 11, 36: 11, 52: 11, 48: 11, 45: 11, 41: 11, 43: 11, 37: 11, 50: 11 },
    { },
    { },
    { 20: 11, 43: 11, 53: 11, 33: 11, 40: 11, 44: 11, 41: 11, 42: 11, 38: 11, 49: 11, 45: 11, 28: 11, 54: 11, 36: 11, 50: 11, 52: 8, 48: 11, 37: 11, 47: 11, 46: 11, 35: 11, 39: 11, 51: 11 },
    { 36: 11, 20: 11, 47: 11, 37: 43, 54: 11, 49: 11, 43: 11, 41: 11, 44: 11, 38: 11, 40: 11, 46: 11, 33: 11, 42: 11, 53: 11, 28: 11, 50: 11, 39: 11, 35: 11, 52: 11, 51: 11, 48: 11, 45: 11 },
    { 25: 55 },
    { 47: 11, 41: 11, 49: 11, 45: 11, 38: 11, 28: 11, 39: 11, 54: 11, 35: 11, 37: 11, 40: 11, 48: 11, 33: 11, 51: 11, 50: 11, 46: 11, 44: 11, 20: 11, 53: 11, 43: 11, 36: 11, 42: 11, 52: 11 },
    { 44: 11, 38: 11, 41: 11, 45: 11, 49: 11, 51: 11, 42: 11, 52: 11, 40: 11, 53: 11, 50: 11, 43: 11, 48: 11, 35: 11, 47: 11, 54: 11, 36: 11, 28: 11, 20: 11, 33: 11, 46: 11, 39: 11, 37: 11 },
    { 17: 10, 34: 10, 38: 10, 42: 10, 57: 10, 27: 10, 33: 10, 1: 10, 51: 10, 44: 10, 49: 10, 22: 10, 21: 10, 9: 10, 4: 10, 40: 10, 20: 10, 41: 10, 45: 10, 28: 10, 55: 10, 54: 10, 11: 10, 14: 10, 16: 10, 2: 10, 12: 10, 35: 10, 43: 10, 23: 10, 25: 10, 50: 10, 47: 10, 46: 10, 13: 10, 18: 10, 36: 10, 15: 10, 53: 10, 31: 33, 30: 22, 37: 10, 52: 10, 32: 10, 56: 10, 8: 10, 7: 10, 48: 10, 10: 10, 29: 10, 6: 10, 24: 10, 19: 10, 39: 10, 26: 10 },
    { 38: 11, 49: 11, 43: 11, 35: 11, 41: 11, 52: 11, 54: 11, 45: 11, 28: 11, 46: 11, 36: 11, 39: 11, 48: 11, 50: 11, 51: 11, 47: 11, 40: 11, 33: 11, 44: 11, 53: 11, 42: 11, 20: 11, 37: 11 },
    { 38: 11, 48: 11, 52: 11, 20: 11, 53: 11, 47: 11, 37: 11, 43: 11, 50: 11, 33: 11, 54: 11, 49: 11, 41: 11, 36: 11, 45: 11, 46: 11, 28: 11, 42: 11, 44: 11, 40: 54, 35: 11, 39: 11, 51: 11 },
    { },
    { 40: 11, 43: 26, 46: 11, 53: 11, 20: 11, 35: 11, 33: 11, 42: 11, 28: 11, 39: 11, 51: 11, 49: 11, 36: 11, 44: 11, 47: 11, 41: 11, 48: 11, 45: 11, 50: 11, 52: 11, 54: 11, 37: 11, 38: 11 },
    { },
    { 21: 16, 30: 16, 39: 16, 16: 16, 55: 16, 37: 16, 29: 16, 18: 16, 22: 16, 4: 16, 11: 16, 57: 16, 28: 16, 6: 16, 42: 16, 41: 16, 31: 16, 5: 16, 9: 16, 24: 16, 48: 16, 43: 16, 49: 16, 33: 16, 8: 16, 13: 16, 12: 16, 10: 16, 53: 16, 44: 16, 56: 16, 32: 16, 26: 16, 27: 16, 52: 16, 34: 16, 45: 16, 47: 16, 3: 16, 50: 16, 35: 16, 54: 16, 19: 16, 38: 16, 17: 16, 2: 16, 15: 16, 23: 16, 14: 37, 7: 16, 40: 16, 51: 16, 46: 16, 1: 16, 36: 16, 20: 16, 25: 16 },
    { },
    { 39: 11, 28: 11, 42: 11, 44: 6, 33: 11, 48: 11, 52: 11, 37: 11, 51: 11, 41: 11, 50: 11, 54: 11, 49: 11, 53: 11, 40: 11, 46: 11, 36: 11, 38: 11, 35: 11, 43: 11, 45: 11, 20: 11, 47: 11 },
    { },
    { 52: 11, 41: 11, 39: 11, 54: 11, 49: 11, 51: 11, 42: 11, 28: 11, 45: 11, 36: 11, 35: 11, 44: 11, 33: 11, 50: 11, 38: 11, 53: 11, 40: 11, 48: 11, 37: 11, 20: 11, 46: 5, 47: 11, 43: 11 },
    { 54: 11, 38: 11, 37: 11, 28: 11, 48: 11, 43: 11, 42: 11, 46: 11, 35: 11, 39: 11, 20: 11, 33: 11, 50: 11, 44: 11, 47: 11, 49: 11, 53: 18, 41: 44, 52: 11, 45: 11, 51: 11, 40: 11, 36: 11 },
    { 18: 10, 1: 10, 6: 10, 45: 10, 36: 10, 14: 10, 57: 10, 42: 10, 49: 10, 23: 10, 37: 10, 19: 10, 33: 10, 35: 10, 38: 10, 41: 10, 16: 10, 27: 10, 56: 10, 17: 10, 22: 10, 31: 10, 43: 10, 50: 10, 51: 10, 7: 10, 52: 10, 28: 10, 24: 10, 15: 10, 2: 10, 13: 10, 39: 10, 29: 10, 8: 10, 12: 10, 11: 10, 53: 10, 47: 10, 25: 10, 46: 10, 54: 10, 30: 10, 4: 10, 55: 10, 21: 10, 10: 10, 32: 10, 9: 10, 20: 10, 26: 10, 44: 10, 34: 10, 40: 10, 48: 10 },
    { 51: 52, 39: 52, 25: 52, 10: 52, 57: 52, 19: 52, 31: 52, 29: 52, 2: 52, 24: 52, 18: 52, 16: 52, 27: 52, 20: 52, 23: 52, 41: 52, 38: 52, 34: 52, 47: 52, 22: 52, 52: 52, 32: 52, 11: 52, 50: 52, 8: 52, 7: 52, 4: 52, 17: 52, 30: 52, 1: 52, 56: 52, 43: 52, 35: 52, 15: 52, 54: 52, 21: 52, 6: 52, 40: 52, 33: 52, 26: 52, 36: 52, 46: 52, 48: 52, 44: 52, 9: 52, 37: 52, 13: 52, 28: 52, 55: 52, 45: 52, 12: 52, 14: 52, 53: 52, 42: 52, 49: 52 },
    { 54: 11, 40: 11, 35: 11, 44: 11, 37: 11, 51: 11, 48: 11, 45: 11, 42: 11, 33: 11, 36: 11, 41: 11, 20: 11, 38: 11, 39: 11, 47: 11, 43: 1, 50: 11, 52: 11, 49: 11, 53: 11, 46: 11, 28: 11 },
    { 44: 11, 54: 11, 35: 11, 41: 11, 49: 11, 51: 11, 37: 42, 36: 11, 38: 11, 47: 11, 50: 11, 20: 11, 53: 11, 42: 11, 48: 11, 33: 11, 39: 11, 46: 11, 52: 11, 40: 11, 43: 11, 45: 11, 28: 11 },
    { 38: 11, 42: 11, 48: 11, 44: 11, 47: 11, 51: 11, 20: 11, 37: 11, 46: 11, 52: 11, 54: 11, 45: 11, 41: 27, 50: 11, 35: 11, 39: 11, 43: 11, 33: 11, 40: 11, 53: 11, 28: 11, 36: 11, 49: 11 },
    { 36: 11, 42: 11, 52: 11, 53: 11, 38: 11, 49: 11, 44: 11, 50: 11, 51: 11, 37: 11, 28: 11, 33: 11, 46: 11, 41: 11, 48: 40, 45: 11, 54: 11, 40: 11, 43: 11, 35: 11, 47: 11, 20: 11, 39: 11 },
    { },
    { 46: 11, 52: 11, 36: 11, 50: 11, 53: 11, 42: 11, 45: 11, 44: 11, 40: 11, 20: 11, 41: 11, 33: 11, 54: 11, 37: 11, 28: 11, 51: 11, 35: 11, 39: 11, 47: 11, 38: 11, 48: 11, 49: 11, 43: 11 },
    { 45: 11, 54: 11, 20: 11, 51: 11, 50: 51, 35: 11, 38: 11, 48: 11, 36: 11, 43: 11, 28: 11, 42: 11, 46: 11, 44: 11, 39: 11, 52: 11, 33: 11, 47: 11, 37: 11, 53: 11, 41: 11, 49: 11, 40: 11 },
    { 49: 11, 48: 11, 45: 11, 35: 11, 39: 47, 38: 11, 28: 11, 50: 11, 43: 11, 46: 11, 54: 11, 36: 11, 41: 11, 40: 11, 42: 11, 33: 11, 52: 11, 37: 11, 44: 11, 51: 11, 53: 11, 47: 11, 20: 11 },
    { 2: 32, 3: 32, 5: 32, 7: 32 },
    { },
    { },
    { },
    { 47: 11, 38: 11, 43: 11, 50: 11, 36: 11, 39: 11, 53: 11, 45: 11, 28: 11, 52: 11, 51: 11, 41: 11, 48: 11, 20: 11, 33: 11, 46: 11, 44: 11, 54: 11, 49: 11, 37: 20, 42: 11, 35: 11, 40: 11 },
    { 27: 16, 30: 16, 45: 16, 19: 46, 14: 16, 25: 16, 40: 16, 10: 16, 53: 16, 44: 16, 11: 16, 41: 16, 54: 16, 38: 16, 51: 16, 5: 16, 32: 16, 57: 16, 13: 16, 42: 16, 3: 16, 36: 16, 49: 16, 20: 16, 9: 16, 50: 16, 7: 16, 34: 16, 31: 16, 16: 16, 6: 16, 26: 16, 12: 16, 37: 16, 23: 16, 8: 16, 28: 16, 17: 16, 2: 16, 29: 16, 43: 16, 55: 16, 48: 16, 21: 16, 52: 16, 4: 16, 22: 16, 18: 16, 24: 16, 56: 16, 33: 16, 15: 16, 35: 16, 39: 16, 47: 16, 46: 16, 1: 16 },
    { 26: 38, 35: 38, 53: 38, 6: 38, 1: 38, 10: 38, 48: 38, 57: 38, 41: 38, 28: 38, 31: 38, 13: 38, 0: 46, 21: 38, 49: 38, 15: 38, 14: 38, 50: 38, 22: 38, 17: 38, 37: 38, 18: 38, 46: 38, 33: 38, 44: 38, 12: 38, 19: 38, 29: 38, 36: 38, 20: 38, 7: 38, 23: 38, 3: 46, 42: 38, 5: 46, 39: 38, 32: 38, 25: 38, 16: 38, 40: 38, 56: 38, 47: 38, 52: 38, 27: 38, 30: 38, 8: 38, 54: 38, 38: 38, 24: 38, 4: 38, 45: 38, 34: 38, 9: 38, 2: 38, 55: 38, 43: 38, 51: 38, 11: 38 },
    { 42: 11, 39: 11, 51: 11, 28: 11, 49: 11, 46: 11, 44: 11, 52: 29, 45: 11, 40: 11, 35: 11, 33: 11, 54: 11, 38: 11, 43: 11, 47: 11, 37: 11, 41: 11, 20: 11, 36: 11, 50: 11, 48: 11, 53: 11 },
    { 48: 11, 44: 11, 38: 11, 40: 11, 39: 11, 35: 11, 54: 11, 43: 11, 37: 11, 49: 11, 50: 11, 52: 11, 46: 11, 42: 11, 47: 11, 45: 11, 28: 11, 41: 11, 20: 11, 53: 11, 36: 11, 33: 11, 51: 11 },
    { },
    { 46: 11, 38: 39, 35: 11, 41: 11, 51: 11, 47: 11, 54: 11, 36: 11, 42: 11, 33: 11, 40: 11, 53: 11, 39: 11, 37: 11, 43: 11, 49: 11, 50: 11, 52: 11, 28: 11, 45: 11, 44: 11, 20: 11, 48: 11 },
    { 33: 11, 28: 11, 39: 11, 40: 11, 44: 11, 47: 11, 50: 11, 48: 11, 53: 11, 42: 11, 20: 11, 51: 11, 43: 11, 45: 11, 46: 11, 41: 11, 52: 11, 54: 11, 49: 11, 35: 11, 38: 11, 36: 11, 37: 11 },
    { 36: 11, 48: 11, 54: 11, 43: 11, 20: 11, 49: 11, 40: 11, 44: 11, 38: 11, 33: 11, 28: 11, 41: 11, 47: 11, 42: 11, 37: 11, 51: 11, 53: 11, 46: 11, 35: 11, 50: 11, 52: 11, 39: 12, 45: 11 },
    { },
    { },
    { 44: 11, 36: 11, 41: 11, 50: 11, 42: 11, 51: 11, 39: 11, 48: 11, 20: 11, 37: 11, 49: 11, 33: 11, 46: 11, 47: 11, 52: 11, 35: 11, 45: 36, 54: 11, 38: 11, 40: 11, 43: 11, 28: 11, 53: 11 },
    { 14: 16, 19: 38 },
    { },
    { 20: 11, 49: 11, 44: 11, 47: 11, 41: 11, 37: 11, 48: 11, 36: 11, 39: 11, 45: 11, 54: 11, 53: 11, 38: 11, 35: 11, 50: 11, 40: 11, 28: 11, 52: 11, 43: 11, 46: 2, 33: 11, 51: 11, 42: 11 },
    { 40: 11, 54: 11, 49: 11, 48: 11, 20: 11, 42: 11, 39: 11, 41: 11, 38: 11, 28: 11, 44: 11, 37: 11, 45: 11, 35: 31, 47: 11, 51: 11, 43: 11, 50: 11, 46: 11, 36: 11, 52: 11, 53: 11, 33: 11 },
    { 55: 52, 23: 52, 53: 52, 24: 52, 16: 52, 31: 52, 2: 52, 8: 52, 10: 52, 35: 52, 14: 52, 50: 52, 17: 52, 21: 52, 45: 52, 20: 52, 34: 52, 25: 52, 12: 52, 44: 52, 32: 52, 28: 52, 7: 52, 37: 52, 36: 52, 39: 52, 52: 52, 9: 49, 15: 52, 29: 52, 18: 52, 48: 52, 47: 52, 54: 52, 46: 52, 49: 52, 19: 52, 4: 52, 30: 23, 1: 52, 38: 52, 51: 52, 27: 52, 13: 52, 43: 52, 57: 52, 42: 52, 11: 52, 26: 52, 22: 52, 6: 52, 40: 52, 33: 52, 56: 52, 41: 52 },
    { 40: 11, 53: 11, 33: 11, 39: 11, 37: 11, 35: 11, 38: 11, 36: 11, 44: 11, 54: 11, 43: 11, 50: 11, 46: 11, 45: 11, 51: 11, 47: 24, 49: 11, 42: 11, 20: 11, 28: 11, 41: 11, 48: 11, 52: 11 },
    { 53: 11, 45: 11, 41: 11, 39: 11, 43: 11, 42: 11, 40: 11, 50: 11, 35: 11, 51: 11, 52: 9, 20: 11, 36: 11, 33: 11, 48: 11, 46: 11, 54: 11, 47: 11, 38: 11, 49: 11, 28: 11, 44: 11, 37: 11 },
    { },
    { },
}
var accept = map[int]TokenType { 39: IDENTIFIER, 42: IDENTIFIER, 29: LEFT, 2: TOKEN, 11: IDENTIFIER, 21: IDENTIFIER, 25: IDENTIFIER, 43: RULE, 45: DOT, 50: IDENTIFIER, 4: EQUAL, 8: FRAGMENT, 13: QUESTION, 14: IDENTIFIER, 19: STAR, 34: L_PAREN, 41: BAR, 9: RIGHT, 31: IDENTIFIER, 44: IDENTIFIER, 46: COMMENT, 53: IDENTIFIER, 3: SEMI, 24: IDENTIFIER, 26: IDENTIFIER, 35: COLON, 32: WHITESPACE, 33: CLASS, 1: IDENTIFIER, 5: IDENTIFIER, 12: IDENTIFIER, 28: PLUS, 36: IDENTIFIER, 49: STRING, 30: IDENTIFIER, 51: IDENTIFIER, 54: IDENTIFIER, 17: R_PAREN, 20: IDENTIFIER, 27: IDENTIFIER, 40: SKIP, 47: IDENTIFIER, 55: ARROW, 56: HASH, 6: IDENTIFIER, 15: EOF, 18: IDENTIFIER }

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
