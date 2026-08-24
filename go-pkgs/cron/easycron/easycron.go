// Package easycron parses and evaluates human-friendly interval expressions
// such as every-1h, every-1h-at-4m, every-5m-until-19h00m, and
// every-5m-not-within-19h00m-to-06h30m.
//
// It is a pure expression library: no ticker, no persistence, no hooks.
package easycron

import (
	"fmt"
	"time"
)

// Clock is a local wall-clock time of day (hour 0–23, minute 0–59).
type Clock struct {
	Hour   int
	Minute int
}

// QuietWindow is a recurring silent band. When Start is after End (e.g. 19:00
// to 06:30), the quiet band wraps overnight.
type QuietWindow struct {
	Start Clock
	End   Clock
}

// Expr is a parsed easy-cron expression. Zero Interval is invalid; use Parse.
type Expr struct {
	Raw      string
	Interval time.Duration
	Align    *time.Duration // nil → fire grid anchored at schedule start
	Until    *Clock         // nil → no hard stop
	Quiet    *QuietWindow   // nil → never quiet
}

// Parse parses an easy-cron expression.
//
// Grammar (fixed suffix order):
//
//	every-<dur>[-at-<offset>][-until-<tod>][-not-within-<tod>-to-<tod>]
//
// <dur> / <offset> are Nh, Nm, or NhNm (hours and/or minutes; no seconds).
// <tod> is NhNm with hour 0–23 and minute 0–59 (both parts required).
func Parse(s string) (Expr, error) {
	return parse(s)
}

// String returns the original input when available.
func (e Expr) String() string {
	if e.Raw != "" {
		return e.Raw
	}
	return fmt.Sprintf("every-%s", formatDuration(e.Interval))
}

// Deadline returns the exclusive hard-stop instant for Until on anchor's local
// calendar day. ok is false when Expr has no Until.
//
// If anchor is already at or past that instant, Deadline is still that (past)
// instant so Next reports expired immediately — starting after today's Until
// does not roll to tomorrow.
func (e Expr) Deadline(anchor time.Time, loc *time.Location) (time.Time, bool) {
	if e.Until == nil {
		return time.Time{}, false
	}
	loc = locationOrLocal(loc)
	anchor = anchor.In(loc)
	return time.Date(anchor.Year(), anchor.Month(), anchor.Day(), e.Until.Hour, e.Until.Minute, 0, 0, loc), true
}

// Active reports whether at falls outside the Quiet window. Until is ignored
// (use Deadline / Next for hard stop). nil Quiet is always active.
func (e Expr) Active(at time.Time, loc *time.Location) bool {
	if e.Quiet == nil {
		return true
	}
	loc = locationOrLocal(loc)
	at = at.In(loc)
	return !inQuiet(clockOf(at), *e.Quiet)
}

// Next returns the earliest fire time >= from that is Active and strictly
// before Deadline (when Until is set).
//
// Without Align, the fire grid is anchor + k*Interval for k = 0,1,2,…
// With Align, fires are on the local wall-clock grid from midnight:
// midnight + Align + k*Interval (Align must be in [0, Interval)).
//
// ok is false when the schedule is already expired (no fire >= from before Deadline).
func (e Expr) Next(anchor, from time.Time, loc *time.Location) (time.Time, bool) {
	if e.Interval <= 0 {
		return time.Time{}, false
	}
	loc = locationOrLocal(loc)
	anchor = anchor.In(loc)
	from = from.In(loc)

	var deadline time.Time
	var hasDeadline bool
	if e.Until != nil {
		deadline, hasDeadline = e.Deadline(anchor, loc)
		if hasDeadline && !from.Before(deadline) {
			return time.Time{}, false
		}
	}

	cand, ok := e.firstCandidate(anchor, from, loc)
	if !ok {
		return time.Time{}, false
	}

	// Bound quiet-skipping so a pathological expr cannot loop forever.
	const maxSteps = 10000
	for i := 0; i < maxSteps; i++ {
		if hasDeadline && !cand.Before(deadline) {
			return time.Time{}, false
		}
		if e.Active(cand, loc) {
			return cand, true
		}
		// Jump to quiet end, then advance to the next grid point >= that instant.
		resume := quietResume(cand, *e.Quiet, loc)
		next, ok := e.firstCandidate(anchor, resume, loc)
		if !ok {
			return time.Time{}, false
		}
		if !next.After(cand) {
			// Ensure forward progress on the grid.
			next, ok = e.firstCandidate(anchor, cand.Add(time.Nanosecond), loc)
			if !ok {
				return time.Time{}, false
			}
		}
		cand = next
	}
	return time.Time{}, false
}

func (e Expr) firstCandidate(anchor, from time.Time, loc *time.Location) (time.Time, bool) {
	if e.Align != nil {
		return firstAligned(from, e.Interval, *e.Align, loc)
	}
	return firstRelative(anchor, from, e.Interval)
}

func firstRelative(anchor, from time.Time, interval time.Duration) (time.Time, bool) {
	if !from.After(anchor) {
		return anchor, true
	}
	elapsed := from.Sub(anchor)
	k := elapsed / interval
	cand := anchor.Add(k * interval)
	if cand.Before(from) {
		cand = cand.Add(interval)
	}
	return cand, true
}

// firstAligned walks the local wall-clock grid that restarts each calendar day
// at midnight+align, then every interval, while staying before the next midnight.
func firstAligned(from time.Time, interval, align time.Duration, loc *time.Location) (time.Time, bool) {
	from = from.In(loc)
	midnight := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, loc)
	for day := 0; day < 400; day++ {
		base := midnight.AddDate(0, 0, day)
		nextMidnight := base.AddDate(0, 0, 1)
		origin := base.Add(align)
		if !origin.Before(nextMidnight) {
			continue
		}
		for k := 0; ; k++ {
			cand := origin.Add(time.Duration(k) * interval)
			if !cand.Before(nextMidnight) {
				break
			}
			if !cand.Before(from) {
				return cand, true
			}
		}
	}
	return time.Time{}, false
}

func inQuiet(c Clock, q QuietWindow) bool {
	start := q.Start.minutes()
	end := q.End.minutes()
	cur := c.minutes()
	if start == end {
		// Degenerate full-day quiet (or empty); treat as always quiet.
		return true
	}
	if start < end {
		return cur >= start && cur < end
	}
	// Overnight wrap: quiet from start through midnight and until end.
	return cur >= start || cur < end
}

// quietResume is the next instant at which the quiet window ends (Active becomes true).
func quietResume(at time.Time, q QuietWindow, loc *time.Location) time.Time {
	at = at.In(loc)
	c := clockOf(at)
	if !inQuiet(c, q) {
		return at
	}
	end := time.Date(at.Year(), at.Month(), at.Day(), q.End.Hour, q.End.Minute, 0, 0, loc)
	if q.Start.minutes() < q.End.minutes() {
		// Same-day quiet: resume at End today (or tomorrow if somehow past — shouldn't happen).
		if !at.Before(end) {
			return end.Add(24 * time.Hour)
		}
		return end
	}
	// Overnight: after Start tonight → End tomorrow; before End this morning → End today.
	if c.minutes() >= q.Start.minutes() {
		return end.Add(24 * time.Hour)
	}
	return end
}

func clockOf(t time.Time) Clock {
	return Clock{Hour: t.Hour(), Minute: t.Minute()}
}

func (c Clock) minutes() int {
	return c.Hour*60 + c.Minute
}

func locationOrLocal(loc *time.Location) *time.Location {
	if loc == nil {
		return time.Local
	}
	return loc
}
