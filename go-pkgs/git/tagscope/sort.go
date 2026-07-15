package tagscope

import (
	"sort"
	"strconv"
	"strings"
)

// SortTagsNewestFirst orders tag names newest-first using git version:refname semantics.
// Numeric releases sort above prereleases at the same version; version components compare numerically.
func SortTagsNewestFirst(tags []string) []string {
	if len(tags) <= 1 {
		out := make([]string, len(tags))
		copy(out, tags)
		return out
	}

	out := append([]string(nil), tags...)
	sort.Slice(out, func(i, j int) bool {
		return compareTagsNewestFirst(out[i], out[j]) < 0
	})
	return out
}

func compareTagsNewestFirst(a, b string) int {
	pa, okA := ParseTagName(a)
	pb, okB := ParseTagName(b)
	if !okA || !okB {
		return strings.Compare(b, a)
	}

	if cmp := compareVersionDesc(pa.Version, pb.Version); cmp != 0 {
		return cmp
	}

	if pa.IsNumericRelease && !pb.IsNumericRelease {
		return -1
	}
	if !pa.IsNumericRelease && pb.IsNumericRelease {
		return 1
	}

	if !pa.IsNumericRelease && !pb.IsNumericRelease {
		if pa.Prerelease > pb.Prerelease {
			return -1
		}
		if pa.Prerelease < pb.Prerelease {
			return 1
		}
	}

	if a < b {
		return 1
	}
	if a > b {
		return -1
	}
	return 0
}

func compareVersionDesc(a, b string) int {
	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")
	for i := 0; i < 3; i++ {
		ai, _ := strconv.Atoi(aParts[i])
		bi, _ := strconv.Atoi(bParts[i])
		if ai > bi {
			return -1
		}
		if ai < bi {
			return 1
		}
	}
	return 0
}