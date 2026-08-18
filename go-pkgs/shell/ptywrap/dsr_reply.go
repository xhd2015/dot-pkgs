package ptywrap

import (
	"bytes"
	"os"
	"strconv"
)

// dsrReplyDisabled reports PTYWRAP_NO_DSR_REPLY=1 (debug / legacy).
func dsrReplyDisabled() bool {
	return os.Getenv("PTYWRAP_NO_DSR_REPLY") == "1"
}

// consumeCSI6nQueries scans buf for complete plain CSI 6n cursor queries
// (ESC [ 6 n). row/col are 1-based CPR coordinates.
//
// replies: concatenated CPR responses to write into the PTY master.
// rest: incomplete trailing bytes to keep for the next read chunk.
func consumeCSI6nQueries(buf []byte, row, col int) (replies, rest []byte) {
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
		end, reply, ok, incomplete := parseOneCSI6nQuery(buf[esc:], row, col)
		if incomplete {
			return out, buf[esc:]
		}
		if ok {
			out = append(out, reply...)
			i = esc + end
			continue
		}
		// Not a plain CSI 6n query: advance past this ESC.
		i = esc + 1
	}
	return out, nil
}

// parseOneCSI6nQuery inspects buf starting at ESC.
// end is bytes consumed from buf when a complete 6n query is recognized.
// incomplete means need more input (rest starts at this ESC).
// ok means a CSI 6n query was found and reply is set.
func parseOneCSI6nQuery(buf []byte, row, col int) (end int, reply []byte, ok bool, incomplete bool) {
	// ESC alone — wait for more.
	if len(buf) < 2 {
		return 0, nil, false, true
	}
	if buf[0] != 0x1b {
		return 1, nil, false, false
	}
	// CSI introducer must be '[' (not OSC ']').
	if buf[1] != '[' {
		return 1, nil, false, false
	}
	// ESC[ alone — incomplete.
	if len(buf) < 3 {
		return 0, nil, false, true
	}
	// DEC private (?6n) and other non-'6' parameters are not plain 6n.
	if buf[2] != '6' {
		return 1, nil, false, false
	}
	// ESC[6 alone — incomplete (need final 'n').
	if len(buf) < 4 {
		return 0, nil, false, true
	}
	if buf[3] != 'n' {
		// e.g. ESC[60m or similar — not cursor DSR.
		return 1, nil, false, false
	}
	return 4, cprReply(row, col), true, false
}

func cprReply(row, col int) []byte {
	var b bytes.Buffer
	b.WriteByte(0x1b)
	b.WriteByte('[')
	b.WriteString(strconv.Itoa(row))
	b.WriteByte(';')
	b.WriteString(strconv.Itoa(col))
	b.WriteByte('R')
	return b.Bytes()
}

// maybeAutoReplyDSR writes CPR replies for CSI 6n queries in data.
// partial carries incomplete sequences across PTY read chunks.
// row/col are 1-based cursor coordinates for the CPR payload.
func maybeAutoReplyDSR(write func([]byte) error, partial, data []byte, row, col int) (nextPartial []byte) {
	if dsrReplyDisabled() {
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
	replies, rest := consumeCSI6nQueries(buf, row, col)
	if len(replies) > 0 && write != nil {
		_ = write(replies)
	}
	if len(rest) > 64 {
		rest = rest[len(rest)-64:]
	}
	return rest
}

// TestExported_ConsumeCSI6nQueries is the sealed test hook for consumeCSI6nQueries.
func TestExported_ConsumeCSI6nQueries(buf []byte, row, col int) (replies, rest []byte) {
	return consumeCSI6nQueries(buf, row, col)
}

// TestExported_MaybeAutoReplyDSR is the sealed test hook for maybeAutoReplyDSR.
func TestExported_MaybeAutoReplyDSR(write func([]byte) error, partial, data []byte, row, col int) (nextPartial []byte) {
	return maybeAutoReplyDSR(write, partial, data, row, col)
}

// TestExported_MaybeAutoReplyDSRDisabled is the PTYWRAP_NO_DSR_REPLY=1 path
// without t.Setenv (doctest leaves always call t.Parallel()).
func TestExported_MaybeAutoReplyDSRDisabled(write func([]byte) error, partial, data []byte, row, col int) (nextPartial []byte) {
	_ = write
	_ = partial
	_ = data
	_ = row
	_ = col
	return nil
}
