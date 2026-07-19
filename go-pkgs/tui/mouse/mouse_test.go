package mouse

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestHitTest(t *testing.T) {
	hits := []Hit{
		{Y0: 3, Y1: 4, X0: 1, X1: 61, ID: "left"},
		{Y0: 3, Y1: 4, X0: 61, X1: 71, ID: "run"},
	}
	h, ok := HitTest(hits, 65, 3)
	if !ok || h.ID != "run" {
		t.Fatalf("got ok=%v id=%q", ok, h.ID)
	}
	_, ok = HitTest(hits, 65, 4)
	if ok {
		t.Fatal("y=4 should miss half-open y1")
	}
}

func TestResolveKnownMidPane(t *testing.T) {
	oy := 6
	hits := []Hit{
		{Y0: 3, Y1: 4, X0: 61, X1: 71, ID: "add-changes"},
		{Y0: 4, Y1: 5, X0: 61, X1: 71, ID: "gen-commit-msg"},
		{Y0: 11, Y1: 12, X0: 61, X1: 71, ID: "tag-next"},
	}
	r := Resolve(ResolveOpts{
		AbsX: 67, AbsY: 9, Height: 26, ViewLines: 20,
		OriginY: &oy, Hits: hits,
	})
	if !r.OK || r.Hit.ID != "add-changes" || r.Kind != "known" || r.LocalY != 3 {
		t.Fatalf("%+v", r)
	}
	r = Resolve(ResolveOpts{
		AbsX: 67, AbsY: 10, Height: 26, ViewLines: 20,
		OriginY: &oy, Hits: hits,
	})
	if !r.OK || r.Hit.ID != "gen-commit-msg" {
		t.Fatalf("%+v", r)
	}
}

func TestResolveDualTopNotTagNext(t *testing.T) {
	hits := []Hit{
		{Y0: 4, Y1: 5, X0: 61, X1: 71, ID: "gen-commit-msg"},
		{Y0: 11, Y1: 12, X0: 61, X1: 71, ID: "tag-next"},
	}
	// Top-anchored: absY = localY
	r := Resolve(ResolveOpts{
		AbsX: 65, AbsY: 4, Height: 40, ViewLines: 20,
		Hits: hits,
	})
	if !r.OK || r.Hit.ID != "gen-commit-msg" {
		t.Fatalf("want gen-commit-msg, got %+v", r)
	}
}

func TestOriginFromCPR(t *testing.T) {
	oy, ok := OriginFromCPR(26, 20)
	if !ok || oy != 6 {
		t.Fatalf("got %d ok=%v", oy, ok)
	}
	if _, ok := OriginFromCPR(9, 20); ok {
		t.Fatal("row1 < viewLines must fail (live rule)")
	}
	oy, ok = OriginFromCPRClamped(5, 20)
	if !ok || oy != 0 {
		t.Fatalf("clamped got %d ok=%v", oy, ok)
	}
}

func TestDemuxCPR_MouseForward(t *testing.T) {
	in := []byte("\x1b[41;1R\x1b[<0;67;25M")
	ev, fwd, rest := DemuxCPR(nil, in)
	if len(rest) != 0 || len(ev) != 1 || ev[0].Row1 != 41 {
		t.Fatalf("ev=%v rest=%q", ev, rest)
	}
	if !bytes.Equal(fwd, []byte("\x1b[<0;67;25M")) {
		t.Fatalf("fwd=%q", fwd)
	}
}

func TestFilter_ReadForwardsMouse(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()

	ch := make(chan CPRMsg, 2)
	f := NewFilter(r, ch)
	go func() {
		_, _ = w.Write([]byte("\x1b[20;1R\x1b[<0;10;5M"))
		_ = w.Close()
	}()

	buf := make([]byte, 64)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		t.Fatal(err)
	}
	if string(buf[:n]) != "\x1b[<0;10;5M" {
		t.Fatalf("got %q", buf[:n])
	}
	select {
	case m := <-ch:
		if m.Row1 != 20 || m.Col1 != 1 {
			t.Fatalf("%+v", m)
		}
	default:
		t.Fatal("expected CPR")
	}
}

