package parser

import "fmt"

func (r Range) String() string {
    if r.Min == r.Max { return FormatChar(r.Min) }
    return fmt.Sprintf("%s-%s", FormatChar(r.Min), FormatChar(r.Max))
}

func FormatChar(char rune) string {
    str := fmt.Sprintf("%q", string(char))
    return str[1:len(str) - 1]
}
