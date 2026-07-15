package tagscope

import "sort"

// CollectFromNames builds inventory from explicit tag names.
// Scopes are sorted lexicographically by VersionPrefix; tags within each scope are newest-first.
func CollectFromNames(names []string) CollectedTags {
	if len(names) == 0 {
		return CollectedTags{
			ByScope: make(map[TagScopeKey]ScopeLineage),
		}
	}

	byScopeNames := make(map[TagScopeKey][]string)
	scopeByKey := make(map[TagScopeKey]TagScope)
	var unparsed []string
	var all []ParsedTag

	for _, name := range names {
		parsed, ok := ParseTagName(name)
		if !ok {
			unparsed = append(unparsed, name)
			continue
		}
		key := scopeKey(parsed.Scope)
		byScopeNames[key] = append(byScopeNames[key], name)
		scopeByKey[key] = parsed.Scope
	}

	if len(byScopeNames) == 0 {
		return CollectedTags{
			Unparsed: unparsed,
			ByScope:  make(map[TagScopeKey]ScopeLineage),
		}
	}

	keys := make([]TagScopeKey, 0, len(byScopeNames))
	for key := range byScopeNames {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		return scopeByKey[keys[i]].VersionPrefix < scopeByKey[keys[j]].VersionPrefix
	})

	scopes := make([]TagScope, len(keys))
	byScope := make(map[TagScopeKey]ScopeLineage, len(keys))

	for i, key := range keys {
		scope := scopeByKey[key]
		scopes[i] = scope

		sortedNames := SortTagsNewestFirst(byScopeNames[key])
		tags := make([]ParsedTag, len(sortedNames))
		for j, name := range sortedNames {
			parsed, _ := ParseTagName(name)
			tags[j] = parsed
		}

		lineage := ScopeLineage{
			Scope: scope,
			Tags:  tags,
		}
		if len(tags) > 0 {
			lineage.Newest = &tags[0]
			lineage.HasPrereleaseHead = tags[0].Prerelease != ""
			for j := range tags {
				if tags[j].IsNumericRelease {
					lineage.LatestRelease = &tags[j]
					break
				}
			}
		}

		byScope[key] = lineage
		all = append(all, tags...)
	}

	return CollectedTags{
		All:      all,
		Scopes:   scopes,
		ByScope:  byScope,
		Unparsed: unparsed,
	}
}