func TestTracker_MidPane(t *testing.T) {
	tr := NewTracker()
	if s := tr.FrameSuffix(26, 20); s != CSI6n {
		t.Fatalf("suffix=%q", s)
	}
	if tr.Phase() != PhasePending {
		t.Fatal(tr.Phase())
	}
	if !tr.OnCPR(26, 1) {
		t.Fatal("OnCPR")
	}
	if tr.Phase() != PhaseKnown || tr.OriginY() == nil || *tr.OriginY() != 6 {
		t.Fatalf("phase=%v oy=%v", tr.Phase(), tr.OriginY())
	}
	// Stale / invalid
	tr2 := NewTracker()
	_ = tr2.FrameSuffix(26, 20)
	if tr2.OnCPR(9, 1) {
		t.Fatal("row1 < viewLines should fail")
	}
	if tr2.Phase() != PhaseFailed || tr2.OriginY() != nil {
		t.Fatalf("phase=%v", tr2.Phase())
	}
}

func TestTracker_ResizeInvalidates(t *testing.T) {
	tr := NewTracker()
	_ = tr.FrameSuffix(40, 20)
	_ = tr.OnCPR(40, 1)
	tr.OnResize()
	if tr.Phase() != PhaseUnknown || tr.OriginY() != nil {
		t.Fatal("resize should invalidate")
	}
	if s := tr.FrameSuffix(40, 20); s != CSI6n {
		t.Fatal("should re-emit after resize")
	}
}

func TestBottomOriginY_Edges(t *testing.T) {
	cases := []struct {
		name             string
		height, viewLines int
		want             int
	}{
		{"normal", 26, 20, 6},
		{"height_zero", 0, 20, 0},
		{"view_zero", 26, 0, 0},
		{"both_zero", 0, 0, 0},
		{"height_neg", -5, 20, 0},
		{"view_neg", 26, -3, 0},
		{"height_lt_view", 10, 20, 0},
		{"equal", 20, 20, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := BottomOriginY(tc.height, tc.viewLines)
			if got != tc.want {
				t.Fatalf("BottomOriginY(%d,%d)=%d want %d", tc.height, tc.viewLines, got, tc.want)
			}
		})
	}
}

func TestOriginFromCPRClamped_Edges(t *testing.T) {
	// row1 < viewLines clamps origin to 0 (legacy path).
	oy, ok := OriginFromCPRClamped(5, 20)
	if !ok || oy != 0 {
		t.Fatalf("row1<viewLines: got %d ok=%v want 0,true", oy, ok)
	}
	oy, ok = OriginFromCPRClamped(1, 20)
	if !ok || oy != 0 {
		t.Fatalf("row1=1: got %d ok=%v want 0,true", oy, ok)
	}
	// Valid mid-pane still returns row1-viewLines.
	oy, ok = OriginFromCPRClamped(26, 20)
	if !ok || oy != 6 {
		t.Fatalf("valid: got %d ok=%v want 6,true", oy, ok)
	}
	// viewLines <= 0 or row1 < 1 fail.
	if _, ok := OriginFromCPRClamped(10, 0); ok {
		t.Fatal("viewLines=0 must fail")
	}
	if _, ok := OriginFromCPRClamped(10, -1); ok {
		t.Fatal("viewLines<0 must fail")
	}
	if _, ok := OriginFromCPRClamped(0, 20); ok {
		t.Fatal("row1=0 must fail")
	}
	if _, ok := OriginFromCPRClamped(-1, 20); ok {
		t.Fatal("row1<0 must fail")
	}
}

