// Package localbin installs scripts under ~/.local/bin and ensures that
// directory is on PATH via a marker block in bash/zsh rc files.
package localbin

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DefaultDir returns $HOME/.local/bin for the given home.
// When home is empty, uses os.UserHomeDir. Errors when home cannot be resolved.
func DefaultDir(home string) (string, error) {
	home = strings.TrimSpace(home)
	if home == "" {
		h, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("HOME is unset; cannot install to ~/.local/bin")
		}
		home = strings.TrimSpace(h)
	}
	if home == "" {
		return "", fmt.Errorf("HOME is unset; cannot install to ~/.local/bin")
	}
	return filepath.Join(home, ".local", "bin"), nil
}

// IsDefaultDest reports whether dir is the default ~/.local/bin for home
// (empty home → UserHomeDir).
func IsDefaultDest(dir, home string) bool {
	want, err := DefaultDir(home)
	if err != nil {
		return false
	}
	return filepath.Clean(dir) == filepath.Clean(want)
}
