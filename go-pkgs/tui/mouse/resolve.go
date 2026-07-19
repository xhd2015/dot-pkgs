package mouse

// ResolveOpts is pure input for absolute mouse → hit resolution.
type ResolveOpts struct {
	AbsX, AbsY int
	Height     int  // terminal rows
	ViewLines  int  // painted line count for the UI frame
	OriginY    *int // nil => dual-origin; non-nil => localY = absY - *OriginY
	Hits       []Hit
	// PreferID non-empty IDs when dual-origin candidates disagree (Run chips).
	// Empty ID is treated as non-preferred (same as wrk RunStage prefer).
}

// ResolveResult is the outcome of Resolve.
type ResolveResult struct {
	OK     bool
	Hit    Hit
	LocalY int    // local Y used for the successful candidate (-1 if miss)
	Kind   string // "known" | "top" | "bottom" | ""
}

// Resolve maps absolute mouse coordinates onto a view-local Hit.
//
// Known origin (OriginY non-nil): localY = absY - *OriginY; single hit-test.
// Dual-origin (OriginY nil): try top (localY=absY) then bottom
// (localY=absY-(height-viewLines)); prefer hits with non-empty ID when they disagree.
func Resolve(opts ResolveOpts) ResolveResult {
	miss := ResolveResult{OK: false, LocalY: -1}

	if opts.OriginY != nil {
		localY := opts.AbsY - *opts.OriginY
		if h, ok := HitTest(opts.Hits, opts.AbsX, localY); ok {
			return ResolveResult{OK: true, Hit: h, LocalY: localY, Kind: "known"}
		}
		return miss
	}

	type cand struct {
		h      Hit
		localY int
		kind   string
	}
	var cands []cand

	if h, ok := HitTest(opts.Hits, opts.AbsX, opts.AbsY); ok {
		cands = append(cands, cand{h: h, localY: opts.AbsY, kind: "top"})
	}

	botY := opts.AbsY - BottomOriginY(opts.Height, opts.ViewLines)
	if h, ok := HitTest(opts.Hits, opts.AbsX, botY); ok {
		if len(cands) == 0 || cands[0].localY != botY || cands[0].h != h {
			cands = append(cands, cand{h: h, localY: botY, kind: "bottom"})
		}
	}

	if len(cands) == 0 {
		return miss
	}

	best := cands[0]
	for _, c := range cands[1:] {
		if c.h.ID != "" && best.h.ID == "" {
			best = c
			continue
		}
		if c.h.ID != "" && best.h.ID != "" && c.h.ID != best.h.ID {
			// Prefer first (top evaluated first).
			continue
		}
	}
	return ResolveResult{OK: true, Hit: best.h, LocalY: best.localY, Kind: best.kind}
}

// BottomOriginY is the dual-origin bottom candidate: height - viewLines (clamped ≥0).
// viewLines must equal the number of lines the renderer paints (no trailing empty segment).
func BottomOriginY(height, viewLines int) int {
	if height <= 0 || viewLines <= 0 {
		return 0
	}
	origin := height - viewLines
	if origin < 0 {
		return 0
	}
	return origin
}
