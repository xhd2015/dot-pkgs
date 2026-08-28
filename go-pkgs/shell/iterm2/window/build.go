package window

import (
	"sort"
	"strings"
)

// BuildWindowStatus lists every tab in loc's window, marking the current tab.
func BuildWindowStatus(refs []PaneRef, loc CurrentLocation) WindowStatus {
	st := WindowStatus{
		WindowID:        loc.WindowID,
		WindowName:      loc.WindowName,
		CurrentTabIndex: loc.TabIndex,
	}
	type agg struct {
		row      TabStatusRow
		haveCurr bool
	}
	byTab := map[int]*agg{}
	var order []int
	for _, ref := range refs {
		if ref.WindowID != loc.WindowID {
			continue
		}
		if ref.TabIndex < 1 {
			continue
		}
		a, ok := byTab[ref.TabIndex]
		if !ok {
			a = &agg{row: TabStatusRow{Index: ref.TabIndex}}
			byTab[ref.TabIndex] = a
			order = append(order, ref.TabIndex)
		}
		isCurrentPane := loc.TabIndex == ref.TabIndex &&
			(loc.SessionID == "" || sessionUUID(ref.SessionID) == sessionUUID(loc.SessionID) ||
				(loc.TTY != "" && normalizeTTY(ref.TTY) == normalizeTTY(loc.TTY)))
		if isCurrentPane {
			a.row.Current = true
			a.row.Name = ref.Name
			a.row.SessionID = ref.SessionID
			a.row.TTY = ref.TTY
			a.haveCurr = true
			continue
		}
		if !a.haveCurr && a.row.Name == "" {
			a.row.Name = ref.Name
			a.row.SessionID = ref.SessionID
			a.row.TTY = ref.TTY
		}
	}
	if loc.TabIndex >= 1 {
		if a, ok := byTab[loc.TabIndex]; ok {
			a.row.Current = true
			if !a.haveCurr {
				if loc.SessionName != "" {
					a.row.Name = loc.SessionName
				}
				if loc.SessionID != "" {
					a.row.SessionID = loc.SessionID
				}
				if loc.TTY != "" {
					a.row.TTY = loc.TTY
				}
			}
		}
	}
	sort.Ints(order)
	for _, ti := range order {
		st.Tabs = append(st.Tabs, byTab[ti].row)
	}
	return st
}

// BuildTabStatus returns the current-tab summary for loc.
func BuildTabStatus(loc CurrentLocation) TabStatus {
	return TabStatus{
		WindowID:   loc.WindowID,
		WindowName: loc.WindowName,
		TabIndex:   loc.TabIndex,
		Name:       loc.SessionName,
		SessionID:  loc.SessionID,
		TTY:        loc.TTY,
	}
}

func sessionUUID(sessionID string) string {
	if i := strings.LastIndex(sessionID, ":"); i >= 0 && i+1 < len(sessionID) {
		return sessionID[i+1:]
	}
	return sessionID
}

func normalizeTTY(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if strings.HasPrefix(s, "/dev/") {
		return s
	}
	return "/dev/" + s
}
