package server

import (
	"fmt"
	"io"
	"os"
	"sync"

	"golang.org/x/term"
)

type synchronizedWriter struct {
	mu  *sync.Mutex
	dst io.Writer
}

func (w *synchronizedWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.dst.Write(p)
}

func message(w io.Writer, color, format string, args ...any) {
	target := w
	if locked, ok := w.(*synchronizedWriter); ok {
		target = locked.dst
	}
	file, ok := target.(*os.File)
	if ok && os.Getenv("NO_COLOR") == "" && term.IsTerminal(int(file.Fd())) {
		fmt.Fprintf(w, "\x1b[%sm%s\x1b[0m", color, fmt.Sprintf(format, args...))
	} else {
		fmt.Fprintf(w, format, args...)
	}
}

// PrintError emits the shared CLI error format, using color only on terminals.
func PrintError(w io.Writer, err error) { message(w, "31", "Error: %v\n", err) }
