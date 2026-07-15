package tagscope

// TagScope identifies a version-tag namespace by optional path prefix and version prefix.
type TagScope struct {
	PathPrefix    string // "" or "sub/" (trailing slash when non-empty)
	VersionPrefix string // "v" or "sub/v"
}

// TagScopeKey is the canonical map key for a scope (PathPrefix).
type TagScopeKey string

// ParsedTag is a recognized scoped semver git tag.
type ParsedTag struct {
	FullName         string
	Scope            TagScope
	Version          string // "0.0.2" without leading v
	Prerelease       string // "" or suffix after hyphen
	IsNumericRelease bool   // true when no prerelease suffix
}

// ScopeLineage holds per-scope tag history and derived head pointers.
type ScopeLineage struct {
	Scope             TagScope
	Tags              []ParsedTag // newest-first
	Newest            *ParsedTag
	LatestRelease     *ParsedTag // newest numeric release only; nil if none
	HasPrereleaseHead bool
}

// CollectedTags is the full tag inventory from CollectFromNames or Collect.
type CollectedTags struct {
	All      []ParsedTag
	Scopes   []TagScope
	ByScope  map[TagScopeKey]ScopeLineage
	Unparsed []string
}

// OwnedTree maps repo paths to blob identity strings (e.g. "100644 abc123...").
type OwnedTree map[string]string

// OwnedTreePair holds owned-file snapshots at a scope's latest release and HEAD.
type OwnedTreePair struct {
	AtRelease OwnedTree
	AtHead    OwnedTree
}

// ChangeCheckInput supplies collected tags, commit refs, and injected owned trees.
type ChangeCheckInput struct {
	Collected      CollectedTags
	HeadCommit     string
	ReleaseCommits map[TagScopeKey]string // commit hash at each scope's LatestRelease
	OwnedTrees     map[TagScopeKey]OwnedTreePair
}

// ScopeDecision is the per-scope outcome from Evaluate.
type ScopeDecision struct {
	Scope         TagScope
	LatestRelease string // full tag name, empty if none
	NextTag       string // empty if no action
	SkipReason    string // "" | "no-baseline" | "prerelease-head" | "same-commit" | "no-changes"
}

// ChangePlan aggregates per-scope decisions for one HEAD commit.
type ChangePlan struct {
	Head      string
	Decisions []ScopeDecision // one per scope in Collected.Scopes order
}

// ScopeTree describes direct parent/child relationships between scopes.
type ScopeTree struct {
	Children map[TagScopeKey][]TagScopeKey
}

func scopeKey(scope TagScope) TagScopeKey {
	return TagScopeKey(scope.PathPrefix)
}