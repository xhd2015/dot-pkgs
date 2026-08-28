package window

import (
	"strings"
	"testing"
)

func fixturePanes() []PaneRef {
	return []PaneRef{
		{WindowID: "23473", WindowName: "Project", TabIndex: 1, SessionID: "E209EDD0-6149-4972-94B0-3816091267B0", TTY: "/dev/ttys143", Name: "grok"},
		{WindowID: "23473", WindowName: "Project", TabIndex: 2, SessionID: "C057E2A3-1111-2222-3333-444444444444", TTY: "/dev/ttys163", Name: "bash"},
		{WindowID: "1592", WindowName: "other", TabIndex: 1, SessionID: "AAAAAAAA-BBBB-CCCC-DDDD-EEEEEEEEEEEE", TTY: "/dev/ttys087", Name: "other"},
	}
}

func locFromPane(ref PaneRef) CurrentLocation {
	return CurrentLocation{
		WindowID:    ref.WindowID,
		WindowName:  ref.WindowName,
		TabIndex:    ref.TabIndex,
		SessionID:   ref.SessionID,
		TTY:         ref.TTY,
		SessionName: ref.Name,
	}
}

func TestBuildWindowStatus_MarksCurrent(t *testing.T) {
	refs := fixturePanes()
	loc := locFromPane(refs[0])
	st := BuildWindowStatus(refs, loc)
	if st.WindowID != "23473" || len(st.Tabs) != 2 {
		t.Fatalf("st=%+v", st)
	}
	if !st.Tabs[0].Current || st.Tabs[1].Current {
		t.Fatalf("current marks: %+v", st.Tabs)
	}
	if st.Tabs[0].Name != "grok" || st.Tabs[1].Name != "bash" {
		t.Fatalf("names: %+v", st.Tabs)
	}
}

func TestFormatWindowStatus_Star(t *testing.T) {
	st := BuildWindowStatus(fixturePanes(), locFromPane(fixturePanes()[0]))
	out := FormatWindowStatus(st)
	if !strings.Contains(out, "window 23473") {
		t.Fatalf("header missing: %q", out)
	}
	if !strings.Contains(out, "* [1]") {
		t.Fatalf("missing * on current tab: %q", out)
	}
	if !strings.Contains(out, "  [2]") {
		t.Fatalf("missing unmarked tab 2: %q", out)
	}
	if !strings.Contains(out, "E209EDD0-6149-4972-94B0-3816091267B0") {
		t.Fatalf("missing full session id for tab 1: %q", out)
	}
	if !strings.Contains(out, "C057E2A3-1111-2222-3333-444444444444") {
		t.Fatalf("missing full session id for tab 2: %q", out)
	}
	if strings.Contains(out, "E209EDD0…") || strings.Contains(out, "C057E2A3…") {
		t.Fatalf("session ids must not be truncated: %q", out)
	}
}

func TestFormatTabStatus(t *testing.T) {
	st := BuildTabStatus(locFromPane(fixturePanes()[0]))
	out := FormatTabStatus(st)
	if !strings.Contains(out, "tab 1 of window 23473") {
		t.Fatalf("header: %q", out)
	}
	if !strings.Contains(out, "E209EDD0-6149-4972-94B0-3816091267B0") {
		t.Fatalf("session: %q", out)
	}
	if !strings.Contains(out, "/dev/ttys143") {
		t.Fatalf("tty: %q", out)
	}
}

func TestBuildWindowStatus_UUIDPrefixMatch(t *testing.T) {
	refs := fixturePanes()
	loc := locFromPane(refs[0])
	loc.SessionID = "w0t0p0:E209EDD0-6149-4972-94B0-3816091267B0"
	st := BuildWindowStatus(refs, loc)
	if !st.Tabs[0].Current {
		t.Fatalf("expected current via UUID suffix match: %+v", st.Tabs[0])
	}
}
