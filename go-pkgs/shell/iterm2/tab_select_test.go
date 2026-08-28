package iterm2

import (
	"strings"
	"testing"
)

func TestParseTabFlag(t *testing.T) {
	cases := []struct {
		in   string
		kind TabSelKind
		n    int
		err  string
	}{
		{"next", TabSelNext, 0, ""},
		{"RIGHT", TabSelNext, 0, ""},
		{"left", TabSelLeft, 0, ""},
		{"2", TabSelAbs1, 2, ""},
		{"", 0, 0, "--tab requires"},
		{"0", 0, 0, ">= 1"},
		{"foo", 0, 0, "1-based index"},
	}
	for _, tc := range cases {
		got, err := ParseTabFlag(tc.in)
		if tc.err != "" {
			if err == nil || !strings.Contains(err.Error(), tc.err) {
				t.Fatalf("%q: err=%v want contain %q", tc.in, err, tc.err)
			}
			continue
		}
		if err != nil {
			t.Fatalf("%q: %v", tc.in, err)
		}
		if got.Kind != tc.kind || got.N != tc.n {
			t.Fatalf("%q: got %+v", tc.in, got)
		}
	}
}

func TestParseTabIndexFlag(t *testing.T) {
	got, err := ParseTabIndexFlag(0)
	if err != nil || got.Kind != TabSelAbs0 || got.N != 0 {
		t.Fatalf("got %+v err=%v", got, err)
	}
	if _, err := ParseTabIndexFlag(-1); err == nil {
		t.Fatal("expected error")
	}
}

func TestSelectWindowTab(t *testing.T) {
	st := WindowStatus{
		WindowID:        "100",
		CurrentTabIndex: 1,
		Tabs: []TabStatusRow{
			{Index: 1, Current: true, SessionID: "AAA", Name: "a"},
			{Index: 2, SessionID: "BBB", Name: "b"},
			{Index: 3, SessionID: "CCC", Name: "c"},
		},
	}

	row, pos, err := SelectWindowTab(st, TabSelector{Kind: TabSelNext})
	if err != nil || pos != 1 || row.SessionID != "BBB" {
		t.Fatalf("next: %+v pos=%d err=%v", row, pos, err)
	}

	row, pos, err = SelectWindowTab(st, TabSelector{Kind: TabSelAbs1, N: 3})
	if err != nil || pos != 2 || row.SessionID != "CCC" {
		t.Fatalf("abs1: %+v pos=%d err=%v", row, pos, err)
	}

	row, pos, err = SelectWindowTab(st, TabSelector{Kind: TabSelAbs0, N: 0})
	if err != nil || pos != 0 || row.SessionID != "AAA" {
		t.Fatalf("abs0: %+v pos=%d err=%v", row, pos, err)
	}

	stLast := st
	stLast.CurrentTabIndex = 3
	stLast.Tabs[0].Current = false
	stLast.Tabs[2].Current = true
	if _, _, err := SelectWindowTab(stLast, TabSelector{Kind: TabSelNext}); err == nil || !strings.Contains(err.Error(), "no tab to the right") {
		t.Fatalf("last next: %v", err)
	}
	if _, _, err := SelectWindowTab(st, TabSelector{Kind: TabSelLeft}); err == nil || !strings.Contains(err.Error(), "no tab to the left") {
		t.Fatalf("first left: %v", err)
	}
	if _, _, err := SelectWindowTab(st, TabSelector{Kind: TabSelAbs0, N: 9}); err == nil || !strings.Contains(err.Error(), "out of range") {
		t.Fatalf("oob: %v", err)
	}
}
