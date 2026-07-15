package tagscope

import (
	"bytes"
	"fmt"
	"os/exec"
	"sort"
	"strings"
)

// LoadOwnedTrees loads per-scope owned trees at LatestRelease and HEAD commits.
func LoadOwnedTrees(repoRoot string, collected CollectedTags, headRef string) (map[TagScopeKey]OwnedTreePair, error) {
	headCommit, err := resolveCommit(repoRoot, headRef)
	if err != nil {
		return nil, err
	}

	headTree, err := loadTreeAtCommit(repoRoot, headCommit)
	if err != nil {
		return nil, err
	}

	scopeTree := BuildScopeTree(collected)
	allPaths := sortedPaths(headTree)
	result := make(map[TagScopeKey]OwnedTreePair, len(collected.Scopes))

	for _, scope := range collected.Scopes {
		key := scopeKey(scope)
		lineage := collected.ByScope[key]
		ownedPaths := OwnedPathsForScope(scope, scopeTree, allPaths)

		atHead := filterOwnedTree(headTree, ownedPaths)
		atRelease := OwnedTree{}

		if lineage.LatestRelease != nil {
			releaseCommit, err := resolveCommit(repoRoot, lineage.LatestRelease.FullName)
			if err != nil {
				return nil, fmt.Errorf("resolve release %q: %w", lineage.LatestRelease.FullName, err)
			}
			releaseTree, err := loadTreeAtCommit(repoRoot, releaseCommit)
			if err != nil {
				return nil, err
			}
			atRelease = filterOwnedTree(releaseTree, ownedPaths)
		}

		result[key] = OwnedTreePair{
			AtRelease: atRelease,
			AtHead:    atHead,
		}
	}

	return result, nil
}

func resolveCommit(repoRoot, ref string) (string, error) {
	cmd := exec.Command("git", "rev-parse", ref)
	cmd.Dir = repoRoot
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse %q in %s: %w", ref, repoRoot, err)
	}
	return strings.TrimSpace(string(out)), nil
}

func loadTreeAtCommit(repoRoot, commit string) (OwnedTree, error) {
	cmd := exec.Command("git", "ls-tree", "-r", "-z", commit)
	cmd.Dir = repoRoot
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git ls-tree -r -z %s in %s: %w", commit, repoRoot, err)
	}

	tree := make(OwnedTree)
	entries := bytes.Split(out, []byte{0})
	for _, entry := range entries {
		if len(entry) == 0 {
			continue
		}
		fields := bytes.Fields(entry)
		if len(fields) < 4 {
			continue
		}
		mode := string(fields[0])
		objectType := string(fields[1])
		oid := string(fields[2])
		path := string(fields[3])
		tree[path] = mode + " " + objectType + " " + oid
	}
	return tree, nil
}

func filterOwnedTree(full OwnedTree, paths []string) OwnedTree {
	filtered := make(OwnedTree, len(paths))
	for _, path := range paths {
		if blob, ok := full[path]; ok {
			filtered[path] = blob
		}
	}
	return filtered
}

func sortedPaths(tree OwnedTree) []string {
	paths := make([]string, 0, len(tree))
	for path := range tree {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}