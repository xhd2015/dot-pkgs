// Package promptspill shortens long agent/shell prompts for channels with
// tight delivery limits (e.g. iTerm AppleScript write text ~1KB).
//
// When full exceeds MaxRunes (default 800), the inline string is truncated and
// only the remainder is written under dir as <FilePrefix>-<id>.txt.
package promptspill

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DefaultMaxRunes is the default max rune length of the returned short string.
const DefaultMaxRunes = 800

// DefaultFilePrefix is used when Options.FilePrefix is empty.
const DefaultFilePrefix = "prompt-spill"

// MarkerOpen is the fixed marker head before the spill path.
const MarkerOpen = ".... (more content check "

// MarkerClose closes the marker after the path.
const MarkerClose = ")"

// Options configures MaybeSpill.
type Options struct {
	// MaxRunes is the max length of short in runes. Zero means DefaultMaxRunes.
	MaxRunes int
	// FilePrefix is the basename prefix before -<id>.txt.
	// Empty means DefaultFilePrefix. Callers should pass a domain-specific
	// prefix (e.g. "open-inject-full") when they care about on-disk names.
	FilePrefix string
}

// MaybeSpill returns short for delivery (OpenSession/Send / write text Prompt).
//
// Under max: short == full, path == "", no file created.
// Over max: writes remainder-only to dir/<FilePrefix>-<id>.txt and returns
//
//	short = prefix + MarkerOpen + path + MarkerClose
//
// with len([]rune(short)) <= max. Path prefers filepath.Abs.
// Length and cut use []rune so multi-byte UTF-8 is never split mid-codepoint.
func MaybeSpill(full, dir string, opts Options) (short string, path string, err error) {
	maxRunes := opts.MaxRunes
	if maxRunes <= 0 {
		maxRunes = DefaultMaxRunes
	}
	prefixName := strings.TrimSpace(opts.FilePrefix)
	if prefixName == "" {
		prefixName = DefaultFilePrefix
	}

	runes := []rune(full)
	if len(runes) <= maxRunes {
		return full, "", nil
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", "", fmt.Errorf("mkdir dir for prompt spill: %w", err)
	}

	id, err := randomID()
	if err != nil {
		return "", "", fmt.Errorf("prompt spill id: %w", err)
	}
	baseName := prefixName + "-" + id + ".txt"
	path = filepath.Join(dir, baseName)
	if abs, aerr := filepath.Abs(path); aerr == nil {
		path = abs
	}

	marker := MarkerOpen + path + MarkerClose
	markerRunes := []rune(marker)
	prefixBudget := maxRunes - len(markerRunes)
	if prefixBudget < 0 {
		return "", "", fmt.Errorf("spill marker too long (%d runes) for max %d", len(markerRunes), maxRunes)
	}
	if prefixBudget > len(runes) {
		prefixBudget = len(runes)
	}

	prefix := string(runes[:prefixBudget])
	remainder := string(runes[prefixBudget:])
	if err := os.WriteFile(path, []byte(remainder), 0o644); err != nil {
		return "", "", fmt.Errorf("write prompt spill %q: %w", path, err)
	}

	return prefix + marker, path, nil
}

func randomID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}
