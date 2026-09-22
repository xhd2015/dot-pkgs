//go:build darwin

package ocr

import (
	_ "embed"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//go:embed vision/RecognizeText.swift
var visionSwiftSource []byte

// ensureVisionHelper returns a path to a compiled Vision OCR helper binary.
func ensureVisionHelper(ctx context.Context, opts Options) (string, error) {
	cacheDir, err := resolveCacheDir(opts)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("mkdir cache: %w", err)
	}

	out := helperPath(cacheDir, visionSwiftSource)
	if st, err := os.Stat(out); err == nil && !st.IsDir() && st.Mode()&0111 != 0 {
		return out, nil
	}

	lockDir := out + ".lock"
	var built string
	err = withDirLock(lockDir, func() error {
		if st, err := os.Stat(out); err == nil && !st.IsDir() && st.Mode()&0111 != 0 {
			built = out
			return nil
		}
		p, err := compileVisionHelper(ctx, opts, cacheDir, out)
		if err != nil {
			return err
		}
		built = p
		return nil
	})
	if err != nil {
		if st, statErr := os.Stat(out); statErr == nil && !st.IsDir() && st.Mode()&0111 != 0 {
			return out, nil
		}
		return "", err
	}
	return built, nil
}

func compileVisionHelper(ctx context.Context, opts Options, cacheDir, out string) (string, error) {
	swiftc := opts.Swiftc
	if swiftc == "" {
		var err error
		swiftc, err = lookPath(opts, "swiftc")
		if err != nil {
			return "", fmt.Errorf("swiftc not found (install Xcode Command Line Tools): %w", err)
		}
	}

	srcPath := filepath.Join(cacheDir, helperName(visionSwiftSource)+".swift")
	if err := os.WriteFile(srcPath, visionSwiftSource, 0600); err != nil {
		return "", fmt.Errorf("write swift source: %w", err)
	}
	defer os.Remove(srcPath)

	tmpOut := out + ".tmp"
	_ = os.Remove(tmpOut)
	_, stderr, err := runCmd(ctx, opts, swiftc, "-O", "-o", tmpOut, srcPath)
	if err != nil {
		msg := strings.TrimSpace(string(stderr))
		if msg == "" {
			msg = err.Error()
		}
		_ = os.Remove(tmpOut)
		return "", fmt.Errorf("swiftc: %s", msg)
	}
	if err := os.Chmod(tmpOut, 0755); err != nil {
		_ = os.Remove(tmpOut)
		return "", fmt.Errorf("chmod helper: %w", err)
	}
	if err := os.Rename(tmpOut, out); err != nil {
		_ = os.Remove(tmpOut)
		return "", fmt.Errorf("install helper: %w", err)
	}
	return out, nil
}
