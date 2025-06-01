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
    { 46: 12, 33: 12, 14: 15, 17: 19, 43: 12, 21: 45, 41: 12, 13: 20, 15: 24, 50: 26, 18: 35, 40: 12, 39: 12, 36: 12, 47: 12, 49: 12, 37: 12, 3: 48, 24: 29, 44: 2, 12: 27, 45: 12, 26: 37, 0: 16, 10: 21, 28: 12, 52: 3, 35: 12, 53: 12, 29: 10, 38: 50, 19: 44, 56: 11, 22: 55, 9: 5, 54: 12, 2: 48, 7: 48, 42: 12, 5: 48, 48: 12, 51: 41 },
    { 50: 12, 35: 12, 49: 12, 47: 12, 20: 12, 51: 12, 53: 12, 40: 34, 54: 12, 48: 12, 28: 12, 39: 12, 52: 12, 38: 12, 42: 12, 46: 12, 41: 12, 43: 12, 36: 12, 45: 12, 44: 12, 37: 12, 33: 12 },
    { 37: 33, 54: 12, 38: 12, 42: 12, 49: 12, 36: 12, 35: 12, 33: 12, 51: 12, 20: 12, 48: 12, 52: 12, 28: 12, 43: 12, 45: 12, 46: 12, 50: 12, 44: 12, 53: 12, 41: 12, 39: 12, 47: 12, 40: 12 },
    { 40: 12, 37: 12, 50: 12, 35: 12, 36: 12, 20: 12, 49: 12, 45: 12, 33: 12, 51: 12, 47: 4, 54: 12, 53: 12, 46: 12, 42: 12, 48: 12, 28: 12, 41: 12, 38: 12, 44: 12, 52: 12, 43: 12, 39: 12 },
    { 53: 12, 28: 12, 44: 12, 52: 12, 43: 42, 33: 12, 38: 12, 45: 12, 51: 12, 54: 12, 49: 12, 41: 12, 50: 12, 46: 12, 39: 12, 40: 12, 20: 12, 42: 12, 37: 12, 48: 12, 47: 12, 35: 12, 36: 12 },
    { 50: 5, 34: 5, 24: 5, 35: 5, 21: 5, 40: 5, 8: 5, 19: 5, 11: 5, 30: 28, 16: 5, 46: 5, 36: 5, 14: 5, 44: 5, 28: 5, 43: 5, 13: 5, 25: 5, 55: 5, 48: 5, 6: 5, 51: 5, 38: 5, 57: 5, 2: 5, 47: 5, 53: 5, 56: 5, 7: 5, 29: 5, 52: 5, 33: 5, 12: 5, 42: 5, 23: 5, 1: 5, 32: 5, 26: 5, 54: 5, 9: 40, 27: 5, 10: 5, 31: 5, 37: 5, 41: 5, 49: 5, 22: 5, 45: 5, 17: 5, 18: 5, 4: 5, 39: 5, 15: 5, 20: 5 },
    { 37: 12, 54: 12, 38: 12, 52: 12, 28: 12, 42: 12, 33: 12, 53: 12, 40: 12, 35: 12, 44: 12, 50: 12, 20: 12, 39: 12, 51: 12, 41: 18, 47: 12, 48: 12, 36: 12, 45: 12, 46: 12, 49: 12, 43: 12 },
    { },
    { 40: 12, 38: 12, 44: 12, 39: 1, 48: 12, 42: 12, 33: 12, 52: 12, 36: 12, 53: 12, 28: 12, 45: 12, 51: 12, 37: 12, 20: 12, 35: 12, 41: 12, 47: 12, 54: 12, 43: 12, 46: 12, 50: 12, 49: 12 },
    { 45: 12, 53: 12, 46: 12, 44: 12, 39: 12, 50: 12, 33: 12, 43: 12, 49: 12, 37: 12, 36: 12, 20: 12, 54: 12, 48: 12, 41: 12, 52: 12, 35: 12, 42: 12, 51: 12, 47: 12, 28: 12, 40: 12, 38: 12 },
    { 26: 10, 43: 10, 12: 10, 36: 10, 19: 10, 51: 10, 50: 10, 47: 10, 27: 10, 54: 10, 23: 10, 30: 14, 52: 10, 32: 10, 40: 10, 29: 10, 8: 10, 9: 10, 41: 10, 18: 10, 13: 10, 6: 10, 20: 10, 44: 10, 45: 10, 34: 10, 7: 10, 14: 10, 17: 10, 21: 10, 38: 10, 11: 10, 25: 10, 39: 10, 2: 10, 55: 10, 28: 10, 56: 10, 48: 10, 15: 10, 57: 10, 53: 10, 31: 49, 46: 10, 42: 10, 1: 10, 37: 10, 49: 10, 33: 10, 35: 10, 4: 10, 10: 10, 16: 10, 24: 10, 22: 10 },
    { },
    { 36: 12, 48: 12, 52: 12, 44: 12, 39: 12, 40: 12, 46: 12, 28: 12, 43: 12, 45: 12, 37: 12, 49: 12, 35: 12, 47: 12, 50: 12, 20: 12, 42: 12, 38: 12, 33: 12, 51: 12, 41: 12, 54: 12, 53: 12 },
    { 52: 12, 53: 12, 37: 12, 49: 12, 51: 12, 44: 12, 50: 12, 36: 12, 54: 12, 28: 12, 38: 12, 47: 12, 39: 12, 20: 12, 43: 12, 42: 12, 46: 43, 33: 12, 48: 12, 40: 12, 35: 12, 41: 12, 45: 12 },
    { 7: 10, 38: 10, 18: 10, 1: 10, 16: 10, 55: 10, 30: 10, 13: 10, 49: 10, 29: 10, 57: 10, 53: 10, 34: 10, 28: 10, 42: 10, 4: 10, 41: 10, 33: 10, 44: 10, 8: 10, 50: 10, 2: 10, 47: 10, 14: 10, 45: 10, 54: 10, 15: 10, 56: 10, 52: 10, 37: 10, 6: 10, 21: 10, 48: 10, 17: 10, 36: 10, 22: 10, 25: 10, 40: 10, 27: 10, 43: 10, 11: 10, 20: 10, 24: 10, 39: 10, 46: 10, 32: 10, 35: 10, 23: 10, 31: 10, 12: 10, 26: 10, 9: 10, 19: 10, 10: 10, 51: 10 },
    { },
    { },
    { 20: 12, 49: 12, 41: 12, 28: 12, 42: 12, 46: 12, 39: 30, 44: 12, 53: 12, 37: 12, 45: 12, 43: 12, 50: 12, 52: 12, 51: 12, 38: 12, 48: 12, 35: 12, 36: 12, 54: 12, 33: 12, 40: 12, 47: 12 },
    { 41: 12, 37: 12, 44: 12, 50: 12, 38: 12, 45: 12, 40: 12, 20: 12, 48: 23, 54: 12, 52: 12, 36: 12, 39: 12, 47: 12, 51: 12, 33: 12, 42: 12, 43: 12, 28: 12, 46: 12, 49: 12, 53: 12, 35: 12 },
    { 25: 7 },
    { },
    { },
    { 50: 12, 54: 12, 42: 12, 47: 12, 38: 12, 41: 12, 28: 12, 52: 12, 35: 17, 53: 12, 37: 12, 45: 12, 46: 12, 40: 12, 20: 12, 44: 12, 39: 12, 43: 12, 36: 12, 48: 12, 49: 12, 51: 12, 33: 12 },
    { 36: 12, 44: 12, 54: 12, 49: 12, 51: 12, 45: 12, 40: 12, 43: 12, 52: 12, 50: 12, 53: 12, 39: 12, 46: 12, 20: 12, 47: 12, 28: 12, 42: 12, 35: 12, 38: 12, 41: 12, 37: 12, 33: 12, 48: 12 },
    { },
    { 39: 12, 50: 12, 36: 12, 20: 12, 51: 12, 40: 12, 35: 12, 33: 12, 38: 12, 49: 12, 53: 12, 45: 12, 28: 12, 42: 12, 46: 12, 44: 12, 54: 12, 52: 52, 43: 12, 37: 12, 41: 12, 47: 12, 48: 12 },
    { 44: 12, 52: 12, 43: 12, 49: 12, 51: 12, 47: 12, 54: 12, 38: 12, 33: 12, 20: 12, 53: 56, 36: 12, 37: 12, 50: 12, 35: 12, 45: 12, 41: 8, 28: 12, 40: 12, 46: 12, 48: 12, 39: 12, 42: 12 },
    { },
    { 12: 5, 49: 5, 42: 5, 13: 5, 11: 5, 37: 5, 1: 5, 6: 5, 29: 5, 45: 5, 21: 5, 36: 5, 41: 5, 10: 5, 2: 5, 27: 5, 28: 5, 23: 5, 30: 5, 39: 5, 16: 5, 32: 5, 50: 5, 18: 5, 22: 5, 57: 5, 14: 5, 33: 5, 38: 5, 26: 5, 47: 5, 19: 5, 55: 5, 31: 5, 35: 5, 24: 5, 9: 5, 43: 5, 15: 5, 25: 5, 7: 5, 17: 5, 4: 5, 56: 5, 46: 5, 34: 5, 54: 5, 44: 5, 40: 5, 48: 5, 51: 5, 52: 5, 8: 5, 53: 5, 20: 5 },
    { },
    { 45: 31, 51: 12, 39: 12, 33: 12, 50: 12, 20: 12, 43: 12, 37: 12, 28: 12, 49: 12, 53: 12, 40: 12, 38: 12, 47: 12, 35: 12, 52: 12, 44: 12, 54: 12, 48: 12, 36: 12, 41: 12, 42: 12, 46: 12 },
    { 38: 12, 28: 12, 40: 12, 50: 12, 54: 12, 53: 12, 47: 12, 36: 12, 33: 12, 49: 12, 35: 12, 41: 12, 52: 12, 37: 51, 48: 12, 39: 12, 46: 12, 51: 12, 42: 12, 43: 12, 45: 12, 44: 12, 20: 12 },
    { 49: 12, 37: 46, 40: 12, 28: 12, 20: 12, 43: 12, 33: 12, 42: 12, 47: 12, 50: 12, 48: 12, 53: 12, 38: 12, 44: 12, 36: 12, 52: 12, 46: 12, 35: 12, 39: 12, 54: 12, 51: 12, 41: 12, 45: 12 },
    { 33: 12, 53: 12, 40: 12, 51: 12, 50: 12, 48: 12, 43: 12, 36: 12, 44: 12, 54: 12, 42: 12, 49: 12, 37: 12, 47: 12, 39: 12, 52: 12, 38: 36, 46: 12, 41: 12, 28: 12, 35: 12, 20: 12, 45: 12 },
    { 41: 12, 36: 12, 37: 12, 38: 12, 47: 12, 48: 12, 52: 47, 33: 12, 43: 12, 28: 12, 39: 12, 51: 12, 45: 12, 50: 12, 40: 12, 44: 12, 53: 12, 46: 12, 35: 12, 20: 12, 42: 12, 54: 12, 49: 12 },
    { },
    { 20: 12, 46: 12, 53: 12, 36: 12, 52: 9, 35: 12, 37: 12, 49: 12, 38: 12, 44: 12, 40: 12, 39: 12, 43: 12, 33: 12, 45: 12, 48: 12, 54: 12, 47: 12, 42: 12, 41: 12, 50: 12, 28: 12, 51: 12 },
    { },
    { 16: 38, 21: 38, 9: 38, 42: 38, 43: 38, 48: 38, 37: 38, 36: 38, 25: 38, 17: 38, 56: 38, 32: 38, 24: 38, 4: 38, 3: 38, 52: 38, 49: 38, 27: 38, 54: 38, 2: 38, 35: 38, 10: 38, 8: 38, 33: 38, 19: 38, 18: 38, 30: 38, 11: 38, 5: 38, 28: 38, 6: 38, 41: 38, 13: 38, 39: 38, 38: 38, 14: 53, 46: 38, 34: 38, 31: 38, 47: 38, 29: 38, 20: 38, 57: 38, 55: 38, 44: 38, 12: 38, 53: 38, 15: 38, 1: 38, 50: 38, 45: 38, 26: 38, 22: 38, 7: 38, 23: 38, 51: 38, 40: 38 },
    { },
    { },
    { 33: 12, 44: 12, 43: 6, 36: 12, 37: 12, 38: 12, 53: 12, 51: 12, 35: 12, 49: 12, 41: 12, 48: 12, 28: 12, 47: 12, 45: 12, 39: 12, 40: 12, 52: 12, 50: 12, 42: 12, 54: 12, 20: 12, 46: 12 },
    { 54: 12, 47: 12, 53: 12, 28: 12, 42: 12, 44: 12, 39: 12, 36: 12, 46: 12, 40: 12, 37: 13, 50: 12, 48: 12, 52: 12, 51: 12, 20: 12, 41: 12, 35: 12, 33: 12, 49: 12, 38: 12, 45: 12, 43: 12 },
    { 50: 12, 41: 12, 52: 12, 53: 12, 39: 12, 43: 12, 46: 12, 35: 12, 51: 12, 45: 12, 33: 12, 48: 12, 47: 12, 40: 12, 54: 12, 28: 12, 49: 12, 20: 12, 44: 12, 36: 12, 42: 12, 37: 12, 38: 12 },
    { 19: 54, 14: 38 },
    { },
    { 42: 12, 20: 12, 51: 12, 45: 12, 50: 12, 43: 12, 33: 12, 52: 12, 28: 12, 35: 12, 47: 12, 44: 12, 48: 12, 40: 12, 53: 12, 38: 12, 49: 12, 36: 12, 37: 12, 41: 12, 54: 12, 46: 12, 39: 12 },
    { 40: 12, 49: 12, 38: 12, 33: 12, 39: 12, 53: 12, 37: 12, 41: 12, 50: 12, 46: 12, 20: 12, 51: 12, 36: 12, 48: 12, 54: 12, 28: 12, 44: 12, 52: 12, 47: 12, 35: 12, 43: 12, 45: 12, 42: 12 },
    { 5: 48, 7: 48, 2: 48, 3: 48 },
    { },
    { 35: 12, 33: 12, 36: 12, 37: 12, 43: 12, 51: 12, 45: 12, 38: 12, 50: 22, 47: 12, 49: 12, 44: 12, 54: 12, 42: 12, 41: 12, 48: 12, 53: 12, 20: 12, 39: 12, 46: 12, 40: 12, 52: 12, 28: 12 },
    { 40: 12, 41: 12, 48: 12, 35: 12, 47: 12, 45: 12, 50: 12, 28: 12, 36: 12, 20: 12, 53: 12, 33: 12, 39: 12, 38: 12, 54: 12, 46: 25, 44: 12, 52: 12, 51: 12, 43: 12, 49: 12, 42: 12, 37: 12 },
    { 41: 12, 20: 12, 33: 12, 39: 12, 42: 12, 47: 12, 50: 12, 48: 12, 40: 12, 51: 12, 49: 12, 36: 12, 35: 12, 53: 12, 46: 12, 54: 12, 44: 12, 28: 12, 43: 12, 38: 12, 45: 12, 52: 12, 37: 12 },
    { 56: 38, 40: 38, 16: 38, 13: 38, 9: 38, 11: 38, 14: 38, 42: 38, 47: 38, 37: 38, 41: 38, 31: 38, 45: 38, 7: 38, 50: 38, 29: 38, 28: 38, 36: 38, 4: 38, 8: 38, 12: 38, 15: 38, 43: 38, 35: 38, 18: 38, 25: 38, 54: 38, 48: 38, 30: 38, 51: 38, 53: 38, 26: 38, 3: 38, 23: 38, 27: 38, 33: 38, 5: 38, 17: 38, 57: 38, 52: 38, 49: 38, 32: 38, 38: 38, 2: 38, 39: 38, 46: 38, 20: 38, 34: 38, 21: 38, 44: 38, 6: 38, 10: 38, 22: 38, 24: 38, 55: 38, 1: 38, 19: 39 },
    { 45: 54, 35: 54, 20: 54, 30: 54, 32: 54, 12: 54, 18: 54, 34: 54, 0: 39, 24: 54, 19: 54, 57: 54, 49: 54, 54: 54, 27: 54, 40: 54, 11: 54, 4: 54, 10: 54, 39: 54, 44: 54, 6: 54, 37: 54, 1: 54, 42: 54, 33: 54, 8: 54, 53: 54, 48: 54, 28: 54, 25: 54, 13: 54, 50: 54, 9: 54, 31: 54, 41: 54, 22: 54, 29: 54, 23: 54, 55: 54, 38: 54, 47: 54, 14: 54, 36: 54, 56: 54, 46: 54, 52: 54, 16: 54, 26: 54, 17: 54, 3: 39, 51: 54, 15: 54, 7: 54, 2: 54, 5: 39, 43: 54, 21: 54 },
    { },
    { 50: 12, 49: 12, 37: 12, 28: 12, 36: 12, 46: 12, 45: 12, 41: 12, 33: 12, 54: 12, 42: 12, 47: 12, 53: 12, 44: 32, 39: 12, 52: 12, 51: 12, 20: 12, 40: 12, 43: 12, 35: 12, 38: 12, 48: 12 },
}
var accept = map[int]TokenType { 18: 20, 30: 20, 35: 12, 46: 2, 56: 20, 16: 23, 21: 14, 37: 11, 2: 20, 25: 20, 29: 8, 36: 20, 39: 1, 42: 20, 45: 16, 48: 0, 3: 20, 6: 20, 24: 9, 26: 20, 51: 20, 22: 20, 27: 17, 34: 20, 40: 21, 50: 20, 55: 15, 1: 20, 7: 19, 9: 5, 15: 10, 17: 20, 31: 20, 33: 20, 43: 3, 8: 20, 11: 13, 13: 20, 20: 18, 23: 7, 32: 20, 47: 6, 49: 22, 4: 20, 12: 20, 41: 20, 52: 4 }

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
