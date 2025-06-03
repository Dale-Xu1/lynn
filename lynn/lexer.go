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

const (WHITESPACE TokenType = iota; COMMENT; RULE; TOKEN; FRAGMENT; LEFT; RIGHT; ERROR; SKIP; EQUAL; PLUS; STAR; QUESTION; DOT; BAR; HASH; SEMI; COLON; L_PAREN; R_PAREN; ARROW; IDENTIFIER; STRING; CLASS; EOF)
func (t TokenType) String() string { return typeName[t] }
var typeName = map[TokenType]string { 0: "WHITESPACE", 1: "COMMENT", 2: "RULE", 3: "TOKEN", 4: "FRAGMENT", 5: "LEFT", 6: "RIGHT", 7: "ERROR", 8: "SKIP", 9: "EQUAL", 10: "PLUS", 11: "STAR", 12: "QUESTION", 13: "DOT", 14: "BAR", 15: "HASH", 16: "SEMI", 17: "COLON", 18: "L_PAREN", 19: "R_PAREN", 20: "ARROW", 21: "IDENTIFIER", 22: "STRING", 23: "CLASS", 24: "EOF" }
var skip = map[TokenType]struct{} { 0: {}, 1: {} }

var ranges = []Range { { '\x00', '\x00' }, { '\x01', '\b' }, { '\t', '\t' }, { '\n', '\n' }, { '\v', '\f' }, { '\r', '\r' }, { '\x0e', '\x1f' }, { ' ', ' ' }, { '!', '!' }, { '"', '"' }, { '#', '#' }, { '$', '\'' }, { '(', '(' }, { ')', ')' }, { '*', '*' }, { '+', '+' }, { ',', ',' }, { '-', '-' }, { '.', '.' }, { '/', '/' }, { '0', '9' }, { ':', ':' }, { ';', ';' }, { '<', '<' }, { '=', '=' }, { '>', '>' }, { '?', '?' }, { '@', '@' }, { 'A', 'Z' }, { '[', '[' }, { '\\', '\\' }, { ']', ']' }, { '^', '^' }, { '_', '_' }, { '`', '`' }, { 'a', 'a' }, { 'b', 'd' }, { 'e', 'e' }, { 'f', 'f' }, { 'g', 'g' }, { 'h', 'h' }, { 'i', 'i' }, { 'j', 'j' }, { 'k', 'k' }, { 'l', 'l' }, { 'm', 'm' }, { 'n', 'n' }, { 'o', 'o' }, { 'p', 'p' }, { 'q', 'q' }, { 'r', 'r' }, { 's', 's' }, { 't', 't' }, { 'u', 'u' }, { 'v', 'z' }, { '{', '{' }, { '|', '|' }, { '}', '\U0010ffff' } }
var transitions = []map[int]int {
    { 33: 52, 54: 52, 44: 38, 17: 59, 49: 52, 39: 52, 0: 4, 36: 52, 24: 16, 3: 29, 41: 52, 35: 52, 14: 26, 38: 55, 45: 52, 18: 5, 40: 52, 47: 52, 13: 32, 51: 47, 26: 61, 15: 48, 5: 29, 43: 52, 50: 44, 48: 52, 28: 52, 12: 40, 37: 50, 22: 1, 9: 13, 52: 45, 56: 2, 19: 53, 2: 29, 46: 52, 42: 52, 10: 14, 29: 12, 21: 30, 7: 29, 53: 52 },
    { },
    { },
    { 36: 52, 35: 52, 42: 52, 38: 15, 48: 52, 45: 52, 47: 52, 37: 52, 33: 52, 40: 52, 44: 52, 46: 52, 39: 52, 28: 52, 43: 52, 54: 52, 49: 52, 52: 52, 50: 52, 41: 52, 53: 52, 51: 52, 20: 52 },
    { },
    { },
    { 2: 13, 46: 13, 35: 13, 19: 13, 7: 13, 29: 13, 39: 13, 57: 13, 14: 13, 34: 13, 42: 13, 37: 13, 55: 13, 1: 13, 30: 13, 16: 13, 24: 13, 23: 13, 26: 13, 6: 13, 12: 13, 49: 13, 8: 13, 36: 13, 51: 13, 33: 13, 4: 13, 17: 13, 48: 13, 32: 13, 9: 13, 15: 13, 47: 13, 40: 13, 13: 13, 41: 13, 22: 13, 31: 13, 18: 13, 53: 13, 27: 13, 52: 13, 11: 13, 20: 13, 21: 13, 56: 13, 28: 13, 50: 13, 10: 13, 45: 13, 44: 13, 54: 13, 38: 13, 25: 13, 43: 13 },
    { 42: 52, 47: 52, 28: 52, 45: 52, 51: 52, 35: 52, 53: 52, 46: 52, 40: 52, 54: 52, 33: 52, 44: 52, 37: 52, 20: 52, 41: 52, 39: 52, 43: 52, 38: 52, 52: 52, 50: 52, 49: 52, 48: 52, 36: 52 },
    { 1: 8, 6: 8, 13: 8, 46: 8, 25: 8, 50: 8, 21: 8, 11: 8, 2: 8, 23: 8, 42: 8, 44: 8, 39: 8, 56: 8, 12: 8, 41: 8, 24: 8, 30: 8, 15: 8, 0: 46, 28: 8, 22: 8, 10: 8, 18: 8, 53: 8, 7: 8, 40: 8, 36: 8, 57: 8, 31: 8, 27: 8, 8: 8, 20: 8, 9: 8, 16: 8, 51: 8, 32: 8, 47: 8, 38: 8, 49: 8, 37: 8, 19: 8, 34: 8, 29: 8, 5: 46, 52: 8, 14: 8, 35: 8, 17: 8, 54: 8, 4: 8, 26: 8, 48: 8, 45: 8, 55: 8, 33: 8, 43: 8, 3: 46 },
    { },
    { 41: 52, 44: 52, 52: 52, 33: 52, 37: 52, 49: 52, 48: 52, 54: 52, 39: 52, 53: 52, 38: 52, 47: 52, 43: 52, 20: 52, 45: 52, 50: 52, 51: 52, 36: 52, 42: 52, 28: 52, 40: 49, 46: 52, 35: 52 },
    { 48: 52, 39: 52, 47: 52, 40: 52, 38: 52, 50: 52, 41: 52, 49: 52, 43: 52, 54: 52, 45: 52, 20: 52, 37: 52, 53: 52, 46: 52, 51: 52, 28: 52, 36: 52, 42: 52, 33: 52, 44: 52, 52: 52, 35: 52 },
    { 34: 12, 54: 12, 33: 12, 37: 12, 14: 12, 13: 12, 53: 12, 40: 12, 12: 12, 8: 12, 15: 12, 2: 12, 41: 12, 57: 12, 6: 12, 28: 12, 11: 12, 51: 12, 7: 12, 18: 12, 23: 12, 52: 12, 46: 12, 44: 12, 39: 12, 38: 12, 22: 12, 17: 12, 55: 12, 50: 12, 24: 12, 47: 12, 45: 12, 10: 12, 30: 58, 36: 12, 42: 12, 31: 42, 9: 12, 49: 12, 56: 12, 4: 12, 43: 12, 21: 12, 27: 12, 29: 12, 48: 12, 25: 12, 19: 12, 32: 12, 1: 12, 26: 12, 16: 12, 20: 12, 35: 12 },
    { 54: 13, 42: 13, 25: 13, 26: 13, 43: 13, 35: 13, 16: 13, 10: 13, 45: 13, 21: 13, 47: 13, 8: 13, 48: 13, 49: 13, 36: 13, 15: 13, 22: 13, 40: 13, 46: 13, 7: 13, 30: 6, 28: 13, 29: 13, 18: 13, 24: 13, 20: 13, 4: 13, 13: 13, 57: 13, 1: 13, 19: 13, 52: 13, 51: 13, 14: 13, 37: 13, 31: 13, 12: 13, 44: 13, 50: 13, 2: 13, 32: 13, 27: 13, 41: 13, 55: 13, 11: 13, 9: 51, 34: 13, 53: 13, 39: 13, 23: 13, 6: 13, 56: 13, 33: 13, 38: 13, 17: 13 },
    { },
    { 36: 52, 42: 52, 44: 52, 53: 52, 46: 52, 20: 52, 35: 52, 54: 52, 45: 52, 49: 52, 39: 52, 28: 52, 48: 52, 41: 52, 37: 52, 40: 52, 33: 52, 51: 52, 38: 52, 52: 54, 50: 52, 47: 52, 43: 52 },
    { },
    { 37: 52, 48: 52, 43: 52, 46: 52, 42: 52, 40: 52, 53: 52, 49: 52, 47: 52, 33: 52, 50: 52, 41: 52, 36: 52, 52: 52, 54: 52, 45: 52, 20: 52, 39: 52, 28: 52, 35: 52, 38: 52, 51: 52, 44: 52 },
    { 40: 52, 53: 52, 54: 52, 28: 52, 33: 52, 50: 52, 48: 52, 41: 52, 45: 52, 51: 52, 38: 52, 44: 52, 35: 52, 49: 52, 46: 52, 47: 52, 42: 52, 43: 52, 20: 52, 39: 10, 37: 52, 36: 52, 52: 52 },
    { 28: 52, 40: 52, 47: 52, 37: 52, 42: 52, 39: 52, 44: 52, 50: 57, 38: 52, 20: 52, 52: 52, 46: 52, 48: 52, 35: 52, 53: 52, 51: 52, 36: 52, 33: 52, 41: 52, 49: 52, 54: 52, 45: 52, 43: 52 },
    { 35: 52, 39: 52, 47: 52, 53: 52, 42: 52, 54: 52, 20: 52, 50: 52, 51: 52, 38: 52, 33: 52, 46: 52, 48: 52, 45: 52, 49: 52, 36: 52, 44: 52, 43: 52, 37: 52, 41: 52, 28: 52, 52: 52, 40: 52 },
    { 36: 52, 48: 52, 43: 52, 52: 52, 49: 52, 28: 52, 38: 52, 51: 52, 39: 52, 37: 52, 50: 52, 47: 52, 54: 52, 44: 24, 33: 52, 42: 52, 35: 52, 40: 52, 53: 52, 20: 52, 41: 52, 45: 52, 46: 52 },
    { 24: 34, 16: 34, 10: 34, 27: 34, 17: 34, 1: 34, 35: 34, 55: 34, 32: 34, 4: 34, 44: 34, 49: 34, 39: 34, 12: 34, 47: 34, 42: 34, 28: 34, 43: 34, 41: 34, 18: 34, 56: 34, 13: 34, 3: 34, 5: 34, 45: 34, 36: 34, 2: 34, 19: 46, 53: 34, 52: 34, 20: 34, 11: 34, 40: 34, 51: 34, 25: 34, 46: 34, 37: 34, 33: 34, 22: 34, 7: 34, 57: 34, 9: 34, 48: 34, 26: 34, 34: 34, 21: 34, 23: 34, 15: 34, 29: 34, 6: 34, 31: 34, 8: 34, 38: 34, 14: 34, 54: 34, 30: 34, 50: 34 },
    { 53: 52, 36: 52, 35: 52, 28: 52, 52: 52, 44: 52, 48: 52, 40: 52, 45: 52, 51: 52, 54: 52, 46: 35, 47: 52, 41: 52, 43: 52, 20: 52, 37: 52, 38: 52, 42: 52, 49: 52, 33: 52, 50: 52, 39: 52 },
    { 28: 52, 46: 52, 20: 52, 36: 52, 47: 52, 50: 52, 40: 52, 51: 52, 53: 52, 48: 52, 41: 52, 42: 52, 33: 52, 54: 52, 38: 52, 52: 52, 39: 52, 43: 52, 37: 7, 45: 52, 49: 52, 44: 52, 35: 52 },
    { 35: 52, 47: 52, 33: 52, 28: 52, 52: 52, 48: 52, 50: 52, 49: 52, 36: 52, 20: 52, 37: 52, 43: 33, 44: 52, 53: 52, 46: 52, 40: 52, 38: 52, 42: 52, 45: 52, 39: 52, 41: 52, 51: 52, 54: 52 },
    { },
    { 50: 52, 52: 52, 38: 52, 54: 52, 46: 52, 40: 52, 53: 52, 28: 52, 41: 52, 45: 52, 39: 52, 43: 52, 42: 52, 44: 52, 20: 52, 51: 52, 37: 52, 49: 52, 47: 52, 33: 52, 35: 52, 48: 36, 36: 52 },
    { 20: 52, 54: 52, 45: 52, 53: 52, 28: 52, 33: 52, 40: 52, 48: 52, 46: 52, 49: 52, 35: 52, 38: 52, 42: 52, 36: 52, 37: 52, 43: 52, 51: 52, 39: 52, 52: 52, 50: 11, 47: 52, 44: 52, 41: 52 },
    { 5: 29, 7: 29, 2: 29, 3: 29 },
    { },
    { 42: 52, 47: 52, 50: 52, 39: 39, 46: 52, 37: 52, 49: 52, 44: 52, 53: 52, 33: 52, 41: 52, 20: 52, 54: 52, 28: 52, 38: 52, 43: 52, 36: 52, 40: 52, 52: 52, 51: 52, 35: 52, 48: 52, 45: 52 },
    { },
    { 44: 52, 33: 52, 48: 52, 45: 52, 35: 52, 51: 52, 36: 52, 41: 52, 43: 52, 28: 52, 42: 52, 20: 52, 49: 52, 47: 52, 53: 52, 54: 52, 46: 52, 52: 52, 38: 52, 37: 37, 50: 52, 40: 52, 39: 52 },
    { 41: 34, 1: 34, 18: 34, 31: 34, 9: 34, 56: 34, 6: 34, 44: 34, 48: 34, 3: 34, 29: 34, 55: 34, 13: 34, 40: 34, 43: 34, 28: 34, 33: 34, 25: 34, 10: 34, 20: 34, 38: 34, 54: 34, 50: 34, 47: 34, 51: 34, 4: 34, 2: 34, 17: 34, 36: 34, 27: 34, 26: 34, 5: 34, 45: 34, 30: 34, 34: 34, 15: 34, 42: 34, 14: 22, 37: 34, 53: 34, 22: 34, 7: 34, 16: 34, 35: 34, 39: 34, 23: 34, 24: 34, 46: 34, 21: 34, 19: 34, 32: 34, 8: 34, 52: 34, 11: 34, 57: 34, 12: 34, 49: 34 },
    { 42: 52, 33: 52, 41: 52, 39: 52, 37: 52, 40: 52, 45: 52, 36: 52, 50: 52, 44: 52, 38: 52, 47: 52, 49: 52, 54: 52, 43: 52, 52: 17, 20: 52, 46: 52, 35: 52, 51: 52, 28: 52, 48: 52, 53: 52 },
    { 35: 52, 37: 52, 50: 52, 39: 52, 44: 52, 20: 52, 43: 52, 36: 52, 51: 52, 45: 52, 33: 52, 48: 52, 53: 52, 42: 52, 40: 52, 41: 52, 28: 52, 38: 52, 49: 52, 47: 52, 46: 52, 52: 52, 54: 52 },
    { 45: 52, 47: 52, 43: 52, 53: 52, 48: 52, 20: 52, 41: 52, 36: 52, 44: 52, 28: 52, 33: 52, 46: 41, 35: 52, 49: 52, 42: 52, 52: 52, 40: 52, 51: 52, 38: 52, 50: 52, 37: 52, 39: 52, 54: 52 },
    { 39: 52, 43: 52, 47: 52, 51: 52, 37: 3, 44: 52, 52: 52, 36: 52, 35: 52, 54: 52, 41: 52, 53: 52, 20: 52, 46: 52, 45: 52, 50: 52, 49: 52, 48: 52, 28: 52, 38: 52, 42: 52, 33: 52, 40: 52 },
    { 48: 52, 45: 43, 39: 52, 44: 52, 36: 52, 41: 52, 37: 52, 49: 52, 33: 52, 40: 52, 53: 52, 46: 52, 50: 52, 47: 52, 38: 52, 42: 52, 35: 52, 43: 52, 52: 52, 20: 52, 28: 52, 51: 52, 54: 52 },
    { },
    { 47: 52, 50: 52, 40: 52, 35: 52, 48: 52, 51: 52, 43: 52, 45: 52, 49: 52, 33: 52, 52: 52, 36: 52, 38: 52, 46: 52, 20: 52, 44: 52, 53: 52, 37: 52, 54: 52, 39: 52, 41: 52, 28: 52, 42: 52 },
    { },
    { 45: 52, 46: 52, 48: 52, 33: 52, 50: 52, 20: 52, 53: 52, 28: 52, 43: 52, 36: 52, 51: 52, 52: 52, 54: 52, 41: 52, 42: 52, 49: 52, 47: 52, 44: 52, 38: 52, 37: 23, 39: 52, 40: 52, 35: 52 },
    { 40: 52, 42: 52, 44: 52, 46: 52, 45: 52, 39: 52, 41: 18, 20: 52, 47: 52, 49: 52, 43: 52, 52: 52, 28: 52, 38: 52, 36: 52, 50: 52, 53: 21, 35: 52, 51: 52, 37: 52, 48: 52, 33: 52, 54: 52 },
    { 43: 52, 45: 52, 28: 52, 47: 25, 39: 52, 53: 52, 42: 52, 51: 52, 37: 52, 35: 52, 54: 52, 41: 52, 49: 52, 44: 52, 36: 52, 50: 52, 46: 52, 33: 52, 20: 52, 38: 52, 52: 52, 48: 52, 40: 52 },
    { },
    { 40: 52, 51: 52, 45: 52, 41: 52, 42: 52, 33: 52, 48: 52, 36: 52, 53: 52, 20: 52, 43: 60, 47: 52, 49: 52, 39: 52, 28: 52, 50: 52, 46: 52, 38: 52, 52: 52, 35: 52, 37: 52, 54: 52, 44: 52 },
    { },
    { 42: 52, 46: 52, 35: 52, 53: 52, 54: 52, 39: 52, 49: 52, 44: 52, 52: 20, 36: 52, 43: 52, 38: 52, 37: 52, 40: 52, 50: 52, 45: 52, 51: 52, 20: 52, 41: 52, 28: 52, 33: 52, 47: 52, 48: 52 },
    { 43: 52, 44: 52, 53: 52, 47: 52, 28: 52, 35: 52, 39: 52, 40: 52, 48: 52, 54: 52, 38: 52, 52: 52, 37: 52, 51: 52, 36: 52, 49: 52, 42: 52, 33: 52, 50: 19, 20: 52, 41: 52, 46: 52, 45: 52 },
    { },
    { 48: 52, 28: 52, 35: 52, 45: 52, 39: 52, 43: 52, 36: 52, 52: 52, 40: 52, 44: 52, 20: 52, 46: 52, 54: 52, 41: 52, 47: 52, 37: 52, 42: 52, 53: 52, 50: 52, 51: 52, 33: 52, 49: 52, 38: 52 },
    { 14: 34, 19: 8 },
    { 42: 52, 50: 52, 20: 52, 36: 52, 37: 52, 48: 52, 49: 52, 39: 52, 44: 52, 46: 52, 45: 52, 28: 52, 52: 52, 54: 52, 41: 52, 38: 52, 47: 52, 43: 52, 53: 52, 40: 52, 35: 52, 51: 52, 33: 52 },
    { 50: 56, 46: 52, 40: 52, 49: 52, 48: 52, 20: 52, 54: 52, 45: 52, 51: 52, 39: 52, 42: 52, 33: 52, 44: 52, 36: 52, 43: 52, 28: 52, 37: 52, 53: 52, 52: 52, 47: 52, 38: 52, 35: 52, 41: 52 },
    { 46: 52, 48: 52, 52: 52, 54: 52, 45: 52, 35: 31, 36: 52, 28: 52, 20: 52, 47: 52, 50: 52, 44: 52, 41: 52, 51: 52, 39: 52, 42: 52, 40: 52, 33: 52, 49: 52, 43: 52, 37: 52, 53: 52, 38: 52 },
    { 48: 52, 28: 52, 36: 52, 45: 52, 37: 52, 20: 52, 44: 52, 40: 52, 39: 52, 49: 52, 42: 52, 50: 52, 41: 52, 47: 28, 54: 52, 51: 52, 53: 52, 46: 52, 38: 52, 33: 52, 35: 52, 43: 52, 52: 52 },
    { 2: 12, 7: 12, 15: 12, 37: 12, 57: 12, 20: 12, 10: 12, 53: 12, 40: 12, 4: 12, 47: 12, 51: 12, 56: 12, 19: 12, 8: 12, 34: 12, 1: 12, 13: 12, 25: 12, 49: 12, 26: 12, 11: 12, 39: 12, 30: 12, 43: 12, 16: 12, 36: 12, 48: 12, 9: 12, 12: 12, 6: 12, 28: 12, 46: 12, 22: 12, 17: 12, 52: 12, 38: 12, 23: 12, 35: 12, 21: 12, 29: 12, 14: 12, 33: 12, 54: 12, 55: 12, 27: 12, 42: 12, 24: 12, 50: 12, 31: 12, 41: 12, 18: 12, 45: 12, 44: 12, 32: 12 },
    { 25: 9 },
    { 44: 52, 38: 52, 28: 52, 41: 27, 49: 52, 39: 52, 48: 52, 54: 52, 51: 52, 37: 52, 35: 52, 50: 52, 43: 52, 42: 52, 45: 52, 40: 52, 20: 52, 36: 52, 46: 52, 53: 52, 47: 52, 33: 52, 52: 52 },
    { },
}
var accept = map[int]TokenType { 18: 21, 23: 21, 32: 19, 33: 21, 38: 21, 41: 3, 50: 21, 16: 9, 37: 21, 40: 18, 43: 21, 45: 21, 52: 21, 54: 5, 55: 21, 7: 2, 15: 21, 20: 6, 35: 21, 47: 21, 56: 21, 60: 21, 14: 15, 19: 21, 24: 21, 28: 21, 31: 21, 46: 1, 48: 10, 5: 13, 27: 21, 51: 22, 57: 21, 4: 24, 9: 20, 11: 7, 25: 21, 30: 17, 61: 12, 1: 16, 2: 14, 3: 21, 21: 21, 29: 0, 36: 8, 39: 21, 17: 4, 26: 11, 42: 23, 44: 21, 49: 21, 10: 21 }

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
