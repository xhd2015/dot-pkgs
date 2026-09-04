package seed

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
)

// Mapping binds a module path's well-known VCS root to a local git checkout.
// Non-well-known ModulePaths are skipped by OverlayEnv.
type Mapping struct {
	// RepoDir is any path inside the git work tree (module dir or toplevel).
	// OverlayEnv resolves it to the git toplevel for file:// insteadOf.
	RepoDir    string
	ModulePath string // e.g. github.com/xhd2015/lib or …/sub
}

// OverlayEnv returns a copy of base (or os.Environ when base is nil) with
// GOPROXY=direct, GONOSUMDB/GOPRIVATE for overlay module paths, and
// process-scoped GIT_CONFIG url.insteadOf rewrites so well-known VCS roots
// resolve to the local git toplevel via file://. No ~/.gitconfig edits.
//
// GOSUMDB is left alone (not forced off) so golang.org/toolchain can still
// verify. Overlay modules skip sumdb/proxy via GONOSUMDB + GOPRIVATE.
//
// RepoDir may be a nested module directory; it is normalized with
// git show-toplevel before building insteadOf. Invalid RepoDir or a
// well-known mapping outside a git work tree is a hard error. The same VCS
// root mapped to two different git tops is a hard error. ModulePaths that
// are not well-known are skipped. When no well-known mappings remain, base
// is returned unchanged (caller GOPROXY preserved — e.g. file:// proxies for
// example.com fixtures).
func OverlayEnv(base []string, maps []Mapping) ([]string, error) {
	if base == nil {
		base = os.Environ()
	}

	// vcsHTTPS → fileURL (dedupe; conflict = hard error)
	byVCS := make(map[string]string)
	var order []string
	var privatePatterns []string
	seenPattern := make(map[string]struct{})

	for _, m := range maps {
		repo := strings.TrimSpace(m.RepoDir)
		modPath := strings.TrimSpace(m.ModulePath)
		if repo == "" || modPath == "" {
			return nil, fmt.Errorf("seed: mapping requires RepoDir and ModulePath")
		}
		abs, err := filepath.Abs(repo)
		if err != nil {
			return nil, fmt.Errorf("seed: resolve repo dir: %w", err)
		}
		st, err := os.Stat(abs)
		if err != nil {
			return nil, fmt.Errorf("seed: repo dir: %w", err)
		}
		if !st.IsDir() {
			return nil, fmt.Errorf("seed: repo dir is not a directory: %s", abs)
		}
		vcsHTTPS, ok := VCSRootHTTPS(modPath)
		if !ok {
			continue
		}
		if _, ok := seenPattern[modPath]; !ok {
			seenPattern[modPath] = struct{}{}
			privatePatterns = append(privatePatterns, modPath)
		}
		top, err := worktree.ShowToplevel(abs)
		if err != nil {
			return nil, fmt.Errorf("seed: git toplevel for %s: %w", abs, err)
		}
		if resolved, err := filepath.EvalSymlinks(top); err == nil {
			top = resolved
		}
		fileURL, err := fileURLForPath(top)
		if err != nil {
			return nil, err
		}
		if prev, exists := byVCS[vcsHTTPS]; exists {
			if prev != fileURL {
				return nil, fmt.Errorf("seed: conflicting RepoDir for %s", vcsHTTPS)
			}
			continue
		}
		byVCS[vcsHTTPS] = fileURL
		order = append(order, vcsHTTPS)
	}

	// No well-known hosts → leave caller env alone (e.g. file:// GOPROXY for
	// example.com fixtures). Still validated RepoDirs above.
	if len(order) == 0 {
		return append([]string(nil), base...), nil
	}

	var prevNoSum, prevPrivate string
	filtered := make([]string, 0, len(base)+32)
	for _, e := range base {
		switch {
		case strings.HasPrefix(e, "GIT_CONFIG_COUNT="),
			strings.HasPrefix(e, "GIT_CONFIG_KEY_"),
			strings.HasPrefix(e, "GIT_CONFIG_VALUE_"),
			strings.HasPrefix(e, "GIT_CONFIG="),
			strings.HasPrefix(e, "GIT_CONFIG_GLOBAL="),
			strings.HasPrefix(e, "GIT_CONFIG_SYSTEM="),
			strings.HasPrefix(e, "GIT_CONFIG_NOSYSTEM="),
			strings.HasPrefix(e, "GOPROXY="):
			continue
		case strings.HasPrefix(e, "GONOSUMDB="):
			prevNoSum = strings.TrimPrefix(e, "GONOSUMDB=")
			continue
		case strings.HasPrefix(e, "GOPRIVATE="):
			prevPrivate = strings.TrimPrefix(e, "GOPRIVATE=")
			continue
		default:
			filtered = append(filtered, e)
		}
	}

	var pairs []envPair
	for _, vcsHTTPS := range order {
		fileURL := byVCS[vcsHTTPS]
		key := fmt.Sprintf("url.%s.insteadOf", fileURL)
		for _, v := range insteadOfValues(vcsHTTPS) {
			pairs = append(pairs, envPair{key: key, value: v})
		}
	}

	noSum := mergeCSV(prevNoSum, privatePatterns)
	private := mergeCSV(prevPrivate, privatePatterns)
	filtered = append(filtered,
		"GOPROXY=direct",
		"GONOSUMDB="+noSum,
		"GOPRIVATE="+private,
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_CONFIG_GLOBAL=/dev/null",
		fmt.Sprintf("GIT_CONFIG_COUNT=%d", len(pairs)),
	)
	for i, p := range pairs {
		filtered = append(filtered,
			fmt.Sprintf("GIT_CONFIG_KEY_%d=%s", i, p.key),
			fmt.Sprintf("GIT_CONFIG_VALUE_%d=%s", i, p.value),
		)
	}
	return filtered, nil
}

type envPair struct {
	key, value string
}

func mergeCSV(existing string, add []string) string {
	seen := make(map[string]struct{})
	var out []string
	for _, part := range strings.Split(existing, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if _, ok := seen[part]; ok {
			continue
		}
		seen[part] = struct{}{}
		out = append(out, part)
	}
	for _, part := range add {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if _, ok := seen[part]; ok {
			continue
		}
		seen[part] = struct{}{}
		out = append(out, part)
	}
	return strings.Join(out, ",")
}

func insteadOfValues(vcsHTTPS string) []string {
	out := []string{
		vcsHTTPS,
		vcsHTTPS + ".git",
	}
	if hostPath, ok := strings.CutPrefix(vcsHTTPS, "https://"); ok {
		out = append(out,
			"ssh://git@"+hostPath,
			"ssh://git@"+hostPath+".git",
			"git@"+strings.Replace(hostPath, "/", ":", 1),
			"git@"+strings.Replace(hostPath, "/", ":", 1)+".git",
		)
	}
	return out
}

func fileURLForPath(abs string) (string, error) {
	abs, err := filepath.Abs(abs)
	if err != nil {
		return "", err
	}
	abs = filepath.ToSlash(abs)
	if !strings.HasPrefix(abs, "/") {
		// Windows: file:///C:/...
		return "file:///" + abs, nil
	}
	return "file://" + abs, nil
}
