package applescript

import (
	"fmt"
	"strings"
)

// WriteTextExpr returns an AppleScript expression that evaluates to s for use
// with iTerm "write text …". Printable ASCII (0x20–0x7e) runs are double-quoted
// (with EscapeString). Other bytes use (ASCII character N). Parts join with &.
func WriteTextExpr(s string) string {
	if s == "" {
		return `""`
	}
	var parts []string
	var run []byte
	flush := func() {
		if len(run) == 0 {
			return
		}
		parts = append(parts, `"`+EscapeString(string(run))+`"`)
		run = run[:0]
	}
	for i := 0; i < len(s); i++ {
		b := s[i]
		if b >= 0x20 && b <= 0x7e {
			run = append(run, b)
			continue
		}
		flush()
		parts = append(parts, fmt.Sprintf("(ASCII character %d)", int(b)))
	}
	flush()
	if len(parts) == 1 {
		return parts[0]
	}
	return strings.Join(parts, " & ")
}
