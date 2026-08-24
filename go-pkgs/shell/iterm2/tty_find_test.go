package iterm2

import (
	"strings"
	"testing"
)

func TestBuildFindByTTYScript_IncludesWantedTTYs(t *testing.T) {
	script := BuildFindByTTYScript([]string{"ttys051", "/dev/ttys052"})
	if !strings.Contains(script, `"/dev/ttys051"`) {
		t.Fatalf("missing normalized tty: %s", script)
	}
	if !strings.Contains(script, `"/dev/ttys052"`) {
		t.Fatalf("missing second tty: %s", script)
	}
	if !strings.Contains(script, "wantTTYs contains normTTY") {
		t.Fatal("expected contains-based TTY filter")
	}
}

func TestBuildFindByTTYScript_Empty(t *testing.T) {
	script := BuildFindByTTYScript(nil)
	if !strings.Contains(script, "set wantTTYs to {}") {
		t.Fatalf("empty want list: %s", script)
	}
}

func TestParseCaptureByTTYOutput(t *testing.T) {
	in := "#meta\t1135\twork\t2\t1\t9AE0162B-AC5B-4BD0-8FCE-E85280C04471\t/dev/ttys051\tpane\nhello\nworld\n"
	got, err := ParseCaptureByTTYOutput(in)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got.Ref.WindowID != "1135" || got.Ref.TabIndex != 2 {
		t.Fatalf("ref = %+v", got.Ref)
	}
	if got.Ref.SessionID != "9AE0162B-AC5B-4BD0-8FCE-E85280C04471" {
		t.Fatalf("session id = %q", got.Ref.SessionID)
	}
	if got.Ref.TTY != "/dev/ttys051" {
		t.Fatalf("tty = %q", got.Ref.TTY)
	}
	if got.Contents != "hello\nworld" {
		t.Fatalf("contents = %q", got.Contents)
	}
}

func TestParseCaptureByTTYOutput_MissingMeta(t *testing.T) {
	_, err := ParseCaptureByTTYOutput("just pane text\n")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBuildContentsByTTYScript_EarlyReturn(t *testing.T) {
	script := BuildContentsByTTYScript("/dev/ttys051", "")
	if !strings.Contains(script, `return contents of aSession`) {
		t.Fatal("expected early-exit contents return")
	}
	if !strings.Contains(script, "tty not found") {
		t.Fatal("expected tty miss error")
	}
}

func TestUniqueNormalizedTTYs(t *testing.T) {
	got := uniqueNormalizedTTYs([]string{"ttys051", "/dev/ttys051", "", "ttys052"})
	if len(got) != 2 || got[0] != "/dev/ttys051" || got[1] != "/dev/ttys052" {
		t.Fatalf("got %#v", got)
	}
}
