package ptywrap

import (
	"bytes"
	"regexp"
)

const maxScrollback = 256 * 1024

// trimLookback is how far before provisional cut we search for an ESC that
// may start a sequence straddling the cut (CSI is short; OSC may be longer).
const trimLookback = 256

// trimScrollback keeps the trailing suffix of buf with length at most max,
// without starting mid escape-sequence. When the provisional cut falls inside
// a CSI/OSC/ESC sequence, the cut advances to just after that sequence ends
// (kept size may be under max).
func trimScrollback(buf []byte, max int) []byte {
	if max <= 0 || len(buf) <= max {
		return buf
	}
	cut := len(buf) - max
	cut = advanceCutPastIncompleteEscape(buf, cut)
	if cut < 0 {
		cut = 0
	}
	if cut > len(buf) {
		cut = len(buf)
	}
	return buf[cut:]
}

// advanceCutPastIncompleteEscape moves cut forward when it lands inside an
// escape sequence that began before cut.
func advanceCutPastIncompleteEscape(buf []byte, cut int) int {
	if cut <= 0 || cut >= len(buf) {
		return cut
	}
	start := cut - trimLookback
	if start < 0 {
		start = 0
	}
	// Nearest ESC before cut that still straddles cut wins.
	for esc := cut - 1; esc >= start; esc-- {
		if buf[esc] != 0x1b {
			continue
		}
		end, complete := escapeSequenceEnd(buf, esc)
		if !complete {
			continue
		}
		if esc < cut && cut < end {
			return end
		}
	}
	return cut
}

// escapeSequenceEnd returns the exclusive end index of the escape sequence
// starting at esc (buf[esc] == ESC). complete=false if the sequence is still
// open at EOF.
func escapeSequenceEnd(buf []byte, esc int) (end int, complete bool) {
	if esc < 0 || esc >= len(buf) || buf[esc] != 0x1b {
		return esc, true
	}
	if esc+1 >= len(buf) {
		return len(buf), false
	}
	switch buf[esc+1] {
	case '[': // CSI: ESC [ … final (0x40–0x7E)
		i := esc + 2
		for i < len(buf) {
			b := buf[i]
			i++
			if b >= 0x40 && b <= 0x7e {
				return i, true
			}
		}
		return len(buf), false
	case ']': // OSC: ESC ] … BEL or ST (ESC \)
		i := esc + 2
		for i < len(buf) {
			if buf[i] == 0x07 {
				return i + 1, true
			}
			if buf[i] == 0x1b && i+1 < len(buf) && buf[i+1] == '\\' {
				return i + 2, true
			}
			i++
		}
		return len(buf), false
	case 'P', 'X', '^', '_': // DCS / SOS / PM / APC — terminated by ST
		i := esc + 2
		for i < len(buf) {
			if buf[i] == 0x1b && i+1 < len(buf) && buf[i+1] == '\\' {
				return i + 2, true
			}
			if buf[i] == 0x07 { // some hosts use BEL
				return i + 1, true
			}
			i++
		}
		return len(buf), false
	default:
		// ESC Fe (2-byte) or ESC + next byte best-effort.
		return esc + 2, true
	}
}

// TestExported_TrimScrollback is the sealed test hook for trimScrollback.
func TestExported_TrimScrollback(buf []byte, max int) []byte {
	return trimScrollback(buf, max)
}

var terminalQueryRe = regexp.MustCompile(
	`\x1b\]\d+;[^\x07\x1b]*\?\x07` +
		`|\x1b\]\d+;[^\x07\x1b]*\?\x1b\\` +
		`|\x1b\[[>=]?c` +
		`|\x1b\[[56]n` +
		`|\x1b\[\?\d+\$p`,
)

func stripTerminalQueries(data []byte) []byte {
	return terminalQueryRe.ReplaceAll(data, nil)
}

func isAlternateScreenActive(data []byte) bool {
	enterSeq := []byte("\x1b[?1049h")
	exitSeq := []byte("\x1b[?1049l")
	lastEnter := bytes.LastIndex(data, enterSeq)
	if lastEnter == -1 {
		return false
	}
	lastExit := bytes.LastIndex(data, exitSeq)
	return lastExit < lastEnter
}

func stripAlternateScreenPairs(data []byte) []byte {
	enterSeq := []byte("\x1b[?1049h")
	exitSeq := []byte("\x1b[?1049l")
	for {
		enterIdx := bytes.Index(data, enterSeq)
		if enterIdx == -1 {
			break
		}
		exitIdx := bytes.Index(data[enterIdx:], exitSeq)
		if exitIdx == -1 {
			break
		}
		endIdx := enterIdx + exitIdx + len(exitSeq)
		data = append(data[:enterIdx], data[endIdx:]...)
	}
	return data
}