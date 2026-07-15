package tagscope

import "regexp"

var tagNamePattern = regexp.MustCompile(`^(.*/)?v(\d+)\.(\d+)\.(\d+)(?:-([^/]+))?$`)

// ParseTagName parses one tag name. Returns ok=false for non-version tags.
func ParseTagName(name string) (ParsedTag, bool) {
	m := tagNamePattern.FindStringSubmatch(name)
	if m == nil {
		return ParsedTag{}, false
	}

	pathPrefix := m[1]
	version := m[2] + "." + m[3] + "." + m[4]
	prerelease := m[5]

	versionPrefix := "v"
	if pathPrefix != "" {
		versionPrefix = pathPrefix + "v"
	}

	return ParsedTag{
		FullName: name,
		Scope: TagScope{
			PathPrefix:    pathPrefix,
			VersionPrefix: versionPrefix,
		},
		Version:          version,
		Prerelease:       prerelease,
		IsNumericRelease: prerelease == "",
	}, true
}