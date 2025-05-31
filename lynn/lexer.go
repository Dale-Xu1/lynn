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
    { 45: 31, 24: 48, 37: 31, 50: 49, 47: 31, 13: 44, 52: 34, 40: 31, 43: 31, 26: 22, 53: 31, 33: 31, 42: 31, 46: 31, 17: 12, 3: 24, 10: 14, 15: 55, 9: 56, 56: 35, 35: 31, 14: 27, 19: 15, 36: 31, 41: 31, 38: 30, 28: 31, 18: 41, 29: 9, 2: 24, 5: 24, 44: 52, 12: 47, 49: 31, 48: 31, 7: 24, 21: 10, 51: 53, 39: 31, 22: 2, 0: 43, 54: 31 },
    { 48: 31, 49: 31, 51: 31, 47: 31, 36: 31, 53: 31, 42: 31, 44: 31, 37: 31, 54: 31, 41: 31, 46: 31, 33: 31, 52: 31, 20: 31, 45: 31, 28: 31, 35: 31, 38: 31, 39: 31, 40: 31, 43: 31, 50: 31 },
    { },
    { },
    { 46: 56, 55: 56, 12: 56, 7: 56, 48: 56, 16: 56, 54: 56, 2: 56, 18: 56, 25: 56, 30: 56, 33: 56, 50: 56, 53: 56, 56: 56, 28: 56, 22: 56, 13: 56, 11: 56, 27: 56, 31: 56, 10: 56, 35: 56, 24: 56, 26: 56, 52: 56, 29: 56, 49: 56, 32: 56, 45: 56, 4: 56, 34: 56, 40: 56, 9: 56, 20: 56, 44: 56, 15: 56, 8: 56, 36: 56, 23: 56, 39: 56, 14: 56, 57: 56, 47: 56, 38: 56, 51: 56, 41: 56, 1: 56, 37: 56, 17: 56, 6: 56, 43: 56, 21: 56, 42: 56, 19: 56 },
    { 36: 31, 53: 31, 40: 31, 54: 31, 48: 31, 28: 31, 52: 31, 43: 31, 44: 31, 39: 31, 45: 31, 37: 31, 33: 31, 49: 31, 47: 31, 41: 31, 51: 31, 46: 31, 20: 31, 50: 31, 35: 31, 42: 31, 38: 31 },
    { 51: 31, 38: 31, 33: 31, 49: 31, 40: 31, 39: 31, 45: 31, 20: 31, 52: 31, 42: 31, 37: 31, 53: 31, 47: 31, 36: 31, 43: 31, 44: 31, 50: 31, 46: 25, 54: 31, 35: 31, 41: 31, 28: 31, 48: 31 },
    { 50: 31, 39: 31, 46: 31, 53: 31, 38: 31, 33: 31, 40: 31, 36: 31, 54: 31, 41: 31, 43: 31, 49: 31, 47: 31, 51: 31, 45: 31, 44: 38, 20: 31, 52: 31, 37: 31, 35: 31, 42: 31, 28: 31, 48: 31 },
    { 20: 31, 46: 31, 48: 31, 43: 31, 47: 31, 40: 42, 33: 31, 35: 31, 49: 31, 42: 31, 38: 31, 36: 31, 50: 31, 52: 31, 54: 31, 37: 31, 28: 31, 45: 31, 53: 31, 41: 31, 51: 31, 44: 31, 39: 31 },
    { 47: 9, 50: 9, 40: 9, 9: 9, 11: 9, 42: 9, 20: 9, 35: 9, 53: 9, 39: 9, 34: 9, 49: 9, 29: 9, 10: 9, 28: 9, 37: 9, 44: 9, 23: 9, 41: 9, 2: 9, 43: 9, 54: 9, 51: 9, 45: 9, 30: 51, 14: 9, 4: 9, 46: 9, 17: 9, 31: 39, 36: 9, 19: 9, 13: 9, 21: 9, 1: 9, 52: 9, 25: 9, 26: 9, 27: 9, 12: 9, 48: 9, 55: 9, 33: 9, 24: 9, 16: 9, 15: 9, 22: 9, 8: 9, 18: 9, 32: 9, 38: 9, 7: 9, 6: 9, 56: 9, 57: 9 },
    { },
    { 38: 31, 41: 31, 45: 31, 46: 31, 40: 31, 52: 31, 50: 31, 48: 31, 42: 31, 33: 31, 39: 31, 28: 31, 35: 31, 20: 31, 51: 31, 53: 31, 37: 31, 36: 31, 43: 50, 47: 31, 49: 31, 54: 31, 44: 31 },
    { 25: 54 },
    { 42: 31, 33: 31, 47: 31, 46: 31, 52: 31, 53: 31, 54: 31, 44: 31, 39: 31, 41: 31, 45: 31, 35: 31, 48: 31, 37: 31, 28: 31, 20: 31, 40: 31, 50: 31, 51: 31, 38: 31, 43: 31, 49: 31, 36: 31 },
    { },
    { 14: 36, 19: 40 },
    { 45: 31, 44: 31, 51: 31, 54: 31, 28: 31, 49: 31, 47: 31, 37: 31, 42: 31, 50: 31, 33: 31, 36: 31, 40: 31, 43: 31, 38: 31, 46: 23, 35: 31, 39: 31, 48: 31, 52: 31, 20: 31, 53: 31, 41: 31 },
    { 36: 31, 53: 31, 20: 31, 44: 31, 46: 31, 35: 31, 39: 31, 49: 31, 51: 31, 43: 31, 38: 31, 33: 31, 52: 1, 40: 31, 48: 31, 54: 31, 28: 31, 42: 31, 50: 31, 41: 31, 47: 31, 37: 31, 45: 31 },
    { 42: 31, 43: 31, 37: 31, 36: 31, 20: 31, 46: 31, 28: 31, 39: 31, 45: 31, 54: 31, 50: 31, 35: 20, 51: 31, 49: 31, 40: 31, 41: 31, 47: 31, 33: 31, 44: 31, 53: 31, 38: 31, 48: 31, 52: 31 },
    { 35: 31, 45: 31, 28: 31, 20: 31, 54: 31, 49: 31, 47: 31, 44: 31, 53: 31, 48: 31, 51: 31, 52: 31, 36: 31, 33: 31, 43: 31, 39: 31, 37: 31, 38: 31, 41: 29, 46: 31, 40: 31, 50: 31, 42: 31 },
    { 36: 31, 33: 31, 53: 31, 46: 31, 20: 31, 51: 31, 37: 31, 50: 31, 38: 31, 45: 31, 42: 31, 54: 31, 28: 31, 52: 31, 35: 31, 40: 31, 44: 31, 41: 31, 48: 31, 39: 45, 49: 31, 43: 31, 47: 31 },
    { 50: 31, 49: 31, 41: 31, 20: 31, 38: 31, 48: 31, 33: 31, 39: 31, 51: 31, 54: 31, 36: 31, 40: 31, 46: 31, 35: 31, 42: 31, 44: 31, 37: 16, 52: 31, 53: 31, 47: 31, 43: 31, 45: 31, 28: 31 },
    { },
    { 36: 31, 53: 31, 54: 31, 48: 31, 52: 26, 49: 31, 42: 31, 51: 31, 38: 31, 20: 31, 50: 31, 37: 31, 46: 31, 33: 31, 41: 31, 47: 31, 35: 31, 39: 31, 45: 31, 40: 31, 44: 31, 43: 31, 28: 31 },
    { 5: 24, 7: 24, 2: 24, 3: 24 },
    { 41: 31, 49: 31, 42: 31, 43: 31, 51: 31, 20: 31, 38: 31, 48: 31, 44: 31, 54: 31, 52: 31, 36: 31, 28: 31, 33: 31, 37: 31, 45: 31, 53: 31, 46: 31, 40: 31, 35: 31, 39: 31, 47: 31, 50: 31 },
    { 49: 31, 42: 31, 46: 31, 43: 31, 53: 31, 28: 31, 33: 31, 35: 31, 39: 31, 52: 31, 36: 31, 38: 31, 47: 31, 54: 31, 20: 31, 48: 31, 45: 31, 40: 31, 44: 31, 50: 31, 51: 31, 37: 31, 41: 31 },
    { },
    { },
    { 20: 31, 48: 13, 42: 31, 39: 31, 45: 31, 35: 31, 53: 31, 41: 31, 50: 31, 46: 31, 49: 31, 33: 31, 40: 31, 54: 31, 28: 31, 36: 31, 37: 31, 51: 31, 43: 31, 44: 31, 52: 31, 47: 31, 38: 31 },
    { 45: 31, 42: 31, 50: 18, 46: 31, 47: 31, 44: 31, 48: 31, 28: 31, 43: 31, 49: 31, 51: 31, 54: 31, 38: 31, 20: 31, 52: 31, 40: 31, 41: 31, 53: 31, 35: 31, 39: 31, 36: 31, 37: 31, 33: 31 },
    { 51: 31, 53: 31, 40: 31, 28: 31, 49: 31, 20: 31, 50: 31, 48: 31, 54: 31, 44: 31, 33: 31, 45: 31, 38: 31, 35: 31, 52: 31, 41: 31, 39: 31, 42: 31, 46: 31, 37: 31, 36: 31, 43: 31, 47: 31 },
    { 45: 31, 54: 31, 44: 31, 35: 31, 46: 31, 43: 31, 37: 31, 33: 31, 39: 31, 36: 31, 41: 31, 48: 31, 49: 31, 47: 31, 28: 31, 40: 31, 50: 31, 53: 31, 38: 17, 52: 31, 42: 31, 20: 31, 51: 31 },
    { 54: 31, 42: 31, 37: 31, 39: 31, 51: 31, 20: 31, 28: 31, 33: 31, 48: 31, 38: 31, 46: 31, 52: 31, 47: 31, 40: 31, 53: 31, 35: 31, 50: 31, 36: 31, 49: 31, 41: 31, 43: 31, 44: 31, 45: 31 },
    { 54: 31, 20: 31, 36: 31, 33: 31, 49: 31, 28: 31, 47: 11, 51: 31, 41: 31, 43: 31, 38: 31, 48: 31, 53: 31, 39: 31, 45: 31, 50: 31, 35: 31, 44: 31, 46: 31, 42: 31, 37: 31, 52: 31, 40: 31 },
    { },
    { 54: 36, 14: 37, 3: 36, 49: 36, 6: 36, 29: 36, 55: 36, 19: 36, 5: 36, 45: 36, 53: 36, 7: 36, 4: 36, 9: 36, 48: 36, 2: 36, 52: 36, 38: 36, 50: 36, 37: 36, 1: 36, 41: 36, 32: 36, 27: 36, 8: 36, 26: 36, 57: 36, 39: 36, 46: 36, 51: 36, 24: 36, 15: 36, 13: 36, 22: 36, 23: 36, 35: 36, 16: 36, 33: 36, 47: 36, 11: 36, 31: 36, 56: 36, 30: 36, 44: 36, 12: 36, 21: 36, 42: 36, 43: 36, 36: 36, 34: 36, 28: 36, 40: 36, 10: 36, 17: 36, 25: 36, 20: 36, 18: 36 },
    { 21: 36, 33: 36, 55: 36, 19: 28, 14: 36, 30: 36, 57: 36, 53: 36, 41: 36, 15: 36, 3: 36, 36: 36, 31: 36, 56: 36, 42: 36, 8: 36, 18: 36, 25: 36, 37: 36, 29: 36, 44: 36, 24: 36, 7: 36, 40: 36, 20: 36, 51: 36, 5: 36, 54: 36, 10: 36, 23: 36, 9: 36, 4: 36, 46: 36, 47: 36, 12: 36, 27: 36, 28: 36, 26: 36, 32: 36, 39: 36, 34: 36, 49: 36, 50: 36, 35: 36, 48: 36, 52: 36, 16: 36, 13: 36, 22: 36, 6: 36, 1: 36, 11: 36, 17: 36, 43: 36, 38: 36, 2: 36, 45: 36 },
    { 41: 31, 28: 31, 42: 31, 35: 31, 47: 31, 51: 31, 45: 31, 33: 31, 36: 31, 53: 31, 38: 31, 44: 31, 54: 31, 20: 31, 50: 31, 40: 31, 39: 31, 48: 31, 43: 31, 52: 31, 46: 31, 37: 5, 49: 31 },
    { },
    { 0: 28, 39: 40, 25: 40, 34: 40, 14: 40, 8: 40, 1: 40, 44: 40, 10: 40, 2: 40, 57: 40, 54: 40, 6: 40, 37: 40, 15: 40, 23: 40, 50: 40, 11: 40, 48: 40, 4: 40, 22: 40, 46: 40, 29: 40, 3: 28, 47: 40, 17: 40, 20: 40, 5: 28, 21: 40, 49: 40, 24: 40, 33: 40, 28: 40, 53: 40, 55: 40, 52: 40, 26: 40, 43: 40, 18: 40, 12: 40, 45: 40, 38: 40, 31: 40, 19: 40, 51: 40, 13: 40, 27: 40, 42: 40, 9: 40, 30: 40, 16: 40, 40: 40, 41: 40, 32: 40, 56: 40, 36: 40, 35: 40, 7: 40 },
    { },
    { 43: 31, 44: 31, 46: 31, 33: 31, 40: 31, 48: 31, 50: 31, 53: 31, 28: 31, 52: 33, 54: 31, 51: 31, 41: 31, 20: 31, 38: 31, 35: 31, 36: 31, 45: 31, 47: 31, 42: 31, 37: 31, 39: 31, 49: 31 },
    { },
    { },
    { 42: 31, 39: 31, 28: 31, 47: 31, 46: 31, 37: 31, 33: 31, 35: 31, 45: 21, 38: 31, 51: 31, 50: 31, 41: 31, 49: 31, 53: 31, 48: 31, 52: 31, 36: 31, 20: 31, 44: 31, 54: 31, 43: 31, 40: 31 },
    { 50: 31, 52: 31, 53: 31, 39: 8, 47: 31, 36: 31, 35: 31, 45: 31, 33: 31, 54: 31, 49: 31, 42: 31, 44: 31, 41: 31, 40: 31, 20: 31, 28: 31, 46: 31, 37: 31, 38: 31, 48: 31, 51: 31, 43: 31 },
    { },
    { },
    { 50: 31, 51: 31, 40: 31, 42: 31, 20: 31, 48: 31, 43: 31, 37: 31, 47: 31, 38: 31, 45: 31, 28: 31, 36: 31, 33: 31, 53: 7, 41: 46, 46: 31, 44: 31, 52: 31, 35: 31, 54: 31, 39: 31, 49: 31 },
    { 43: 31, 20: 31, 37: 6, 44: 31, 54: 31, 51: 31, 49: 31, 50: 31, 53: 31, 45: 31, 38: 31, 41: 31, 48: 31, 52: 31, 28: 31, 46: 31, 35: 31, 36: 31, 39: 31, 40: 31, 47: 31, 33: 31, 42: 31 },
    { 17: 9, 15: 9, 10: 9, 16: 9, 34: 9, 40: 9, 25: 9, 55: 9, 39: 9, 4: 9, 30: 9, 11: 9, 47: 9, 6: 9, 19: 9, 43: 9, 8: 9, 24: 9, 52: 9, 20: 9, 56: 9, 27: 9, 12: 9, 13: 9, 42: 9, 37: 9, 28: 9, 33: 9, 57: 9, 31: 9, 45: 9, 21: 9, 48: 9, 50: 9, 23: 9, 9: 9, 54: 9, 29: 9, 44: 9, 7: 9, 14: 9, 32: 9, 46: 9, 51: 9, 22: 9, 18: 9, 36: 9, 41: 9, 26: 9, 2: 9, 1: 9, 35: 9, 49: 9, 53: 9, 38: 9 },
    { 35: 31, 53: 31, 28: 31, 33: 31, 48: 31, 52: 31, 20: 31, 51: 31, 36: 31, 43: 31, 41: 31, 38: 31, 44: 31, 50: 31, 46: 31, 49: 31, 42: 31, 45: 31, 37: 32, 47: 31, 39: 31, 40: 31, 54: 31 },
    { 40: 31, 50: 31, 38: 31, 53: 31, 33: 31, 39: 31, 51: 31, 28: 31, 46: 31, 37: 31, 41: 31, 49: 31, 48: 31, 44: 31, 36: 31, 54: 31, 35: 31, 52: 31, 43: 19, 45: 31, 47: 31, 42: 31, 20: 31 },
    { },
    { },
    { 16: 56, 46: 56, 7: 56, 25: 56, 27: 56, 54: 56, 28: 56, 26: 56, 24: 56, 44: 56, 51: 56, 36: 56, 43: 56, 17: 56, 13: 56, 53: 56, 11: 56, 42: 56, 6: 56, 37: 56, 55: 56, 2: 56, 33: 56, 35: 56, 56: 56, 21: 56, 4: 56, 39: 56, 8: 56, 57: 56, 29: 56, 38: 56, 47: 56, 45: 56, 1: 56, 19: 56, 31: 56, 23: 56, 30: 4, 48: 56, 15: 56, 32: 56, 40: 56, 18: 56, 12: 56, 41: 56, 9: 3, 10: 56, 50: 56, 49: 56, 22: 56, 34: 56, 52: 56, 20: 56, 14: 56 },
}
var accept = map[int]TokenType { 49: 20, 11: 20, 23: 20, 35: 13, 42: 20, 27: 10, 48: 8, 53: 20, 54: 19, 17: 20, 34: 20, 43: 23, 46: 20, 55: 9, 30: 20, 3: 21, 8: 20, 20: 20, 33: 6, 38: 20, 47: 17, 2: 15, 19: 20, 32: 20, 39: 22, 41: 12, 45: 20, 52: 20, 1: 5, 21: 20, 25: 3, 26: 4, 31: 20, 50: 20, 5: 2, 6: 20, 7: 20, 10: 16, 13: 7, 14: 14, 22: 11, 24: 0, 16: 20, 18: 20, 28: 1, 29: 20, 44: 18 }

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
