package portmap

import (
	"net/http"
	"strings"
)

// relayResponseCacheControl honors explicit origin cache policies for safe,
// successful reads, without allowing a shared cache to bypass the relay's Wrap
// authentication. It never guesses a TTL from an asset's name or content type.
// Missing/malformed policies, errors, redirects and writes remain non-storable.
func relayResponseCacheControl(response *http.Response) error {
	if response.StatusCode == http.StatusSwitchingProtocols {
		return nil
	}
	// These targeted policies can override Cache-Control at an edge cache.
	// Reusable responses must remain browser-private even when an origin was
	// configured for a public CDN before being exposed through this relay.
	for _, name := range []string{"CDN-Cache-Control", "Cloudflare-CDN-Cache-Control", "Surrogate-Control"} {
		response.Header.Del(name)
	}
	noStore := func() error {
		response.Header.Set("Cache-Control", "no-store")
		return nil
	}
	if response.Request == nil || (response.Request.Method != http.MethodGet && response.Request.Method != http.MethodHead) {
		return noStore()
	}
	if (response.StatusCode < 200 || response.StatusCode >= 300) && response.StatusCode != http.StatusNotModified {
		return noStore()
	}
	directives, ok := splitCacheControl(response.Header.Values("Cache-Control"))
	if !ok || len(directives) == 0 {
		return noStore()
	}
	private := []string{"private"}
	originNoStore := false
	for _, directive := range directives {
		name, hasValue, valid := parseCacheDirective(directive)
		if !valid {
			return noStore()
		}
		switch name {
		case "no-store":
			if hasValue {
				return noStore()
			}
			originNoStore = true
		case "private", "public", "s-maxage":
			// Normalize even qualified private="field, names" to an
			// unqualified private: the entire representation is protected.
			// Public/shared-cache freshness must not reach the edge.
		default:
			private = append(private, directive)
		}
	}
	if originNoStore {
		// Preserve an origin's already restrictive, valid policy verbatim.
		return nil
	}
	response.Header.Set("Cache-Control", strings.Join(private, ", "))
	return nil
}

// Validate directive syntax before retaining any origin value. A malformed
// suffix must not make an edge ignore the entire header, including private.
func parseCacheDirective(directive string) (string, bool, bool) {
	name, value, hasValue := strings.Cut(directive, "=")
	name, value = strings.TrimSpace(name), strings.TrimSpace(value)
	if !cacheToken(name) {
		return "", false, false
	}
	if hasValue && !cacheToken(value) {
		if len(value) < 2 || value[0] != '"' || value[len(value)-1] != '"' {
			return "", false, false
		}
		for i := 1; i < len(value)-1; i++ {
			c := value[i]
			if c == '\\' {
				i++
				if i >= len(value)-1 {
					return "", false, false
				}
				c = value[i]
			} else if c == '"' {
				return "", false, false
			}
			if (c < ' ' && c != '\t') || c == 127 {
				return "", false, false
			}
		}
	}
	return strings.ToLower(name), hasValue, true
}

func cacheToken(value string) bool {
	if value == "" {
		return false
	}
	for i := 0; i < len(value); i++ {
		c := value[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || strings.ContainsRune("!#$%&'*+-.^_`|~", rune(c)) {
			continue
		}
		return false
	}
	return true
}

// Cache-Control permits quoted extension values and qualified field lists.
// Splitting blindly on commas would corrupt them (or miss shared-cache flags).
func splitCacheControl(values []string) ([]string, bool) {
	value := strings.Join(values, ",")
	var directives []string
	start := 0
	quoted, escaped := false, false
	for i := 0; i < len(value); i++ {
		if escaped {
			escaped = false
			continue
		}
		switch value[i] {
		case '\\':
			if quoted {
				escaped = true
			}
		case '"':
			quoted = !quoted
		case ',':
			if !quoted {
				if directive := strings.TrimSpace(value[start:i]); directive != "" {
					directives = append(directives, directive)
				}
				start = i + 1
			}
		}
	}
	if quoted || escaped {
		return nil, false
	}
	if directive := strings.TrimSpace(value[start:]); directive != "" {
		directives = append(directives, directive)
	}
	return directives, true
}
