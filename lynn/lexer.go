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
    { 28: 27, 17: 24, 22: 41, 37: 27, 26: 47, 48: 27, 0: 7, 36: 27, 13: 17, 15: 18, 51: 10, 39: 27, 50: 11, 38: 35, 45: 27, 3: 48, 43: 27, 7: 48, 21: 29, 56: 3, 54: 27, 29: 36, 5: 48, 10: 45, 9: 52, 53: 27, 12: 46, 19: 4, 35: 27, 42: 27, 14: 9, 49: 27, 24: 40, 18: 20, 2: 48, 46: 27, 47: 27, 52: 21, 40: 27, 33: 27, 44: 19, 41: 27 },
    { 28: 27, 54: 27, 39: 27, 42: 27, 35: 27, 46: 27, 20: 27, 37: 2, 44: 27, 41: 27, 33: 27, 48: 27, 47: 27, 50: 27, 36: 27, 43: 27, 51: 27, 38: 27, 49: 27, 45: 27, 53: 27, 40: 27, 52: 27 },
    { 46: 27, 53: 27, 49: 27, 42: 27, 47: 27, 51: 27, 54: 27, 50: 27, 28: 27, 37: 27, 52: 27, 41: 27, 20: 27, 44: 27, 38: 27, 36: 27, 33: 27, 48: 27, 43: 27, 45: 27, 39: 27, 35: 27, 40: 27 },
    { },
    { 19: 13, 14: 5 },
    { 56: 5, 40: 5, 33: 5, 52: 5, 25: 5, 2: 5, 48: 5, 23: 5, 27: 5, 19: 5, 29: 5, 39: 5, 11: 5, 8: 5, 22: 5, 42: 5, 21: 5, 50: 5, 14: 14, 31: 5, 26: 5, 49: 5, 1: 5, 28: 5, 57: 5, 4: 5, 38: 5, 13: 5, 30: 5, 46: 5, 54: 5, 43: 5, 47: 5, 24: 5, 10: 5, 17: 5, 36: 5, 45: 5, 35: 5, 12: 5, 5: 5, 7: 5, 55: 5, 51: 5, 3: 5, 34: 5, 18: 5, 9: 5, 20: 5, 37: 5, 41: 5, 16: 5, 32: 5, 44: 5, 6: 5, 53: 5, 15: 5 },
    { 28: 27, 41: 27, 33: 27, 38: 27, 51: 27, 49: 27, 42: 27, 20: 27, 35: 27, 44: 27, 54: 27, 52: 27, 45: 27, 48: 27, 36: 27, 40: 27, 46: 27, 50: 27, 47: 27, 43: 27, 53: 27, 37: 38, 39: 27 },
    { },
    { 38: 27, 42: 27, 43: 27, 54: 27, 33: 27, 50: 27, 37: 27, 51: 27, 53: 27, 44: 27, 47: 27, 28: 27, 39: 27, 35: 27, 49: 27, 40: 27, 41: 27, 48: 27, 52: 32, 20: 27, 36: 27, 45: 27, 46: 27 },
    { },
    { 44: 27, 38: 27, 53: 27, 42: 27, 47: 27, 43: 30, 49: 27, 54: 27, 39: 27, 50: 27, 36: 27, 46: 27, 20: 27, 40: 27, 41: 27, 37: 27, 45: 27, 35: 27, 52: 27, 51: 27, 48: 27, 28: 27, 33: 27 },
    { 38: 27, 28: 27, 37: 27, 39: 27, 48: 27, 36: 27, 46: 27, 20: 27, 44: 27, 42: 27, 54: 27, 33: 27, 52: 27, 53: 16, 35: 27, 45: 27, 50: 27, 40: 27, 47: 27, 49: 27, 41: 55, 43: 27, 51: 27 },
    { 52: 27, 39: 27, 42: 27, 45: 27, 53: 27, 48: 27, 43: 27, 51: 27, 50: 27, 37: 27, 36: 27, 40: 27, 41: 27, 35: 23, 33: 27, 28: 27, 44: 27, 49: 27, 47: 27, 20: 27, 46: 27, 38: 27, 54: 27 },
    { 20: 13, 5: 53, 25: 13, 16: 13, 34: 13, 12: 13, 1: 13, 24: 13, 54: 13, 17: 13, 48: 13, 32: 13, 39: 13, 21: 13, 37: 13, 27: 13, 9: 13, 8: 13, 33: 13, 55: 13, 2: 13, 44: 13, 31: 13, 13: 13, 28: 13, 11: 13, 42: 13, 0: 53, 29: 13, 19: 13, 14: 13, 26: 13, 30: 13, 6: 13, 18: 13, 43: 13, 45: 13, 38: 13, 23: 13, 53: 13, 3: 53, 41: 13, 49: 13, 40: 13, 15: 13, 10: 13, 52: 13, 7: 13, 56: 13, 22: 13, 47: 13, 4: 13, 36: 13, 51: 13, 35: 13, 46: 13, 57: 13, 50: 13 },
    { 27: 5, 48: 5, 52: 5, 11: 5, 35: 5, 16: 5, 24: 5, 8: 5, 9: 5, 4: 5, 45: 5, 43: 5, 10: 5, 3: 5, 56: 5, 38: 5, 31: 5, 2: 5, 41: 5, 30: 5, 49: 5, 17: 5, 54: 5, 40: 5, 51: 5, 5: 5, 23: 5, 7: 5, 15: 5, 25: 5, 57: 5, 34: 5, 1: 5, 47: 5, 6: 5, 42: 5, 37: 5, 12: 5, 19: 53, 21: 5, 39: 5, 22: 5, 50: 5, 13: 5, 33: 5, 36: 5, 20: 5, 46: 5, 32: 5, 55: 5, 44: 5, 26: 5, 53: 5, 18: 5, 14: 5, 28: 5, 29: 5 },
    { },
    { 37: 27, 47: 27, 50: 27, 28: 27, 52: 27, 36: 27, 45: 27, 39: 27, 48: 27, 43: 27, 40: 27, 35: 27, 33: 27, 54: 27, 42: 27, 44: 1, 51: 27, 41: 27, 20: 27, 53: 27, 38: 27, 49: 27, 46: 27 },
    { },
    { },
    { 20: 27, 49: 27, 48: 27, 38: 27, 39: 27, 52: 27, 46: 27, 43: 27, 33: 27, 40: 27, 35: 27, 44: 27, 36: 27, 45: 27, 51: 27, 53: 27, 42: 27, 41: 27, 47: 27, 54: 27, 37: 37, 50: 27, 28: 27 },
    { },
    { 33: 27, 37: 27, 52: 27, 20: 27, 40: 27, 54: 27, 35: 27, 43: 27, 39: 27, 36: 27, 51: 27, 42: 27, 47: 22, 41: 27, 28: 27, 49: 27, 46: 27, 48: 27, 45: 27, 38: 27, 53: 27, 44: 27, 50: 27 },
    { 50: 27, 38: 27, 40: 27, 46: 27, 49: 27, 53: 27, 35: 27, 37: 27, 41: 27, 51: 27, 54: 27, 52: 27, 33: 27, 36: 27, 39: 27, 20: 27, 44: 27, 48: 27, 28: 27, 42: 27, 43: 6, 45: 27, 47: 27 },
    { 36: 27, 41: 27, 51: 27, 52: 27, 49: 27, 53: 27, 37: 27, 33: 27, 44: 27, 40: 27, 28: 27, 38: 27, 35: 27, 42: 27, 43: 27, 54: 27, 50: 27, 47: 27, 48: 27, 45: 27, 39: 39, 46: 27, 20: 27 },
    { 25: 15 },
    { 45: 27, 38: 27, 48: 54, 39: 27, 47: 27, 52: 27, 33: 27, 43: 27, 28: 27, 37: 27, 54: 27, 35: 27, 49: 27, 20: 27, 36: 27, 50: 27, 42: 27, 44: 27, 41: 27, 46: 27, 51: 27, 53: 27, 40: 27 },
    { 6: 52, 24: 52, 16: 52, 31: 52, 37: 52, 40: 52, 21: 52, 17: 52, 39: 52, 55: 52, 56: 52, 7: 52, 46: 52, 19: 52, 44: 52, 15: 52, 43: 52, 26: 52, 51: 52, 32: 52, 49: 52, 13: 52, 14: 52, 54: 52, 23: 52, 45: 52, 29: 52, 38: 52, 57: 52, 30: 52, 28: 52, 48: 52, 52: 52, 53: 52, 8: 52, 4: 52, 25: 52, 9: 52, 27: 52, 12: 52, 35: 52, 22: 52, 18: 52, 42: 52, 50: 52, 34: 52, 1: 52, 41: 52, 20: 52, 36: 52, 47: 52, 2: 52, 10: 52, 11: 52, 33: 52 },
    { 51: 27, 48: 27, 36: 27, 45: 27, 33: 27, 40: 27, 47: 27, 42: 27, 37: 27, 54: 27, 38: 27, 28: 27, 50: 27, 35: 27, 46: 27, 53: 27, 44: 27, 41: 27, 52: 27, 39: 27, 43: 27, 49: 27, 20: 27 },
    { 43: 27, 33: 27, 36: 27, 37: 27, 49: 27, 45: 27, 51: 27, 50: 27, 46: 8, 42: 27, 48: 27, 20: 27, 38: 27, 47: 27, 53: 27, 28: 27, 54: 27, 39: 27, 40: 27, 35: 27, 52: 27, 44: 27, 41: 27 },
    { },
    { 46: 27, 39: 27, 40: 27, 33: 27, 38: 27, 37: 27, 20: 27, 48: 27, 52: 27, 49: 27, 44: 27, 35: 27, 47: 27, 36: 27, 50: 27, 42: 27, 41: 25, 51: 27, 28: 27, 45: 27, 54: 27, 43: 27, 53: 27 },
    { 49: 27, 40: 27, 41: 27, 43: 27, 54: 27, 47: 27, 50: 27, 52: 27, 35: 27, 37: 27, 48: 27, 39: 27, 20: 27, 51: 27, 38: 27, 46: 27, 33: 27, 44: 27, 53: 27, 42: 27, 36: 27, 45: 27, 28: 27 },
    { 33: 27, 42: 27, 41: 27, 51: 27, 49: 27, 37: 27, 48: 27, 35: 27, 47: 27, 36: 27, 44: 27, 54: 27, 53: 27, 40: 27, 45: 27, 52: 27, 50: 27, 43: 27, 46: 27, 38: 27, 39: 27, 20: 27, 28: 27 },
    { 20: 36, 35: 36, 50: 36, 32: 36, 49: 36, 8: 36, 24: 36, 11: 36, 51: 36, 1: 36, 55: 36, 29: 36, 33: 36, 21: 36, 41: 36, 56: 36, 26: 36, 42: 36, 12: 36, 46: 36, 57: 36, 38: 36, 47: 36, 23: 36, 4: 36, 25: 36, 31: 36, 14: 36, 22: 36, 34: 36, 54: 36, 18: 36, 36: 36, 2: 36, 44: 36, 27: 36, 37: 36, 6: 36, 28: 36, 13: 36, 19: 36, 52: 36, 45: 36, 7: 36, 39: 36, 48: 36, 17: 36, 43: 36, 53: 36, 30: 36, 40: 36, 15: 36, 9: 36, 10: 36, 16: 36 },
    { },
    { 52: 27, 48: 27, 28: 27, 35: 27, 42: 27, 44: 27, 46: 27, 39: 27, 36: 27, 47: 27, 53: 27, 20: 27, 50: 12, 49: 27, 54: 27, 33: 27, 51: 27, 41: 27, 38: 27, 45: 27, 37: 27, 40: 27, 43: 27 },
    { 43: 36, 13: 36, 23: 36, 4: 36, 46: 36, 11: 36, 47: 36, 51: 36, 49: 36, 54: 36, 17: 36, 44: 36, 16: 36, 42: 36, 27: 36, 36: 36, 28: 36, 6: 36, 18: 36, 33: 36, 41: 36, 1: 36, 19: 36, 39: 36, 7: 36, 12: 36, 21: 36, 24: 36, 55: 36, 29: 36, 32: 36, 30: 33, 25: 36, 31: 51, 14: 36, 45: 36, 35: 36, 53: 36, 2: 36, 50: 36, 10: 36, 9: 36, 57: 36, 48: 36, 22: 36, 15: 36, 56: 36, 40: 36, 26: 36, 52: 36, 38: 36, 20: 36, 37: 36, 34: 36, 8: 36 },
    { 47: 27, 39: 27, 52: 27, 53: 27, 33: 27, 49: 27, 50: 27, 20: 27, 51: 27, 38: 49, 44: 27, 48: 27, 35: 27, 43: 27, 46: 27, 40: 27, 45: 27, 41: 27, 28: 27, 42: 27, 36: 27, 54: 27, 37: 27 },
    { 47: 27, 43: 27, 42: 27, 45: 27, 28: 27, 33: 27, 36: 27, 35: 27, 48: 27, 54: 27, 40: 27, 20: 27, 41: 27, 53: 27, 37: 27, 46: 31, 38: 27, 39: 27, 49: 27, 44: 27, 51: 27, 50: 27, 52: 27 },
    { 53: 27, 33: 27, 48: 27, 54: 27, 37: 27, 47: 27, 50: 27, 28: 27, 38: 27, 52: 27, 45: 43, 41: 27, 42: 27, 40: 27, 43: 27, 39: 27, 20: 27, 46: 27, 49: 27, 35: 27, 51: 27, 44: 27, 36: 27 },
    { },
    { },
    { 20: 27, 36: 27, 40: 27, 45: 27, 52: 27, 43: 27, 39: 27, 42: 27, 41: 27, 53: 27, 33: 27, 37: 27, 50: 27, 44: 27, 54: 27, 38: 27, 28: 27, 35: 27, 51: 27, 48: 27, 47: 27, 46: 27, 49: 27 },
    { 20: 27, 46: 27, 33: 27, 53: 27, 39: 27, 28: 27, 37: 28, 47: 27, 48: 27, 40: 27, 41: 27, 44: 27, 50: 27, 38: 27, 51: 27, 42: 27, 45: 27, 52: 27, 43: 27, 35: 27, 54: 27, 49: 27, 36: 27 },
    { 48: 27, 51: 27, 54: 27, 33: 27, 44: 27, 36: 27, 45: 27, 46: 27, 38: 27, 52: 42, 37: 27, 47: 27, 41: 27, 40: 27, 42: 27, 43: 27, 39: 27, 20: 27, 50: 27, 28: 27, 35: 27, 49: 27, 53: 27 },
    { },
    { },
    { },
    { 2: 48, 3: 48, 5: 48, 7: 48 },
    { 39: 27, 48: 27, 38: 27, 20: 27, 49: 27, 41: 27, 33: 27, 44: 27, 28: 27, 40: 27, 51: 27, 47: 27, 52: 56, 50: 27, 43: 27, 45: 27, 53: 27, 35: 27, 54: 27, 36: 27, 46: 27, 42: 27, 37: 27 },
    { 45: 27, 35: 27, 41: 27, 54: 27, 38: 27, 47: 27, 51: 27, 48: 27, 39: 27, 44: 27, 49: 27, 46: 27, 52: 27, 28: 27, 36: 27, 42: 27, 50: 27, 53: 27, 37: 27, 33: 27, 40: 44, 43: 27, 20: 27 },
    { },
    { 29: 52, 15: 52, 34: 52, 39: 52, 53: 52, 4: 52, 22: 52, 52: 52, 16: 52, 50: 52, 23: 52, 21: 52, 18: 52, 35: 52, 12: 52, 36: 52, 7: 52, 41: 52, 2: 52, 47: 52, 45: 52, 19: 52, 11: 52, 57: 52, 13: 52, 55: 52, 40: 52, 26: 52, 10: 52, 9: 34, 51: 52, 48: 52, 25: 52, 6: 52, 46: 52, 8: 52, 32: 52, 56: 52, 20: 52, 49: 52, 33: 52, 24: 52, 17: 52, 28: 52, 43: 52, 54: 52, 37: 52, 14: 52, 31: 52, 44: 52, 30: 26, 38: 52, 42: 52, 1: 52, 27: 52 },
    { },
    { 43: 27, 28: 27, 44: 27, 36: 27, 37: 27, 46: 27, 53: 27, 35: 27, 40: 27, 42: 27, 48: 27, 20: 27, 52: 27, 49: 27, 51: 27, 47: 27, 41: 27, 33: 27, 39: 27, 50: 27, 45: 27, 54: 27, 38: 27 },
    { 35: 27, 46: 27, 51: 27, 42: 27, 40: 27, 39: 50, 49: 27, 44: 27, 36: 27, 43: 27, 38: 27, 33: 27, 47: 27, 48: 27, 45: 27, 37: 27, 41: 27, 28: 27, 20: 27, 54: 27, 52: 27, 50: 27, 53: 27 },
    { 39: 27, 50: 27, 20: 27, 36: 27, 46: 27, 35: 27, 37: 27, 47: 27, 41: 27, 33: 27, 44: 27, 38: 27, 53: 27, 49: 27, 42: 27, 45: 27, 48: 27, 51: 27, 52: 27, 40: 27, 54: 27, 28: 27, 43: 27 },
}
var accept = map[int]TokenType { 31: 3, 6: 20, 7: 23, 11: 20, 15: 19, 16: 20, 17: 18, 25: 20, 8: 20, 10: 20, 35: 20, 40: 8, 43: 20, 55: 20, 27: 20, 39: 20, 42: 6, 47: 11, 49: 20, 3: 13, 34: 21, 45: 14, 48: 0, 30: 20, 9: 10, 19: 20, 20: 12, 28: 20, 44: 20, 50: 20, 51: 22, 1: 20, 12: 20, 21: 20, 29: 16, 32: 4, 37: 20, 38: 20, 46: 17, 22: 20, 41: 15, 53: 1, 2: 2, 18: 9, 23: 20, 54: 7, 56: 5 }

// Base lexer interface.
type BaseLexer interface { Next() Token }
// Lexer struct. Produces token stream.
type Lexer struct {
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
    lexer := &Lexer { stream, handler }
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
    // Create token struct
    return Token { token, string(input[:i]), start, end }
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
        token := l.Next()
        location := fmt.Sprintf("%d:%d-%d:%d", token.Start.Line, token.Start.Col, token.End.Line, token.End.Col)
        fmt.Printf("%-16s | %-16s %-16s\n", location, token.Type, token.Value)
        if token.Type == EOF { break }
    }
}
