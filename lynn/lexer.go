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
    { 42: 36, 40: 36, 7: 1, 21: 38, 2: 1, 0: 54, 22: 2, 43: 36, 19: 14, 3: 1, 50: 49, 54: 36, 38: 17, 5: 1, 17: 24, 51: 4, 44: 21, 49: 36, 14: 29, 33: 36, 39: 36, 18: 5, 52: 46, 35: 36, 56: 55, 12: 41, 28: 36, 9: 35, 29: 33, 37: 36, 15: 10, 26: 53, 53: 36, 13: 23, 46: 36, 24: 19, 47: 36, 41: 36, 10: 27, 48: 36, 45: 36, 36: 36 },
    { 7: 1, 2: 1, 3: 1, 5: 1 },
    { },
    { 6: 26, 54: 26, 31: 26, 52: 26, 12: 26, 22: 26, 48: 26, 15: 26, 11: 26, 16: 26, 14: 26, 42: 26, 50: 26, 43: 26, 57: 26, 33: 26, 5: 26, 7: 26, 18: 26, 51: 26, 4: 26, 44: 26, 36: 26, 1: 26, 38: 26, 45: 26, 49: 26, 19: 7, 23: 26, 21: 26, 55: 26, 29: 26, 2: 26, 46: 26, 8: 26, 40: 26, 53: 26, 20: 26, 17: 26, 13: 26, 41: 26, 47: 26, 26: 26, 24: 26, 32: 26, 34: 26, 39: 26, 56: 26, 28: 26, 30: 26, 10: 26, 3: 26, 25: 26, 27: 26, 37: 26, 9: 26, 35: 26 },
    { 40: 36, 33: 36, 35: 36, 43: 31, 48: 36, 53: 36, 36: 36, 47: 36, 20: 36, 51: 36, 42: 36, 46: 36, 39: 36, 49: 36, 54: 36, 37: 36, 28: 36, 44: 36, 52: 36, 41: 36, 50: 36, 45: 36, 38: 36 },
    { },
    { 33: 36, 40: 36, 48: 36, 39: 36, 35: 36, 46: 36, 44: 36, 42: 36, 37: 36, 47: 36, 41: 36, 43: 36, 52: 36, 49: 36, 54: 36, 38: 36, 51: 36, 28: 36, 50: 36, 53: 36, 20: 36, 36: 36, 45: 36 },
    { },
    { 8: 8, 16: 8, 11: 8, 20: 8, 31: 8, 55: 8, 13: 8, 53: 8, 51: 8, 44: 8, 56: 8, 27: 8, 48: 8, 47: 8, 23: 8, 39: 8, 17: 8, 4: 8, 36: 8, 2: 8, 35: 8, 45: 8, 15: 8, 57: 8, 21: 8, 7: 8, 22: 8, 54: 8, 24: 8, 50: 8, 25: 8, 5: 7, 9: 8, 38: 8, 19: 8, 26: 8, 34: 8, 30: 8, 3: 7, 6: 8, 41: 8, 32: 8, 14: 8, 40: 8, 12: 8, 46: 8, 42: 8, 18: 8, 29: 8, 33: 8, 37: 8, 1: 8, 28: 8, 49: 8, 10: 8, 43: 8, 0: 7, 52: 8 },
    { 46: 36, 28: 36, 43: 36, 20: 36, 33: 36, 39: 36, 35: 36, 50: 36, 37: 36, 47: 36, 40: 15, 51: 36, 45: 36, 36: 36, 41: 36, 48: 36, 38: 36, 44: 36, 54: 36, 53: 36, 49: 36, 52: 36, 42: 36 },
    { },
    { 28: 36, 50: 36, 43: 36, 36: 36, 44: 36, 20: 36, 41: 36, 35: 36, 52: 36, 33: 36, 38: 36, 48: 36, 49: 36, 54: 36, 46: 36, 40: 36, 37: 36, 42: 36, 45: 43, 53: 36, 51: 36, 47: 36, 39: 36 },
    { 37: 36, 45: 36, 47: 36, 52: 36, 50: 36, 51: 36, 35: 36, 36: 36, 39: 36, 49: 36, 43: 36, 44: 36, 40: 36, 53: 36, 48: 36, 42: 36, 20: 36, 38: 36, 46: 44, 54: 36, 28: 36, 33: 36, 41: 36 },
    { 37: 36, 45: 36, 38: 32, 50: 36, 35: 36, 48: 36, 20: 36, 36: 36, 42: 36, 46: 36, 47: 36, 53: 36, 41: 36, 43: 36, 44: 36, 33: 36, 51: 36, 54: 36, 49: 36, 39: 36, 52: 36, 28: 36, 40: 36 },
    { 19: 8, 14: 26 },
    { 37: 36, 44: 36, 50: 36, 45: 36, 41: 36, 28: 36, 39: 36, 40: 36, 47: 36, 48: 36, 42: 36, 33: 36, 53: 36, 49: 36, 20: 36, 52: 30, 43: 36, 54: 36, 35: 36, 51: 36, 38: 36, 36: 36, 46: 36 },
    { 35: 36, 41: 36, 36: 36, 49: 36, 53: 36, 48: 36, 54: 36, 40: 36, 43: 36, 37: 36, 42: 36, 44: 36, 38: 36, 45: 36, 52: 36, 20: 36, 46: 36, 39: 36, 50: 36, 28: 36, 47: 36, 51: 36, 33: 36 },
    { 54: 36, 38: 36, 28: 36, 49: 36, 42: 36, 47: 36, 44: 36, 46: 36, 33: 36, 45: 36, 41: 36, 35: 36, 51: 36, 52: 36, 53: 36, 48: 36, 43: 36, 36: 36, 37: 36, 20: 36, 40: 36, 50: 28, 39: 36 },
    { },
    { },
    { },
    { 52: 36, 40: 36, 35: 36, 33: 36, 42: 36, 20: 36, 41: 36, 43: 36, 51: 36, 28: 36, 38: 36, 50: 36, 49: 36, 44: 36, 39: 36, 36: 36, 37: 13, 54: 36, 53: 36, 45: 36, 46: 36, 47: 36, 48: 36 },
    { 53: 36, 45: 36, 40: 36, 51: 36, 54: 36, 41: 36, 52: 36, 33: 36, 46: 6, 20: 36, 28: 36, 44: 36, 39: 36, 49: 36, 50: 36, 42: 36, 47: 36, 48: 36, 35: 36, 37: 36, 36: 36, 38: 36, 43: 36 },
    { },
    { 25: 20 },
    { 41: 33, 1: 33, 10: 33, 56: 33, 53: 33, 16: 33, 22: 33, 32: 33, 55: 33, 15: 33, 11: 33, 43: 33, 44: 33, 24: 33, 49: 33, 50: 33, 17: 33, 42: 33, 57: 33, 23: 33, 6: 33, 12: 33, 14: 33, 35: 33, 4: 33, 39: 33, 21: 33, 20: 33, 36: 33, 2: 33, 34: 33, 40: 33, 46: 33, 38: 33, 28: 33, 48: 33, 30: 33, 18: 33, 19: 33, 9: 33, 45: 33, 54: 33, 29: 33, 52: 33, 26: 33, 13: 33, 33: 33, 31: 33, 7: 33, 51: 33, 27: 33, 25: 33, 37: 33, 8: 33, 47: 33 },
    { 21: 26, 48: 26, 57: 26, 32: 26, 17: 26, 53: 26, 7: 26, 6: 26, 44: 26, 13: 26, 14: 3, 55: 26, 50: 26, 51: 26, 19: 26, 29: 26, 41: 26, 40: 26, 35: 26, 9: 26, 16: 26, 8: 26, 46: 26, 25: 26, 12: 26, 43: 26, 39: 26, 33: 26, 23: 26, 20: 26, 28: 26, 11: 26, 56: 26, 10: 26, 5: 26, 24: 26, 27: 26, 54: 26, 42: 26, 4: 26, 31: 26, 18: 26, 47: 26, 26: 26, 30: 26, 3: 26, 38: 26, 36: 26, 52: 26, 1: 26, 22: 26, 34: 26, 49: 26, 2: 26, 37: 26, 15: 26, 45: 26 },
    { },
    { 47: 36, 38: 36, 43: 36, 36: 36, 46: 36, 28: 36, 20: 36, 39: 36, 49: 36, 52: 36, 45: 36, 44: 36, 33: 36, 40: 36, 35: 40, 51: 36, 53: 36, 54: 36, 42: 36, 41: 36, 50: 36, 48: 36, 37: 36 },
    { },
    { 46: 36, 37: 36, 20: 36, 53: 36, 39: 36, 41: 36, 48: 36, 33: 36, 44: 36, 38: 36, 49: 36, 35: 36, 51: 36, 36: 36, 42: 36, 50: 36, 52: 36, 28: 36, 43: 36, 54: 36, 45: 36, 47: 36, 40: 36 },
    { 52: 36, 45: 36, 53: 36, 40: 36, 49: 36, 35: 36, 46: 36, 37: 36, 50: 36, 39: 36, 48: 36, 44: 36, 42: 36, 20: 36, 43: 36, 47: 36, 51: 36, 41: 45, 38: 36, 54: 36, 36: 36, 28: 36, 33: 36 },
    { 43: 36, 35: 36, 51: 36, 37: 36, 33: 36, 49: 36, 47: 36, 44: 36, 52: 16, 53: 36, 46: 36, 54: 36, 40: 36, 50: 36, 36: 36, 48: 36, 45: 36, 41: 36, 42: 36, 20: 36, 28: 36, 38: 36, 39: 36 },
    { 46: 33, 52: 33, 29: 33, 42: 33, 27: 33, 34: 33, 50: 33, 41: 33, 18: 33, 15: 33, 9: 33, 24: 33, 53: 33, 51: 33, 11: 33, 1: 33, 25: 33, 30: 25, 6: 33, 12: 33, 2: 33, 8: 33, 20: 33, 35: 33, 13: 33, 56: 33, 40: 33, 32: 33, 54: 33, 4: 33, 7: 33, 28: 33, 19: 33, 37: 33, 23: 33, 36: 33, 57: 33, 48: 33, 21: 33, 43: 33, 31: 48, 49: 33, 17: 33, 38: 33, 47: 33, 44: 33, 55: 33, 10: 33, 16: 33, 14: 33, 45: 33, 26: 33, 22: 33, 39: 33, 33: 33 },
    { 42: 36, 20: 36, 40: 36, 54: 36, 33: 36, 37: 22, 38: 36, 35: 36, 52: 36, 53: 36, 41: 36, 47: 36, 49: 36, 43: 36, 51: 36, 48: 36, 50: 36, 45: 36, 39: 36, 36: 36, 44: 36, 46: 36, 28: 36 },
    { 53: 35, 15: 35, 12: 35, 27: 35, 38: 35, 47: 35, 7: 35, 31: 35, 16: 35, 17: 35, 19: 35, 30: 56, 25: 35, 20: 35, 41: 35, 55: 35, 45: 35, 43: 35, 6: 35, 57: 35, 36: 35, 35: 35, 11: 35, 51: 35, 28: 35, 39: 35, 4: 35, 2: 35, 13: 35, 1: 35, 40: 35, 34: 35, 23: 35, 26: 35, 46: 35, 52: 35, 56: 35, 21: 35, 10: 35, 32: 35, 49: 35, 33: 35, 14: 35, 22: 35, 48: 35, 18: 35, 42: 35, 50: 35, 54: 35, 37: 35, 24: 35, 9: 18, 29: 35, 8: 35, 44: 35 },
    { 47: 36, 49: 36, 52: 36, 28: 36, 46: 36, 54: 36, 51: 36, 45: 36, 39: 36, 50: 36, 33: 36, 43: 36, 37: 36, 44: 36, 41: 36, 20: 36, 35: 36, 40: 36, 42: 36, 38: 36, 53: 36, 48: 36, 36: 36 },
    { 54: 36, 47: 36, 42: 36, 28: 36, 33: 36, 37: 36, 51: 36, 53: 36, 38: 36, 43: 36, 36: 36, 41: 36, 20: 36, 52: 36, 46: 36, 40: 36, 49: 36, 35: 36, 39: 36, 45: 36, 50: 36, 48: 36, 44: 36 },
    { },
    { 49: 36, 20: 36, 28: 36, 36: 36, 44: 50, 35: 36, 33: 36, 51: 36, 53: 36, 50: 36, 45: 36, 41: 36, 43: 36, 46: 36, 37: 36, 47: 36, 38: 36, 52: 36, 40: 36, 48: 36, 42: 36, 54: 36, 39: 36 },
    { 54: 36, 33: 36, 41: 36, 20: 36, 37: 36, 45: 36, 35: 36, 47: 36, 39: 11, 40: 36, 48: 36, 43: 36, 49: 36, 44: 36, 38: 36, 51: 36, 52: 36, 53: 36, 28: 36, 50: 36, 42: 36, 46: 36, 36: 36 },
    { },
    { 37: 36, 48: 36, 20: 36, 50: 36, 44: 36, 52: 36, 49: 36, 47: 36, 40: 36, 46: 36, 43: 36, 42: 36, 28: 36, 36: 36, 33: 36, 38: 36, 41: 36, 35: 36, 45: 36, 39: 9, 54: 36, 51: 36, 53: 36 },
    { 44: 36, 45: 36, 53: 36, 20: 36, 49: 36, 52: 36, 35: 36, 46: 36, 28: 36, 48: 36, 43: 36, 40: 36, 47: 36, 33: 36, 51: 36, 54: 36, 37: 12, 39: 36, 36: 36, 42: 36, 38: 36, 50: 36, 41: 36 },
    { 35: 36, 43: 36, 44: 36, 48: 36, 52: 52, 54: 36, 45: 36, 49: 36, 51: 36, 38: 36, 50: 36, 37: 36, 20: 36, 47: 36, 46: 36, 40: 36, 53: 36, 33: 36, 36: 36, 28: 36, 39: 36, 41: 36, 42: 36 },
    { 45: 36, 37: 36, 48: 37, 50: 36, 52: 36, 42: 36, 49: 36, 36: 36, 54: 36, 28: 36, 40: 36, 38: 36, 51: 36, 43: 36, 41: 36, 44: 36, 46: 36, 47: 36, 35: 36, 33: 36, 20: 36, 53: 36, 39: 36 },
    { 41: 36, 42: 36, 45: 36, 40: 36, 50: 36, 36: 36, 35: 36, 49: 36, 53: 36, 44: 36, 43: 36, 28: 36, 37: 36, 47: 47, 33: 36, 39: 36, 48: 36, 46: 36, 52: 36, 38: 36, 54: 36, 51: 36, 20: 36 },
    { 39: 36, 41: 36, 48: 36, 44: 36, 20: 36, 43: 34, 37: 36, 40: 36, 53: 36, 38: 36, 52: 36, 49: 36, 54: 36, 46: 36, 35: 36, 36: 36, 28: 36, 51: 36, 45: 36, 50: 36, 47: 36, 33: 36, 42: 36 },
    { },
    { 37: 36, 44: 36, 28: 36, 33: 36, 43: 36, 38: 36, 42: 36, 49: 36, 45: 36, 50: 36, 51: 36, 53: 39, 54: 36, 48: 36, 41: 42, 46: 36, 40: 36, 20: 36, 39: 36, 36: 36, 35: 36, 47: 36, 52: 36 },
    { 39: 36, 49: 36, 45: 36, 35: 36, 33: 36, 44: 36, 28: 36, 37: 51, 51: 36, 52: 36, 41: 36, 47: 36, 50: 36, 36: 36, 42: 36, 20: 36, 40: 36, 48: 36, 43: 36, 46: 36, 54: 36, 53: 36, 38: 36 },
    { 53: 36, 38: 36, 36: 36, 44: 36, 54: 36, 33: 36, 41: 36, 48: 36, 35: 36, 42: 36, 46: 36, 50: 36, 52: 36, 40: 36, 45: 36, 47: 36, 43: 36, 51: 36, 20: 36, 28: 36, 37: 36, 39: 36, 49: 36 },
    { 20: 36, 44: 36, 49: 36, 43: 36, 54: 36, 45: 36, 53: 36, 35: 36, 47: 36, 37: 36, 42: 36, 46: 36, 38: 36, 40: 36, 50: 36, 33: 36, 52: 36, 39: 36, 41: 36, 48: 36, 36: 36, 51: 36, 28: 36 },
    { },
    { },
    { },
    { 1: 35, 20: 35, 34: 35, 16: 35, 56: 35, 21: 35, 50: 35, 42: 35, 24: 35, 30: 35, 32: 35, 8: 35, 44: 35, 29: 35, 28: 35, 57: 35, 37: 35, 43: 35, 53: 35, 55: 35, 54: 35, 7: 35, 48: 35, 11: 35, 38: 35, 12: 35, 2: 35, 31: 35, 17: 35, 9: 35, 13: 35, 18: 35, 40: 35, 26: 35, 36: 35, 45: 35, 46: 35, 27: 35, 4: 35, 14: 35, 39: 35, 6: 35, 41: 35, 47: 35, 19: 35, 15: 35, 22: 35, 52: 35, 49: 35, 23: 35, 35: 35, 33: 35, 51: 35, 10: 35, 25: 35 },
}
var accept = map[int]TokenType { 6: 3, 7: 1, 9: 20, 15: 20, 17: 20, 27: 14, 19: 8, 32: 20, 50: 20, 4: 20, 47: 20, 51: 2, 1: 0, 10: 9, 13: 20, 21: 20, 29: 10, 30: 6, 36: 20, 37: 7, 2: 15, 16: 5, 22: 20, 31: 20, 38: 16, 39: 20, 40: 20, 43: 20, 18: 21, 23: 18, 34: 20, 41: 17, 46: 20, 48: 22, 54: 23, 55: 13, 5: 12, 12: 20, 20: 19, 28: 20, 52: 4, 53: 11, 11: 20, 42: 20, 44: 20, 45: 20, 49: 20 }

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
