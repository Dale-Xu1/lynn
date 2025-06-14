package lynn

import (
	"bufio"
	"fmt"
	"os"
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

const (WHITESPACE TokenType = iota; COMMENT; RULE; PRECEDENCE; TOKEN; FRAGMENT; LEFT; RIGHT; ERROR; SKIP; EQUAL; PLUS; STAR; QUESTION; DOT; BAR; HASH; PERCENT; SEMI; COLON; L_PAREN; R_PAREN; ARROW; IDENTIFIER; STRING; CLASS; EOF)
func (t TokenType) String() string { return typeName[t] }
var typeName = map[TokenType]string { 0: "WHITESPACE", 1: "COMMENT", 2: "RULE", 3: "PRECEDENCE", 4: "TOKEN", 5: "FRAGMENT", 6: "LEFT", 7: "RIGHT", 8: "ERROR", 9: "SKIP", 10: "EQUAL", 11: "PLUS", 12: "STAR", 13: "QUESTION", 14: "DOT", 15: "BAR", 16: "HASH", 17: "PERCENT", 18: "SEMI", 19: "COLON", 20: "L_PAREN", 21: "R_PAREN", 22: "ARROW", 23: "IDENTIFIER", 24: "STRING", 25: "CLASS", 26: "EOF" }
var skip = map[TokenType]struct{} { 0: {}, 1: {} }

var ranges = []Range { { '\x00', '\x00' }, { '\x01', '\b' }, { '\t', '\t' }, { '\n', '\n' }, { '\v', '\f' }, { '\r', '\r' }, { '\x0e', '\x1f' }, { ' ', ' ' }, { '!', '!' }, { '"', '"' }, { '#', '#' }, { '$', '$' }, { '%', '%' }, { '&', '\'' }, { '(', '(' }, { ')', ')' }, { '*', '*' }, { '+', '+' }, { ',', ',' }, { '-', '-' }, { '.', '.' }, { '/', '/' }, { '0', '9' }, { ':', ':' }, { ';', ';' }, { '<', '<' }, { '=', '=' }, { '>', '>' }, { '?', '?' }, { '@', '@' }, { 'A', 'F' }, { 'G', 'T' }, { 'U', 'U' }, { 'V', 'Z' }, { '[', '[' }, { '\\', '\\' }, { ']', ']' }, { '^', '^' }, { '_', '_' }, { '`', '`' }, { 'a', 'a' }, { 'b', 'b' }, { 'c', 'c' }, { 'd', 'd' }, { 'e', 'e' }, { 'f', 'f' }, { 'g', 'g' }, { 'h', 'h' }, { 'i', 'i' }, { 'j', 'j' }, { 'k', 'k' }, { 'l', 'l' }, { 'm', 'm' }, { 'n', 'n' }, { 'o', 'o' }, { 'p', 'p' }, { 'q', 'q' }, { 'r', 'r' }, { 's', 's' }, { 't', 't' }, { 'u', 'u' }, { 'v', 'w' }, { 'x', 'x' }, { 'y', 'z' }, { '{', '{' }, { '|', '|' }, { '}', '\U0010ffff' } }
var transitions = []map[int]int {
    { 31: 66, 63: 66, 0: 39, 47: 66, 48: 66, 57: 57, 33: 66, 26: 50, 46: 66, 19: 77, 58: 26, 51: 53, 24: 60, 49: 66, 17: 40, 56: 66, 52: 66, 10: 17, 41: 66, 34: 3, 16: 74, 55: 41, 62: 66, 21: 78, 2: 1, 20: 14, 45: 4, 32: 66, 61: 66, 12: 65, 14: 43, 43: 66, 38: 66, 3: 1, 40: 66, 28: 55, 53: 66, 65: 23, 42: 66, 50: 66, 60: 66, 7: 1, 9: 45, 5: 1, 44: 68, 59: 63, 23: 44, 15: 9, 30: 66, 54: 66 },
    { 7: 1, 2: 1, 3: 1, 5: 1 },
    { 45: 47, 22: 47, 30: 47, 40: 47, 41: 47, 42: 47, 43: 47, 44: 47 },
    { 2: 3, 1: 3, 11: 3, 41: 3, 58: 3, 44: 3, 61: 3, 39: 3, 19: 3, 62: 3, 45: 3, 56: 3, 50: 3, 31: 3, 12: 3, 22: 3, 34: 3, 4: 3, 57: 3, 15: 3, 8: 3, 35: 19, 7: 3, 48: 3, 24: 3, 21: 3, 25: 3, 59: 3, 46: 3, 38: 3, 29: 3, 53: 3, 49: 3, 37: 3, 60: 3, 52: 3, 36: 18, 27: 3, 55: 3, 13: 3, 51: 3, 33: 3, 6: 3, 26: 3, 16: 3, 65: 3, 17: 3, 20: 3, 54: 3, 30: 3, 9: 3, 10: 3, 40: 3, 18: 3, 64: 3, 43: 3, 47: 3, 63: 3, 23: 3, 28: 3, 42: 3, 32: 3, 66: 3, 14: 3 },
    { 42: 66, 53: 66, 60: 66, 55: 66, 31: 66, 33: 66, 59: 66, 62: 66, 54: 66, 38: 66, 46: 66, 43: 66, 56: 66, 52: 66, 30: 66, 58: 66, 44: 66, 61: 66, 49: 66, 48: 66, 40: 66, 32: 66, 57: 75, 22: 66, 50: 66, 51: 66, 47: 66, 45: 66, 41: 66, 63: 66 },
    { 61: 66, 52: 66, 43: 66, 53: 66, 48: 66, 41: 66, 22: 66, 59: 66, 40: 66, 62: 66, 44: 66, 58: 66, 54: 66, 51: 66, 42: 66, 56: 66, 60: 66, 49: 66, 45: 66, 32: 66, 38: 66, 50: 66, 57: 66, 55: 66, 31: 66, 30: 66, 47: 34, 63: 66, 46: 66, 33: 66 },
    { 54: 66, 52: 66, 31: 66, 38: 66, 44: 66, 46: 66, 41: 66, 47: 66, 40: 66, 48: 66, 22: 66, 43: 66, 60: 66, 62: 66, 50: 66, 33: 66, 63: 66, 58: 66, 59: 66, 45: 66, 51: 66, 55: 22, 56: 66, 32: 66, 57: 66, 53: 66, 49: 66, 30: 66, 61: 66, 42: 66 },
    { 51: 66, 54: 66, 44: 66, 61: 66, 59: 29, 62: 66, 31: 66, 38: 66, 48: 66, 32: 66, 33: 66, 58: 66, 56: 66, 53: 66, 46: 66, 63: 66, 49: 66, 41: 66, 47: 66, 22: 66, 57: 66, 60: 66, 40: 66, 42: 66, 52: 66, 45: 66, 50: 66, 43: 66, 55: 66, 30: 66 },
    { 43: 62, 44: 62, 30: 62, 40: 62, 45: 62, 22: 62, 41: 62, 42: 62 },
    { },
    { 40: 2, 41: 2, 42: 2, 43: 2, 44: 2, 45: 2, 22: 2, 30: 2 },
    { 41: 3, 42: 3, 44: 3, 45: 3, 40: 3, 43: 3, 22: 3, 30: 3 },
    { 51: 66, 48: 66, 43: 66, 53: 66, 47: 66, 38: 66, 30: 66, 46: 66, 40: 66, 31: 66, 63: 66, 56: 66, 44: 66, 32: 66, 50: 66, 54: 66, 60: 66, 41: 66, 52: 66, 57: 66, 61: 66, 62: 66, 59: 66, 42: 66, 33: 66, 55: 66, 49: 66, 45: 66, 22: 66, 58: 66 },
    { },
    { },
    { 43: 24, 45: 24, 30: 24, 40: 24, 42: 24, 44: 24, 22: 24, 41: 24 },
    { 30: 48, 40: 48, 41: 48, 42: 48, 43: 48, 44: 48, 45: 48, 22: 48 },
    { },
    { },
    { 38: 3, 13: 3, 44: 3, 55: 3, 53: 3, 27: 3, 23: 3, 14: 3, 45: 3, 36: 3, 56: 3, 59: 3, 47: 3, 41: 3, 51: 3, 18: 3, 26: 3, 28: 3, 19: 3, 17: 3, 34: 3, 32: 61, 63: 3, 11: 3, 25: 3, 10: 3, 52: 3, 39: 3, 24: 3, 58: 3, 21: 3, 60: 8, 61: 3, 6: 3, 48: 3, 20: 3, 33: 3, 65: 3, 29: 3, 16: 3, 15: 3, 43: 3, 66: 3, 49: 3, 31: 3, 7: 3, 37: 3, 54: 3, 8: 3, 4: 3, 30: 3, 40: 3, 64: 3, 50: 3, 62: 27, 1: 3, 22: 3, 2: 3, 46: 3, 12: 3, 57: 3, 35: 3, 9: 3, 42: 3 },
    { 63: 66, 41: 66, 38: 66, 30: 66, 32: 66, 46: 66, 57: 66, 50: 66, 54: 66, 60: 66, 58: 66, 40: 66, 49: 66, 44: 30, 48: 66, 61: 66, 62: 66, 43: 66, 33: 66, 42: 66, 53: 66, 59: 66, 55: 66, 45: 66, 47: 66, 52: 66, 56: 66, 31: 66, 22: 66, 51: 66 },
    { 47: 66, 43: 66, 44: 52, 58: 66, 56: 66, 57: 66, 32: 66, 60: 66, 22: 66, 40: 66, 54: 66, 42: 66, 53: 66, 62: 66, 63: 66, 52: 66, 46: 66, 30: 66, 55: 66, 38: 66, 41: 66, 48: 66, 33: 66, 51: 66, 61: 66, 49: 66, 59: 66, 45: 66, 50: 66, 31: 66 },
    { 61: 66, 54: 66, 47: 66, 50: 66, 56: 66, 38: 66, 43: 66, 46: 66, 33: 66, 52: 66, 57: 66, 42: 66, 45: 66, 30: 66, 40: 66, 48: 66, 62: 66, 55: 66, 51: 66, 41: 66, 49: 66, 22: 66, 60: 66, 44: 66, 59: 66, 32: 66, 53: 66, 63: 66, 31: 66, 58: 66 },
    { },
    { 30: 56, 40: 56, 41: 56, 42: 56, 43: 56, 44: 56, 45: 56, 22: 56 },
    { 42: 66, 44: 66, 57: 49, 47: 66, 50: 66, 54: 66, 55: 66, 31: 66, 43: 66, 61: 66, 59: 66, 30: 66, 33: 66, 45: 66, 32: 66, 56: 66, 48: 66, 41: 66, 52: 66, 53: 66, 58: 66, 22: 66, 62: 66, 49: 66, 46: 66, 60: 66, 63: 66, 38: 66, 51: 66, 40: 66 },
    { 58: 66, 22: 66, 61: 66, 43: 66, 44: 66, 42: 66, 62: 66, 55: 66, 30: 66, 31: 66, 38: 66, 32: 66, 63: 66, 45: 66, 47: 66, 53: 66, 50: 51, 52: 66, 60: 66, 54: 66, 57: 66, 33: 66, 59: 66, 46: 66, 48: 66, 56: 66, 51: 66, 40: 66, 49: 66, 41: 66 },
    { 41: 11, 42: 11, 43: 11, 44: 11, 45: 11, 22: 11, 30: 11, 40: 11 },
    { 55: 66, 45: 66, 43: 66, 62: 66, 56: 66, 61: 66, 47: 66, 58: 66, 63: 66, 38: 66, 57: 66, 50: 66, 30: 66, 54: 66, 48: 66, 59: 66, 32: 66, 40: 66, 46: 66, 49: 66, 52: 66, 41: 66, 60: 66, 31: 66, 53: 66, 33: 66, 42: 66, 51: 66, 44: 66, 22: 66 },
    { 22: 66, 62: 66, 56: 66, 48: 66, 63: 66, 59: 66, 50: 66, 52: 66, 33: 66, 40: 66, 45: 66, 38: 66, 43: 66, 30: 66, 53: 66, 54: 66, 44: 66, 58: 66, 49: 66, 55: 66, 57: 66, 60: 66, 46: 66, 41: 66, 47: 66, 31: 66, 32: 66, 51: 66, 42: 66, 61: 66 },
    { 59: 66, 51: 66, 22: 66, 31: 66, 61: 66, 50: 66, 43: 66, 56: 66, 46: 66, 48: 66, 49: 66, 32: 66, 53: 66, 57: 66, 45: 66, 44: 66, 47: 66, 63: 66, 33: 66, 60: 66, 55: 66, 52: 66, 30: 66, 62: 66, 41: 66, 42: 12, 54: 66, 38: 66, 40: 66, 58: 66 },
    { 33: 31, 35: 31, 27: 31, 1: 31, 62: 31, 50: 31, 37: 31, 6: 31, 29: 31, 13: 31, 55: 31, 25: 31, 11: 31, 45: 31, 44: 31, 52: 31, 17: 31, 66: 31, 40: 31, 63: 31, 59: 31, 34: 31, 36: 31, 54: 31, 43: 31, 42: 31, 16: 32, 38: 31, 4: 31, 64: 31, 32: 31, 23: 31, 39: 31, 57: 31, 61: 31, 24: 31, 2: 31, 21: 31, 53: 31, 19: 31, 20: 31, 47: 31, 46: 31, 26: 31, 8: 31, 31: 31, 41: 31, 48: 31, 60: 31, 15: 31, 3: 31, 18: 31, 10: 31, 9: 31, 58: 31, 12: 31, 14: 31, 22: 31, 49: 31, 30: 31, 65: 31, 7: 31, 28: 31, 5: 31, 51: 31, 56: 31 },
    { 7: 31, 53: 31, 65: 31, 23: 31, 44: 31, 54: 31, 43: 31, 41: 31, 60: 31, 37: 31, 57: 31, 12: 31, 11: 31, 42: 31, 17: 31, 64: 31, 28: 31, 39: 31, 38: 31, 40: 31, 48: 31, 30: 31, 19: 31, 46: 31, 33: 31, 45: 31, 10: 31, 66: 31, 16: 31, 21: 13, 25: 31, 49: 31, 9: 31, 51: 31, 4: 31, 52: 31, 32: 31, 27: 31, 13: 31, 58: 31, 2: 31, 24: 31, 56: 31, 35: 31, 47: 31, 15: 31, 5: 31, 6: 31, 29: 31, 31: 31, 18: 31, 59: 31, 62: 31, 50: 31, 55: 31, 8: 31, 14: 31, 22: 31, 63: 31, 34: 31, 36: 31, 26: 31, 3: 31, 1: 31, 61: 31, 20: 31 },
    { 57: 28, 50: 66, 59: 66, 61: 66, 49: 66, 58: 66, 40: 66, 33: 66, 54: 66, 46: 66, 62: 66, 53: 66, 48: 66, 45: 66, 47: 66, 55: 66, 32: 66, 51: 66, 56: 66, 22: 66, 41: 66, 42: 66, 30: 66, 31: 66, 52: 66, 43: 66, 60: 66, 63: 66, 38: 66, 44: 66 },
    { 42: 66, 51: 66, 44: 66, 22: 66, 31: 66, 47: 66, 63: 66, 43: 66, 45: 66, 32: 66, 30: 66, 49: 66, 38: 66, 60: 66, 58: 66, 52: 66, 57: 66, 41: 66, 40: 66, 53: 66, 48: 66, 33: 66, 50: 66, 62: 66, 46: 66, 54: 66, 55: 66, 59: 35, 61: 66, 56: 66 },
    { 54: 66, 22: 66, 32: 66, 59: 66, 50: 66, 42: 66, 40: 66, 30: 66, 55: 66, 57: 66, 52: 66, 47: 66, 46: 66, 62: 66, 63: 66, 44: 66, 45: 66, 49: 66, 48: 66, 33: 66, 38: 66, 58: 66, 51: 66, 60: 66, 41: 66, 53: 66, 43: 66, 61: 66, 56: 66, 31: 66 },
    { 48: 66, 57: 66, 40: 66, 53: 66, 61: 66, 50: 66, 42: 66, 60: 66, 49: 66, 22: 66, 54: 66, 45: 66, 38: 66, 63: 66, 33: 66, 52: 66, 47: 66, 55: 66, 32: 66, 51: 66, 62: 66, 31: 66, 30: 66, 41: 66, 44: 64, 59: 66, 58: 66, 56: 66, 46: 66, 43: 66 },
    { 31: 37, 26: 37, 49: 37, 22: 37, 8: 37, 6: 37, 50: 37, 9: 37, 46: 37, 28: 37, 27: 37, 0: 13, 41: 37, 61: 37, 10: 37, 57: 37, 14: 37, 47: 37, 36: 37, 18: 37, 20: 37, 3: 13, 7: 37, 25: 37, 59: 37, 5: 13, 54: 37, 12: 37, 56: 37, 52: 37, 63: 37, 29: 37, 1: 37, 16: 37, 62: 37, 44: 37, 48: 37, 30: 37, 66: 37, 43: 37, 53: 37, 51: 37, 23: 37, 19: 37, 21: 37, 15: 37, 17: 37, 34: 37, 35: 37, 33: 37, 45: 37, 58: 37, 32: 37, 60: 37, 64: 37, 40: 37, 42: 37, 11: 37, 55: 37, 2: 37, 24: 37, 4: 37, 39: 37, 13: 37, 38: 37, 65: 37, 37: 37 },
    { 52: 66, 22: 66, 32: 66, 41: 66, 42: 66, 47: 66, 57: 66, 33: 66, 43: 66, 51: 66, 62: 66, 54: 66, 49: 66, 55: 66, 45: 66, 44: 66, 48: 66, 30: 66, 58: 66, 38: 66, 40: 66, 60: 66, 53: 66, 63: 66, 59: 66, 46: 66, 50: 21, 31: 66, 56: 66, 61: 66 },
    { },
    { },
    { 22: 66, 49: 66, 63: 66, 42: 66, 44: 66, 56: 66, 53: 66, 60: 66, 48: 66, 51: 66, 31: 66, 54: 66, 32: 66, 30: 66, 55: 66, 38: 66, 41: 66, 47: 66, 33: 66, 50: 66, 58: 66, 46: 66, 57: 20, 52: 66, 40: 66, 59: 66, 43: 66, 61: 66, 62: 66, 45: 66 },
    { 55: 66, 43: 66, 51: 66, 42: 66, 59: 66, 50: 66, 38: 66, 48: 66, 57: 66, 60: 66, 56: 66, 49: 66, 32: 66, 30: 66, 61: 66, 63: 66, 58: 66, 31: 66, 45: 66, 44: 66, 52: 66, 62: 66, 41: 66, 47: 66, 46: 54, 54: 66, 33: 66, 22: 66, 40: 66, 53: 66 },
    { },
    { },
    { 38: 45, 56: 45, 64: 45, 4: 45, 19: 45, 15: 45, 63: 45, 17: 45, 34: 45, 57: 45, 23: 45, 43: 45, 2: 45, 48: 45, 9: 46, 58: 45, 32: 45, 8: 45, 60: 45, 50: 45, 55: 45, 39: 45, 29: 45, 7: 45, 21: 45, 24: 45, 22: 45, 45: 45, 18: 45, 52: 45, 35: 67, 61: 45, 46: 45, 36: 45, 16: 45, 1: 45, 54: 45, 51: 45, 62: 45, 6: 45, 44: 45, 53: 45, 40: 45, 12: 45, 66: 45, 47: 45, 28: 45, 59: 45, 37: 45, 26: 45, 13: 45, 20: 45, 25: 45, 49: 45, 14: 45, 41: 45, 27: 45, 10: 45, 31: 45, 30: 45, 42: 45, 65: 45, 11: 45, 33: 45 },
    { },
    { 40: 15, 41: 15, 42: 15, 43: 15, 44: 15, 45: 15, 22: 15, 30: 15 },
    { 40: 10, 41: 10, 42: 10, 43: 10, 44: 10, 45: 10, 22: 10, 30: 10 },
    { 40: 66, 55: 66, 53: 66, 30: 66, 44: 66, 60: 66, 41: 66, 45: 66, 46: 66, 62: 66, 48: 66, 47: 66, 42: 66, 43: 66, 33: 66, 61: 66, 58: 66, 32: 66, 22: 66, 56: 66, 54: 33, 59: 66, 57: 66, 63: 66, 49: 66, 51: 66, 52: 66, 50: 66, 31: 66, 38: 66 },
    { },
    { 40: 66, 54: 66, 33: 66, 42: 66, 62: 66, 63: 66, 56: 66, 22: 66, 58: 66, 30: 66, 38: 66, 44: 66, 51: 66, 43: 66, 52: 66, 41: 66, 32: 66, 45: 66, 61: 66, 47: 66, 46: 66, 49: 66, 50: 66, 53: 66, 48: 6, 59: 66, 55: 66, 57: 66, 60: 66, 31: 66 },
    { 22: 66, 46: 66, 41: 66, 60: 66, 51: 66, 62: 66, 49: 66, 52: 66, 32: 66, 30: 66, 55: 66, 40: 66, 54: 66, 56: 66, 44: 66, 59: 66, 61: 66, 43: 66, 50: 66, 33: 66, 45: 66, 53: 71, 48: 66, 58: 66, 38: 66, 42: 66, 31: 66, 47: 66, 63: 66, 57: 66 },
    { 51: 66, 44: 59, 49: 66, 59: 66, 32: 66, 41: 66, 30: 66, 60: 66, 38: 66, 33: 66, 57: 66, 47: 66, 50: 66, 22: 66, 53: 66, 62: 66, 55: 66, 45: 66, 46: 66, 43: 66, 54: 66, 63: 66, 61: 66, 52: 66, 31: 66, 40: 66, 42: 66, 56: 66, 58: 66, 48: 66 },
    { 44: 66, 40: 66, 49: 66, 54: 66, 62: 66, 31: 66, 41: 66, 63: 66, 33: 66, 52: 66, 56: 66, 60: 66, 45: 66, 30: 66, 51: 66, 61: 66, 46: 66, 42: 66, 59: 66, 58: 66, 38: 66, 43: 66, 22: 66, 50: 66, 47: 66, 48: 66, 55: 66, 32: 66, 53: 66, 57: 66 },
    { },
    { 30: 45, 43: 45, 45: 45, 40: 45, 41: 45, 42: 45, 44: 45, 22: 45 },
    { 30: 66, 53: 66, 46: 66, 41: 66, 22: 66, 48: 58, 38: 66, 58: 66, 40: 66, 47: 66, 51: 66, 63: 66, 57: 66, 54: 66, 60: 76, 45: 66, 42: 66, 33: 66, 62: 66, 49: 66, 52: 66, 55: 66, 56: 66, 32: 66, 43: 66, 44: 66, 50: 66, 31: 66, 61: 66, 59: 66 },
    { 60: 66, 43: 66, 62: 66, 63: 66, 53: 66, 59: 66, 49: 66, 40: 66, 54: 66, 46: 5, 48: 66, 44: 66, 41: 66, 22: 66, 47: 66, 42: 66, 52: 66, 45: 66, 30: 66, 32: 66, 55: 66, 58: 66, 61: 66, 38: 66, 50: 66, 51: 66, 33: 66, 57: 66, 31: 66, 56: 66 },
    { 61: 66, 53: 66, 59: 66, 51: 66, 55: 66, 44: 66, 57: 66, 47: 66, 49: 66, 31: 66, 63: 66, 46: 66, 30: 66, 48: 66, 32: 66, 40: 66, 56: 66, 42: 66, 50: 66, 62: 66, 58: 66, 22: 66, 41: 66, 60: 66, 33: 66, 38: 66, 45: 7, 54: 66, 52: 66, 43: 66 },
    { },
    { 42: 69, 43: 69, 44: 69, 45: 69, 22: 69, 30: 69, 40: 69, 41: 69 },
    { 22: 27, 30: 27, 43: 27, 44: 27, 45: 27, 40: 27, 41: 27, 42: 27 },
    { 30: 66, 41: 66, 40: 66, 45: 66, 53: 66, 22: 66, 52: 66, 38: 66, 44: 66, 31: 66, 56: 66, 47: 66, 61: 66, 63: 66, 57: 66, 58: 66, 54: 38, 43: 66, 59: 66, 51: 66, 60: 66, 32: 66, 46: 66, 33: 66, 42: 66, 55: 66, 50: 66, 48: 66, 62: 66, 49: 66 },
    { 38: 66, 57: 66, 62: 66, 55: 66, 58: 66, 56: 66, 47: 66, 31: 66, 32: 66, 53: 66, 61: 66, 33: 66, 50: 66, 52: 66, 40: 66, 46: 66, 45: 66, 48: 66, 42: 66, 30: 66, 54: 66, 63: 66, 44: 66, 43: 66, 59: 66, 41: 66, 49: 66, 22: 66, 51: 66, 60: 66 },
    { },
    { 31: 66, 38: 66, 53: 66, 63: 66, 40: 66, 30: 66, 44: 66, 41: 66, 49: 66, 51: 66, 58: 66, 54: 66, 61: 66, 50: 66, 33: 66, 45: 66, 56: 66, 43: 66, 59: 66, 47: 66, 60: 66, 55: 66, 42: 66, 62: 66, 48: 66, 57: 66, 32: 66, 46: 66, 22: 66, 52: 66 },
    { 51: 45, 9: 45, 24: 45, 6: 45, 17: 45, 11: 45, 66: 45, 55: 45, 14: 45, 10: 45, 61: 45, 23: 45, 60: 47, 58: 45, 33: 45, 59: 45, 27: 45, 19: 45, 48: 45, 7: 45, 50: 45, 4: 45, 18: 45, 56: 45, 43: 45, 63: 45, 34: 45, 65: 45, 16: 45, 35: 45, 47: 45, 8: 45, 32: 16, 21: 45, 20: 45, 29: 45, 2: 45, 31: 45, 64: 45, 28: 45, 38: 45, 12: 45, 30: 45, 57: 45, 52: 45, 37: 45, 44: 45, 39: 45, 22: 45, 25: 45, 54: 45, 26: 45, 53: 45, 62: 24, 42: 45, 41: 45, 1: 45, 45: 45, 46: 45, 49: 45, 36: 45, 40: 45, 13: 45, 15: 45 },
    { 46: 66, 44: 66, 38: 66, 41: 66, 33: 66, 56: 66, 50: 66, 51: 66, 49: 66, 32: 66, 47: 66, 58: 66, 52: 66, 30: 66, 42: 66, 40: 66, 57: 25, 54: 66, 53: 66, 59: 66, 31: 66, 62: 66, 55: 66, 43: 66, 45: 66, 61: 66, 22: 66, 63: 66, 60: 66, 48: 66 },
    { 41: 73, 42: 73, 43: 73, 44: 73, 45: 73, 22: 73, 30: 73, 40: 73 },
    { 22: 8, 30: 8, 40: 8, 42: 8, 44: 8, 41: 8, 43: 8, 45: 8 },
    { 42: 66, 38: 66, 61: 66, 48: 66, 30: 66, 57: 66, 40: 66, 53: 66, 41: 66, 31: 66, 22: 66, 32: 66, 60: 66, 55: 66, 46: 66, 54: 66, 49: 66, 50: 66, 59: 66, 62: 66, 56: 66, 44: 66, 33: 66, 63: 66, 47: 66, 58: 66, 45: 66, 52: 66, 43: 66, 51: 66 },
    { },
    { 30: 70, 40: 70, 41: 70, 42: 70, 43: 70, 44: 70, 45: 70, 22: 70 },
    { },
    { 57: 66, 42: 66, 51: 66, 49: 66, 45: 66, 46: 66, 33: 66, 63: 66, 44: 66, 22: 66, 60: 66, 40: 42, 43: 66, 56: 66, 54: 66, 48: 66, 53: 66, 59: 66, 61: 66, 58: 66, 31: 66, 30: 66, 41: 66, 52: 66, 62: 66, 47: 66, 32: 66, 55: 66, 50: 66, 38: 66 },
    { 49: 66, 22: 66, 46: 66, 59: 66, 48: 66, 55: 66, 33: 66, 58: 66, 45: 66, 50: 66, 62: 66, 43: 66, 54: 66, 52: 66, 63: 66, 31: 66, 38: 66, 47: 66, 44: 66, 42: 66, 57: 66, 60: 66, 41: 66, 51: 36, 56: 66, 30: 66, 61: 66, 32: 66, 40: 66, 53: 66 },
    { 27: 72 },
    { 16: 31, 21: 37 },
}
var accept = map[int]TokenType { 9: 21, 28: 8, 35: 7, 59: 23, 50: 10, 57: 23, 58: 23, 63: 23, 68: 23, 76: 23, 23: 15, 26: 23, 66: 23, 14: 14, 17: 16, 21: 23, 29: 6, 30: 23, 36: 23, 13: 1, 20: 23, 25: 23, 33: 23, 38: 23, 40: 11, 42: 23, 52: 23, 41: 23, 51: 23, 53: 23, 55: 13, 65: 17, 72: 22, 74: 12, 7: 23, 12: 3, 18: 25, 22: 9, 43: 20, 46: 24, 64: 2, 71: 4, 34: 23, 39: 26, 44: 19, 49: 23, 54: 5, 60: 18, 75: 23, 1: 0, 4: 23, 5: 23, 6: 23 }

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
    fmt.Fprintf(os.Stderr, "Syntax error: Unexpected %s - %d:%d\n", str, location.Line, location.Col)
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
    var char rune
    for {
        // Read current character in stream and add to input
        char = l.stream.Read()
        input = append(input, char)
        next, ok := transitions[state][searchRange(char)]
        // Exit loop if we cannot transition from this state on the character
        if !ok { l.stream.Unread(); break }
        // Store the visited states since the last occurring accepting state
        if _, ok := accept[state]; ok { stack = stack[:0] }
        stack = append(stack, state)
        state = next
        i++
    }
    // Backtrack to last accepting state
    location := l.stream.location
    var token TokenType
    for {
        // Unread current character
        if t, ok := accept[state]; ok { token = t; break }
        if len(stack) == 0 {
            // If no accepting state was encountered, raise error and synchronize
            l.stream.synchronize(l.handler, char, location)
            return l.Next() // Attempt to read token again
        }
        // Restore previously visited states
        state, stack = stack[len(stack) - 1], stack[:len(stack) - 1]
        l.stream.Unread()
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
    case '\t': l.Col += 5 - l.Col % 4
    default: l.Col++
    }
    return char
}

// Unreads the current character in the input stream while maintaining location.
func (i *InputStream) Unread() {
    if len(i.stack) == 0 { return }
    data := i.stack[len(i.stack) - 1]; i.stack = i.stack[:len(i.stack) - 1]
    l := i.location; i.location = data.location
    i.buffer = append(i.buffer, streamData { data.char, l })
}

// Releases previously read characters.
func (i *InputStream) reset() { i.stack = i.stack[:0] }
func (i *InputStream) synchronize(handler LexerErrorHandler, char rune, location Location) {
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
