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
    { 43: 53, 17: 32, 45: 53, 10: 1, 15: 47, 50: 12, 53: 53, 40: 53, 29: 20, 38: 27, 7: 30, 26: 2, 52: 29, 28: 53, 21: 36, 2: 30, 5: 30, 22: 25, 41: 53, 24: 43, 14: 51, 36: 53, 54: 53, 35: 53, 51: 42, 33: 53, 49: 53, 12: 26, 56: 10, 13: 56, 39: 53, 37: 53, 19: 45, 18: 31, 44: 11, 42: 53, 3: 30, 46: 53, 47: 53, 48: 53, 0: 19, 9: 38 },
    { },
    { },
    { 35: 53, 42: 53, 38: 53, 45: 53, 20: 53, 36: 53, 52: 53, 49: 53, 51: 53, 41: 53, 50: 53, 53: 53, 37: 53, 43: 53, 48: 53, 54: 53, 46: 53, 47: 53, 28: 53, 33: 53, 39: 53, 40: 53, 44: 53 },
    { 5: 18, 24: 18, 22: 18, 47: 18, 13: 18, 27: 18, 6: 18, 52: 18, 26: 18, 3: 18, 15: 18, 31: 18, 17: 18, 43: 18, 4: 18, 50: 18, 40: 18, 32: 18, 18: 18, 36: 18, 7: 18, 2: 18, 46: 18, 21: 18, 55: 18, 48: 18, 11: 18, 33: 18, 49: 18, 14: 18, 25: 18, 20: 18, 19: 5, 30: 18, 10: 18, 8: 18, 29: 18, 51: 18, 28: 18, 34: 18, 39: 18, 44: 18, 56: 18, 42: 18, 45: 18, 1: 18, 9: 18, 57: 18, 37: 18, 35: 18, 23: 18, 16: 18, 54: 18, 12: 18, 53: 18, 38: 18, 41: 18 },
    { },
    { 5: 5, 28: 6, 53: 6, 0: 5, 20: 6, 30: 6, 32: 6, 18: 6, 12: 6, 41: 6, 7: 6, 55: 6, 57: 6, 37: 6, 3: 5, 49: 6, 56: 6, 33: 6, 22: 6, 19: 6, 42: 6, 4: 6, 34: 6, 36: 6, 15: 6, 21: 6, 43: 6, 6: 6, 52: 6, 1: 6, 48: 6, 54: 6, 26: 6, 39: 6, 13: 6, 25: 6, 47: 6, 29: 6, 23: 6, 44: 6, 9: 6, 16: 6, 17: 6, 35: 6, 51: 6, 38: 6, 46: 6, 40: 6, 31: 6, 50: 6, 24: 6, 14: 6, 10: 6, 11: 6, 45: 6, 2: 6, 27: 6, 8: 6 },
    { 48: 53, 54: 53, 51: 53, 43: 53, 37: 53, 53: 53, 42: 53, 20: 53, 45: 53, 47: 53, 38: 53, 28: 53, 33: 53, 46: 53, 52: 53, 44: 53, 41: 53, 40: 53, 50: 53, 35: 53, 39: 53, 49: 53, 36: 53 },
    { 47: 53, 37: 53, 38: 53, 45: 53, 42: 53, 48: 53, 43: 53, 44: 53, 46: 53, 39: 53, 35: 53, 49: 53, 36: 53, 20: 53, 41: 53, 28: 53, 40: 53, 33: 53, 51: 53, 53: 53, 50: 53, 52: 53, 54: 53 },
    { 33: 53, 20: 53, 48: 53, 43: 21, 40: 53, 49: 53, 51: 53, 36: 53, 35: 53, 42: 53, 37: 53, 54: 53, 28: 53, 47: 53, 38: 53, 39: 53, 44: 53, 46: 53, 52: 53, 45: 53, 53: 53, 41: 53, 50: 53 },
    { },
    { 46: 53, 33: 53, 42: 53, 49: 53, 45: 53, 41: 53, 43: 53, 52: 53, 47: 53, 50: 53, 38: 53, 36: 53, 37: 34, 40: 53, 51: 53, 35: 53, 39: 53, 53: 53, 20: 53, 28: 53, 44: 53, 48: 53, 54: 53 },
    { 28: 53, 41: 48, 42: 53, 53: 55, 33: 53, 45: 53, 51: 53, 48: 53, 50: 53, 39: 53, 44: 53, 20: 53, 46: 53, 38: 53, 47: 53, 36: 53, 49: 53, 35: 53, 40: 53, 43: 53, 37: 53, 52: 53, 54: 53 },
    { 20: 53, 50: 53, 28: 53, 54: 53, 47: 53, 33: 53, 44: 53, 36: 53, 35: 53, 46: 53, 45: 53, 37: 53, 49: 53, 41: 53, 51: 53, 52: 53, 40: 33, 53: 53, 39: 53, 42: 53, 48: 53, 43: 53, 38: 53 },
    { 41: 53, 49: 53, 53: 53, 52: 53, 33: 53, 44: 53, 48: 53, 38: 53, 20: 53, 46: 53, 40: 53, 54: 53, 35: 53, 45: 53, 51: 53, 42: 53, 43: 53, 50: 53, 28: 53, 47: 53, 39: 53, 37: 53, 36: 53 },
    { 46: 53, 40: 53, 20: 53, 41: 53, 42: 53, 38: 53, 48: 53, 44: 53, 52: 53, 47: 53, 50: 53, 35: 53, 53: 53, 33: 53, 43: 53, 37: 53, 49: 53, 45: 53, 51: 53, 39: 53, 54: 53, 36: 53, 28: 53 },
    { },
    { 43: 53, 42: 53, 50: 53, 37: 53, 45: 24, 40: 53, 44: 53, 47: 53, 38: 53, 46: 53, 48: 53, 51: 53, 28: 53, 54: 53, 33: 53, 52: 53, 49: 53, 20: 53, 39: 53, 53: 53, 36: 53, 41: 53, 35: 53 },
    { 20: 18, 52: 18, 43: 18, 21: 18, 15: 18, 34: 18, 31: 18, 29: 18, 8: 18, 7: 18, 45: 18, 14: 4, 13: 18, 25: 18, 57: 18, 41: 18, 36: 18, 30: 18, 53: 18, 39: 18, 50: 18, 54: 18, 2: 18, 18: 18, 48: 18, 32: 18, 19: 18, 24: 18, 27: 18, 49: 18, 22: 18, 1: 18, 40: 18, 6: 18, 55: 18, 16: 18, 44: 18, 37: 18, 5: 18, 23: 18, 33: 18, 51: 18, 28: 18, 12: 18, 3: 18, 35: 18, 9: 18, 38: 18, 10: 18, 11: 18, 56: 18, 4: 18, 17: 18, 47: 18, 26: 18, 46: 18, 42: 18 },
    { },
    { 8: 20, 19: 20, 46: 20, 28: 20, 32: 20, 53: 20, 37: 20, 39: 20, 11: 20, 9: 20, 55: 20, 29: 20, 20: 20, 42: 20, 1: 20, 54: 20, 36: 20, 13: 20, 48: 20, 17: 20, 33: 20, 27: 20, 44: 20, 35: 20, 57: 20, 47: 20, 31: 39, 30: 49, 49: 20, 14: 20, 40: 20, 25: 20, 45: 20, 43: 20, 10: 20, 22: 20, 7: 20, 23: 20, 56: 20, 16: 20, 12: 20, 52: 20, 51: 20, 24: 20, 21: 20, 15: 20, 34: 20, 38: 20, 41: 20, 26: 20, 50: 20, 2: 20, 4: 20, 6: 20, 18: 20 },
    { 35: 53, 46: 53, 20: 53, 44: 53, 54: 53, 47: 53, 38: 53, 49: 53, 36: 53, 28: 53, 33: 53, 48: 53, 53: 53, 37: 52, 52: 53, 39: 53, 40: 53, 50: 53, 51: 53, 41: 53, 42: 53, 45: 53, 43: 53 },
    { 43: 53, 46: 53, 41: 53, 33: 53, 44: 53, 48: 53, 40: 53, 42: 53, 47: 53, 37: 53, 36: 53, 35: 53, 39: 53, 53: 53, 49: 53, 54: 53, 20: 53, 45: 53, 51: 53, 50: 53, 38: 53, 28: 53, 52: 53 },
    { 41: 53, 43: 53, 50: 53, 49: 53, 54: 53, 39: 53, 45: 53, 36: 53, 46: 53, 47: 53, 33: 53, 44: 53, 52: 53, 37: 8, 53: 53, 51: 53, 38: 53, 20: 53, 48: 53, 35: 53, 42: 53, 28: 53, 40: 53 },
    { 46: 53, 44: 53, 20: 53, 49: 53, 40: 53, 38: 53, 54: 53, 48: 53, 53: 53, 28: 53, 33: 53, 39: 53, 43: 53, 35: 53, 36: 53, 41: 53, 50: 53, 45: 53, 52: 53, 37: 41, 47: 53, 42: 53, 51: 53 },
    { },
    { },
    { 40: 53, 52: 53, 33: 53, 44: 53, 51: 53, 42: 53, 28: 53, 39: 53, 53: 53, 36: 53, 47: 53, 35: 53, 43: 53, 46: 53, 49: 53, 50: 40, 48: 53, 41: 53, 45: 53, 20: 53, 37: 53, 38: 53, 54: 53 },
    { 40: 53, 51: 53, 47: 53, 36: 53, 41: 53, 28: 53, 20: 53, 33: 53, 42: 53, 43: 53, 53: 53, 49: 53, 45: 53, 50: 53, 35: 53, 46: 53, 39: 17, 48: 53, 37: 53, 38: 53, 52: 53, 44: 53, 54: 53 },
    { 50: 53, 35: 53, 33: 53, 43: 53, 36: 53, 47: 9, 38: 53, 53: 53, 42: 53, 28: 53, 45: 53, 37: 53, 46: 53, 52: 53, 48: 53, 40: 53, 44: 53, 54: 53, 41: 53, 20: 53, 49: 53, 39: 53, 51: 53 },
    { 2: 30, 3: 30, 5: 30, 7: 30 },
    { },
    { 25: 16 },
    { 53: 53, 36: 53, 20: 53, 52: 7, 39: 53, 48: 53, 41: 53, 50: 53, 49: 53, 35: 53, 40: 53, 46: 53, 33: 53, 28: 53, 37: 53, 51: 53, 45: 53, 44: 53, 54: 53, 43: 53, 47: 53, 38: 53, 42: 53 },
    { 33: 53, 38: 35, 37: 53, 36: 53, 41: 53, 42: 53, 20: 53, 52: 53, 51: 53, 48: 53, 35: 53, 39: 53, 43: 53, 44: 53, 50: 53, 28: 53, 46: 53, 54: 53, 53: 53, 47: 53, 40: 53, 49: 53, 45: 53 },
    { 41: 53, 50: 53, 35: 53, 54: 53, 42: 53, 38: 53, 44: 53, 39: 53, 53: 53, 43: 53, 20: 53, 45: 53, 36: 53, 46: 53, 33: 53, 47: 53, 40: 53, 49: 53, 28: 53, 52: 22, 48: 53, 37: 53, 51: 53 },
    { },
    { 53: 53, 28: 53, 49: 53, 54: 53, 41: 44, 47: 53, 36: 53, 37: 53, 38: 53, 45: 53, 50: 53, 46: 53, 42: 53, 40: 53, 48: 53, 20: 53, 44: 53, 39: 53, 35: 53, 43: 53, 52: 53, 33: 53, 51: 53 },
    { 50: 38, 56: 38, 36: 38, 2: 38, 55: 38, 47: 38, 19: 38, 45: 38, 34: 38, 39: 38, 52: 38, 53: 38, 35: 38, 42: 38, 28: 38, 7: 38, 4: 38, 31: 38, 30: 54, 23: 38, 20: 38, 13: 38, 14: 38, 44: 38, 49: 38, 54: 38, 1: 38, 38: 38, 15: 38, 51: 38, 8: 38, 40: 38, 25: 38, 33: 38, 11: 38, 41: 38, 46: 38, 9: 46, 16: 38, 24: 38, 48: 38, 27: 38, 21: 38, 10: 38, 18: 38, 32: 38, 37: 38, 57: 38, 43: 38, 22: 38, 17: 38, 6: 38, 12: 38, 26: 38, 29: 38 },
    { },
    { 36: 53, 42: 53, 50: 53, 28: 53, 20: 53, 53: 53, 44: 53, 48: 53, 37: 53, 47: 53, 41: 53, 51: 53, 43: 53, 35: 28, 45: 53, 38: 53, 52: 53, 46: 53, 33: 53, 49: 53, 54: 53, 39: 53, 40: 53 },
    { 50: 53, 20: 53, 54: 53, 47: 53, 28: 53, 38: 53, 40: 53, 42: 53, 35: 53, 39: 53, 52: 53, 53: 53, 48: 53, 44: 53, 49: 53, 45: 53, 36: 53, 43: 53, 33: 53, 37: 53, 46: 50, 51: 53, 41: 53 },
    { 33: 53, 36: 53, 42: 53, 20: 53, 52: 53, 28: 53, 38: 53, 44: 53, 49: 53, 40: 53, 54: 53, 35: 53, 46: 53, 53: 53, 39: 53, 48: 53, 43: 37, 47: 53, 51: 53, 41: 53, 37: 53, 50: 53, 45: 53 },
    { },
    { 35: 53, 47: 53, 37: 53, 46: 53, 49: 53, 52: 53, 42: 53, 50: 53, 28: 53, 39: 53, 38: 53, 45: 53, 40: 53, 48: 3, 54: 53, 53: 53, 41: 53, 51: 53, 33: 53, 44: 53, 20: 53, 36: 53, 43: 53 },
    { 19: 6, 14: 18 },
    { },
    { },
    { 20: 53, 44: 53, 49: 53, 42: 53, 37: 53, 52: 53, 33: 53, 46: 53, 50: 53, 28: 53, 54: 53, 51: 53, 41: 53, 39: 13, 38: 53, 45: 53, 48: 53, 36: 53, 53: 53, 35: 53, 43: 53, 40: 53, 47: 53 },
    { 21: 20, 20: 20, 52: 20, 8: 20, 31: 20, 45: 20, 6: 20, 39: 20, 47: 20, 46: 20, 57: 20, 53: 20, 18: 20, 25: 20, 38: 20, 16: 20, 10: 20, 13: 20, 42: 20, 56: 20, 1: 20, 34: 20, 22: 20, 44: 20, 2: 20, 50: 20, 17: 20, 32: 20, 12: 20, 33: 20, 29: 20, 7: 20, 55: 20, 19: 20, 24: 20, 41: 20, 43: 20, 9: 20, 36: 20, 37: 20, 27: 20, 49: 20, 54: 20, 11: 20, 28: 20, 48: 20, 51: 20, 4: 20, 14: 20, 23: 20, 35: 20, 15: 20, 40: 20, 26: 20, 30: 20 },
    { 33: 53, 44: 53, 45: 53, 37: 53, 38: 53, 28: 53, 49: 53, 39: 53, 40: 53, 47: 53, 36: 53, 41: 53, 46: 53, 35: 53, 52: 14, 53: 53, 51: 53, 54: 53, 42: 53, 43: 53, 20: 53, 50: 53, 48: 53 },
    { },
    { 52: 53, 51: 53, 53: 53, 54: 53, 50: 53, 42: 53, 20: 53, 46: 15, 39: 53, 36: 53, 43: 53, 40: 53, 41: 53, 45: 53, 28: 53, 37: 53, 48: 53, 33: 53, 44: 53, 49: 53, 47: 53, 35: 53, 38: 53 },
    { 46: 53, 52: 53, 51: 53, 53: 53, 37: 53, 47: 53, 36: 53, 35: 53, 43: 53, 39: 53, 28: 53, 20: 53, 40: 53, 41: 53, 50: 53, 38: 53, 45: 53, 42: 53, 48: 53, 49: 53, 54: 53, 33: 53, 44: 53 },
    { 40: 38, 24: 38, 13: 38, 16: 38, 35: 38, 27: 38, 20: 38, 11: 38, 38: 38, 17: 38, 46: 38, 57: 38, 41: 38, 1: 38, 51: 38, 43: 38, 23: 38, 45: 38, 9: 38, 10: 38, 18: 38, 49: 38, 21: 38, 8: 38, 44: 38, 26: 38, 42: 38, 15: 38, 52: 38, 2: 38, 56: 38, 32: 38, 29: 38, 39: 38, 50: 38, 19: 38, 22: 38, 25: 38, 53: 38, 48: 38, 37: 38, 12: 38, 6: 38, 55: 38, 33: 38, 7: 38, 14: 38, 30: 38, 36: 38, 4: 38, 54: 38, 34: 38, 47: 38, 28: 38, 31: 38 },
    { 20: 53, 28: 53, 46: 53, 48: 53, 51: 53, 41: 53, 47: 53, 54: 53, 45: 53, 38: 53, 42: 53, 33: 53, 44: 23, 50: 53, 43: 53, 40: 53, 49: 53, 35: 53, 37: 53, 52: 53, 53: 53, 39: 53, 36: 53 },
    { },
}
var accept = map[int]TokenType { 25: 15, 34: 20, 39: 22, 2: 11, 22: 5, 30: 0, 33: 20, 42: 20, 43: 8, 55: 20, 56: 18, 5: 1, 12: 20, 17: 20, 21: 20, 29: 20, 36: 16, 44: 20, 48: 20, 11: 20, 19: 23, 23: 20, 46: 21, 15: 3, 8: 2, 47: 9, 1: 14, 13: 20, 26: 17, 27: 20, 31: 12, 40: 20, 52: 20, 53: 20, 16: 19, 14: 4, 28: 20, 35: 20, 37: 20, 41: 20, 50: 20, 51: 10, 3: 7, 7: 6, 9: 20, 10: 13, 24: 20 }

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
