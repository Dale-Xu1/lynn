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
    Type       TokenType
    Value      string
    Start, End Location
}

// Represents a range between characters.
type Range struct { Min, Max rune }

const (WHITESPACE TokenType = iota; COMMENT; RULE; TOKEN; FRAGMENT; LEFT; RIGHT; SKIP; EQUAL; PLUS; STAR; QUESTION; DOT; BAR; HASH; SEMI; COLON; L_PAREN; R_PAREN; ARROW; IDENTIFIER; STRING; CLASS; EOF)
func (t TokenType) String() string { return typeName[t] }
var typeName = map[TokenType]string { 0: "WHITESPACE", 1: "COMMENT", 2: "RULE", 3: "TOKEN", 4: "FRAGMENT", 5: "LEFT", 6: "RIGHT", 7: "SKIP", 8: "EQUAL", 9: "PLUS", 10: "STAR", 11: "QUESTION", 12: "DOT", 13: "BAR", 14: "HASH", 15: "SEMI", 16: "COLON", 17: "L_PAREN", 18: "R_PAREN", 19: "ARROW", 20: "IDENTIFIER", 21: "STRING", 22: "CLASS", 23: "EOF" }
var skip = map[TokenType]struct{} { 0: {}, 1: {} }

var ranges = []Range { { '\x00', '\x00' }, { '\x01', '\b' }, { '\t', '\t' }, { '\n', '\n' }, { '\v', '\f' }, { '\r', '\r' }, { '\x0e', '\x1f' }, { ' ', ' ' }, { '!', '!' }, { '"', '"' }, { '#', '#' }, { '$', '\'' }, { '(', '(' }, { ')', ')' }, { '*', '*' }, { '+', '+' }, { ',', ',' }, { '-', '-' }, { '.', '.' }, { '/', '/' }, { '0', '9' }, { ':', ':' }, { ';', ';' }, { '<', '<' }, { '=', '=' }, { '>', '>' }, { '?', '?' }, { '@', '@' }, { 'A', 'Z' }, { '[', '[' }, { '\\', '\\' }, { ']', ']' }, { '^', '^' }, { '_', '_' }, { '`', '`' }, { 'a', 'a' }, { 'b', 'd' }, { 'e', 'e' }, { 'f', 'f' }, { 'g', 'g' }, { 'h', 'h' }, { 'i', 'i' }, { 'j', 'j' }, { 'k', 'k' }, { 'l', 'l' }, { 'm', 'm' }, { 'n', 'n' }, { 'o', 'o' }, { 'p', 'p' }, { 'q', 'q' }, { 'r', 'r' }, { 's', 's' }, { 't', 't' }, { 'u', 'u' }, { 'v', 'z' }, { '{', '{' }, { '|', '|' }, { '}', '\U0010ffff' } }
var transitions = []map[int]int {
    { 42: 43, 46: 43, 10: 16, 35: 43, 7: 11, 38: 25, 15: 26, 50: 6, 39: 43, 14: 27, 47: 43, 56: 33, 5: 11, 12: 21, 48: 43, 52: 8, 2: 11, 45: 43, 43: 43, 40: 43, 51: 34, 36: 43, 3: 11, 0: 35, 19: 28, 22: 13, 44: 19, 53: 43, 33: 43, 37: 43, 26: 10, 9: 49, 49: 43, 13: 54, 28: 43, 29: 1, 41: 43, 18: 42, 17: 22, 24: 45, 21: 46, 54: 43 },
    { 10: 1, 18: 1, 35: 1, 27: 1, 28: 1, 16: 1, 25: 1, 55: 1, 30: 50, 1: 1, 12: 1, 37: 1, 39: 1, 7: 1, 48: 1, 50: 1, 54: 1, 17: 1, 23: 1, 22: 1, 40: 1, 11: 1, 24: 1, 42: 1, 46: 1, 4: 1, 52: 1, 14: 1, 41: 1, 15: 1, 29: 1, 20: 1, 13: 1, 57: 1, 19: 1, 34: 1, 53: 1, 31: 5, 2: 1, 36: 1, 47: 1, 51: 1, 6: 1, 38: 1, 8: 1, 33: 1, 43: 1, 21: 1, 56: 1, 9: 1, 49: 1, 32: 1, 44: 1, 45: 1, 26: 1 },
    { },
    { 42: 3, 20: 3, 8: 3, 51: 3, 34: 3, 46: 3, 19: 3, 54: 3, 35: 3, 31: 3, 47: 3, 22: 3, 48: 3, 17: 3, 10: 3, 36: 3, 16: 3, 14: 3, 0: 32, 37: 3, 41: 3, 50: 3, 9: 3, 43: 3, 5: 32, 24: 3, 32: 3, 30: 3, 33: 3, 25: 3, 23: 3, 12: 3, 55: 3, 6: 3, 3: 32, 39: 3, 11: 3, 21: 3, 15: 3, 44: 3, 7: 3, 28: 3, 49: 3, 26: 3, 45: 3, 18: 3, 1: 3, 38: 3, 4: 3, 40: 3, 13: 3, 56: 3, 2: 3, 29: 3, 53: 3, 27: 3, 52: 3, 57: 3 },
    { 47: 43, 54: 43, 45: 43, 35: 43, 48: 43, 37: 43, 42: 43, 44: 43, 53: 43, 40: 43, 38: 24, 50: 43, 51: 43, 43: 43, 33: 43, 20: 43, 52: 43, 49: 43, 28: 43, 39: 43, 46: 43, 41: 43, 36: 43 },
    { },
    { 52: 43, 42: 43, 38: 43, 40: 43, 35: 43, 36: 43, 50: 43, 45: 43, 44: 43, 53: 7, 51: 43, 33: 43, 46: 43, 43: 43, 20: 43, 54: 43, 39: 43, 41: 36, 37: 43, 47: 43, 49: 43, 28: 43, 48: 43 },
    { 51: 43, 54: 43, 28: 43, 33: 43, 45: 43, 47: 43, 53: 43, 40: 43, 50: 43, 20: 43, 48: 43, 49: 43, 44: 15, 52: 43, 46: 43, 39: 43, 43: 43, 36: 43, 41: 43, 35: 43, 37: 43, 38: 43, 42: 43 },
    { 40: 43, 51: 43, 20: 43, 39: 43, 50: 43, 28: 43, 33: 43, 53: 43, 44: 43, 54: 43, 49: 43, 43: 43, 52: 43, 36: 43, 41: 43, 35: 43, 48: 43, 45: 43, 42: 43, 46: 43, 38: 43, 47: 37, 37: 43 },
    { 28: 43, 42: 43, 40: 43, 53: 43, 36: 43, 38: 43, 48: 43, 41: 43, 52: 43, 39: 43, 35: 43, 43: 43, 37: 43, 49: 43, 20: 43, 33: 43, 45: 43, 47: 43, 50: 43, 51: 43, 54: 43, 46: 43, 44: 43 },
    { },
    { 5: 11, 7: 11, 2: 11, 3: 11 },
    { 14: 56, 8: 56, 18: 56, 13: 56, 27: 56, 43: 56, 49: 56, 34: 56, 48: 56, 35: 56, 38: 56, 21: 56, 40: 56, 6: 56, 2: 56, 44: 56, 25: 56, 30: 56, 26: 56, 31: 56, 50: 56, 12: 56, 15: 56, 9: 56, 10: 56, 36: 56, 33: 56, 20: 56, 24: 56, 3: 56, 51: 56, 11: 56, 22: 56, 29: 56, 1: 56, 39: 56, 32: 56, 55: 56, 52: 56, 45: 56, 47: 56, 46: 56, 56: 56, 54: 56, 19: 32, 16: 56, 42: 56, 53: 56, 41: 56, 37: 56, 7: 56, 5: 56, 57: 56, 23: 56, 4: 56, 28: 56, 17: 56 },
    { },
    { 53: 43, 37: 43, 50: 43, 28: 43, 52: 43, 46: 43, 43: 43, 44: 43, 38: 43, 41: 43, 35: 43, 42: 43, 47: 43, 48: 43, 40: 43, 33: 43, 39: 43, 54: 43, 49: 43, 36: 43, 45: 18, 51: 43, 20: 43 },
    { 43: 43, 53: 43, 28: 43, 44: 43, 46: 43, 41: 43, 33: 43, 20: 43, 52: 43, 54: 43, 49: 43, 36: 43, 35: 43, 40: 43, 37: 44, 42: 43, 51: 43, 39: 43, 45: 43, 48: 43, 50: 43, 38: 43, 47: 43 },
    { },
    { 51: 43, 38: 43, 33: 43, 35: 43, 44: 43, 52: 43, 41: 43, 28: 43, 47: 43, 48: 43, 54: 43, 36: 43, 50: 43, 42: 43, 40: 43, 37: 43, 53: 43, 39: 14, 49: 43, 43: 43, 46: 43, 45: 43, 20: 43 },
    { 33: 43, 50: 43, 44: 43, 45: 43, 35: 43, 37: 47, 42: 43, 49: 43, 40: 43, 20: 43, 28: 43, 36: 43, 48: 43, 43: 43, 39: 43, 41: 43, 52: 43, 46: 43, 53: 43, 47: 43, 51: 43, 54: 43, 38: 43 },
    { 54: 43, 46: 43, 38: 43, 49: 43, 48: 43, 20: 43, 52: 43, 45: 43, 33: 43, 50: 43, 43: 43, 36: 43, 47: 43, 39: 43, 41: 43, 35: 43, 44: 43, 28: 43, 40: 43, 37: 4, 42: 43, 53: 43, 51: 43 },
    { 48: 43, 41: 43, 49: 43, 20: 43, 51: 43, 53: 43, 39: 43, 35: 43, 43: 43, 38: 43, 28: 43, 40: 52, 36: 43, 47: 43, 42: 43, 50: 43, 52: 43, 54: 43, 45: 43, 33: 43, 46: 43, 37: 43, 44: 43 },
    { },
    { 25: 2 },
    { 54: 43, 36: 43, 43: 43, 49: 43, 51: 43, 28: 43, 45: 43, 38: 43, 33: 43, 48: 43, 52: 43, 46: 43, 42: 43, 39: 43, 40: 43, 41: 48, 37: 43, 50: 43, 35: 43, 53: 43, 47: 43, 44: 43, 20: 43 },
    { 47: 43, 51: 43, 52: 53, 54: 43, 20: 43, 50: 43, 33: 43, 36: 43, 38: 43, 40: 43, 46: 43, 35: 43, 42: 43, 53: 43, 37: 43, 41: 43, 28: 43, 48: 43, 45: 43, 43: 43, 49: 43, 44: 43, 39: 43 },
    { 53: 43, 44: 43, 46: 43, 38: 43, 37: 43, 20: 43, 50: 55, 43: 43, 41: 43, 48: 43, 36: 43, 51: 43, 28: 43, 40: 43, 39: 43, 33: 43, 35: 43, 42: 43, 47: 43, 54: 43, 45: 43, 49: 43, 52: 43 },
    { },
    { },
    { 19: 3, 14: 56 },
    { 33: 43, 40: 43, 47: 43, 48: 43, 39: 43, 36: 43, 52: 51, 41: 43, 51: 43, 54: 43, 35: 43, 49: 43, 45: 43, 42: 43, 20: 43, 46: 43, 50: 43, 37: 43, 44: 43, 28: 43, 43: 43, 53: 43, 38: 43 },
    { 48: 43, 43: 43, 41: 43, 40: 43, 53: 43, 52: 43, 37: 43, 42: 43, 20: 43, 46: 43, 38: 43, 50: 43, 35: 43, 36: 43, 49: 43, 39: 43, 47: 43, 44: 43, 28: 43, 45: 43, 33: 43, 51: 43, 54: 43 },
    { 52: 43, 43: 43, 53: 43, 54: 43, 44: 43, 47: 43, 46: 43, 41: 43, 49: 43, 35: 43, 20: 43, 28: 43, 40: 43, 36: 43, 42: 43, 38: 43, 48: 43, 37: 38, 39: 43, 45: 43, 50: 43, 33: 43, 51: 43 },
    { },
    { },
    { 42: 43, 48: 43, 28: 43, 53: 43, 54: 43, 46: 43, 40: 43, 45: 43, 50: 43, 52: 43, 43: 23, 51: 43, 33: 43, 36: 43, 20: 43, 38: 43, 49: 43, 44: 43, 41: 43, 47: 43, 39: 43, 35: 43, 37: 43 },
    { },
    { 36: 43, 52: 43, 39: 20, 43: 43, 51: 43, 54: 43, 46: 43, 42: 43, 20: 43, 41: 43, 47: 43, 33: 43, 40: 43, 49: 43, 37: 43, 53: 43, 38: 43, 45: 43, 48: 43, 50: 43, 35: 43, 28: 43, 44: 43 },
    { 36: 43, 43: 31, 46: 43, 33: 43, 20: 43, 53: 43, 38: 43, 42: 43, 52: 43, 44: 43, 50: 43, 28: 43, 39: 43, 48: 43, 54: 43, 41: 43, 51: 43, 45: 43, 37: 43, 35: 43, 49: 43, 47: 43, 40: 43 },
    { 45: 43, 50: 43, 44: 43, 36: 43, 48: 43, 37: 43, 20: 43, 40: 43, 33: 43, 52: 43, 41: 43, 38: 43, 42: 43, 43: 43, 35: 43, 39: 43, 54: 43, 28: 43, 49: 43, 51: 43, 47: 43, 53: 43, 46: 9 },
    { 37: 43, 42: 43, 49: 43, 50: 43, 40: 43, 33: 43, 52: 43, 28: 43, 45: 43, 38: 43, 35: 43, 54: 43, 39: 43, 48: 43, 47: 43, 20: 43, 36: 43, 44: 43, 41: 43, 51: 43, 46: 43, 53: 43, 43: 43 },
    { 48: 49, 43: 49, 52: 49, 11: 49, 32: 49, 38: 49, 10: 49, 50: 49, 18: 49, 7: 49, 25: 49, 23: 49, 15: 49, 20: 49, 40: 49, 1: 49, 56: 49, 41: 49, 51: 49, 2: 49, 28: 49, 16: 49, 30: 49, 31: 49, 33: 49, 6: 49, 17: 49, 45: 49, 22: 49, 29: 49, 35: 49, 39: 49, 54: 49, 9: 49, 26: 49, 27: 49, 8: 49, 53: 49, 12: 49, 37: 49, 14: 49, 42: 49, 44: 49, 46: 49, 34: 49, 24: 49, 57: 49, 13: 49, 19: 49, 4: 49, 21: 49, 49: 49, 47: 49, 36: 49, 55: 49 },
    { },
    { },
    { 28: 43, 37: 43, 45: 43, 51: 43, 35: 43, 36: 43, 43: 43, 47: 43, 48: 43, 42: 43, 33: 43, 46: 43, 39: 43, 53: 43, 54: 43, 44: 43, 40: 43, 49: 43, 20: 43, 38: 43, 52: 43, 50: 43, 41: 43 },
    { 41: 43, 20: 43, 46: 43, 45: 43, 52: 43, 53: 43, 49: 43, 48: 43, 40: 43, 50: 43, 43: 43, 28: 43, 47: 43, 51: 43, 33: 43, 37: 43, 44: 43, 39: 43, 42: 43, 38: 43, 54: 43, 35: 43, 36: 43 },
    { },
    { },
    { 35: 43, 49: 43, 54: 43, 33: 43, 36: 43, 45: 43, 20: 43, 41: 43, 53: 43, 37: 43, 48: 43, 52: 43, 40: 43, 43: 43, 39: 43, 44: 43, 51: 43, 42: 43, 46: 29, 47: 43, 50: 43, 38: 43, 28: 43 },
    { 38: 43, 35: 43, 41: 43, 33: 43, 49: 43, 39: 43, 48: 39, 43: 43, 52: 43, 40: 43, 50: 43, 44: 43, 37: 43, 53: 43, 20: 43, 54: 43, 28: 43, 36: 43, 45: 43, 51: 43, 47: 43, 42: 43, 46: 43 },
    { 52: 49, 31: 49, 34: 49, 14: 49, 16: 49, 53: 49, 36: 49, 28: 49, 12: 49, 20: 49, 57: 49, 56: 49, 23: 49, 42: 49, 13: 49, 19: 49, 21: 49, 43: 49, 30: 40, 6: 49, 25: 49, 35: 49, 37: 49, 54: 49, 27: 49, 44: 49, 41: 49, 29: 49, 39: 49, 49: 49, 26: 49, 7: 49, 17: 49, 2: 49, 47: 49, 33: 49, 4: 49, 11: 49, 38: 49, 40: 49, 55: 49, 10: 49, 18: 49, 24: 49, 32: 49, 48: 49, 8: 49, 51: 49, 46: 49, 15: 49, 50: 49, 9: 41, 45: 49, 1: 49, 22: 49 },
    { 11: 1, 53: 1, 38: 1, 46: 1, 49: 1, 35: 1, 1: 1, 24: 1, 13: 1, 25: 1, 26: 1, 52: 1, 41: 1, 44: 1, 23: 1, 39: 1, 56: 1, 47: 1, 19: 1, 10: 1, 16: 1, 31: 1, 7: 1, 4: 1, 17: 1, 43: 1, 20: 1, 50: 1, 32: 1, 22: 1, 27: 1, 34: 1, 51: 1, 28: 1, 36: 1, 12: 1, 2: 1, 6: 1, 30: 1, 33: 1, 37: 1, 9: 1, 18: 1, 8: 1, 29: 1, 14: 1, 40: 1, 21: 1, 54: 1, 42: 1, 57: 1, 15: 1, 48: 1, 45: 1, 55: 1 },
    { 33: 43, 48: 43, 43: 43, 20: 43, 36: 43, 35: 43, 37: 43, 51: 43, 52: 43, 40: 43, 44: 43, 54: 43, 39: 43, 46: 43, 41: 43, 45: 43, 47: 43, 38: 43, 42: 43, 49: 43, 53: 43, 28: 43, 50: 43 },
    { 28: 43, 42: 43, 48: 43, 50: 43, 52: 30, 43: 43, 53: 43, 46: 43, 51: 43, 36: 43, 37: 43, 33: 43, 40: 43, 35: 43, 44: 43, 54: 43, 38: 43, 41: 43, 45: 43, 20: 43, 39: 43, 49: 43, 47: 43 },
    { 44: 43, 41: 43, 36: 43, 39: 43, 42: 43, 20: 43, 49: 43, 52: 43, 54: 43, 46: 43, 40: 43, 53: 43, 45: 43, 35: 43, 50: 43, 48: 43, 28: 43, 43: 43, 51: 43, 38: 43, 33: 43, 37: 43, 47: 43 },
    { },
    { 42: 43, 37: 43, 41: 43, 51: 43, 44: 43, 38: 43, 39: 43, 45: 43, 54: 43, 28: 43, 35: 17, 47: 43, 46: 43, 53: 43, 33: 43, 36: 43, 50: 43, 20: 43, 48: 43, 43: 43, 49: 43, 40: 43, 52: 43 },
    { 14: 12, 16: 56, 57: 56, 51: 56, 40: 56, 2: 56, 5: 56, 7: 56, 10: 56, 43: 56, 30: 56, 55: 56, 8: 56, 32: 56, 53: 56, 26: 56, 50: 56, 33: 56, 12: 56, 31: 56, 3: 56, 34: 56, 45: 56, 13: 56, 28: 56, 1: 56, 24: 56, 29: 56, 39: 56, 36: 56, 49: 56, 41: 56, 44: 56, 42: 56, 17: 56, 56: 56, 48: 56, 19: 56, 25: 56, 4: 56, 35: 56, 22: 56, 18: 56, 46: 56, 23: 56, 20: 56, 11: 56, 54: 56, 37: 56, 9: 56, 6: 56, 38: 56, 21: 56, 52: 56, 27: 56, 47: 56, 15: 56 },
}
var accept = map[int]TokenType { 18: 20, 23: 20, 39: 7, 48: 20, 15: 20, 20: 20, 38: 20, 43: 20, 4: 20, 34: 20, 36: 20, 37: 20, 51: 4, 8: 20, 11: 0, 19: 20, 25: 20, 27: 10, 45: 8, 55: 20, 6: 20, 7: 20, 13: 15, 16: 14, 24: 20, 30: 6, 33: 13, 41: 21, 10: 11, 17: 20, 21: 17, 29: 20, 32: 1, 42: 12, 46: 16, 53: 5, 2: 19, 5: 22, 31: 20, 35: 23, 44: 2, 47: 20, 52: 20, 9: 3, 14: 20, 26: 9, 54: 18 }