func TestParseCPR_Edges(t *testing.T) {
	// Empty / incomplete
	if _, _, ok := ParseCPR(nil); ok {
		t.Fatal("nil buf should fail")
	}
	if _, _, ok := ParseCPR([]byte{}); ok {
		t.Fatal("empty should fail")
	}
	if _, _, ok := ParseCPR([]byte("\x1b[")); ok {
		t.Fatal("incomplete ESC[ should fail")
	}
	if _, _, ok := ParseCPR([]byte("\x1b[20;")); ok {
		t.Fatal("incomplete after semicolon should fail")
	}
	if _, _, ok := ParseCPR([]byte("\x1b[20;1")); ok {
		t.Fatal("missing R should fail")
	}

	// Malformed: not a CPR shape
	if _, _, ok := ParseCPR([]byte("\x1b[A")); ok {
		t.Fatal("CSI A should not parse as CPR")
	}
	if _, _, ok := ParseCPR([]byte("\x1b[<0;10;5M")); ok {
		t.Fatal("mouse SGR must not parse as CPR")
	}
	if _, _, ok := ParseCPR([]byte("hello world")); ok {
		t.Fatal("plain text should fail")
	}

	// Noise around a valid CPR
	row, col, ok := ParseCPR([]byte("noise\x1b[20;1Rtrail"))
	if !ok || row != 20 || col != 1 {
		t.Fatalf("noise around CPR: row=%d col=%d ok=%v", row, col, ok)
	}

	// Multi first-wins
	row, col, ok = ParseCPR([]byte("\x1b[10;2R\x1b[99;3R"))
	if !ok || row != 10 || col != 2 {
		t.Fatalf("first-wins: row=%d col=%d ok=%v", row, col, ok)
	}

	// Valid single
	row, col, ok = ParseCPR([]byte("\x1b[26;1R"))
	if !ok || row != 26 || col != 1 {
		t.Fatalf("valid: row=%d col=%d ok=%v", row, col, ok)
	}
}

func TestResolve_Miss(t *testing.T) {
	hits := []Hit{
		{Y0: 3, Y1: 4, X0: 61, X1: 71, ID: "add-changes"},
		{Y0: 4, Y1: 5, X0: 61, X1: 71, ID: "gen-commit-msg"},
	}
	oy := 6

	// Known origin: click outside chip x-range
	r := Resolve(ResolveOpts{
		AbsX: 10, AbsY: 9, Height: 26, ViewLines: 20,
		OriginY: &oy, Hits: hits,
	})
	if r.OK || r.LocalY != -1 || r.Kind != "" {
		t.Fatalf("known miss: %+v", r)
	}

	// Known origin: localY maps outside any hit y-range
	r = Resolve(ResolveOpts{
		AbsX: 65, AbsY: 20, Height: 26, ViewLines: 20,
		OriginY: &oy, Hits: hits,
	})
	if r.OK || r.LocalY != -1 {
		t.Fatalf("known miss far absY: %+v", r)
	}

	// Dual-origin miss: neither top nor bottom hits
	r = Resolve(ResolveOpts{
		AbsX: 1, AbsY: 1, Height: 40, ViewLines: 20,
		Hits: hits,
	})
	if r.OK || r.LocalY != -1 || r.Kind != "" {
		t.Fatalf("dual miss: %+v", r)
	}

	// Dual miss with empty hit list
	r = Resolve(ResolveOpts{
		AbsX: 65, AbsY: 4, Height: 40, ViewLines: 20,
		Hits: nil,
	})
	if r.OK || r.LocalY != -1 {
		t.Fatalf("empty hits miss: %+v", r)
	}
}

func TestTracker_FrameSuffixOnceWhilePending(t *testing.T) {
	tr := NewTracker()
	s1 := tr.FrameSuffix(26, 20)
	if s1 != CSI6n {
		t.Fatalf("first suffix=%q want CSI6n", s1)
	}
	if tr.Phase() != PhasePending {
		t.Fatalf("phase after first: %v", tr.Phase())
	}
	// Second View while still Pending must not re-emit CSI 6n.
	s2 := tr.FrameSuffix(26, 20)
	if s2 != "" {
		t.Fatalf("second suffix while Pending=%q want empty", s2)
	}
	if tr.Phase() != PhasePending {
		t.Fatalf("phase after second: %v", tr.Phase())
	}
	// Third call still empty.
	if s3 := tr.FrameSuffix(26, 20); s3 != "" {
		t.Fatalf("third suffix=%q want empty", s3)
	}
}
