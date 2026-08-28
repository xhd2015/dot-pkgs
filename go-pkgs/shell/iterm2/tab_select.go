package iterm2

import (
	"fmt"
	"strconv"
	"strings"
)

// TabSelKind is how a tab selector was specified.
type TabSelKind int

const (
	// TabSelAbs1 is --tab N (1-based iTerm tab index).
	TabSelAbs1 TabSelKind = iota
	// TabSelAbs0 is --tab-index N (0-based into the window's ordered tab list).
	TabSelAbs0
	// TabSelNext is --tab next (alias: right); no wrap.
	TabSelNext
	// TabSelLeft is --tab left; no wrap.
	TabSelLeft
)

// TabSelector identifies a tab in the current iTerm2 window.
type TabSelector struct {
	Kind TabSelKind
	N    int // for TabSelAbs1 / TabSelAbs0
}

// ParseTabFlag parses --tab values: 1-based N, or next|right|left.
func ParseTabFlag(raw string) (TabSelector, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return TabSelector{}, fmt.Errorf("--tab requires a value (1-based index, or next|left|right)")
	}
	switch strings.ToLower(s) {
	case "next", "right":
		return TabSelector{Kind: TabSelNext}, nil
	case "left":
		return TabSelector{Kind: TabSelLeft}, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return TabSelector{}, fmt.Errorf("--tab must be a 1-based index or next|left|right, got %q", raw)
	}
	if n < 1 {
		return TabSelector{}, fmt.Errorf("--tab index must be >= 1 (1-based); use --tab-index for 0-based")
	}
	return TabSelector{Kind: TabSelAbs1, N: n}, nil
}

// ParseTabIndexFlag parses --tab-index N (0-based).
func ParseTabIndexFlag(n int) (TabSelector, error) {
	if n < 0 {
		return TabSelector{}, fmt.Errorf("--tab-index must be >= 0")
	}
	return TabSelector{Kind: TabSelAbs0, N: n}, nil
}

// SelectWindowTab picks a tab from st according to sel.
// Returns the tab row and its 0-based position in st.Tabs.
func SelectWindowTab(st WindowStatus, sel TabSelector) (TabStatusRow, int, error) {
	tabs := st.Tabs
	if len(tabs) == 0 {
		return TabStatusRow{}, -1, fmt.Errorf("no tabs in window %s", st.WindowID)
	}

	curPos := -1
	for i, t := range tabs {
		if t.Current || t.Index == st.CurrentTabIndex {
			curPos = i
			break
		}
	}
	if curPos < 0 {
		for i, t := range tabs {
			if t.Index == st.CurrentTabIndex {
				curPos = i
				break
			}
		}
	}
	if curPos < 0 && st.CurrentTabIndex >= 1 {
		return TabStatusRow{}, -1, fmt.Errorf("current tab %d not found in window %s", st.CurrentTabIndex, st.WindowID)
	}

	switch sel.Kind {
	case TabSelAbs1:
		for i, t := range tabs {
			if t.Index == sel.N {
				return t, i, nil
			}
		}
		return TabStatusRow{}, -1, fmt.Errorf("tab %d not found in window %s", sel.N, st.WindowID)
	case TabSelAbs0:
		if sel.N < 0 || sel.N >= len(tabs) {
			return TabStatusRow{}, -1, fmt.Errorf("--tab-index %d out of range (valid: 0..%d)", sel.N, len(tabs)-1)
		}
		return tabs[sel.N], sel.N, nil
	case TabSelNext:
		if curPos < 0 {
			return TabStatusRow{}, -1, fmt.Errorf("current tab unknown; cannot resolve --tab next")
		}
		if curPos+1 >= len(tabs) {
			return TabStatusRow{}, -1, fmt.Errorf("no tab to the right (current tab is last in window)")
		}
		return tabs[curPos+1], curPos + 1, nil
	case TabSelLeft:
		if curPos < 0 {
			return TabStatusRow{}, -1, fmt.Errorf("current tab unknown; cannot resolve --tab left")
		}
		if curPos == 0 {
			return TabStatusRow{}, -1, fmt.Errorf("no tab to the left (current tab is first in window)")
		}
		return tabs[curPos-1], curPos - 1, nil
	default:
		return TabStatusRow{}, -1, fmt.Errorf("invalid tab selector")
	}
}
