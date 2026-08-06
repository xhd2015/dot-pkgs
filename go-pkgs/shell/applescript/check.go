package applescript

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// Empirical limits for a single iTerm-style AppleScript write text payload:
//
//	write text "<command>"
//
// Measured 2026-08-06 via ForceNew + write text of shell-quoted printf bodies
// (see tests/scripts/measure-write-text-limit):
//
//   - follow command length ≲ 950 bytes: reliable PASS
//   - ≳ 1050–1100: EMPTY / MISMATCH / UTF-8 corruption
//   - short write text (e.g. bash script.sh) + large body on disk: multi-KB OK
//
// SoftMax is the failure-band round number; SafeMax is under the last stable PASS.
const (
	// WriteTextSafeMaxBytes: clearly safe for write text "…" delivery.
	WriteTextSafeMaxBytes = 900
	// WriteTextSoftMaxBytes: into the flaky/failure band (≥ ~1050 often fails).
	WriteTextSoftMaxBytes = 1024
)

// Stable reason codes for WriteTextCheck.Reasons.
const (
	ReasonSoftMaxBytes = "soft_max_bytes"
	ReasonNearLimit    = "near_limit"
	ReasonContainsNUL  = "contains_nul"
)

// WriteTextCheck is the result of CheckWriteText.
type WriteTextCheck struct {
	// OK is true only when ByteLen ≤ WriteTextSafeMaxBytes and no hard risks.
	OK bool
	// SoftExceeded is true when ByteLen > WriteTextSoftMaxBytes.
	SoftExceeded bool
	// NearLimit is true when SafeMax < ByteLen ≤ SoftMax.
	NearLimit bool
	ByteLen   int
	RuneLen   int
	Newlines  int
	// Reasons lists stable codes (and optional human suffixes after ':').
	Reasons []string
}

// CheckWriteText evaluates s as the command string that will be embedded in:
//
//	write text "EscapeString(s)"
//
// Pure: does not run osascript or iTerm. Use before sending long FollowUp
// shell lines (e.g. multi-KB open inject).
func CheckWriteText(s string) WriteTextCheck {
	c := WriteTextCheck{
		ByteLen:  len(s),
		RuneLen:  utf8.RuneCountInString(s),
		Newlines: strings.Count(s, "\n"),
	}
	if strings.Contains(s, "\x00") {
		c.Reasons = append(c.Reasons, ReasonContainsNUL+": NUL byte not safe in write text embedding")
	}
	if c.ByteLen > WriteTextSoftMaxBytes {
		c.SoftExceeded = true
		c.Reasons = append(c.Reasons, fmt.Sprintf(
			"%s: %d bytes > SoftMax %d (write text FollowUp often EMPTY/MISMATCH)",
			ReasonSoftMaxBytes, c.ByteLen, WriteTextSoftMaxBytes,
		))
	} else if c.ByteLen > WriteTextSafeMaxBytes {
		c.NearLimit = true
		c.Reasons = append(c.Reasons, fmt.Sprintf(
			"%s: %d bytes in (%d, %d] — possible but flaky",
			ReasonNearLimit, c.ByteLen, WriteTextSafeMaxBytes, WriteTextSoftMaxBytes,
		))
	}
	c.OK = !c.SoftExceeded && !c.NearLimit && len(c.Reasons) == 0
	// Hard risks (nul) also clear OK.
	if strings.Contains(s, "\x00") {
		c.OK = false
	}
	return c
}
