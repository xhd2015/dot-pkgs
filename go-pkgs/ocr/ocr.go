// Package ocr recognizes text in images (OCR).
// On macOS the default engine is Apple Vision via an embedded Swift helper
// compiled on first use into the user cache.
package ocr

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const EngineVision = "vision"

// Result is recognized text from an image.
type Result struct {
	Text   string `json:"text"`
	Engine string `json:"engine"`
}

// Options configures Recognize / EnsureHelper. Nil injectables use production defaults.
type Options struct {
	// CacheDir holds the compiled Vision helper. Empty → UserCacheDir()/dot-pkgs/ocr.
	CacheDir string

	// UserCacheDir overrides os.UserCacheDir when CacheDir is empty.
	UserCacheDir func() (string, error)

	// LookPath resolves tool names (swiftc). Nil → exec.LookPath.
	LookPath func(file string) (string, error)

	// Swiftc is an absolute path to swiftc; when set, LookPath is skipped.
	Swiftc string

	// RunCmd runs a command and returns combined-style capture of stdout/stderr.
	// Nil → exec.CommandContext.
	RunCmd func(ctx context.Context, name string, args []string) (stdout, stderr []byte, err error)

	// HelperPath, when set, skips ensure/compile and runs this binary directly (tests).
	HelperPath string
}

// RecognizeFile runs OCR on an image at path.
func RecognizeFile(ctx context.Context, path string, opts Options) (*Result, error) {
	if path == "" {
		return nil, fmt.Errorf("empty image path")
	}
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}
	return recognizePath(ctx, path, opts)
}

// RecognizeBytes writes data to a temp file and runs OCR.
// ext is a file extension without dot (e.g. "png"); default "png".
func RecognizeBytes(ctx context.Context, data []byte, ext string, opts Options) (*Result, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty image data")
	}
	if ext == "" {
		ext = "png"
	}
	ext = strings.TrimPrefix(ext, ".")
	dir, err := os.MkdirTemp("", "dot-pkgs-ocr-*")
	if err != nil {
		return nil, fmt.Errorf("temp dir: %w", err)
	}
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "image."+ext)
	if err := os.WriteFile(path, data, 0600); err != nil {
		return nil, fmt.Errorf("write temp: %w", err)
	}
	return recognizePath(ctx, path, opts)
}
