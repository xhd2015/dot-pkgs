// Command fixture-inline is the P2 headless geometry fixture for tui/mouse.
//
// Product contract (sealed by tui/mouse/tests/headless DSN):
//
//	--anchor=top|mid|bottom (default mid)
//	Paint pad + 5-line UI; btn-a @ localY=3, btn-b @ localY=4, X 2–12
//	Embed CSI 6n via mouse.Tracker; on CPR print:
//	  ORIGIN=<n> VIEW=<v>
//	On SGR click Resolve; print:
//	  HIT id=<id> localY=<y> kind=<k>
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/tui/mouse"
	"golang.org/x/term"
)

const (
	viewLines = 5
	// Default headless PTY size used by tty-watch / creack/pty.
	defaultHeight = 24
	defaultWidth  = 80
)

func main() {
	anchor := flag.String("anchor", "mid", "top|mid|bottom pad before UI")
	flag.Parse()

	height, width := termSize()
	pad := padForAnchor(strings.ToLower(*anchor), height)

	hits := []mouse.Hit{
		{Y0: 3, Y1: 4, X0: 2, X1: 12, ID: "btn-a"},
		{Y0: 4, Y1: 5, X0: 2, X1: 12, ID: "btn-b"},
	}

	tracker := mouse.NewTracker()
	tracker.OnViewLines(viewLines)

	// Pad blanks (anchor geometry).
	for i := 0; i < pad; i++ {
		fmt.Print("\n")
	}

	// 5-line UI frame. Last line has no trailing newline so CSI 6n reports
	// the cursor on the last view line (originY = row1 - viewLines).
	var body strings.Builder
	body.WriteString("fixture-inline\n")
	body.WriteString("anchor=" + strings.ToLower(*anchor) + "\n")
	body.WriteString("---\n")
	body.WriteString("  [btn-a]  \n") // localY=3
	body.WriteString("  [btn-b]  ")  // localY=4 — no trailing \n
	body.WriteString(tracker.FrameSuffix(height, viewLines))
	fmt.Print(body.String())
	_ = os.Stdout.Sync()

	// Put stdin in raw mode when it is a terminal (PTY slave under tty-watch)
	// so CPR/SGR bytes are readable without line buffering.
	var restore func()
	if term.IsTerminal(int(os.Stdin.Fd())) {
		old, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err == nil {
			restore = func() { _ = term.Restore(int(os.Stdin.Fd()), old) }
			defer restore()
		}
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)

	done := make(chan struct{})
	go func() {
		select {
		case <-sig:
		case <-time.After(2 * time.Minute):
		}
		close(done)
	}()

	// Input loop: peel CPR into Tracker; forward SGR mouse → Resolve → HIT.
	buf := make([]byte, 256)
	var hold []byte
	originPrinted := false

	for {
		select {
		case <-done:
			return
		default:
		}

		// Short read deadline via non-blocking-ish select with SetReadDeadline
		// is not available on all platforms for PTY; use a small poll via
		// concurrent read instead.
		type readResult struct {
			n   int
			err error
			b   []byte
		}
		ch := make(chan readResult, 1)
		go func() {
			tmp := make([]byte, len(buf))
			n, err := os.Stdin.Read(tmp)
			if n > 0 {
				ch <- readResult{n: n, err: err, b: append([]byte(nil), tmp[:n]...)}
				return
			}
			ch <- readResult{n: n, err: err}
		}()

		select {
		case <-done:
			return
		case rr := <-ch:
			if rr.n == 0 && rr.err != nil {
				if rr.err == io.EOF {
					// Keep process alive until kill (EOF can happen without kill).
					select {
					case <-done:
						return
					case <-time.After(200 * time.Millisecond):
						continue
					}
				}
				// Transient errors: keep going until kill.
				select {
				case <-done:
					return
				case <-time.After(50 * time.Millisecond):
					continue
				}
			}
			if rr.n <= 0 {
				continue
			}

			events, forward, rest := mouse.DemuxCPR(hold, rr.b)
			hold = rest
			for _, ev := range events {
				if tracker.OnCPR(ev.Row1, ev.Col1) {
					oy := tracker.OriginY()
					if oy != nil && !originPrinted {
						// Machine protocol line (scrollback). Then restore
						// cursor to the last UI line so host query-cursor
						// stays near originY+viewLines-1 (0-based).
						fmt.Printf("\nORIGIN=%d VIEW=%d\n", *oy, viewLines)
						lastUIRow1 := *oy + viewLines // 1-based last view row
						if lastUIRow1 < 1 {
							lastUIRow1 = 1
						}
						fmt.Printf("\x1b[%d;1H", lastUIRow1)
						_ = os.Stdout.Sync()
						originPrinted = true
					}
				}
			}
			if len(forward) > 0 {
				handleSGRForward(forward, tracker, hits, height, width)
			}
		}
	}
}

func padForAnchor(anchor string, height int) int {
	switch anchor {
	case "top":
		return 0
	case "bottom":
		// Enough blanks that UI sits near bottom of PTY.
		p := height - viewLines
		if p < 0 {
			return 0
		}
		// Prefer classic 19 on 24-row (DSN); clamp to height-viewLines.
		if p > 19 && height >= 24 {
			return height - viewLines
		}
		return p
	default: // mid
		return 8
	}
}

func termSize() (height, width int) {
	height, width = defaultHeight, defaultWidth
	if term.IsTerminal(int(os.Stdout.Fd())) {
		if w, h, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
			if h > 0 {
				height = h
			}
			if w > 0 {
				width = w
			}
		}
	}
	return height, width
}

// handleSGRForward scans forwarded (non-CPR) bytes for SGR mouse presses and
// resolves hits. release events (lowercase m) are ignored.
func handleSGRForward(data []byte, tracker *mouse.Tracker, hits []mouse.Hit, height, width int) {
	_ = width
	i := 0
	for i < len(data) {
		// ESC [ < btn ; col ; row M/m
		if data[i] != 0x1b {
			i++
			continue
		}
		if i+2 >= len(data) || data[i+1] != '[' || data[i+2] != '<' {
			i++
			continue
		}
		j := i + 3
		for j < len(data) && data[j] != 'M' && data[j] != 'm' {
			j++
		}
		if j >= len(data) {
			return
		}
		payload := string(data[i+3 : j])
		press := data[j] == 'M'
		i = j + 1
		if !press {
			continue // release
		}
		btn, col1, row1, ok := parseSGRParams(payload)
		if !ok {
			continue
		}
		_ = btn
		// SGR is 1-based; Resolve uses 0-based abs coords.
		absX := col1 - 1
		absY := row1 - 1
		var originY *int
		if tracker.Phase() == mouse.PhaseKnown {
			originY = tracker.OriginY()
		}
		res := mouse.Resolve(mouse.ResolveOpts{
			AbsX:      absX,
			AbsY:      absY,
			Height:    height,
			ViewLines: viewLines,
			OriginY:   originY,
			Hits:      hits,
		})
		if !res.OK {
			continue
		}
		fmt.Printf("HIT id=%s localY=%d kind=%s\n", res.Hit.ID, res.LocalY, res.Kind)
		_ = os.Stdout.Sync()
	}
}

func parseSGRParams(s string) (btn, col1, row1 int, ok bool) {
	// btn;col;row
	parts := strings.Split(s, ";")
	if len(parts) != 3 {
		return 0, 0, 0, false
	}
	var err error
	btn, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, 0, false
	}
	col1, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, 0, false
	}
	row1, err = strconv.Atoi(parts[2])
	if err != nil {
		return 0, 0, 0, false
	}
	return btn, col1, row1, true
}
