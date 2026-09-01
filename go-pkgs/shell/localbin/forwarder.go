package localbin

import (
	"strings"
)

// ForwarderBody returns a #!/bin/sh wrapper that execs argv then "$@".
// Empty argv yields only the shebang and exec of "$@" (unusual; callers
// should pass at least one program name).
func ForwarderBody(argv []string) string {
	var b strings.Builder
	b.WriteString("#!/bin/sh\n")
	b.WriteString("exec")
	for _, a := range argv {
		b.WriteByte(' ')
		b.WriteString(shellQuote(a))
	}
	b.WriteString(" \"$@\"\n")
	return b.String()
}

func shellQuote(s string) string {
	safe := s != ""
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '/' || c == '.' || c == '-' || c == '_') {
			safe = false
			break
		}
	}
	if safe {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
