package stringvocaboverlap

import (
	"strings"
)

func ExtractHost(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if idx := strings.LastIndex(raw, "@"); idx >= 0 {
		raw = raw[idx+1:]
	}
	if idx := strings.Index(raw, ":"); idx >= 0 {
		raw = raw[:idx]
	}
	return strings.ToLower(raw)
}
