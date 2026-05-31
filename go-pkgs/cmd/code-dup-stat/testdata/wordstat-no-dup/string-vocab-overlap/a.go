package stringvocaboverlap

import "strings"

func ParseConfig(line string) (string, string, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return "", "", nil
	}
	if idx := strings.Index(line, "="); idx >= 0 {
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		return key, val, nil
	}
	return line, "", nil
}
