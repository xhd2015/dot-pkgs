package window

import (
	"fmt"
	"strings"
)

// FormatWindowStatus renders WindowStatus for human CLI stdout.
func FormatWindowStatus(st WindowStatus) string {
	var b strings.Builder
	title := st.WindowName
	if title == "" {
		title = "(untitled)"
	}
	fmt.Fprintf(&b, "window %s  %s\n", st.WindowID, title)
	maxName := 0
	for _, tab := range st.Tabs {
		if n := len(tab.Name); n > maxName {
			maxName = n
		}
	}
	if maxName > 48 {
		maxName = 48
	}
	for _, tab := range st.Tabs {
		mark := " "
		if tab.Current {
			mark = "*"
		}
		name := tab.Name
		if name == "" {
			name = "(untitled)"
		}
		if maxName > 0 && len(name) > maxName {
			if maxName <= 1 {
				name = name[:maxName]
			} else {
				name = name[:maxName-1] + "…"
			}
		}
		fmt.Fprintf(&b, "  %s [%d] %-*s  %s  %s\n",
			mark, tab.Index, maxName, name, displaySessionID(tab.SessionID), displayTTY(tab.TTY))
	}
	return b.String()
}

// FormatTabStatus renders TabStatus for human CLI stdout.
func FormatTabStatus(st TabStatus) string {
	var b strings.Builder
	wtitle := st.WindowName
	if wtitle == "" {
		wtitle = "(untitled)"
	}
	fmt.Fprintf(&b, "tab %d of window %s  %s\n", st.TabIndex, st.WindowID, wtitle)
	name := st.Name
	if name == "" {
		name = "(untitled)"
	}
	fmt.Fprintf(&b, "  name:     %s\n", name)
	fmt.Fprintf(&b, "  session:  %s\n", st.SessionID)
	fmt.Fprintf(&b, "  tty:      %s\n", displayTTY(st.TTY))
	return b.String()
}

// displaySessionID returns the full session unique ID for CLI copy/paste.
// Empty → "-".
func displaySessionID(id string) string {
	id = strings.TrimSpace(sessionUUID(id))
	if id == "" {
		return "-"
	}
	return id
}

func displayTTY(tty string) string {
	tty = strings.TrimSpace(tty)
	if tty == "" {
		return "-"
	}
	return tty
}
