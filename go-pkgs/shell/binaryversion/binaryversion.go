// Package binaryversion selects the newest versioned executable from caller-supplied candidates.
package binaryversion

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var semverRe = regexp.MustCompile(`\d+\.\d+\.\d+`)

// Probe reports a normalized semantic version for binary.
// Implementations may use any CLI-specific command or parsing convention.
type Probe func(ctx context.Context, binary string) (string, error)

// Candidate is one executable path supplied by the caller. Via is optional
// provenance such as "gopath" or "login_shell".
type Candidate struct {
	Path string
	Via  string
}

// Found is a candidate with a successfully probed, normalized version.
type Found struct {
	Candidate
	Version string
}

// Command returns a Probe that runs binary with args and parses the first X.Y.Z
// semantic version from stdout. Different CLIs can supply different args, for
// example Command("--version") or Command("version").
func Command(args ...string) Probe {
	args = append([]string(nil), args...)
	return func(ctx context.Context, binary string) (string, error) {
		if strings.TrimSpace(binary) == "" {
			return "", fmt.Errorf("probe version: empty binary")
		}
		out, err := exec.CommandContext(ctx, binary, args...).Output()
		if err != nil {
			return "", err
		}
		return ParseSemver(string(out))
	}
}

// ParseSemver extracts the first X.Y.Z semantic version from output.
func ParseSemver(output string) (string, error) {
	if strings.TrimSpace(output) == "" {
		return "", fmt.Errorf("parse version: empty output")
	}
	v := semverRe.FindString(output)
	if v == "" {
		return "", fmt.Errorf("parse version: no semver in %q", output)
	}
	return v, nil
}

// CompareSemver compares semantic versions embedded in left and right.
// It returns -1, 0, or 1 when left is less than, equal to, or greater than right.
func CompareSemver(left, right string) (int, error) {
	l, err := parse(left)
	if err != nil {
		return 0, err
	}
	r, err := parse(right)
	if err != nil {
		return 0, err
	}
	for i := range l {
		if l[i] < r[i] {
			return -1, nil
		}
		if l[i] > r[i] {
			return 1, nil
		}
	}
	return 0, nil
}

// Find probes candidates in order. Empty paths, duplicate paths, and failed
// probes are skipped. Its result preserves candidate order for stable tie breaks.
func Find(ctx context.Context, candidates []Candidate, probe Probe) []Found {
	if ctx == nil {
		ctx = context.Background()
	}
	if probe == nil {
		return []Found{}
	}
	seen := make(map[string]struct{}, len(candidates))
	out := make([]Found, 0, len(candidates))
	for _, candidate := range candidates {
		candidate.Path = strings.TrimSpace(candidate.Path)
		if candidate.Path == "" {
			continue
		}
		if _, ok := seen[candidate.Path]; ok {
			continue
		}
		seen[candidate.Path] = struct{}{}
		version, err := probe(ctx, candidate.Path)
		if err != nil {
			continue
		}
		version, err = ParseSemver(version)
		if err != nil {
			continue
		}
		out = append(out, Found{Candidate: candidate, Version: version})
	}
	return out
}

// Newest returns the highest-version candidate. Version ties retain input order.
func Newest(ctx context.Context, candidates []Candidate, probe Probe) (Found, error) {
	found := Find(ctx, candidates, probe)
	if len(found) == 0 {
		return Found{}, fmt.Errorf("newest binary: none found")
	}
	best := found[0]
	for _, candidate := range found[1:] {
		cmp, err := CompareSemver(best.Version, candidate.Version)
		if err == nil && cmp < 0 {
			best = candidate
		}
	}
	return best, nil
}

func parse(raw string) ([3]int, error) {
	v, err := ParseSemver(raw)
	if err != nil {
		return [3]int{}, err
	}
	parts := strings.Split(v, ".")
	var result [3]int
	for i, part := range parts {
		n, err := strconv.Atoi(part)
		if err != nil || n < 0 {
			return [3]int{}, fmt.Errorf("parse version: invalid semver %q", v)
		}
		result[i] = n
	}
	return result, nil
}
