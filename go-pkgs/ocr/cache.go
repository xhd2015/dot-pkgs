package ocr

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	cacheSubdir   = "dot-pkgs/ocr"
	helperPrefix  = "vision-ocr-"
	buildLockWait = 30 * time.Second
	buildLockPoll = 50 * time.Millisecond
)

func resolveCacheDir(opts Options) (string, error) {
	if opts.CacheDir != "" {
		return opts.CacheDir, nil
	}
	userCache := opts.UserCacheDir
	if userCache == nil {
		userCache = os.UserCacheDir
	}
	base, err := userCache()
	if err != nil {
		return "", fmt.Errorf("cache dir: %w", err)
	}
	return filepath.Join(base, cacheSubdir), nil
}

func helperName(source []byte) string {
	sum := sha256.Sum256(source)
	return helperPrefix + hex.EncodeToString(sum[:8])
}

func helperPath(cacheDir string, source []byte) string {
	return filepath.Join(cacheDir, helperName(source))
}

// withDirLock creates exclusive lockDir, runs fn, then removes lockDir.
// If lockDir already exists, waits until it disappears or timeout.
func withDirLock(lockDir string, fn func() error) error {
	deadline := time.Now().Add(buildLockWait)
	for {
		err := os.Mkdir(lockDir, 0700)
		if err == nil {
			defer os.RemoveAll(lockDir)
			return fn()
		}
		if !os.IsExist(err) {
			return fmt.Errorf("build lock: %w", err)
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for helper build lock")
		}
		time.Sleep(buildLockPoll)
	}
}
