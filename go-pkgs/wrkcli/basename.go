package wrkcli

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/wrkcli/storage"
	"golang.org/x/term"
)

func isBasename(dir string) bool {
	if dir == "" || filepath.IsAbs(dir) {
		return false
	}
	if strings.ContainsRune(dir, filepath.Separator) {
		return false
	}
	// Also reject forward slashes on all platforms (wrk accepts Unix-style args).
	if strings.Contains(dir, "/") {
		return false
	}
	return true
}

func isCreateMode(projects, addFlagSet, removeFlagSet, setTaskFlagSet, whereFlagSet, repos, status bool, depPath string, allDeps, list, done, mergeBack bool) bool {
	if projects || addFlagSet || removeFlagSet || setTaskFlagSet || whereFlagSet || repos || status {
		return false
	}
	if depPath != "" || allDeps || list || done || mergeBack {
		return false
	}
	return true
}

// resolveDirArg resolves dir to an absolute path: Abs → stat → optional basename
// fallback via resolveBasenameFromProjects when allowBasenameFallback is true.
func resolveDirArg(dir string, allowBasenameFallback bool, wrkHome string) (string, error) {
	absCandidate, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("resolve dir: %w", err)
	}
	if _, err := os.Stat(absCandidate); err == nil {
		return absCandidate, nil
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("stat dir: %w", err)
	}

	if allowBasenameFallback && isBasename(dir) {
		resolved, fallbackErr := resolveBasenameFromProjects(wrkHome, dir)
		if fallbackErr != nil {
			return "", fallbackErr
		}
		if resolved != "" {
			if _, err := os.Stat(resolved); err != nil {
				if os.IsNotExist(err) {
					return "", fmt.Errorf("wrk: %s does not exist", resolved)
				}
				return "", fmt.Errorf("stat dir: %w", err)
			}
			return resolved, nil
		}
	}

	return "", fmt.Errorf("wrk: %s does not exist", absCandidate)
}

// resolveSourceWorkDir resolves the effective workDir from an optional sourceDir
// positional. When sourceDir is absent, returns the process cwd.
func resolveSourceWorkDir(origWd, sourceDir string, allowBasenameFallback bool, wrkHome string) (string, error) {
	if sourceDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("get cwd: %w", err)
		}
		return wd, nil
	}

	_ = origWd // resolveDirArg already resolves relative paths against process cwd.
	return resolveDirArg(sourceDir, allowBasenameFallback, wrkHome)
}

func resolveBasenameFromProjects(wrkHome, basename string) (string, error) {
	matches, err := storage.FindProjectsByBasename(wrkHome, basename)
	if err != nil {
		return "", err
	}
	switch len(matches) {
	case 0:
		return "", nil
	case 1:
		return matches[0], nil
	default:
		return pickAmbiguousBasename(basename, matches)
	}
}

func pickAmbiguousBasename(basename string, matches []string) (string, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "Multiple projects match %q:\n", basename)
	for i, p := range matches {
		fmt.Fprintf(&b, "  %d) %s\n", i+1, p)
	}
	listing := b.String()

	bypass := os.Getenv("WRK_BASENAME_CONFIRM") == "1"
	if bypass || term.IsTerminal(int(os.Stdin.Fd())) {
		n := len(matches)
		fmt.Fprint(os.Stderr, listing)
		fmt.Fprintf(os.Stderr, "Select [1-%d]: ", n)
		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("wrk: read selection: %w", err)
		}
		choice, err := strconv.Atoi(strings.TrimSpace(line))
		if err != nil || choice < 1 || choice > n {
			return "", fmt.Errorf("wrk: invalid selection")
		}
		return matches[choice-1], nil
	}

	return "", errors.New(strings.TrimRight(listing, "\n"))
}