// Package tidy runs go mod tidy with optional local-VCS overlays.
package tidy

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/mod/seed"
)

// SeededRequest runs go mod tidy in Dir, optionally rewriting well-known VCS
// roots to local git checkouts via seed.OverlayEnv.
type SeededRequest struct {
	Dir      string         // consumer module directory
	Locals   []seed.Mapping // deps to prefer from local trees; empty → plain tidy
	GoCmd    string         // optional; default "go"
	Environ  []string       // optional base env; nil → os.Environ()
	Stdout   io.Writer
	Stderr   io.Writer
}

// Seeded runs go mod tidy in req.Dir. When Locals is non-empty, the child env
// uses GOPROXY=direct, GONOSUMDB/GOPRIVATE for those modules, and GIT_CONFIG
// url.insteadOf so fetches for well-known hosts hit the local repos (cache
// fill happens inside tidy if needed). GOSUMDB is not forced off.
func Seeded(ctx context.Context, req SeededRequest) error {
	if ctx == nil {
		ctx = context.Background()
	}
	dir := strings.TrimSpace(req.Dir)
	if dir == "" {
		return fmt.Errorf("tidy: Dir required")
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("tidy: resolve Dir: %w", err)
	}
	st, err := os.Stat(absDir)
	if err != nil {
		return fmt.Errorf("tidy: Dir: %w", err)
	}
	if !st.IsDir() {
		return fmt.Errorf("tidy: Dir is not a directory: %s", absDir)
	}

	goCmd := strings.TrimSpace(req.GoCmd)
	if goCmd == "" {
		goCmd = "go"
	}

	base := req.Environ
	if base == nil {
		base = os.Environ()
	}
	var env []string
	if len(req.Locals) == 0 {
		env = append([]string(nil), base...)
	} else {
		env, err = seed.OverlayEnv(base, req.Locals)
		if err != nil {
			return err
		}
	}

	cmd := exec.CommandContext(ctx, goCmd, "mod", "tidy")
	cmd.Dir = absDir
	cmd.Env = env
	cmd.Stdout = req.Stdout
	cmd.Stderr = req.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go mod tidy: %w", err)
	}
	return nil
}
