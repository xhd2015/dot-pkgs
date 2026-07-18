package ptywrap

import (
	"bytes"
	"os"
	"strconv"
)

// Dark-terminal defaults for termenv/lipgloss OSC color probes under headless
// PTYs (tty-watch). RGB uses 16-bit xterm channels.
const (
	oscReplyFG = "rgb:c0c0/c0c0/c0c0"
	oscReplyBG = "rgb:1e1e/1e1e/1e1e"
)

// oscReplyDisabled reports PTYWRAP_NO_OSC_REPLY=1 (debug / legacy).
func oscReplyDisabled() bool {
	return os.Getenv("PTYWRAP_NO_OSC_REPLY") == "1"
}

// consumeOSCColorQueries scans buf for complete OSC 10/11 color queries
// (ESC ] 10 ; ? BEL|ST and ESC ] 11 ; ? BEL|ST).
//
// replies: concatenated responses to write into the PTY master.
// rest: incomplete trailing bytes to keep for the next read chunk.
func consumeOSCColorQueries(buf []byte) (replies, rest []byte) {
	if len(buf) == 0 {
		return nil, nil
	}
	var out []byte
	i := 0
	for i < len(buf) {
		rel := bytes.IndexByte(buf[i:], 0x1b)
		if rel < 0 {
			return out, nil
		}
		esc := i + rel
		end, reply, ok, incomplete := parseOneOSCColorQuery(buf[esc:])
		if incomplete {
			return out, buf[esc:]
		}
		if ok {
			out = append(out, reply...)
			i = esc + end
			continue
		}
		// Not a color query (or not OSC): advance past this ESC.
		i = esc + 1
	}
	return out, nil
}

// parseOneOSCColorQuery inspects buf starting at ESC.
// end is bytes consumed from buf when a complete sequence is recognized (query or not).
// incomplete means need more input (rest starts at this ESC).
// ok means a color query was found and reply is set.
func parseOneOSCColorQuery(buf []byte) (end int, reply []byte, ok bool, incomplete bool) {
	if len(buf) < 2 {
		return 0, nil, false, true
	}
	if buf[0] != 0x1b || buf[1] != ']' {
		return 1, nil, false, false
	}
	j := 2
	psStart := j
	for j < len(buf) && buf[j] >= '0' && buf[j] <= '9' {
		j++
	}
	if j == psStart {
		return 1, nil, false, false
	}
	if j >= len(buf) {
		return 0, nil, false, true
	}
	if buf[j] != ';' {
		return 1, nil, false, false
	}
	ps, err := strconv.Atoi(string(buf[psStart:j]))
	if err != nil {
		return 1, nil, false, false
	}
	j++ // ';'
	ptStart := j
	for j < len(buf) {
		switch buf[j] {
		case 0x07: // BEL
			pt := buf[ptStart:j]
			end = j + 1
			if isOSCColorQuery(ps, pt) {
				return end, oscColorReply(ps), true, false
			}
			return end, nil, false, false
		case 0x1b:
			if j+1 >= len(buf) {
				return 0, nil, false, true
			}
			if buf[j+1] == '\\' { // ST
				pt := buf[ptStart:j]
				end = j + 2
				if isOSCColorQuery(ps, pt) {
					return end, oscColorReply(ps), true, false
				}
				return end, nil, false, false
			}
			// Nested ESC: not our OSC terminator; treat as not matching.
			return 1, nil, false, false
		default:
			j++
		}
	}
	return 0, nil, false, true
}

func isOSCColorQuery(ps int, pt []byte) bool {
	if ps != 10 && ps != 11 {
		return false
	}
	return bytes.Equal(bytes.TrimSpace(pt), []byte("?"))
}

func oscColorReply(ps int) []byte {
	rgb := oscReplyBG
	if ps == 10 {
		rgb = oscReplyFG
	}
	var b bytes.Buffer
	b.WriteByte(0x1b)
	b.WriteByte(']')
	b.WriteString(strconv.Itoa(ps))
	b.WriteByte(';')
	b.WriteString(rgb)
	b.WriteByte(0x07)
	return b.Bytes()
}

// maybeAutoReplyOSC writes OSC 10/11 replies for queries in data.
// partial carries incomplete sequences across PTY read chunks.
func maybeAutoReplyOSC(write func([]byte) error, partial, data []byte) (nextPartial []byte) {
	if oscReplyDisabled() {
		return nil
	}
	var buf []byte
	switch {
	case len(partial) == 0:
		buf = data
	case len(data) == 0:
		buf = partial
	default:
		buf = append(append([]byte{}, partial...), data...)
	}
	replies, rest := consumeOSCColorQueries(buf)
	if len(replies) > 0 && write != nil {
		_ = write(replies)
	}
	if len(rest) > 64 {
		rest = rest[len(rest)-64:]
	}
	return rest
}