// Lexer struct. Produces token stream.
type Lexer struct {
    Token   Token
    stream  *InputStream
    handler LexerErrorHandler
}

// Input stream struct. Produces character stream.
type InputStream struct {
    reader        *bufio.Reader
    location      Location
    buffer, stack []streamData
}
type streamData struct { char rune; location Location }

// Function called when the lexer encounters an error. Expected to bring input stream to synchronization point.
type LexerErrorHandler func (stream *InputStream, char rune, location Location)
var DEFAULT_LEXER_HANDLER = func (stream *InputStream, char rune, location Location) {
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
func NewLexer(reader *bufio.Reader, handler LexerErrorHandler) *Lexer {
    stream := &InputStream { reader, Location { 1, 1 }, make([]streamData, 0), make([]streamData, 0) }
    lexer := &Lexer { Token { }, stream, handler }
    lexer.Next()
    return lexer
}

// Emits next token in stream.
func (l *Lexer) Next() Token {
    start := l.stream.location
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
    end := l.stream.stack[len(l.stream.stack) - 1].location
    l.stream.reset()
    if _, ok := skip[token]; ok { return l.Next() } // Skip token
    // Create token and store as current token
    l.Token = Token { token, string(input[:i]), start, end }
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
func (i *InputStream) synchronize(handler LexerErrorHandler, char rune, location Location) {
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

// FOR DEBUG PURPOSES:
// Consumes all tokens emitted by lexer and prints them to the standard output.
func (l *Lexer) PrintTokens() {
    for {
        location := fmt.Sprintf("%d:%d-%d:%d", l.Token.Start.Line, l.Token.Start.Col, l.Token.End.Line, l.Token.End.Col)
        fmt.Printf("%-16s | %-16s %-16s\n", location, l.Token.Type, l.Token.Value)
        if l.Token.Type == EOF { break }
        l.Next()
    }
}
