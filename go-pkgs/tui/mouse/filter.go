package mouse

import (
	"os"
)

// CPRMsg is a peeled cursor position report for the app message loop
// (row/col are 1-based DEC coordinates).
type CPRMsg struct {
	Row1, Col1 int
}

// Filter wraps a TTY *os.File: peels CPR (ESC [ r ; c R) into Ch and forwards
// every other byte. Implements term.File (Read/Write/Close/Fd) so Bubble Tea
// (and similar) still MakeRaw the real TTY — a plain io.Reader breaks raw mode.
//
// Single consumer of the TTY: do not open a parallel /dev/tty CPR reader.
type Filter struct {
	f       *os.File
	ch      chan<- CPRMsg
	hold    []byte
	pending []byte
	// OnDrop is optional; called when Ch is full (non-blocking send failed).
	OnDrop func(CPRMsg)
}

// NewFilter wraps f (usually os.Stdin). ch may be nil (CPR discarded).
func NewFilter(f *os.File, ch chan<- CPRMsg) *Filter {
	return &Filter{f: f, ch: ch}
}

// Fd implements term.File so the host can put the terminal in raw mode.
func (f *Filter) Fd() uintptr { return f.f.Fd() }

// Write implements term.File.
func (f *Filter) Write(p []byte) (int, error) { return f.f.Write(p) }

// Close implements term.File. Does not close process stdin.
func (f *Filter) Close() error { return nil }

func (f *Filter) emit(c CPR) {
	msg := CPRMsg{Row1: c.Row1, Col1: c.Col1}
	if f.ch != nil {
		select {
		case f.ch <- msg:
		default:
			if f.OnDrop != nil {
				f.OnDrop(msg)
			}
		}
	}
}

// Read implements io.Reader for the host input loop.
func (f *Filter) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if len(f.pending) > 0 {
		n := copy(p, f.pending)
		f.pending = f.pending[n:]
		return n, nil
	}

	for {
		tmp := make([]byte, 512)
		n, err := f.f.Read(tmp)
		if n > 0 {
			events, forward, rest := DemuxCPR(f.hold, tmp[:n])
			f.hold = rest
			for _, ev := range events {
				f.emit(ev)
			}
			if len(forward) > 0 {
				copied := copy(p, forward)
				if copied < len(forward) {
					f.pending = append([]byte{}, forward[copied:]...)
				}
				if copied > 0 {
					return copied, nil
				}
			}
			if err == nil {
				continue
			}
			return 0, err
		}
		if err != nil {
			return 0, err
		}
		continue
	}
}
