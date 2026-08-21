package iterm2

import (
	"errors"
	"strings"
	"testing"
)

func fixtureRefs() []SessionRef {
	return []SessionRef{
		{WindowID: "23473", WindowName: "Project", TabIndex: 1, SessionID: "E209EDD0-6149-4972-94B0-3816091267B0", TTY: "/dev/ttys143", Name: "grok"},
		{WindowID: "23473", WindowName: "Project", TabIndex: 2, SessionID: "C057E2A3-1111-2222-3333-444444444444", TTY: "/dev/ttys163", Name: "bash"},
		{WindowID: "1592", WindowName: "other", TabIndex: 1, SessionID: "AAAAAAAA-BBBB-CCCC-DDDD-EEEEEEEEEEEE", TTY: "/dev/ttys087", Name: "other"},
	}
}

func TestBuildSessionListScript_SoftSkipIndexes(t *testing.T) {
	script := BuildSessionListScript()
	if !strings.Contains(script, "count of windows") {
		t.Fatal("expected count-of-windows snapshot")
	}
	if !strings.Contains(script, "repeat with wi from 1 to windowCount") {
		t.Fatal("expected indexed window loop")
	}
	if strings.Contains(script, "repeat with aWindow in windows") {
		t.Fatal("live `repeat with … in windows` iterator is race-prone")
	}
	if strings.Contains(script, "repeat with aTab in tabs") {
		t.Fatal("live `repeat with … in tabs` iterator is race-prone")
	}
	if !strings.Contains(script, "on error") {
		t.Fatal("expected per-item soft-skip on error")
	}
}

func TestBuildSessionsInWindowByUUIDScript(t *testing.T) {
	script := BuildSessionsInWindowByUUIDScript("w0t0p0:E209EDD0-6149-4972-94B0-3816091267B0")
	if !strings.Contains(script, "E209EDD0-6149-4972-94B0-3816091267B0") {
		t.Fatal("script should embed session UUID")
	}
	if !strings.Contains(script, "foundInWindow") {
		t.Fatal("expected window-scoped dump after UUID match")
	}
	if strings.Contains(script, "repeat with aWindow in windows") {
		t.Fatal("must not use live window iterator")
	}
}

func TestAppleScriptInvalidIndex(t *testing.T) {
	if !appleScriptInvalidIndex(errors.New(`execution error: Can’t get item 9 of every tab of item 7 of every window. Invalid index. (-1719)`)) {
		t.Fatal("expected -1719 Invalid index detection")
	}
	if appleScriptInvalidIndex(errors.New("session not found")) {
		t.Fatal("non-index errors must not match")
	}
}

func TestFindBySessionUUID(t *testing.T) {
	refs := fixtureRefs()
	got := FindBySessionUUID(refs, "w24t0p0:E209EDD0-6149-4972-94B0-3816091267B0")
	if len(got) != 1 || got[0].WindowID != "23473" || got[0].TabIndex != 1 {
		t.Fatalf("uuid match = %+v", got)
	}
	if len(FindBySessionUUID(refs, "")) != 0 {
		t.Fatal("empty uuid should yield empty")
	}
	if len(FindBySessionUUID(refs, "nope")) != 0 {
		t.Fatal("miss should be empty")
	}
}

func TestResolveCurrentLocation_UUID(t *testing.T) {
	refs := fixtureRefs()
	loc, err := resolveCurrentLocationFrom(refs, CurrentStatusConfig{
		SessionID:      func() string { return "w0t0p0:E209EDD0-6149-4972-94B0-3816091267B0" },
		ControllingTTY: func() string { return "" },
		AncestorTTYs:   func() []string { return nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	if loc.WindowID != "23473" || loc.TabIndex != 1 || loc.TTY != "/dev/ttys143" {
		t.Fatalf("loc=%+v", loc)
	}
}

func TestResolveCurrentLocation_ControllingTTY(t *testing.T) {
	refs := fixtureRefs()
	loc, err := resolveCurrentLocationFrom(refs, CurrentStatusConfig{
		SessionID:      func() string { return "" },
		ControllingTTY: func() string { return "ttys163" },
		AncestorTTYs:   func() []string { return []string{"/dev/ttys157"} },
	})
	if err != nil {
		t.Fatal(err)
	}
	if loc.TabIndex != 2 || loc.SessionID != "C057E2A3-1111-2222-3333-444444444444" {
		t.Fatalf("want tab 2 via controlling tty, got %+v", loc)
	}
}

func TestResolveCurrentLocation_AncestorTTYSkipsUnlisted(t *testing.T) {
	refs := fixtureRefs()
	loc, err := resolveCurrentLocationFrom(refs, CurrentStatusConfig{
		SessionID:      func() string { return "" },
		ControllingTTY: func() string { return "" },
		// First ancestor is grok-only PTY (not in session list); second is session tty.
		AncestorTTYs: func() []string { return []string{"/dev/ttys157", "/dev/ttys143"} },
	})
	if err != nil {
		t.Fatal(err)
	}
	if loc.TTY != "/dev/ttys143" || loc.TabIndex != 1 {
		t.Fatalf("want listed ancestor ttys143, got %+v", loc)
	}
}

func TestResolveCurrentLocation_NotInSession(t *testing.T) {
	_, err := resolveCurrentLocationFrom(fixtureRefs(), CurrentStatusConfig{
		SessionID:      func() string { return "" },
		ControllingTTY: func() string { return "" },
		AncestorTTYs:   func() []string { return []string{"/dev/ttys157"} },
	})
	if !errors.Is(err, ErrNotInSession) {
		t.Fatalf("err=%v, want ErrNotInSession", err)
	}
}

func TestBuildWindowStatus_MarksCurrent(t *testing.T) {
	refs := fixtureRefs()
	loc := locationFromRef(refs[0])
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
	st := BuildWindowStatus(fixtureRefs(), locationFromRef(fixtureRefs()[0]))
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
	// Full session IDs so callers can copy into session send/status.
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
	st := BuildTabStatus(locationFromRef(fixtureRefs()[0]))
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

func TestCurrentWindowStatusWith_Inject(t *testing.T) {
	refs := fixtureRefs()
	st, err := CurrentWindowStatusWith(&CurrentStatusConfig{
		SessionID:    func() string { return "E209EDD0-6149-4972-94B0-3816091267B0" },
		ListSessions: func() ([]SessionRef, error) { return refs, nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(st.Tabs) != 2 || !st.Tabs[0].Current {
		t.Fatalf("%+v", st)
	}
}

func TestCurrentTabStatusWith_Inject(t *testing.T) {
	refs := fixtureRefs()
	st, err := CurrentTabStatusWith(&CurrentStatusConfig{
		SessionID:    func() string { return "E209EDD0-6149-4972-94B0-3816091267B0" },
		ListSessions: func() ([]SessionRef, error) { return refs, nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	if st.TabIndex != 1 || st.WindowID != "23473" {
		t.Fatalf("%+v", st)
	}
}
