package applescript

import "fmt"

// DocumentWriteTextLimitation returns multi-line English documentation of the
// iTerm write-text delivery limit for logs, help text, and skills.
func DocumentWriteTextLimitation() string {
	return fmt.Sprintf(`AppleScript / iTerm write-text delivery limit

When a shell command is delivered to a new iTerm session via AppleScript:

  write text "<command>"

the full <command> string (after EscapeString) is typed into the terminal.
This is reliable only for short commands.

Measured (2026-08-06, ForceNew + printf of shell-quoted bodies):

  - follow/write-text length ≤ ~%d bytes: reliable PASS (SafeMax=%d)
  - ~%d–1050: gray / flaky band
  - ≳ ~1050–1100: often EMPTY (no run) or MISMATCH / UTF-8 corruption
  - SoftMax constant: %d

Chinese / multi-line / <<'EOF' / __seq_ paths are fine when short.
Large payloads (multi-KB open inject) fail when embedded in write text itself.

Workaround (proven control): keep write text SHORT (e.g. bash /path/to/script.sh
or --prompt-file=…) and put the large body on disk. Same UTF-8 multi-KB body
then arrives intact.

See also: shell/applescript.CheckWriteText, tests/scripts/measure-write-text-limit.
`, WriteTextSafeMaxBytes, WriteTextSafeMaxBytes, WriteTextSafeMaxBytes, WriteTextSoftMaxBytes)
}
