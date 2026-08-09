package ptywrap

import "strings"

// MergeProcessEnv builds a process environ from base by removing Unset keys,
// then applying Set assignments (KEY=value). Duplicate keys in set: last wins.
// Unset of a missing key is a no-op. Does not apply spawn TERM policy.
func MergeProcessEnv(base, set, unset []string) []string {
	drop := make(map[string]struct{}, len(unset))
	for _, k := range unset {
		if k == "" {
			continue
		}
		drop[k] = struct{}{}
	}

	out := make([]string, 0, len(base)+len(set))
	for _, e := range base {
		if e == "" {
			continue
		}
		if _, ok := drop[envEntryKey(e)]; ok {
			continue
		}
		out = append(out, e)
	}

	for _, e := range set {
		if e == "" {
			continue
		}
		key := envEntryKey(e)
		out = filterEnvKey(out, key)
		out = append(out, e)
	}
	return out
}

// EnsureSpawnTERM applies the PTY spawn TERM policy: if TERM is missing,
// empty, or exactly "dumb", set TERM=xterm-256color; otherwise leave env as-is.
func EnsureSpawnTERM(env []string) []string {
	val, ok := envLookup(env, "TERM")
	if ok && val != "" && val != "dumb" {
		return env
	}
	out := filterEnvKey(env, "TERM")
	return append(out, "TERM=xterm-256color")
}

func envEntryKey(e string) string {
	if i := strings.IndexByte(e, '='); i >= 0 {
		return e[:i]
	}
	return e
}

// envLookup returns the last value for key in a KEY=value environ slice.
func envLookup(env []string, key string) (string, bool) {
	prefix := key + "="
	found := false
	val := ""
	for _, e := range env {
		if strings.HasPrefix(e, prefix) {
			found = true
			val = e[len(prefix):]
		} else if e == key {
			found = true
			val = ""
		}
	}
	return val, found
}

// filterEnvKey returns a new slice without any entries for key.
func filterEnvKey(env []string, key string) []string {
	if len(env) == 0 {
		return nil
	}
	prefix := key + "="
	out := make([]string, 0, len(env))
	for _, e := range env {
		if strings.HasPrefix(e, prefix) || e == key {
			continue
		}
		out = append(out, e)
	}
	return out
}
