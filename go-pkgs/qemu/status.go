package qemu

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseKV parses qemu script status lines into a key=value map.
// It strips ANSI, takes the payload after the last ": " SMC prefix, then k=v.
func ParseKV(out string) map[string]string {
	m := make(map[string]string)
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(stripANSISimple(line))
		if line == "" {
			continue
		}
		if i := strings.LastIndex(line, ": "); i >= 0 {
			rest := strings.TrimSpace(line[i+2:])
			if rest != "" {
				line = rest
			}
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		m[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	return m
}

func stripANSISimple(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			i += 2
			for i < len(s) && !((s[i] >= 'A' && s[i] <= 'Z') || (s[i] >= 'a' && s[i] <= 'z')) {
				i++
			}
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

// HumanBytes formats a decimal byte-count string for display.
func HumanBytes(s string) string {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil || n <= 0 {
		if s == "" {
			return "missing"
		}
		return s
	}
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%dB", n)
	}
	div, exp := int64(unit), 0
	for n/div >= unit && exp < 3 {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.0f%ciB", float64(n)/float64(div), "KMGT"[exp])
}
