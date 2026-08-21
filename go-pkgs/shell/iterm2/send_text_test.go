package iterm2

import (
	"errors"
	"strings"
	"testing"
)

func TestBuildSendTextScriptDefault(t *testing.T) {
	s := BuildSendTextScript("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee", "echo hi", SendTextOptions{}, "")
	if !strings.Contains(s, `tell application "iTerm2"`) {
		t.Fatalf("bare tell:\n%s", s)
	}
	if !strings.Contains(s, "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee") {
		t.Fatalf("missing uuid:\n%s", s)
	}
	if !strings.Contains(s, `write text ((ASCII character 21) & "echo hi")`) {
		t.Fatalf("default Ctrl-U write:\n%s", s)
	}
	if strings.Contains(s, "without newline") {
		t.Fatalf("default must submit:\n%s", s)
	}
	if !strings.Contains(s, "repeat with wi from 1 to windowCount") {
		t.Fatalf("expected indexed soft-skip loops:\n%s", s)
	}
	if strings.Contains(s, "repeat with aWindow in windows") {
		t.Fatalf("live window iterator is race-prone:\n%s", s)
	}
	lower := strings.ToLower(s)
	if strings.Contains(lower, "activate") {
		t.Fatalf("default must not activate:\n%s", s)
	}
	if strings.Contains(lower, "select ") {
		t.Fatalf("default must not select:\n%s", s)
	}
}

func TestBuildSendTextScriptNoCtrlUNoSubmit(t *testing.T) {
	s := BuildSendTextScript("aabb", `say "hi"`, SendTextOptions{NoCtrlU: true, NoSubmit: true}, "/Applications/iTerm.app")
	if !strings.Contains(s, `tell application "/Applications/iTerm.app"`) {
		t.Fatalf("path tell:\n%s", s)
	}
	want := `write text "say \"hi\"" without newline`
	if !strings.Contains(s, want) {
		t.Fatalf("want %q in:\n%s", want, s)
	}
	if strings.Contains(s, "ASCII character 21") {
		t.Fatalf("--no-ctrl-u must omit Ctrl-U:\n%s", s)
	}
}

func TestBuildSendTextScriptFocus(t *testing.T) {
	s := BuildSendTextScript("uuid-1", "ls", SendTextOptions{Focus: true}, "")
	lower := strings.ToLower(s)
	if !strings.Contains(lower, "activate") {
		t.Fatalf("focus must activate:\n%s", s)
	}
	if !strings.Contains(lower, "select targetwindow") {
		t.Fatalf("focus must select window:\n%s", s)
	}
	if !strings.Contains(lower, "select targettab") {
		t.Fatalf("focus must select tab:\n%s", s)
	}
	if !strings.Contains(lower, "select targetsession") {
		t.Fatalf("focus must select session:\n%s", s)
	}
	if !strings.Contains(s, `write text ((ASCII character 21) & "ls")`) {
		t.Fatalf("write line:\n%s", s)
	}
}

func TestSendTextEmptyID(t *testing.T) {
	err := SendText("", "x", SendTextOptions{}, &SendTextConfig{
		Apps: []ContentsApp{},
	})
	if err == nil || !strings.Contains(err.Error(), "session id is required") {
		t.Fatalf("got %v", err)
	}
}

func TestSendTextNotFound(t *testing.T) {
	err := SendText("deadbeef", "x", SendTextOptions{}, &SendTextConfig{
		Apps: []ContentsApp{{Canonical: "t", Path: "/tmp/iTerm.app"}},
		Exec: func(script string) error {
			return errors.New("session not found: deadbeef")
		},
	})
	if !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("got %v", err)
	}
}

func TestSendTextSuccessFirstApp(t *testing.T) {
	var got string
	err := SendText("11111111-2222-3333-4444-555555555555", "echo hi", SendTextOptions{}, &SendTextConfig{
		Apps: []ContentsApp{{Canonical: "home", Path: "/home/iTerm.app"}},
		Exec: func(script string) error {
			got = script
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "echo hi") {
		t.Fatalf("script=%q", got)
	}
	if !strings.Contains(got, `tell application "/home/iTerm.app"`) {
		t.Fatalf("script=%q", got)
	}
}
