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
    { 37: 24, 9: 6, 44: 21, 12: 37, 2: 55, 5: 55, 46: 24, 21: 53, 47: 24, 29: 45, 39: 24, 14: 34, 41: 24, 24: 12, 26: 25, 51: 13, 53: 24, 13: 43, 42: 24, 54: 24, 18: 26, 56: 15, 35: 24, 22: 2, 38: 35, 45: 24, 0: 40, 50: 28, 10: 8, 52: 44, 7: 55, 48: 24, 28: 24, 17: 11, 36: 24, 19: 56, 40: 24, 49: 24, 15: 20, 33: 24, 3: 55, 43: 24 },
    { 52: 24, 39: 24, 33: 24, 37: 24, 49: 24, 46: 24, 44: 24, 47: 24, 50: 24, 51: 24, 38: 24, 36: 24, 41: 24, 53: 24, 45: 24, 20: 24, 43: 24, 35: 24, 28: 24, 54: 24, 48: 49, 42: 24, 40: 24 },
    { },
    { 54: 24, 49: 24, 44: 24, 50: 24, 39: 24, 46: 24, 28: 24, 51: 24, 38: 24, 36: 24, 41: 24, 43: 24, 20: 24, 48: 24, 42: 24, 47: 24, 40: 24, 37: 39, 33: 24, 35: 24, 53: 24, 52: 24, 45: 24 },
    { 20: 24, 35: 24, 54: 24, 45: 24, 47: 24, 53: 24, 48: 24, 36: 24, 37: 24, 40: 24, 46: 24, 44: 24, 50: 24, 28: 24, 41: 24, 38: 24, 39: 24, 43: 24, 52: 31, 42: 24, 51: 24, 33: 24, 49: 24 },
    { 45: 24, 33: 24, 20: 24, 37: 24, 48: 24, 53: 24, 36: 24, 28: 24, 40: 24, 43: 24, 49: 24, 52: 24, 47: 24, 51: 24, 44: 17, 46: 24, 35: 24, 54: 24, 42: 24, 38: 24, 39: 24, 50: 24, 41: 24 },
    { 48: 6, 52: 6, 27: 6, 31: 6, 13: 6, 10: 6, 26: 6, 38: 6, 43: 6, 41: 6, 57: 6, 24: 6, 28: 6, 55: 6, 51: 6, 9: 52, 22: 6, 36: 6, 54: 6, 44: 6, 12: 6, 33: 6, 14: 6, 8: 6, 29: 6, 19: 6, 37: 6, 1: 6, 30: 41, 46: 6, 39: 6, 21: 6, 2: 6, 20: 6, 4: 6, 32: 6, 45: 6, 15: 6, 25: 6, 42: 6, 40: 6, 6: 6, 16: 6, 7: 6, 35: 6, 17: 6, 23: 6, 53: 6, 49: 6, 56: 6, 34: 6, 11: 6, 47: 6, 18: 6, 50: 6 },
    { 54: 24, 47: 24, 38: 24, 37: 24, 20: 24, 36: 24, 40: 24, 28: 24, 44: 24, 39: 24, 41: 24, 43: 24, 49: 24, 42: 24, 50: 24, 45: 24, 53: 24, 46: 24, 52: 24, 51: 24, 48: 24, 33: 24, 35: 24 },
    { },
    { 36: 24, 37: 24, 44: 24, 51: 24, 35: 24, 40: 24, 20: 24, 38: 24, 41: 24, 28: 24, 49: 24, 42: 24, 46: 24, 52: 30, 33: 24, 39: 24, 48: 24, 43: 24, 53: 24, 50: 24, 47: 24, 54: 24, 45: 24 },
    { 43: 24, 54: 24, 41: 24, 52: 24, 44: 24, 47: 24, 46: 24, 33: 24, 42: 24, 35: 24, 49: 24, 45: 24, 50: 24, 51: 24, 20: 24, 39: 24, 36: 24, 38: 24, 28: 24, 48: 24, 53: 24, 37: 24, 40: 24 },
    { 25: 51 },
    { },
    { 48: 24, 44: 24, 45: 24, 40: 24, 41: 24, 52: 24, 50: 24, 42: 24, 20: 24, 43: 14, 47: 24, 39: 24, 53: 24, 49: 24, 51: 24, 28: 24, 37: 24, 38: 24, 54: 24, 36: 24, 35: 24, 46: 24, 33: 24 },
    { 38: 24, 52: 24, 48: 24, 50: 24, 53: 24, 51: 24, 46: 24, 37: 24, 43: 24, 42: 24, 28: 24, 20: 24, 40: 24, 54: 24, 49: 24, 45: 24, 33: 24, 41: 1, 35: 24, 44: 24, 36: 24, 39: 24, 47: 24 },
    { },
    { 36: 24, 41: 24, 50: 24, 43: 24, 40: 24, 39: 24, 44: 24, 37: 24, 54: 24, 46: 24, 51: 24, 53: 24, 47: 24, 49: 24, 42: 24, 38: 24, 35: 24, 45: 24, 33: 24, 52: 24, 20: 24, 28: 24, 48: 24 },
    { 53: 24, 54: 24, 36: 24, 20: 24, 33: 24, 37: 10, 51: 24, 42: 24, 28: 24, 50: 24, 38: 24, 52: 24, 45: 24, 40: 24, 49: 24, 48: 24, 43: 24, 41: 24, 46: 24, 44: 24, 39: 24, 35: 24, 47: 24 },
    { 33: 24, 40: 24, 41: 24, 37: 24, 49: 24, 43: 24, 20: 24, 35: 24, 47: 24, 39: 23, 52: 24, 54: 24, 45: 24, 46: 24, 28: 24, 50: 24, 38: 24, 51: 24, 42: 24, 36: 24, 48: 24, 53: 24, 44: 24 },
    { 1: 29, 55: 29, 50: 29, 2: 29, 51: 29, 7: 29, 23: 29, 25: 29, 13: 29, 18: 29, 42: 29, 10: 29, 17: 29, 30: 29, 44: 29, 45: 29, 20: 29, 28: 29, 21: 29, 33: 29, 27: 29, 41: 29, 16: 29, 57: 29, 39: 29, 54: 29, 32: 29, 38: 29, 36: 29, 53: 29, 46: 29, 6: 29, 49: 29, 37: 29, 48: 29, 43: 29, 3: 29, 35: 29, 15: 29, 52: 29, 11: 29, 24: 29, 26: 29, 8: 29, 19: 46, 5: 29, 47: 29, 34: 29, 9: 29, 22: 29, 31: 29, 40: 29, 56: 29, 12: 29, 14: 29, 4: 29, 29: 29 },
    { },
    { 50: 24, 40: 24, 44: 24, 46: 24, 45: 24, 41: 24, 49: 24, 51: 24, 36: 24, 20: 24, 43: 24, 52: 24, 33: 24, 35: 24, 48: 24, 28: 24, 53: 24, 38: 24, 42: 24, 47: 24, 39: 24, 37: 42, 54: 24 },
    { 44: 24, 28: 24, 49: 24, 39: 27, 35: 24, 52: 24, 50: 24, 33: 24, 37: 24, 40: 24, 20: 24, 42: 24, 38: 24, 51: 24, 43: 24, 41: 24, 48: 24, 46: 24, 36: 24, 53: 24, 54: 24, 45: 24, 47: 24 },
    { 54: 24, 43: 24, 40: 50, 49: 24, 37: 24, 52: 24, 28: 24, 45: 24, 38: 24, 33: 24, 35: 24, 36: 24, 51: 24, 39: 24, 20: 24, 44: 24, 47: 24, 41: 24, 46: 24, 50: 24, 53: 24, 48: 24, 42: 24 },
    { 37: 24, 38: 24, 39: 24, 46: 24, 33: 24, 42: 24, 20: 24, 36: 24, 44: 24, 51: 24, 43: 24, 28: 24, 53: 24, 47: 24, 35: 24, 41: 24, 54: 24, 45: 24, 40: 24, 48: 24, 49: 24, 52: 24, 50: 24 },
    { },
    { },
    { 43: 24, 36: 24, 47: 24, 41: 24, 20: 24, 35: 24, 37: 24, 49: 24, 28: 24, 40: 24, 46: 24, 39: 24, 38: 24, 52: 24, 50: 24, 48: 24, 44: 24, 33: 24, 45: 3, 53: 24, 54: 24, 42: 24, 51: 24 },
    { 42: 24, 40: 24, 49: 24, 38: 24, 48: 24, 44: 24, 51: 24, 33: 24, 39: 24, 53: 5, 28: 24, 46: 24, 47: 24, 35: 24, 52: 24, 54: 24, 50: 24, 43: 24, 36: 24, 37: 24, 20: 24, 41: 18, 45: 24 },
    { 30: 29, 31: 29, 24: 29, 10: 29, 51: 29, 35: 29, 23: 29, 1: 29, 28: 29, 13: 29, 21: 29, 33: 29, 26: 29, 5: 29, 16: 29, 8: 29, 43: 29, 22: 29, 15: 29, 45: 29, 9: 29, 7: 29, 41: 29, 2: 29, 55: 29, 17: 29, 34: 29, 39: 29, 48: 29, 6: 29, 36: 29, 47: 29, 56: 29, 18: 29, 3: 29, 19: 29, 32: 29, 50: 29, 54: 29, 46: 29, 42: 29, 44: 29, 40: 29, 53: 29, 37: 29, 20: 29, 38: 29, 25: 29, 27: 29, 52: 29, 14: 19, 49: 29, 12: 29, 57: 29, 11: 29, 29: 29, 4: 29 },
    { 53: 24, 51: 24, 49: 24, 41: 24, 46: 24, 20: 24, 48: 24, 44: 24, 52: 24, 54: 24, 35: 24, 39: 24, 36: 24, 43: 24, 45: 24, 28: 24, 33: 24, 38: 24, 40: 24, 50: 24, 37: 24, 42: 24, 47: 24 },
    { 50: 24, 28: 24, 39: 24, 54: 24, 45: 24, 48: 24, 46: 24, 53: 24, 37: 24, 47: 24, 35: 24, 49: 24, 51: 24, 38: 24, 42: 24, 33: 24, 52: 24, 40: 24, 20: 24, 43: 24, 36: 24, 44: 24, 41: 24 },
    { 49: 24, 38: 24, 43: 24, 42: 24, 48: 24, 52: 24, 50: 24, 39: 24, 41: 24, 44: 24, 46: 16, 36: 24, 47: 24, 33: 24, 40: 24, 20: 24, 37: 24, 28: 24, 53: 24, 51: 24, 54: 24, 45: 24, 35: 24 },
    { 38: 33, 55: 33, 46: 33, 18: 33, 53: 33, 28: 33, 10: 33, 20: 33, 33: 33, 42: 33, 23: 33, 44: 33, 17: 33, 2: 33, 9: 33, 0: 46, 22: 33, 13: 33, 8: 33, 34: 33, 37: 33, 12: 33, 52: 33, 16: 33, 47: 33, 15: 33, 43: 33, 56: 33, 24: 33, 3: 46, 45: 33, 57: 33, 36: 33, 41: 33, 49: 33, 4: 33, 32: 33, 14: 33, 35: 33, 25: 33, 29: 33, 48: 33, 30: 33, 6: 33, 19: 33, 26: 33, 50: 33, 40: 33, 1: 33, 7: 33, 51: 33, 54: 33, 27: 33, 11: 33, 21: 33, 31: 33, 39: 33, 5: 46 },
    { },
    { 40: 24, 44: 24, 51: 24, 43: 24, 38: 24, 20: 24, 54: 24, 49: 24, 52: 24, 33: 24, 36: 24, 48: 24, 37: 24, 53: 24, 46: 24, 28: 24, 42: 24, 47: 24, 35: 24, 45: 24, 41: 24, 50: 54, 39: 24 },
    { },
    { },
    { 47: 45, 11: 45, 37: 45, 15: 45, 20: 45, 34: 45, 10: 45, 12: 45, 25: 45, 48: 45, 57: 45, 53: 45, 32: 45, 39: 45, 6: 45, 29: 45, 4: 45, 52: 45, 41: 45, 33: 45, 38: 45, 2: 45, 35: 45, 31: 45, 44: 45, 50: 45, 43: 45, 26: 45, 56: 45, 22: 45, 28: 45, 36: 45, 51: 45, 42: 45, 21: 45, 45: 45, 30: 45, 1: 45, 46: 45, 8: 45, 49: 45, 54: 45, 17: 45, 55: 45, 7: 45, 40: 45, 23: 45, 19: 45, 13: 45, 9: 45, 14: 45, 27: 45, 24: 45, 18: 45, 16: 45 },
    { 42: 24, 33: 24, 47: 24, 39: 24, 45: 24, 46: 4, 41: 24, 37: 24, 43: 24, 50: 24, 38: 24, 35: 24, 49: 24, 54: 24, 52: 24, 48: 24, 44: 24, 40: 24, 20: 24, 51: 24, 36: 24, 53: 24, 28: 24 },
    { },
    { 34: 6, 23: 6, 20: 6, 4: 6, 17: 6, 2: 6, 26: 6, 49: 6, 15: 6, 33: 6, 8: 6, 32: 6, 19: 6, 40: 6, 37: 6, 9: 6, 7: 6, 36: 6, 57: 6, 52: 6, 13: 6, 38: 6, 30: 6, 11: 6, 55: 6, 27: 6, 29: 6, 50: 6, 6: 6, 1: 6, 18: 6, 10: 6, 44: 6, 21: 6, 47: 6, 51: 6, 56: 6, 45: 6, 43: 6, 46: 6, 22: 6, 14: 6, 41: 6, 31: 6, 53: 6, 39: 6, 16: 6, 48: 6, 35: 6, 12: 6, 28: 6, 54: 6, 24: 6, 42: 6, 25: 6 },
    { 47: 24, 50: 24, 48: 24, 46: 24, 40: 24, 36: 24, 41: 24, 53: 24, 38: 9, 35: 24, 49: 24, 44: 24, 54: 24, 37: 24, 45: 24, 20: 24, 43: 24, 39: 24, 28: 24, 52: 24, 51: 24, 33: 24, 42: 24 },
    { },
    { 46: 24, 36: 24, 41: 24, 38: 24, 42: 24, 49: 24, 54: 24, 20: 24, 43: 24, 39: 24, 50: 24, 28: 24, 47: 47, 35: 24, 44: 24, 45: 24, 40: 24, 48: 24, 51: 24, 33: 24, 37: 24, 52: 24, 53: 24 },
    { 32: 45, 50: 45, 55: 45, 35: 45, 22: 45, 2: 45, 49: 45, 16: 45, 21: 45, 31: 36, 39: 45, 10: 45, 46: 45, 14: 45, 12: 45, 18: 45, 47: 45, 48: 45, 6: 45, 41: 45, 15: 45, 19: 45, 4: 45, 29: 45, 42: 45, 23: 45, 33: 45, 52: 45, 51: 45, 26: 45, 11: 45, 54: 45, 56: 45, 13: 45, 24: 45, 36: 45, 44: 45, 34: 45, 45: 45, 9: 45, 43: 45, 53: 45, 30: 38, 37: 45, 25: 45, 17: 45, 27: 45, 57: 45, 40: 45, 7: 45, 8: 45, 20: 45, 28: 45, 1: 45, 38: 45 },
    { },
    { 45: 24, 35: 24, 48: 24, 47: 24, 39: 24, 53: 24, 28: 24, 40: 24, 43: 48, 36: 24, 46: 24, 50: 24, 54: 24, 49: 24, 51: 24, 38: 24, 41: 24, 33: 24, 44: 24, 37: 24, 52: 24, 20: 24, 42: 24 },
    { 39: 24, 43: 24, 38: 24, 44: 24, 47: 24, 37: 32, 54: 24, 51: 24, 40: 24, 48: 24, 45: 24, 33: 24, 52: 24, 36: 24, 20: 24, 28: 24, 53: 24, 35: 24, 50: 24, 49: 24, 42: 24, 41: 24, 46: 24 },
    { 45: 24, 47: 24, 50: 24, 39: 24, 41: 24, 46: 24, 28: 24, 42: 24, 38: 24, 53: 24, 35: 24, 54: 24, 37: 24, 40: 24, 51: 24, 52: 24, 44: 24, 20: 24, 48: 24, 49: 24, 43: 24, 33: 24, 36: 24 },
    { 39: 24, 28: 24, 35: 24, 40: 24, 43: 24, 54: 24, 37: 24, 42: 24, 33: 24, 36: 24, 41: 24, 47: 24, 48: 24, 53: 24, 44: 24, 45: 24, 46: 24, 20: 24, 50: 24, 51: 24, 38: 24, 49: 24, 52: 7 },
    { },
    { },
    { },
    { 39: 24, 35: 22, 47: 24, 50: 24, 49: 24, 54: 24, 42: 24, 44: 24, 40: 24, 48: 24, 38: 24, 28: 24, 43: 24, 45: 24, 46: 24, 20: 24, 37: 24, 52: 24, 53: 24, 51: 24, 33: 24, 36: 24, 41: 24 },
    { 2: 55, 3: 55, 5: 55, 7: 55 },
    { 19: 33, 14: 29 },
}
var accept = map[int]TokenType { 48: 20, 1: 20, 8: 14, 12: 8, 31: 4, 50: 20, 7: 6, 21: 20, 23: 20, 25: 11, 36: 22, 49: 7, 3: 20, 15: 13, 17: 20, 27: 20, 43: 18, 52: 21, 54: 20, 5: 20, 14: 20, 16: 3, 24: 20, 35: 20, 39: 20, 22: 20, 32: 20, 42: 20, 47: 20, 53: 16, 13: 20, 20: 9, 30: 5, 40: 23, 51: 19, 55: 0, 18: 20, 26: 12, 34: 10, 37: 17, 44: 20, 46: 1, 2: 15, 4: 20, 9: 20, 10: 2, 28: 20 }

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
