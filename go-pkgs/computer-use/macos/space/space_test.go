package space

import (
	"errors"
	"strings"
	"testing"
)

func TestUnsupportedPlatform(t *testing.T) {
	SetGOOSForTest("linux")
	defer SetGOOSForTest("")
	if err := Create(nil); !errors.Is(err, ErrUnsupportedPlatform) {
		t.Fatalf("Create: %v", err)
	}
	if _, err := List(nil); !errors.Is(err, ErrUnsupportedPlatform) {
		t.Fatalf("List: %v", err)
	}
	if _, err := Highest(nil); !errors.Is(err, ErrUnsupportedPlatform) {
		t.Fatalf("Highest: %v", err)
	}
	if err := Switch(1, nil); !errors.Is(err, ErrUnsupportedPlatform) {
		t.Fatalf("Switch: %v", err)
	}
}

func TestCreateInjected(t *testing.T) {
	SetGOOSForTest("darwin")
	defer SetGOOSForTest("")
	var gotScript string
	cfg := &Config{
		Settle: -1,
		Osascript: func(script string, args ...string) (string, error) {
			gotScript = script
			if len(args) != 0 {
				t.Fatalf("create args=%v", args)
			}
			return "OK: created", nil
		},
	}
	if err := Create(cfg); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gotScript, "add desktop") {
		t.Fatalf("script missing add desktop: %s", gotScript[:min(80, len(gotScript))])
	}
}

func TestSwitchInjected(t *testing.T) {
	SetGOOSForTest("darwin")
	defer SetGOOSForTest("")
	var gotArgs []string
	cfg := &Config{
		Settle: -1,
		Osascript: func(script string, args ...string) (string, error) {
			gotArgs = append([]string(nil), args...)
			if !strings.Contains(script, "exit to Desktop") {
				t.Fatal("switch script unexpected")
			}
			return "OK: switched", nil
		},
	}
	if err := Switch(12, cfg); err != nil {
		t.Fatal(err)
	}
	if len(gotArgs) != 1 || gotArgs[0] != "12" {
		t.Fatalf("args=%v", gotArgs)
	}
}

func TestSwitchInvalid(t *testing.T) {
	SetGOOSForTest("darwin")
	defer SetGOOSForTest("")
	if err := Switch(0, &Config{Settle: -1}); err == nil {
		t.Fatal("expected error")
	}
}

func TestHighestAndList(t *testing.T) {
	SetGOOSForTest("darwin")
	defer SetGOOSForTest("")
	cfg := &Config{
		Settle: -1,
		Osascript: func(script string, args ...string) (string, error) {
			if strings.Contains(script, "maxN") {
				return "7", nil
			}
			return "count=2 desktops=[Desktop 1, Desktop 2]", nil
		},
	}
	n, err := Highest(cfg)
	if err != nil || n != 7 {
		t.Fatalf("Highest: %d %v", n, err)
	}
	list, err := List(cfg)
	if err != nil || len(list) != 2 || list[0].Number != 1 {
		t.Fatalf("List: %+v %v", list, err)
	}
}

func TestCreateAndActivate(t *testing.T) {
	SetGOOSForTest("darwin")
	defer SetGOOSForTest("")
	var calls []string
	cfg := &Config{
		Settle: -1,
		Osascript: func(script string, args ...string) (string, error) {
			switch {
			case strings.Contains(script, "add desktop"):
				calls = append(calls, "create")
				return "OK", nil
			case strings.Contains(script, "maxN"):
				calls = append(calls, "highest")
				return "4", nil
			case strings.Contains(script, "exit to Desktop"):
				calls = append(calls, "switch:"+strings.Join(args, ","))
				return "OK", nil
			default:
				t.Fatalf("unexpected script")
				return "", nil
			}
		},
	}
	n, err := CreateAndActivate(cfg)
	if err != nil || n != 4 {
		t.Fatalf("n=%d err=%v", n, err)
	}
	want := []string{"create", "highest", "switch:4"}
	if strings.Join(calls, "|") != strings.Join(want, "|") {
		t.Fatalf("calls=%v", calls)
	}
}

func TestCreateAndActivate_RetriesSwitchNotCreate(t *testing.T) {
	SetGOOSForTest("darwin")
	defer SetGOOSForTest("")
	var calls []string
	switchN := 0
	cfg := &Config{
		Settle: -1,
		Osascript: func(script string, args ...string) (string, error) {
			switch {
			case strings.Contains(script, "add desktop"):
				calls = append(calls, "create")
				return "OK", nil
			case strings.Contains(script, "maxN"):
				calls = append(calls, "highest")
				return "5", nil
			case strings.Contains(script, "exit to Desktop"):
				switchN++
				calls = append(calls, "switch")
				if switchN == 1 {
					return "FAIL: desktop not found: Desktop 5", nil
				}
				return "OK", nil
			default:
				t.Fatalf("unexpected script")
				return "", nil
			}
		},
	}
	n, err := CreateAndActivate(cfg)
	if err != nil || n != 5 {
		t.Fatalf("n=%d err=%v calls=%v", n, err, calls)
	}
	// One create only; highest+switch retried after first switch fail.
	want := []string{"create", "highest", "switch", "highest", "switch"}
	if strings.Join(calls, "|") != strings.Join(want, "|") {
		t.Fatalf("calls=%v want=%v", calls, want)
	}
}

func TestCreateAndActivate_RetriesTransientCreate(t *testing.T) {
	SetGOOSForTest("darwin")
	defer SetGOOSForTest("")
	var creates int
	cfg := &Config{
		Settle: -1,
		Osascript: func(script string, args ...string) (string, error) {
			switch {
			case strings.Contains(script, "add desktop"):
				creates++
				if creates == 1 {
					return "FAIL: System Events got an error: Can’t get group 1 of application process \"Dock\" whose name = \"Mission Control\". Invalid index. (-1719)", nil
				}
				return "OK", nil
			case strings.Contains(script, "maxN"):
				return "3", nil
			case strings.Contains(script, "exit to Desktop"):
				return "OK", nil
			default:
				t.Fatalf("unexpected script")
				return "", nil
			}
		},
	}
	n, err := CreateAndActivate(cfg)
	if err != nil || n != 3 {
		t.Fatalf("n=%d err=%v creates=%d", n, err, creates)
	}
	if creates != 2 {
		t.Fatalf("creates=%d want 2", creates)
	}
}

func TestCreateFail(t *testing.T) {
	SetGOOSForTest("darwin")
	defer SetGOOSForTest("")
	cfg := &Config{
		Settle: -1,
		Osascript: func(script string, args ...string) (string, error) {
			return "FAIL: no button", nil
		},
	}
	if err := Create(cfg); err == nil || !strings.Contains(err.Error(), "FAIL:") {
		t.Fatalf("err=%v", err)
	}
}

func TestParseListOutput(t *testing.T) {
	d := ParseListOutput("count=2 desktops=[Desktop 1, Desktop 2]")
	if len(d) != 2 || d[1].Number != 2 {
		t.Fatalf("%+v", d)
	}
}

func TestMockBackend(t *testing.T) {
	m := &MockBackend{Desktops: []int{1, 2}}
	if err := m.Create(); err != nil {
		t.Fatal(err)
	}
	h, err := m.Highest()
	if err != nil || h != 3 {
		t.Fatalf("h=%d err=%v", h, err)
	}
	if err := m.Switch(3); err != nil {
		t.Fatal(err)
	}
	list, err := m.List()
	if err != nil || len(list) != 3 {
		t.Fatalf("%+v %v", list, err)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
