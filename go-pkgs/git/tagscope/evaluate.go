package tagscope

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// BuildScopeTree derives direct parent/child scope links from collected scopes.
func BuildScopeTree(collected CollectedTags) ScopeTree {
	children := make(map[TagScopeKey][]TagScopeKey)
	for _, parent := range collected.Scopes {
		parentKey := scopeKey(parent)
		for _, candidate := range collected.Scopes {
			if candidate.PathPrefix == parent.PathPrefix {
				continue
			}
			if !strings.HasPrefix(candidate.PathPrefix, parent.PathPrefix) {
				continue
			}
			if isDirectChild(parent.PathPrefix, candidate.PathPrefix, collected.Scopes) {
				children[parentKey] = append(children[parentKey], scopeKey(candidate))
			}
		}
	}
	return ScopeTree{Children: children}
}

func isDirectChild(parentPrefix, childPrefix string, scopes []TagScope) bool {
	for _, mid := range scopes {
		if mid.PathPrefix == parentPrefix || mid.PathPrefix == childPrefix {
			continue
		}
		if strings.HasPrefix(mid.PathPrefix, parentPrefix) && strings.HasPrefix(childPrefix, mid.PathPrefix) {
			return false
		}
	}
	return true
}

// OwnedPathsForScope returns paths owned by scope, excluding nested child subtrees.
func OwnedPathsForScope(scope TagScope, tree ScopeTree, allPaths []string) []string {
	key := scopeKey(scope)
	childKeys := tree.Children[key]

	var owned []string
	for _, path := range allPaths {
		if pathOwnedByScope(path, scope.PathPrefix, childKeys) {
			owned = append(owned, path)
		}
	}
	return owned
}

func pathOwnedByScope(path, scopePrefix string, childKeys []TagScopeKey) bool {
	if scopePrefix != "" && !strings.HasPrefix(path, scopePrefix) {
		return false
	}
	for _, childKey := range childKeys {
		if strings.HasPrefix(path, string(childKey)) {
			return false
		}
	}
	return true
}

// DiffOwnedTrees reports whether any path was added, removed, or changed blob identity.
func DiffOwnedTrees(old, new OwnedTree) bool {
	for path, oldBlob := range old {
		newBlob, ok := new[path]
		if !ok || newBlob != oldBlob {
			return true
		}
	}
	for path := range new {
		if _, ok := old[path]; !ok {
			return true
		}
	}
	return false
}

// IncrementTag bumps the trailing numeric segment of a release tag name.
func IncrementTag(tag string) (string, error) {
	m := tagNamePattern.FindStringSubmatch(tag)
	if m == nil {
		return "", fmt.Errorf("tag %q is not a numeric release tag", tag)
	}
	if m[5] != "" {
		return "", fmt.Errorf("tag %q has prerelease suffix", tag)
	}

	patch, err := strconv.Atoi(m[4])
	if err != nil {
		return "", fmt.Errorf("tag %q: invalid patch segment: %w", tag, err)
	}

	prefix := m[1] + "v" + m[2] + "." + m[3] + "."
	return fmt.Sprintf("%s%d", prefix, patch+1), nil
}

func scopesInEvaluateOrder(scopes []TagScope) []TagScope {
	ordered := append([]TagScope(nil), scopes...)
	sort.Slice(ordered, func(i, j int) bool {
		pi := ordered[i].PathPrefix
		pj := ordered[j].PathPrefix
		if len(pi) != len(pj) {
			return len(pi) < len(pj)
		}
		return pi < pj
	})
	return ordered
}

// Evaluate decides per scope whether to skip or plan NextTag from injected owned trees.
func Evaluate(in ChangeCheckInput) ChangePlan {
	plan := ChangePlan{
		Head:      in.HeadCommit,
		Decisions: make([]ScopeDecision, 0, len(in.Collected.Scopes)),
	}

	for _, scope := range scopesInEvaluateOrder(in.Collected.Scopes) {
		key := scopeKey(scope)
		lineage := in.Collected.ByScope[key]
		decision := ScopeDecision{Scope: scope}

		if lineage.LatestRelease == nil {
			decision.SkipReason = "no-baseline"
			plan.Decisions = append(plan.Decisions, decision)
			continue
		}

		decision.LatestRelease = lineage.LatestRelease.FullName

		if lineage.HasPrereleaseHead {
			decision.SkipReason = "prerelease-head"
			plan.Decisions = append(plan.Decisions, decision)
			continue
		}

		if in.ReleaseCommits != nil && in.ReleaseCommits[key] == in.HeadCommit {
			decision.SkipReason = "same-commit"
			plan.Decisions = append(plan.Decisions, decision)
			continue
		}

		pair := in.OwnedTrees[key]
		if !DiffOwnedTrees(pair.AtRelease, pair.AtHead) {
			decision.SkipReason = "no-changes"
			plan.Decisions = append(plan.Decisions, decision)
			continue
		}

		nextTag, err := IncrementTag(lineage.LatestRelease.FullName)
		if err != nil {
			decision.SkipReason = "no-baseline"
			plan.Decisions = append(plan.Decisions, decision)
			continue
		}

		decision.NextTag = nextTag
		plan.Decisions = append(plan.Decisions, decision)
	}

	return plan
}