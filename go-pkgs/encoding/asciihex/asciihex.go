package asciihex

import (
	"fmt"
	"strconv"
	"strings"
)

// Encode formats each byte as a lowercase `\xHH` group and does not append `\n`.
// Empty and nil inputs yield "".
func Encode(data []byte) string {
	var b strings.Builder
	b.Grow(len(data) * 4)
	for _, c := range data {
		fmt.Fprintf(&b, "\\x%02x", c)
	}
	return b.String()
}

// Decode walks s in steps of 4 (`\xHH`) and returns raw bytes (`\xff` is 0xff).
func Decode(s string) ([]byte, error) {
	if len(s) < 4 || s[0] != '\\' || s[1] != 'x' {
		return nil, fmt.Errorf("invalid hex escape sequence")
	}
	out := make([]byte, 0, len(s)/4)
	for i := 0; i < len(s); i += 4 {
		if i+4 > len(s) || s[i] != '\\' || s[i+1] != 'x' {
			return nil, fmt.Errorf("malformed hex escape sequence at position %d", i)
		}
		hex := s[i+2 : i+4]
		value, err := strconv.ParseInt(hex, 16, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid hex value %s: %v", hex, err)
		}
		out = append(out, byte(value))
	}
	return out, nil
}
