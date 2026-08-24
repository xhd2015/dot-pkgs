package easycron

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func parse(s string) (Expr, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Expr{}, fmt.Errorf("easycron: empty expression")
	}
	const prefix = "every-"
	if !strings.HasPrefix(s, prefix) {
		return Expr{}, fmt.Errorf("easycron: must start with %q", prefix)
	}
	rest := s[len(prefix):]
	if rest == "" {
		return Expr{}, fmt.Errorf("easycron: missing interval after %q", prefix)
	}

	durPart, rest, err := splitDurationPrefix(rest)
	if err != nil {
		return Expr{}, err
	}
	interval, err := parseDuration(durPart)
	if err != nil {
		return Expr{}, fmt.Errorf("easycron: invalid interval %q: %w", durPart, err)
	}
	if interval <= 0 {
		return Expr{}, fmt.Errorf("easycron: interval must be positive")
	}

	expr := Expr{Raw: s, Interval: interval}

	if strings.HasPrefix(rest, "-at-") {
		rest = rest[len("-at-"):]
		offPart, next, err := splitDurationPrefix(rest)
		if err != nil {
			return Expr{}, fmt.Errorf("easycron: invalid -at- offset: %w", err)
		}
		rest = next
		offset, err := parseDuration(offPart)
		if err != nil {
			return Expr{}, fmt.Errorf("easycron: invalid -at- offset %q: %w", offPart, err)
		}
		if offset < 0 || offset >= interval {
			return Expr{}, fmt.Errorf("easycron: offset %s must be in [0, %s)", formatDuration(offset), formatDuration(interval))
		}
		expr.Align = &offset
	}

	if strings.HasPrefix(rest, "-until-") {
		rest = rest[len("-until-"):]
		todPart, next, err := splitTODPrefix(rest)
		if err != nil {
			return Expr{}, fmt.Errorf("easycron: invalid -until- time: %w", err)
		}
		rest = next
		clk, err := parseTOD(todPart)
		if err != nil {
			return Expr{}, fmt.Errorf("easycron: invalid -until- time %q: %w", todPart, err)
		}
		expr.Until = &clk
	}

	if strings.HasPrefix(rest, "-not-within-") {
		rest = rest[len("-not-within-"):]
		startPart, afterStart, ok := strings.Cut(rest, "-to-")
		if !ok {
			return Expr{}, fmt.Errorf("easycron: -not-within- requires <tod>-to-<tod>")
		}
		startPart = strings.TrimSpace(startPart)
		if startPart == "" {
			return Expr{}, fmt.Errorf("easycron: -not-within- missing start time")
		}
		endPart, next, err := splitTODPrefix(afterStart)
		if err != nil {
			return Expr{}, fmt.Errorf("easycron: invalid -not-within- end time: %w", err)
		}
		rest = next
		start, err := parseTOD(startPart)
		if err != nil {
			return Expr{}, fmt.Errorf("easycron: invalid -not-within- start %q: %w", startPart, err)
		}
		end, err := parseTOD(endPart)
		if err != nil {
			return Expr{}, fmt.Errorf("easycron: invalid -not-within- end %q: %w", endPart, err)
		}
		expr.Quiet = &QuietWindow{Start: start, End: end}
	}

	if rest != "" {
		return Expr{}, fmt.Errorf("easycron: unexpected trailing %q", rest)
	}
	return expr, nil
}

// splitDurationPrefix reads a leading NhNm / Nh / Nm token and returns the remainder.
func splitDurationPrefix(s string) (token, rest string, err error) {
	if s == "" {
		return "", "", fmt.Errorf("missing duration")
	}
	i := 0
	for i < len(s) {
		r := rune(s[i])
		if unicode.IsDigit(r) || r == 'h' || r == 'm' {
			i++
			continue
		}
		break
	}
	if i == 0 {
		return "", s, fmt.Errorf("missing duration")
	}
	token = s[:i]
	rest = s[i:]
	// Duration tokens end before a modifier boundary; if we consumed into "-at" etc.,
	// the loop already stopped at '-'.
	if _, err := parseDuration(token); err != nil {
		return "", s, err
	}
	return token, rest, nil
}

