package cursor

import "fmt"

const (
	// CR is a carriage return (same line).
	CR = "\r"
	// ClearLine erases the entire current line (CSI EL2).
	ClearLine = "\x1b[2K"
)

// Up moves the cursor up n lines. n < 1 returns "".
func Up(n int) string {
	if n < 1 {
		return ""
	}
	return fmt.Sprintf("\x1b[%dA", n)
}

// Rewrite returns CR + ClearLine + text (in-place line replace).
func Rewrite(text string) string {
	return CR + ClearLine + text
}

// Clear returns CR + ClearLine (erase the in-place line).
func Clear() string {
	return CR + ClearLine
}
