package ptywrap

import (
	"strconv"
	"strings"
)

const (
	outputProbeRestCap = 64
	inputReportRestCap = 96
)

// stripOutputProbes removes complete OSC 10/11 color queries and CSI 6n
// cursor queries from PTY output that attach clients would otherwise see.
// Call only for probe kinds the session auto-replies (stripOSC / stripDSR).
// rest is an incomplete trailing prefix to prepend on the next chunk.
func stripOutputProbes(partial, data []byte, stripOSC, stripDSR bool) (display, rest []byte) {
	buf := mergePartial(partial, data)
	if len(buf) == 0 {
		return nil, nil
	}
	if !stripOSC && !stripDSR {
		return buf, nil
	}
	var out []byte
	i := 0
	for i < len(buf) {
		if buf[i] != 0x1b {
			out = append(out, buf[i])
			i++
			continue
		}
		rem := buf[i:]
		if stripOSC {
			end, _, ok, inc := parseOneOSCColorQuery(rem)
			if inc {
				return capRest(out, rem, outputProbeRestCap)
			}
			if ok {
				i += end
				continue
			}
		}
		if stripDSR {
			end, _, ok, inc := parseOneCSI6nQuery(rem, 1, 1)
			if inc {
				return capRest(out, rem, outputProbeRestCap)
			}
			if ok {
				i += end
				continue
			}
		}
		if couldBeHeldOutputProbePrefix(rem, stripOSC, stripDSR) {
			return capRest(out, rem, outputProbeRestCap)
		}
		out = append(out, 0x1b)
		i++
	}
	return out, nil
}

func couldBeHeldOutputProbePrefix(b []byte, stripOSC, stripDSR bool) bool {
	if len(b) == 0 || b[0] != 0x1b {
		return false
	}
	s := string(b)
	if stripOSC {
		for _, p := range []string{
			"\x1b]10;?\x07",
			"\x1b]11;?\x07",
			"\x1b]10;?\x1b\\",
			"\x1b]11;?\x1b\\",
		} {
			if strings.HasPrefix(p, s) && s != p {
				return true
			}
		}
	}
	if stripDSR {
		p := "\x1b[6n"
		if strings.HasPrefix(p, s) && s != p {
			return true
		}
	}
	return false
}

// filterInputReports drops duplicate OSC 10/11 reports and CPR sequences
// that a local terminal emits after seeing forwarded probes. Real keystrokes
// (including arrows ESC[A) are kept. rest is an incomplete trailing prefix.
func filterInputReports(partial, data []byte, dropOSC, dropCPR bool) (keep, rest []byte) {
	buf := mergePartial(partial, data)
	if len(buf) == 0 {
		return nil, nil
	}
	if !dropOSC && !dropCPR {
		return buf, nil
	}
	var out []byte
	i := 0
	for i < len(buf) {
		if buf[i] != 0x1b {
			out = append(out, buf[i])
			i++
			continue
		}
		rem := buf[i:]
		if dropOSC {
			end, ok, inc := parseOneOSC10or11(rem)
			if inc {
				return capRest(out, rem, inputReportRestCap)
			}
			if ok {
				i += end
				continue
			}
		}
		if dropCPR {
			end, ok, inc := parseOneCPR(rem)
			if inc {
				return capRest(out, rem, inputReportRestCap)
			}
			if ok {
				i += end
				continue
			}
		}
		out = append(out, 0x1b)
		i++
	}
	return out, nil
}

// parseOneOSC10or11 recognizes a complete OSC 10 or 11 sequence (query or
// report). incomplete means the bytes from this ESC could still become one.
func parseOneOSC10or11(buf []byte) (end int, ok, incomplete bool) {
	if len(buf) == 0 {
		return 0, false, true
	}
	if buf[0] != 0x1b {
		return 1, false, false
	}
	if len(buf) < 2 {
		return 0, false, true
	}
	if buf[1] != ']' {
		return 1, false, false
	}
	j := 2
	if j >= len(buf) {
		return 0, false, true
	}
	psStart := j
	for j < len(buf) && buf[j] >= '0' && buf[j] <= '9' {
		j++
	}
	if j == psStart {
		return 1, false, false
	}
	if j >= len(buf) {
		return 0, false, true
	}
	ps, err := strconv.Atoi(string(buf[psStart:j]))
	if err != nil || (ps != 10 && ps != 11) {
		return 1, false, false
	}
	if buf[j] != ';' {
		return 1, false, false
	}
	j++
	for j < len(buf) {
		switch buf[j] {
		case 0x07:
			return j + 1, true, false
		case 0x1b:
			if j+1 >= len(buf) {
				return 0, false, true
			}
			if buf[j+1] == '\\' {
				return j + 2, true, false
			}
			return 1, false, false
		default:
			j++
		}
	}
	return 0, false, true
}

// parseOneCPR recognizes CSI cursor-position reports: ESC [ digits ; digits R.
// At least one digit is required so ESC[R is not treated as CPR.
func parseOneCPR(buf []byte) (end int, ok, incomplete bool) {
	if len(buf) == 0 {
		return 0, false, true
	}
	if buf[0] != 0x1b {
		return 1, false, false
	}
	if len(buf) < 2 {
		return 0, false, true
	}
	if buf[1] != '[' {
		return 1, false, false
	}
	if len(buf) < 3 {
		return 0, false, true
	}
	if buf[2] == '?' {
		return 1, false, false
	}
	j := 2
	sawDigit := false
	for j < len(buf) {
		b := buf[j]
		switch {
		case b >= '0' && b <= '9':
			sawDigit = true
			j++
		case b == ';':
			j++
		case b == 'R':
			if !sawDigit {
				return 1, false, false
			}
			return j + 1, true, false
		default:
			return 1, false, false
		}
	}
	return 0, false, true
}

func mergePartial(partial, data []byte) []byte {
	switch {
	case len(partial) == 0:
		return data
	case len(data) == 0:
		return partial
	default:
		return append(append([]byte{}, partial...), data...)
	}
}

func capRest(out, rem []byte, capLen int) (display, rest []byte) {
	if len(rem) > capLen {
		out = append(out, rem...)
		return out, nil
	}
	return out, rem
}

// TestExported_StripOutputProbes is the sealed test hook for stripOutputProbes.
func TestExported_StripOutputProbes(partial, data []byte, stripOSC, stripDSR bool) (display, rest []byte) {
	return stripOutputProbes(partial, data, stripOSC, stripDSR)
}

// TestExported_FilterInputReports is the sealed test hook for filterInputReports.
func TestExported_FilterInputReports(partial, data []byte, dropOSC, dropCPR bool) (keep, rest []byte) {
	return filterInputReports(partial, data, dropOSC, dropCPR)
}
