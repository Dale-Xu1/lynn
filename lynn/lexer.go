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
    { 24: 6, 14: 18, 39: 5, 52: 8, 38: 38, 56: 4, 12: 32, 19: 55, 9: 44, 28: 5, 29: 36, 26: 23, 21: 47, 42: 5, 3: 25, 53: 5, 45: 5, 18: 3, 49: 5, 13: 24, 33: 5, 35: 5, 15: 11, 2: 25, 37: 5, 48: 5, 50: 30, 43: 5, 46: 5, 10: 28, 54: 5, 22: 45, 44: 12, 41: 5, 5: 25, 17: 37, 7: 25, 36: 5, 47: 5, 0: 46, 40: 5, 51: 17 },
    { 46: 5, 38: 5, 43: 5, 54: 5, 35: 5, 41: 5, 50: 5, 49: 5, 37: 5, 48: 5, 20: 5, 40: 40, 45: 5, 52: 5, 28: 5, 39: 5, 53: 5, 42: 5, 44: 5, 33: 5, 47: 5, 51: 5, 36: 5 },
    { 41: 5, 28: 5, 20: 5, 50: 5, 36: 5, 33: 5, 38: 5, 51: 5, 52: 5, 53: 5, 45: 5, 48: 5, 39: 1, 35: 5, 46: 5, 49: 5, 37: 5, 44: 5, 54: 5, 47: 5, 42: 5, 40: 5, 43: 5 },
    { },
    { },
    { 37: 5, 54: 5, 36: 5, 45: 5, 48: 5, 50: 5, 49: 5, 52: 5, 42: 5, 35: 5, 47: 5, 51: 5, 46: 5, 44: 5, 41: 5, 38: 5, 28: 5, 33: 5, 40: 5, 43: 5, 39: 5, 53: 5, 20: 5 },
    { },
    { 40: 5, 38: 5, 51: 5, 28: 5, 48: 5, 46: 5, 20: 5, 39: 5, 47: 5, 50: 5, 41: 5, 42: 5, 35: 5, 54: 5, 33: 5, 43: 5, 52: 5, 44: 5, 36: 5, 49: 5, 37: 5, 45: 5, 53: 5 },
    { 44: 5, 47: 19, 46: 5, 42: 5, 54: 5, 36: 5, 52: 5, 33: 5, 38: 5, 50: 5, 20: 5, 39: 5, 28: 5, 35: 5, 43: 5, 53: 5, 49: 5, 41: 5, 45: 5, 40: 5, 51: 5, 48: 5, 37: 5 },
    { 50: 5, 46: 51, 39: 5, 33: 5, 52: 5, 45: 5, 42: 5, 36: 5, 54: 5, 40: 5, 35: 5, 53: 5, 44: 5, 47: 5, 51: 5, 48: 5, 20: 5, 37: 5, 43: 5, 49: 5, 38: 5, 41: 5, 28: 5 },
    { },
    { },
    { 52: 5, 40: 5, 36: 5, 46: 5, 37: 31, 44: 5, 20: 5, 39: 5, 49: 5, 53: 5, 28: 5, 43: 5, 47: 5, 38: 5, 42: 5, 51: 5, 35: 5, 45: 5, 41: 5, 33: 5, 48: 5, 54: 5, 50: 5 },
    { },
    { 14: 35, 46: 35, 11: 35, 9: 35, 35: 35, 15: 35, 32: 35, 44: 35, 54: 35, 10: 35, 52: 35, 7: 35, 57: 35, 50: 35, 17: 35, 24: 35, 56: 35, 19: 43, 33: 35, 30: 35, 27: 35, 1: 35, 53: 35, 45: 35, 25: 35, 3: 35, 41: 35, 39: 35, 47: 35, 37: 35, 4: 35, 43: 35, 34: 35, 12: 35, 20: 35, 36: 35, 49: 35, 26: 35, 16: 35, 51: 35, 8: 35, 40: 35, 5: 35, 22: 35, 18: 35, 29: 35, 55: 35, 31: 35, 48: 35, 6: 35, 28: 35, 13: 35, 23: 35, 38: 35, 42: 35, 2: 35, 21: 35 },
    { 28: 5, 47: 5, 46: 5, 42: 5, 33: 5, 54: 5, 45: 5, 40: 5, 41: 5, 35: 5, 49: 5, 48: 5, 53: 5, 36: 5, 43: 5, 52: 5, 20: 5, 51: 5, 37: 16, 44: 5, 50: 5, 39: 5, 38: 5 },
    { 50: 5, 33: 5, 39: 5, 54: 5, 38: 5, 52: 5, 44: 5, 40: 5, 43: 5, 20: 5, 41: 5, 37: 5, 42: 5, 35: 5, 28: 5, 36: 5, 45: 5, 53: 5, 46: 20, 47: 5, 49: 5, 51: 5, 48: 5 },
    { 40: 5, 37: 5, 41: 5, 51: 5, 48: 5, 47: 5, 46: 5, 50: 5, 36: 5, 43: 52, 53: 5, 39: 5, 49: 5, 52: 5, 45: 5, 20: 5, 44: 5, 54: 5, 28: 5, 33: 5, 35: 5, 38: 5, 42: 5 },
    { },
    { 52: 5, 46: 5, 33: 5, 49: 5, 54: 5, 20: 5, 40: 5, 47: 5, 41: 5, 35: 5, 45: 5, 38: 5, 50: 5, 39: 5, 51: 5, 42: 5, 48: 5, 53: 5, 37: 5, 28: 5, 44: 5, 36: 5, 43: 15 },
    { 50: 5, 33: 5, 46: 5, 39: 5, 52: 5, 44: 5, 42: 5, 28: 5, 43: 5, 37: 5, 54: 5, 51: 5, 20: 5, 35: 5, 47: 5, 45: 5, 49: 5, 36: 5, 40: 5, 48: 5, 53: 5, 38: 5, 41: 5 },
    { 43: 5, 28: 5, 44: 5, 39: 5, 38: 5, 20: 5, 51: 5, 42: 5, 54: 5, 40: 5, 50: 5, 49: 5, 46: 5, 35: 5, 53: 5, 48: 5, 41: 5, 36: 5, 45: 5, 47: 5, 37: 9, 52: 5, 33: 5 },
    { 33: 44, 19: 44, 53: 44, 17: 44, 10: 44, 45: 44, 25: 44, 40: 44, 38: 44, 32: 44, 49: 44, 26: 44, 47: 44, 57: 44, 50: 44, 13: 44, 52: 44, 39: 44, 15: 44, 23: 44, 16: 44, 21: 44, 31: 44, 44: 44, 6: 44, 36: 44, 42: 44, 7: 44, 35: 44, 30: 44, 54: 44, 1: 44, 51: 44, 41: 44, 34: 44, 9: 44, 27: 44, 29: 44, 43: 44, 4: 44, 2: 44, 24: 44, 22: 44, 20: 44, 18: 44, 28: 44, 48: 44, 46: 44, 11: 44, 56: 44, 14: 44, 55: 44, 8: 44, 37: 44, 12: 44 },
    { },
    { },
    { 3: 25, 5: 25, 7: 25, 2: 25 },
    { 36: 5, 48: 5, 49: 5, 37: 5, 45: 5, 28: 5, 52: 5, 42: 5, 35: 5, 43: 5, 51: 5, 50: 5, 47: 5, 20: 5, 38: 5, 46: 5, 39: 5, 53: 5, 40: 5, 54: 5, 33: 5, 41: 5, 44: 33 },
    { 37: 5, 35: 5, 46: 5, 28: 5, 51: 5, 48: 5, 45: 5, 53: 5, 54: 5, 40: 5, 47: 5, 33: 5, 41: 5, 39: 5, 49: 5, 44: 5, 42: 5, 36: 5, 38: 5, 43: 5, 50: 5, 20: 5, 52: 5 },
    { },
    { 37: 5, 33: 5, 43: 5, 39: 5, 53: 5, 46: 5, 48: 5, 50: 5, 47: 5, 49: 5, 41: 5, 36: 5, 42: 5, 51: 5, 28: 5, 52: 5, 35: 5, 45: 5, 38: 5, 54: 5, 40: 5, 20: 5, 44: 5 },
    { 47: 5, 20: 5, 40: 5, 54: 5, 35: 5, 38: 5, 42: 5, 53: 26, 51: 5, 45: 5, 37: 5, 36: 5, 39: 5, 48: 5, 50: 5, 49: 5, 46: 5, 28: 5, 33: 5, 52: 5, 41: 2, 43: 5, 44: 5 },
    { 48: 5, 43: 5, 37: 5, 52: 5, 28: 5, 39: 5, 36: 5, 51: 5, 35: 5, 41: 5, 40: 5, 50: 5, 47: 5, 45: 5, 33: 5, 46: 5, 49: 5, 38: 48, 44: 5, 54: 5, 42: 5, 53: 5, 20: 5 },
    { },
    { 40: 5, 39: 5, 44: 5, 20: 5, 35: 5, 43: 5, 51: 5, 33: 5, 37: 27, 54: 5, 45: 5, 46: 5, 42: 5, 28: 5, 52: 5, 53: 5, 50: 5, 49: 5, 38: 5, 36: 5, 41: 5, 48: 5, 47: 5 },
    { 33: 5, 47: 5, 38: 5, 28: 5, 52: 5, 39: 5, 46: 5, 20: 5, 42: 5, 36: 5, 53: 5, 40: 5, 48: 5, 50: 5, 49: 5, 44: 5, 35: 5, 45: 5, 43: 5, 51: 5, 37: 5, 41: 5, 54: 5 },
    { 36: 35, 45: 35, 47: 35, 11: 35, 8: 35, 10: 35, 41: 35, 29: 35, 20: 35, 57: 35, 12: 35, 1: 35, 26: 35, 52: 35, 19: 35, 46: 35, 43: 35, 23: 35, 38: 35, 9: 35, 13: 35, 27: 35, 35: 35, 55: 35, 34: 35, 25: 35, 5: 35, 2: 35, 42: 35, 37: 35, 40: 35, 18: 35, 24: 35, 50: 35, 6: 35, 53: 35, 21: 35, 54: 35, 30: 35, 3: 35, 28: 35, 7: 35, 4: 35, 44: 35, 48: 35, 31: 35, 32: 35, 22: 35, 17: 35, 15: 35, 39: 35, 51: 35, 16: 35, 56: 35, 49: 35, 33: 35, 14: 14 },
    { 41: 36, 7: 36, 11: 36, 9: 36, 33: 36, 56: 36, 30: 56, 21: 36, 19: 36, 54: 36, 45: 36, 13: 36, 15: 36, 31: 39, 6: 36, 46: 36, 22: 36, 25: 36, 48: 36, 8: 36, 2: 36, 26: 36, 57: 36, 49: 36, 51: 36, 47: 36, 20: 36, 39: 36, 27: 36, 32: 36, 52: 36, 50: 36, 17: 36, 40: 36, 38: 36, 53: 36, 23: 36, 35: 36, 12: 36, 43: 36, 29: 36, 18: 36, 55: 36, 28: 36, 44: 36, 42: 36, 14: 36, 24: 36, 37: 36, 4: 36, 10: 36, 1: 36, 34: 36, 36: 36, 16: 36 },
    { 25: 13 },
    { 36: 5, 33: 5, 44: 5, 48: 5, 53: 5, 38: 5, 37: 5, 20: 5, 50: 53, 52: 5, 42: 5, 40: 5, 43: 5, 28: 5, 47: 5, 54: 5, 49: 5, 51: 5, 35: 5, 39: 5, 41: 5, 45: 5, 46: 5 },
    { },
    { 28: 5, 42: 5, 52: 34, 35: 5, 53: 5, 47: 5, 49: 5, 36: 5, 48: 5, 38: 5, 46: 5, 54: 5, 40: 5, 37: 5, 43: 5, 44: 5, 45: 5, 20: 5, 39: 5, 41: 5, 33: 5, 51: 5, 50: 5 },
    { 45: 5, 40: 5, 39: 5, 41: 5, 42: 5, 53: 5, 46: 5, 38: 5, 33: 5, 37: 5, 36: 5, 52: 5, 44: 5, 54: 5, 20: 5, 35: 5, 48: 5, 50: 5, 51: 5, 43: 5, 28: 5, 47: 5, 49: 5 },
    { 35: 42, 8: 42, 37: 42, 24: 42, 9: 42, 11: 42, 30: 42, 12: 42, 15: 42, 49: 42, 29: 42, 51: 42, 4: 42, 18: 42, 56: 42, 0: 43, 33: 42, 38: 42, 39: 42, 41: 42, 17: 42, 10: 42, 57: 42, 6: 42, 26: 42, 40: 42, 1: 42, 2: 42, 48: 42, 7: 42, 21: 42, 19: 42, 55: 42, 13: 42, 46: 42, 31: 42, 16: 42, 23: 42, 45: 42, 53: 42, 44: 42, 28: 42, 3: 43, 22: 42, 50: 42, 20: 42, 34: 42, 5: 43, 36: 42, 25: 42, 54: 42, 27: 42, 43: 42, 42: 42, 47: 42, 52: 42, 14: 42, 32: 42 },
    { },
    { 14: 44, 30: 22, 15: 44, 42: 44, 52: 44, 36: 44, 10: 44, 21: 44, 26: 44, 45: 44, 31: 44, 18: 44, 32: 44, 34: 44, 29: 44, 33: 44, 55: 44, 7: 44, 48: 44, 25: 44, 49: 44, 28: 44, 12: 44, 4: 44, 6: 44, 1: 44, 19: 44, 27: 44, 57: 44, 40: 44, 50: 44, 39: 44, 16: 44, 43: 44, 22: 44, 38: 44, 17: 44, 23: 44, 41: 44, 47: 44, 44: 44, 46: 44, 54: 44, 56: 44, 9: 10, 20: 44, 24: 44, 11: 44, 13: 44, 37: 44, 51: 44, 2: 44, 53: 44, 8: 44, 35: 44 },
    { },
    { },
    { },
    { 46: 5, 37: 5, 42: 5, 54: 5, 50: 5, 39: 5, 49: 5, 51: 5, 53: 5, 33: 5, 28: 5, 38: 5, 47: 5, 45: 5, 40: 5, 20: 5, 43: 5, 44: 5, 35: 5, 36: 5, 52: 7, 41: 5, 48: 5 },
    { 54: 5, 35: 5, 36: 5, 50: 5, 52: 5, 47: 5, 40: 5, 43: 5, 37: 5, 53: 5, 38: 5, 28: 5, 39: 5, 49: 5, 42: 5, 41: 5, 51: 5, 48: 29, 33: 5, 44: 5, 46: 5, 45: 5, 20: 5 },
    { 39: 5, 47: 5, 38: 5, 37: 5, 54: 5, 35: 5, 40: 5, 49: 5, 52: 5, 41: 5, 20: 5, 36: 5, 48: 5, 42: 5, 43: 5, 44: 5, 50: 5, 33: 5, 51: 5, 45: 21, 53: 5, 46: 5, 28: 5 },
    { 20: 5, 47: 5, 48: 5, 35: 5, 33: 5, 43: 5, 42: 5, 54: 5, 45: 5, 39: 5, 50: 5, 53: 5, 46: 5, 40: 5, 51: 5, 28: 5, 37: 5, 41: 5, 38: 5, 49: 5, 44: 5, 52: 41, 36: 5 },
    { 45: 5, 40: 5, 47: 5, 48: 5, 53: 5, 49: 5, 35: 5, 33: 5, 28: 5, 44: 5, 43: 5, 52: 5, 54: 5, 46: 5, 42: 5, 50: 5, 41: 49, 38: 5, 39: 5, 36: 5, 20: 5, 51: 5, 37: 5 },
    { 37: 5, 40: 5, 51: 5, 50: 5, 20: 5, 33: 5, 43: 5, 47: 5, 42: 5, 48: 5, 45: 5, 35: 54, 28: 5, 53: 5, 38: 5, 46: 5, 52: 5, 54: 5, 36: 5, 49: 5, 41: 5, 44: 5, 39: 5 },
    { 37: 5, 45: 5, 47: 5, 53: 5, 49: 5, 36: 5, 44: 5, 50: 5, 38: 5, 28: 5, 35: 5, 43: 5, 20: 5, 33: 5, 51: 5, 48: 5, 39: 50, 41: 5, 52: 5, 54: 5, 40: 5, 42: 5, 46: 5 },
    { 19: 42, 14: 35 },
    { 18: 36, 43: 36, 6: 36, 36: 36, 55: 36, 42: 36, 44: 36, 37: 36, 54: 36, 31: 36, 19: 36, 47: 36, 8: 36, 24: 36, 21: 36, 2: 36, 20: 36, 39: 36, 46: 36, 4: 36, 23: 36, 41: 36, 9: 36, 32: 36, 10: 36, 56: 36, 48: 36, 26: 36, 30: 36, 13: 36, 25: 36, 35: 36, 45: 36, 33: 36, 1: 36, 17: 36, 57: 36, 49: 36, 15: 36, 7: 36, 50: 36, 29: 36, 53: 36, 12: 36, 51: 36, 11: 36, 38: 36, 16: 36, 28: 36, 40: 36, 52: 36, 27: 36, 34: 36, 22: 36, 14: 36 },
}
var accept = map[int]TokenType { 2: 20, 7: 5, 11: 9, 16: 20, 19: 20, 21: 20, 26: 20, 29: 7, 12: 20, 25: 0, 34: 6, 41: 4, 45: 15, 46: 23, 9: 20, 13: 19, 20: 3, 28: 14, 51: 20, 54: 20, 1: 20, 10: 21, 18: 10, 33: 20, 48: 20, 49: 20, 8: 20, 17: 20, 30: 20, 31: 20, 32: 17, 39: 22, 43: 1, 47: 16, 6: 8, 15: 20, 50: 20, 52: 20, 3: 12, 4: 13, 5: 20, 23: 11, 24: 18, 27: 2, 40: 20, 53: 20, 38: 20 }

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
