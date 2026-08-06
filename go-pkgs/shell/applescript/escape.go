// Package applescript provides shared helpers for building and validating
// AppleScript string literals and iTerm-style write text payloads.
//
// Escape rules match historical iterm2.EscapeCommandForAppleScript /
// EscapePathForAppleScript: backslash and double-quote only.
package applescript

import "strings"

// EscapeString escapes s for embedding inside a double-quoted AppleScript
// string literal: "...."
//
// Only \ and " are escaped. Newlines and other bytes are left unchanged
// (callers that need multi-line write text should prefer short commands
// and CheckWriteText).
func EscapeString(s string) string {
	return strings.NewReplacer(
		`\`, `\\`,
		`"`, `\"`,
	).Replace(s)
}

// QuoteString returns a double-quoted AppleScript string literal containing s.
func QuoteString(s string) string {
	return `"` + EscapeString(s) + `"`
}
