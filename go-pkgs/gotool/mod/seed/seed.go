// Package seed fills GOMODCACHE via go mod download from a local git repo,
// using ephemeral GIT_CONFIG url.insteadOf rewrites (no ~/.gitconfig edits).
package seed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Module is one path@version to download into the module cache.
type Module struct {
	Path    string // e.g. github.com/xhd2015/lib or …/sub
	Version string // go require version, e.g. v0.0.2
}

// Request downloads Modules from RepoDir's local git tags.
type Request struct {
	RepoDir string
	Modules []Module
}

// ModuleResult is the outcome for one module.
type ModuleResult struct {
	Path     string
	Version  string
	Zip      string
	Sum      string
	GoModSum string
	Hash     string
	Ref      string
	Skipped  bool  // true when VCS host is not well-known (no download attempted)
	Err      error // download/parse error (nil when Skipped or success)
}

// Result aggregates per-module outcomes.
type Result struct {
	Modules []ModuleResult
}

type downloadJSON struct {
	Path     string `json:"Path"`
	Version  string `json:"Version"`
	Zip      string `json:"Zip"`
	Sum      string `json:"Sum"`
	GoModSum string `json:"GoModSum"`
	Error    string `json:"Error"`
	Origin   *struct {
		Hash string `json:"Hash"`
		Ref  string `json:"Ref"`
	} `json:"Origin"`
}

// Download runs go mod download -json for each module, rewriting the VCS root
// URL to file://RepoDir via GIT_CONFIG_COUNT (process-scoped only).
//
// Unknown / non-well-known module hosts are skipped (Skipped=true).
// Per-module failures set ModuleResult.Err; the returned error is non-nil only
// when RepoDir is invalid or the go toolchain cannot be started.
func Download(ctx context.Context, req Request) (Result, error) {
	var out Result
	if ctx == nil {
		ctx = context.Background()
	}
	repoDir, err := filepath.Abs(strings.TrimSpace(req.RepoDir))
	if err != nil {
		return out, fmt.Errorf("seed: resolve repo dir: %w", err)
	}
	st, err := os.Stat(repoDir)
	if err != nil {
		return out, fmt.Errorf("seed: repo dir: %w", err)
	}
	if !st.IsDir() {
		return out, fmt.Errorf("seed: repo dir is not a directory: %s", repoDir)
	}

	tmp, err := os.MkdirTemp("", "mod-seed-*")
	if err != nil {
		return out, fmt.Errorf("seed: temp dir: %w", err)
	}
	defer os.RemoveAll(tmp)
	if err := os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module mod-seed\n\ngo 1.22\n"), 0o644); err != nil {
		return out, fmt.Errorf("seed: write go.mod: %w", err)
	}

	for _, m := range req.Modules {
		mr := ModuleResult{Path: m.Path, Version: m.Version}
		if m.Path == "" || m.Version == "" {
			mr.Err = fmt.Errorf("path and version required")
			out.Modules = append(out.Modules, mr)
			continue
		}
		if !IsWellKnown(m.Path) {
			mr.Skipped = true
			out.Modules = append(out.Modules, mr)
			continue
		}
		env, err := OverlayEnv(nil, []Mapping{{RepoDir: repoDir, ModulePath: m.Path}})
		if err != nil {
			mr.Err = err
			out.Modules = append(out.Modules, mr)
			continue
		}
		spec := m.Path + "@" + m.Version
		cmd := exec.CommandContext(ctx, "go", "mod", "download", "-json", spec)
		cmd.Dir = tmp
		cmd.Env = env
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		runErr := cmd.Run()
		var dj downloadJSON
		if uerr := json.Unmarshal(stdout.Bytes(), &dj); uerr != nil {
			if runErr != nil {
				msg := strings.TrimSpace(stderr.String())
				if msg == "" {
					msg = runErr.Error()
				}
				mr.Err = fmt.Errorf("go mod download %s: %s", spec, msg)
			} else {
				mr.Err = fmt.Errorf("go mod download %s: parse json: %w", spec, uerr)
			}
			out.Modules = append(out.Modules, mr)
			continue
		}
		if dj.Error != "" {
			mr.Err = fmt.Errorf("%s", dj.Error)
			out.Modules = append(out.Modules, mr)
			continue
		}
		if runErr != nil {
			msg := strings.TrimSpace(stderr.String())
			if msg == "" {
				msg = runErr.Error()
			}
			mr.Err = fmt.Errorf("go mod download %s: %s", spec, msg)
			out.Modules = append(out.Modules, mr)
			continue
		}
		mr.Zip = dj.Zip
		mr.Sum = dj.Sum
		mr.GoModSum = dj.GoModSum
		if dj.Origin != nil {
			mr.Hash = dj.Origin.Hash
			mr.Ref = dj.Origin.Ref
		}
		out.Modules = append(out.Modules, mr)
	}
	return out, nil
}
