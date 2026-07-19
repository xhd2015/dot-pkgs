package mouse

// Phase is the CSI 6n measurement state.
type Phase int

const (
	// PhaseUnknown: need to emit CSI 6n on next frame.
	PhaseUnknown Phase = iota
	// PhasePending: query already embedded in a frame; waiting for CPR.
	PhasePending
	// PhaseKnown: OriginY is valid.
	PhaseKnown
	// PhaseFailed: no usable CPR for this layout; use dual-origin Resolve.
	PhaseFailed
)

// Tracker owns origin measurement state for an inline TUI.
// The app appends FrameSuffix() to the same paint string as the UI frame.
type Tracker struct {
	phase            Phase
	originY          *int
	layoutGen        int
	pendingGen       int
	pendingViewLines int
	knownViewLines   int
}

// NewTracker returns a Tracker in PhaseUnknown.
func NewTracker() *Tracker {
	return &Tracker{phase: PhaseUnknown}
}

// Phase returns the current measurement phase.
func (t *Tracker) Phase() Phase {
	if t == nil {
		return PhaseUnknown
	}
	return t.phase
}

// OriginY returns the measured origin (nil if not Known).
func (t *Tracker) OriginY() *int {
	if t == nil || t.phase != PhaseKnown {
		return nil
	}
	return t.originY
}

// LayoutGen returns the current layout generation counter.
func (t *Tracker) LayoutGen() int {
	if t == nil {
		return 0
	}
	return t.layoutGen
}

// OnResize invalidates origin (call on terminal size change).
func (t *Tracker) OnResize() {
	if t == nil {
		return
	}
	t.invalidate()
}

// OnViewLines notes the painted line count; invalidates if it changed while Known.
func (t *Tracker) OnViewLines(viewLines int) {
	if t == nil {
		return
	}
	if t.phase == PhaseKnown && t.knownViewLines != viewLines {
		t.invalidate()
	}
}

func (t *Tracker) invalidate() {
	t.originY = nil
	t.phase = PhaseUnknown
	t.layoutGen++
}

// FrameSuffix returns CSI6n when a probe should be embedded in the current
// paint string, otherwise "". Call once per View after viewLines is known:
//
//	body := render()
//	body += tracker.FrameSuffix(height, viewLines)
//
// Transitions Unknown → Pending. Not a wall-clock delay: the query rides in
// the same buffer as the frame so the async renderer cannot flush UI without it.
func (t *Tracker) FrameSuffix(height, viewLines int) string {
	if t == nil {
		return ""
	}
	if t.phase == PhaseKnown && t.knownViewLines != viewLines {
		t.invalidate()
	}
	if t.phase != PhaseUnknown || height <= 0 || viewLines <= 0 {
		return ""
	}
	t.phase = PhasePending
	t.pendingGen = t.layoutGen
	t.pendingViewLines = viewLines
	return CSI6n
}

// OnCPR applies a peeled CPR. Returns true if origin became Known.
// Ignores stale replies (wrong phase or layout generation).
func (t *Tracker) OnCPR(row1, col1 int) bool {
	if t == nil {
		return false
	}
	_ = col1
	if t.phase != PhasePending {
		return false
	}
	if t.pendingGen != t.layoutGen {
		return false
	}
	oy, ok := OriginFromCPR(row1, t.pendingViewLines)
	if !ok {
		t.phase = PhaseFailed
		t.originY = nil
		return false
	}
	t.originY = &oy
	t.phase = PhaseKnown
	t.knownViewLines = t.pendingViewLines
	return true
}