// splitTODPrefix reads a leading NhNm TOD token (both h and m required).
func splitTODPrefix(s string) (token, rest string, err error) {
	if s == "" {
		return "", "", fmt.Errorf("missing time of day")
	}
	i := 0
	for i < len(s) {
		r := rune(s[i])
		if unicode.IsDigit(r) || r == 'h' || r == 'm' {
			i++
			continue
		}
		break
	}
	if i == 0 {
		return "", s, fmt.Errorf("missing time of day")
	}
	token = s[:i]
	rest = s[i:]
	if _, err := parseTOD(token); err != nil {
		return "", s, err
	}
	return token, rest, nil
}

func parseDuration(s string) (time.Duration, error) {
	if s == "" {
		return 0, fmt.Errorf("empty duration")
	}
	var hours, mins int
	var sawH, sawM bool
	rest := s
	for rest != "" {
		n, next, ok := readInt(rest)
		if !ok {
			return 0, fmt.Errorf("expected digits in %q", s)
		}
		rest = next
		if rest == "" {
			return 0, fmt.Errorf("duration %q missing unit", s)
		}
		switch rest[0] {
		case 'h':
			if sawH {
				return 0, fmt.Errorf("duplicate hour in %q", s)
			}
			if sawM {
				return 0, fmt.Errorf("hour after minute in %q", s)
			}
			hours = n
			sawH = true
			rest = rest[1:]
		case 'm':
			if sawM {
				return 0, fmt.Errorf("duplicate minute in %q", s)
			}
			mins = n
			sawM = true
			rest = rest[1:]
		default:
			return 0, fmt.Errorf("unknown unit in %q", s)
		}
	}
	if !sawH && !sawM {
		return 0, fmt.Errorf("duration %q missing unit", s)
	}
	return time.Duration(hours)*time.Hour + time.Duration(mins)*time.Minute, nil
}

func parseTOD(s string) (Clock, error) {
	// Require both hour and minute parts: 19h00m, 6h30m, 0h0m.
	var hours, mins int
	var sawH, sawM bool
	rest := s
	for rest != "" {
		n, next, ok := readInt(rest)
		if !ok {
			return Clock{}, fmt.Errorf("expected digits")
		}
		rest = next
		if rest == "" {
			return Clock{}, fmt.Errorf("time of day must look like 19h00m")
		}
		switch rest[0] {
		case 'h':
			if sawH || sawM {
				return Clock{}, fmt.Errorf("time of day must look like 19h00m")
			}
			hours = n
			sawH = true
			rest = rest[1:]
		case 'm':
			if !sawH || sawM {
				return Clock{}, fmt.Errorf("time of day must look like 19h00m")
			}
			mins = n
			sawM = true
			rest = rest[1:]
		default:
			return Clock{}, fmt.Errorf("time of day must look like 19h00m")
		}
	}
	if !sawH || !sawM {
		return Clock{}, fmt.Errorf("time of day must look like 19h00m")
	}
	if hours < 0 || hours > 23 {
		return Clock{}, fmt.Errorf("hour %d out of range 0–23", hours)
	}
	if mins < 0 || mins > 59 {
		return Clock{}, fmt.Errorf("minute %d out of range 0–59", mins)
	}
	return Clock{Hour: hours, Minute: mins}, nil
}

func readInt(s string) (n int, rest string, ok bool) {
	i := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	if i == 0 {
		return 0, s, false
	}
	v, err := strconv.Atoi(s[:i])
	if err != nil {
		return 0, s, false
	}
	return v, s[i:], true
}

func formatDuration(d time.Duration) string {
	if d < 0 {
		d = -d
	}
	h := d / time.Hour
	m := (d % time.Hour) / time.Minute
	switch {
	case h > 0 && m > 0:
		return fmt.Sprintf("%dh%dm", h, m)
	case h > 0:
		return fmt.Sprintf("%dh", h)
	default:
		return fmt.Sprintf("%dm", m)
	}
}
