// Package mouse provides pure hit-testing and CSI 6n origin measurement for
// inline (non-alt-screen) terminal UIs. It has no dependency on wrk or Bubble Tea.
//
// See go-best-practice skill --show cli/inline-tui-mouse.
package mouse

// Hit is a rectangular region in view-local coordinates (Y = 0 at the first
// line of the UI frame). Half-open: y0 ≤ y < y1; if x1 > x0 then x0 ≤ x < x1.
type Hit struct {
	Y0, Y1 int
	X0, X1 int
	// ID is an app-defined action or focus token (e.g. "gen-commit-msg").
	// Empty means geometry-only (still matchable).
	ID string
}

// HitTest returns the first Hit containing (x, localY).
func HitTest(hits []Hit, x, localY int) (Hit, bool) {
	for _, cand := range hits {
		if localY < cand.Y0 || localY >= cand.Y1 {
			continue
		}
		if cand.X1 > cand.X0 && (x < cand.X0 || x >= cand.X1) {
			continue
		}
		return cand, true
	}
	return Hit{}, false
}
