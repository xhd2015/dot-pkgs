package mouse

// CSI6n is the Device Status Report query for cursor position.
// Terminal replies with CPR: ESC [ <row> ; <col> R (1-based).
// Append to the same paint buffer as the UI frame (not a separate delayed Write).
const CSI6n = "\x1b[6n"

// CPR is one peeled cursor position report (1-based DEC coordinates).
type CPR struct {
	Row1, Col1 int
}

// ParseCPR finds the first complete CPR in buf.
// Leading/trailing noise is ignored. Multiple complete CPRs → first wins.
func ParseCPR(buf []byte) (row1, col1 int, ok bool) {
	events, _, _ := DemuxCPR(nil, buf)
	if len(events) == 0 {
		return 0, 0, false
	}
	return events[0].Row1, events[0].Col1, true
}

// OriginFromCPR derives 0-based absolute originY when the cursor sits on the
// last line of a viewLines-tall frame:
//
//	originY0 = row1 - viewLines
//
// Live rule: ok only when viewLines > 0, row1 >= 1, and row1 >= viewLines.
// Do not treat row1 < viewLines as origin 0 (that marks a stale probe as top-anchored).
func OriginFromCPR(row1, viewLines int) (originY0 int, ok bool) {
	if viewLines <= 0 || row1 < 1 || row1 < viewLines {
		return 0, false
	}
	return row1 - viewLines, true
}

// OriginFromCPRClamped allows row1 < viewLines by clamping originY0 to 0.
// Prefer OriginFromCPR for live TUI origin; clamped form is for legacy tests only.
func OriginFromCPRClamped(row1, viewLines int) (originY0 int, ok bool) {
	if viewLines <= 0 || row1 < 1 {
		return 0, false
	}
	oy := row1 - viewLines
	if oy < 0 {
		oy = 0
	}
	return oy, true
}

// DemuxCPR peels complete CPR sequences (ESC [ row ; col R) out of a byte stream.
// Mouse SGR (ESC [ < … M/m) and all other bytes are forwarded.
// hold is incomplete trailing data from the previous read; rest is the new hold.
func DemuxCPR(hold, data []byte) (events []CPR, forward, rest []byte) {
	buf := make([]byte, 0, len(hold)+len(data))
	buf = append(buf, hold...)
	buf = append(buf, data...)
	if len(buf) == 0 {
		return nil, nil, nil
	}

	i := 0
	for i < len(buf) {
		if buf[i] != 0x1b {
			j := i + 1
			for j < len(buf) && buf[j] != 0x1b {
				j++
			}
			forward = append(forward, buf[i:j]...)
			i = j
			continue
		}
		if i+1 >= len(buf) {
			rest = append([]byte{}, buf[i:]...)
			break
		}
		if buf[i+1] != '[' {
			forward = append(forward, buf[i])
			i++
			continue
		}
		j := i + 2
		if j >= len(buf) {
			rest = append([]byte{}, buf[i:]...)
			break
		}
		// Mouse SGR: ESC [ < … M/m
		if buf[j] == '<' {
			k := j + 1
			for k < len(buf) && buf[k] != 'M' && buf[k] != 'm' {
				k++
			}
			if k >= len(buf) {
				rest = append([]byte{}, buf[i:]...)
				break
			}
			forward = append(forward, buf[i:k+1]...)
			i = k + 1
			continue
		}
		if buf[j] < '0' || buf[j] > '9' {
			forward = append(forward, buf[i])
			i++
			continue
		}
		row := 0
		for j < len(buf) && buf[j] >= '0' && buf[j] <= '9' {
			row = row*10 + int(buf[j]-'0')
			j++
		}
		if j >= len(buf) {
			rest = append([]byte{}, buf[i:]...)
			break
		}
		if buf[j] != ';' {
			forward = append(forward, buf[i])
			i++
			continue
		}
		j++
		if j >= len(buf) {
			rest = append([]byte{}, buf[i:]...)
			break
		}
		if buf[j] < '0' || buf[j] > '9' {
			forward = append(forward, buf[i])
			i++
			continue
		}
		col := 0
		for j < len(buf) && buf[j] >= '0' && buf[j] <= '9' {
			col = col*10 + int(buf[j]-'0')
			j++
		}
		if j >= len(buf) {
			rest = append([]byte{}, buf[i:]...)
			break
		}
		if buf[j] != 'R' {
			forward = append(forward, buf[i])
			i++
			continue
		}
		events = append(events, CPR{Row1: row, Col1: col})
		i = j + 1
	}
	return events, forward, rest
}
