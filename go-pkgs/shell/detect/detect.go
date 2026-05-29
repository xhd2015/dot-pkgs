package detect

import (
	"os"
	"path/filepath"
	"strings"
)

// Shell returns the user's current shell name — "bash", "zsh", or "fish"
// — by inspecting $SHELL. Returns "" if the shell is unknown or unset.
func Shell() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		return ""
	}
	base := strings.TrimSuffix(filepath.Base(shell), ".exe")
	switch base {
	case "bash":
		return "bash"
	case "zsh":
		return "zsh"
	case "fish":
		return "fish"
	default:
		return ""
	}
}
