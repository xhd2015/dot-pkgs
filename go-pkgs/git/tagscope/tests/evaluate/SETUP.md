# Scenario

**Feature**: tagscope Phase 2 evaluates owned-file changes and plans next tags

```
tag names -> CollectFromNames -> Evaluate(OwnedTreePair, gates) -> ChangePlan
owned trees -> DiffOwnedTrees | BuildScopeTree | OwnedPathsForScope
tag name -> IncrementTag -> bumped release tag
```

## Preconditions

- Package `github.com/xhd2015/dot-pkgs/go-pkgs/git/tagscope` is importable.
- Phase 2 API (`Evaluate`, `DiffOwnedTrees`, `IncrementTag`, `BuildScopeTree`,
  `OwnedPathsForScope`) is not implemented yet — tests expect RED.
- Most leaves inject `OwnedTreePair` maps (no git subprocess).
- `ReleaseCommits` maps scope keys to the commit hash at each scope's
  `LatestRelease` tag (required for `same-commit` gate in injected tests).

## Context

- Blob identity strings use `"100644 <oid>"` form (mode type oid).
- `Evaluate` decisions follow `Collected.Scopes` order.
- Helpers below are shared across all ops in this tree.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/tagscope"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	return nil
}

func scopeKey(scope tagscope.TagScope) tagscope.TagScopeKey {
	return tagscope.TagScopeKey(scope.PathPrefix)
}

func lineageFor(t *testing.T, collected tagscope.CollectedTags, prefix string) tagscope.ScopeLineage {
	t.Helper()
	key := tagscope.TagScopeKey(prefix)
	lineage, ok := collected.ByScope[key]
	if !ok {
		t.Fatalf("scope %q missing from ByScope", prefix)
	}
	return lineage
}

func scopeForPrefix(t *testing.T, collected tagscope.CollectedTags, prefix string) tagscope.TagScope {
	t.Helper()
	for _, scope := range collected.Scopes {
		if scope.PathPrefix == prefix {
			return scope
		}
	}
	t.Fatalf("scope prefix %q not found in Collected.Scopes", prefix)
	return tagscope.TagScope{}
}

func decisionFor(t *testing.T, plan tagscope.ChangePlan, prefix string) tagscope.ScopeDecision {
	t.Helper()
	for _, d := range plan.Decisions {
		if d.Scope.PathPrefix == prefix {
			return d
		}
	}
	t.Fatalf("decision for scope %q not found in plan", prefix)
	return tagscope.ScopeDecision{}
}

func collectedFromReq(t *testing.T, req *Request) tagscope.CollectedTags {
	t.Helper()
	return tagscope.CollectFromNames(req.Names)
}

func ownedPair(release, head tagscope.OwnedTree) tagscope.OwnedTreePair {
	return tagscope.OwnedTreePair{AtRelease: release, AtHead: head}
}
```